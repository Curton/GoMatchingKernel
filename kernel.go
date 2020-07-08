/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/6 11:24
 */

package exchangeKernel

import (
	"container/list"
	"exchangeKernel/types"
	"math"
	"sync"
	"time"
)

type kernel struct {
	ask             *SkipList
	bid             *SkipList
	ask1Price       int64
	bid1Price       int64
	matchedInfoChan chan *matchedInfo
	errorInfoChan   chan *KernelErr
	ask1PriceMux    sync.Mutex
	bid1PriceMux    sync.Mutex
}

type matchedInfo struct {
	makerOrders    []types.KernelOrder
	matchedSizeMap map[uint64]int64
	takerOrder     types.KernelOrder
}

type KernelErr error

type Kernel struct {
}

// new order at the head of the list, old order at the tail of the list
// a.k.a price level
type priceBucket struct {
	l    *list.List
	Left int64
}

type orderBook struct {
	ask []orderBookItem
	bid []orderBookItem
}

type orderBookItem struct {
	Price int64
	Size  int64
}

// after price & amount checked, 附条件异步, (不同价格)的(不可成交)订单可以同时插入, 以上两个条件需同时满足
func (k *kernel) insertCheckedOrder(order *types.KernelOrder) bool {
	if order.Amount < 0 {
		get := k.ask.Get(float64(order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			bucket.Left += order.Left
			return true
		} else {
			l := list.New()
			l.PushFront(order)
			k.ask.Set(float64(order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL 减少锁开销
			if k.ask1Price == math.MaxInt64 {
				k.ask1PriceMux.Lock()
				if k.ask1Price == math.MaxInt64 {
					k.ask1Price = order.Price
				} else if k.ask1Price > order.Price {
					k.ask1Price = order.Price
				}
				k.ask1PriceMux.Unlock()
			} else if k.ask1Price > order.Price {
				k.ask1PriceMux.Lock()
				if k.ask1Price > order.Price {
					k.ask1Price = order.Price
				}
				k.ask1PriceMux.Unlock()
			}
			return true
		}
	} else {
		get := k.bid.Get(float64(-order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			bucket.Left += order.Left
			return true
		} else {
			l := list.New()
			l.PushFront(order)
			k.bid.Set(float64(-order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL 减少锁开销
			if k.bid1Price == math.MinInt64 {
				k.bid1PriceMux.Lock()
				if k.bid1Price == math.MinInt64 {
					k.bid1Price = order.Price
				} else if k.bid1Price < order.Price {
					k.bid1Price = order.Price
				}
				k.bid1PriceMux.Unlock()
			} else if k.bid1Price < order.Price {
				k.bid1PriceMux.Lock()
				if k.bid1Price < order.Price {
					k.bid1Price = order.Price
				}
				k.bid1PriceMux.Unlock()
			}
			return true
		}
	}
}

// clear a price level/bucket
func (k *kernel) clearBucket(e *Element, takerOrder types.KernelOrder, wg *sync.WaitGroup, took int64) {
	defer wg.Done()
	bucket := e.Value().(*priceBucket)
	matchingInfo := &matchedInfo{
		makerOrders:    nil,
		takerOrder:     takerOrder,
		matchedSizeMap: make(map[uint64]int64),
	}
	matchingInfo.matchedSizeMap[takerOrder.KernelOrderID] = -took
	makerOrders := make([]types.KernelOrder, 0, bucket.l.Len())
	element := bucket.l.Back()
	for v := element; v != nil; v = v.Prev() {
		matchedOrder := v.Value.(*types.KernelOrder)
		matchedOrder.FilledTotal += matchedOrder.Left * matchedOrder.Price
		matchingInfo.matchedSizeMap[matchedOrder.KernelOrderID] = matchedOrder.Left
		matchedOrder.Left = 0
		matchedOrder.Status = types.CLOSED
		matchedOrder.UpdateTime = takerOrder.UpdateTime
		makerOrders = append(makerOrders, *matchedOrder)
	}
	matchingInfo.makerOrders = makerOrders
	k.matchedInfoChan <- matchingInfo
}

// run in single thread,  需确证可撮合的订单进入
func (k *kernel) matchingOrder(side *SkipList, takerOrder *types.KernelOrder, isAsk bool) {
	wg := sync.WaitGroup{}

	// GTC takerOrder
	removeBucketKeyList := list.New()
	if takerOrder.TimeInForce == types.GTC {
	Loop:
		for skipListElement := side.Front(); skipListElement != nil; skipListElement = skipListElement.Next() {
			bucket := skipListElement.Value().(*priceBucket)
			bucketListHead := bucket.l.Front().Value.(*types.KernelOrder)
			// check price
			if (isAsk && bucketListHead.Price < takerOrder.Price) || (!isAsk && bucketListHead.Price > takerOrder.Price) {
				break Loop
			}
			// check if enough amount of left order in a bucket
			if (isAsk && bucket.Left <= -takerOrder.Left) || (!isAsk && bucket.Left >= -takerOrder.Left) {
				// async clear price bucket
				takerOrder.Left += bucket.Left
				takerOrder.FilledTotal -= bucket.Left * bucketListHead.Price
				// todo: check if this can remove
				if takerOrder.Left == 0 {
					takerOrder.Status = types.CLOSED
				}
				takerOrder.UpdateTime = time.Now().UnixNano()
				removeBucketKeyList.PushBack(skipListElement.key)
				wg.Add(1)
				go k.clearBucket(skipListElement, *takerOrder, &wg, bucket.Left)
			} else { // matching remaining order
				if takerOrder.Left == 0 {
					break Loop
				}
				matchingInfo := &matchedInfo{
					makerOrders:    make([]types.KernelOrder, 0, bucket.l.Len()),
					matchedSizeMap: make(map[uint64]int64),
				}
				matchingInfo.matchedSizeMap[takerOrder.KernelOrderID] = takerOrder.Left

				for listElement := bucket.l.Back(); ; /*  listElement != nil */ /* listElement = listElement.Prev() */ {
					matchedOrder := listElement.Value.(*types.KernelOrder)
					unixNano := time.Now().UnixNano()
					if (isAsk && matchedOrder.Left <= -takerOrder.Left) || (!isAsk && matchedOrder.Left >= -takerOrder.Left) {
						// clear matched maker Order
						bucket.Left -= matchedOrder.Left
						takerOrder.Left += matchedOrder.Left
						matchedOrder.FilledTotal += matchedOrder.Left * matchedOrder.Price
						takerOrder.FilledTotal -= matchedOrder.Left * matchedOrder.Price
						matchingInfo.matchedSizeMap[matchedOrder.KernelOrderID] = matchedOrder.Left
						matchedOrder.Left = 0
						matchedOrder.Status = types.CLOSED
						matchedOrder.UpdateTime = unixNano
						matchingInfo.makerOrders = append(matchingInfo.makerOrders, *matchedOrder)
						rm := listElement
						listElement = listElement.Prev()
						bucket.l.Remove(rm)
					} else {
						// 吃了完了matchedOrder还有剩余
						matchedOrder.FilledTotal -= takerOrder.Left * matchedOrder.Price
						takerOrder.FilledTotal += takerOrder.Left * matchedOrder.Price
						matchingInfo.matchedSizeMap[matchedOrder.KernelOrderID] = -takerOrder.Left
						matchedOrder.Left += takerOrder.Left
						bucket.Left += takerOrder.Left
						takerOrder.Left = 0
						matchedOrder.UpdateTime = unixNano
						matchingInfo.makerOrders = append(matchingInfo.makerOrders, *matchedOrder)
					}
					if takerOrder.Left == 0 {
						takerOrder.UpdateTime = unixNano
						takerOrder.Status = types.CLOSED
						matchingInfo.takerOrder = *takerOrder
						// send matched info.
						k.matchedInfoChan <- matchingInfo
						break Loop
					}
				}
			}
		}
		// Loop end

		// 还有剩余的不能成交, 插入卖单队列
		if takerOrder.Left != 0 {
			k.insertCheckedOrder(takerOrder)
		}
	}

	// remove cleared bucket
	for e := removeBucketKeyList.Front(); e != nil; e = e.Next() {
		side.Remove(e.Value.(float64))
	}

	// 等待异步处理完成
	wg.Wait()

	// 必须等待异步处理完成后, 更新买一/卖一价格
	if side.Length != 0 {
		bucket := side.Front().value.(*priceBucket)
		kernelOrder := bucket.l.Front().Value.(*types.KernelOrder)
		price := kernelOrder.Price
		if isAsk {
			k.bid1Price = price
		} else {
			k.ask1Price = price
		}
	} else {
		if isAsk {
			k.bid1Price = math.MinInt64
		} else {
			k.ask1Price = math.MaxInt64
		}
	}
}

func NewKernel() *kernel {
	return &kernel{
		ask:             NewSkipList(),
		bid:             NewSkipList(),
		ask1Price:       math.MaxInt64,
		bid1Price:       math.MinInt64,
		matchedInfoChan: make(chan *matchedInfo),
		errorInfoChan:   make(chan *KernelErr),
		ask1PriceMux:    sync.Mutex{},
		bid1PriceMux:    sync.Mutex{},
	}
}
