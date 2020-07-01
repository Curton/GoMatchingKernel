/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 *
 * 撮合系统核心, 必须由 kernel_scheduler 调度使用, 不完全线程安全
 *
 */

package exchangeKernel

import (
	"container/list"
	"exchangeKernel/types"
	"math"
	"strconv"
	"sync"
	"time"
)

var (
	ask                    = New() // skip-list
	bid                    = New() // skip-list
	ask1Price        int64 = math.MaxInt64
	bid1Price        int64 = math.MinInt64
	matchingInfoChan       = make(chan *matchingInfo)
	ask1PriceMux           = sync.Mutex{}
	bid1PriceMux           = sync.Mutex{}
)

type matchingInfo struct {
	makerOrders    []types.KernelOrder
	takerOrder     types.KernelOrder
	matchedSizeMap map[uint64]int64
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

func takeOrderBook() *orderBook {
	orderBook := &orderBook{
		ask: make([]orderBookItem, 0, ask.Length+100), // ensure not reallocate array
		bid: make([]orderBookItem, 0, bid.Length+100),
	}
	// todo
	return orderBook
}

// after price & amount checked, (不同价格)的(不可成交)订单可以同时插入, 以上两个条件需同时满足
func insertCheckedOrder(order *types.KernelOrder) bool {
	if order.Amount < 0 {
		get := ask.Get(float64(order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			bucket.Left += order.Left
			return true
		} else {
			l := list.New()
			l.PushFront(order)
			ask.Set(float64(order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL 减少锁开销
			if ask1Price == math.MaxInt64 {
				ask1PriceMux.Lock()
				if ask1Price == math.MaxInt64 {
					ask1Price = order.Price
				} else if ask1Price > order.Price {
					ask1Price = order.Price
				}
				ask1PriceMux.Unlock()
			} else if ask1Price > order.Price {
				ask1PriceMux.Lock()
				if ask1Price > order.Price {
					ask1Price = order.Price
				}
				ask1PriceMux.Unlock()
			}
			return true
		}
	} else {
		get := bid.Get(float64(-order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			bucket.Left += order.Left
			return true
		} else {
			l := list.New()
			l.PushFront(order)
			bid.Set(float64(-order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL 减少锁开销
			if bid1Price == math.MinInt64 {
				bid1PriceMux.Lock()
				if bid1Price == math.MinInt64 {
					bid1Price = order.Price
				} else if bid1Price < order.Price {
					bid1Price = order.Price
				}
				bid1PriceMux.Unlock()
			} else if bid1Price < order.Price {
				bid1PriceMux.Lock()
				if bid1Price < order.Price {
					bid1Price = order.Price
				}
				bid1PriceMux.Unlock()
			}
			return true
		}
	}
}

// clear a price level/bucket
func clearBucket(e *Element, takerOrder types.KernelOrder, wg *sync.WaitGroup, took int64) {
	defer wg.Done()
	bucket := e.Value().(*priceBucket)
	matchingInfo := &matchingInfo{
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
	bucket = nil
	matchingInfoChan <- matchingInfo
}

// run in single thread,  需确证可撮合的订单进入
func matchingOrder(side *SkipList, takerOrder *types.KernelOrder, isAsk bool) {
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
				// 可异步清空价价格篮子
				takerOrder.Left += bucket.Left
				takerOrder.FilledTotal -= bucket.Left * bucketListHead.Price
				if takerOrder.Left == 0 {
					takerOrder.Status = types.CLOSED
				}
				takerOrder.UpdateTime = time.Now().UnixNano()
				removeBucketKeyList.PushBack(skipListElement.key)
				wg.Add(1)
				//clearBucket(skipListElement, *takerOrder, &wg)
				go clearBucket(skipListElement, *takerOrder, &wg, bucket.Left)
				// todo
				//time.Sleep(time.Millisecond)
			} else { // 匹配剩余订单
				if takerOrder.Left == 0 {
					break Loop
				}
				matchingInfo := &matchingInfo{
					makerOrders:    make([]types.KernelOrder, 0, bucket.l.Len()),
					matchedSizeMap: make(map[uint64]int64),
				}
				matchingInfo.matchedSizeMap[takerOrder.KernelOrderID] = takerOrder.Left

				for listElement := bucket.l.Back(); ; /*  listElement != nil */ /* listElement = listElement.Prev() */ {
					matchedOrder := listElement.Value.(*types.KernelOrder)
					unixNano := time.Now().UnixNano()
					if (isAsk && matchedOrder.Left <= -takerOrder.Left) || (!isAsk && matchedOrder.Left >= -takerOrder.Left) {
						// 全部吃完
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
						// send matched
						matchingInfoChan <- matchingInfo
						// todo
						//time.Sleep(time.Millisecond)
						break Loop
					}
				}
				// fail fast
				panic("takerOrder.Left expected to be 0, actual value is " + strconv.FormatInt(takerOrder.Left, 10))
			}
		}
		// Loop end

		// 还有剩余的不能成交, 插入卖单队列
		if takerOrder.Left != 0 {
			insertCheckedOrder(takerOrder)
		}
	}

	// 清理 bucket
	for e := removeBucketKeyList.Front(); e != nil; e = e.Next() {
		side.Remove(e.Value.(float64))
	}

	// 等待异步处理完成
	wg.Wait()

	//fmt.Println("----------------------------------------------------------------")
	// 必须等待异步处理完成后, 更新买一/卖一价格
	if side.Length != 0 {
		bucket := side.Front().value.(*priceBucket)
		kernelOrder := bucket.l.Front().Value.(*types.KernelOrder)
		price := kernelOrder.Price
		if isAsk {
			bid1Price = price
		} else {
			ask1Price = price
		}
	} else {
		if isAsk {
			bid1Price = math.MinInt64
		} else {
			ask1Price = math.MaxInt64
		}
	}
}
