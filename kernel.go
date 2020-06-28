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
	"sync"
	"time"
)

var (
	ask                    = New() // Thread-Safe skip-list
	bid                    = New() // Thread-Safe skip-list
	ask1Price        int64 = math.MaxInt64
	bid1Price        int64 = math.MinInt64
	matchingInfoChan       = make(chan *matchingInfo, 1000)
	ask1PriceMux           = sync.Mutex{}
	bid1PriceMux           = sync.Mutex{}
)

type matchingInfo struct {
	makerOrders []types.KernelOrder
	takerOrder  types.KernelOrder
}

// new order at the head of the list, old order at the tail of the list
// a.k.a price level
type priceBucket struct {
	l    list.List
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
			l := list.List{}
			l.PushFront(order)
			ask.Set(float64(order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL
			if ask1Price == math.MaxInt64 {
				ask1PriceMux.Lock()
				if ask1Price == math.MaxInt64 {
					ask1Price = order.Price
				} else if ask1Price > order.Price {
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
			l := list.List{}
			l.PushFront(order)
			bid.Set(float64(-order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL
			if bid1Price == math.MinInt64 {
				bid1PriceMux.Lock()
				if bid1Price == math.MinInt64 {
					bid1Price = order.Price
				} else if bid1Price < order.Price {
					bid1Price = order.Price
				}
				bid1PriceMux.Unlock()
			}
			return true
		}
	}
}

// clear a price level/bucket
func clearBucket(bucket *priceBucket, order *types.KernelOrder, wg *sync.WaitGroup) {
	defer wg.Done()
	if order.Left == 0 {
		order.Status = types.CLOSED
	}
	matchingInfo := &matchingInfo{
		makerOrders: nil,
		takerOrder:  *order,
	}
	makerOrders := make([]types.KernelOrder, 0, bucket.l.Len())
	element := bucket.l.Back()
	for v := element; v != nil; v = v.Prev() {
		matchedOrder := v.Value.(*types.KernelOrder)
		matchedOrder.FilledTotal = matchedOrder.Amount * matchedOrder.Price
		matchedOrder.Left = 0
		matchedOrder.Status = types.CLOSED
		matchedOrder.UpdateTime = time.Now().UnixNano()
		makerOrders = append(makerOrders, *matchedOrder)
	}
	matchingInfo.makerOrders = makerOrders
	order.UpdateTime = time.Now().UnixNano()
	// remove bucket
	price := element.Value.(*types.KernelOrder).Price
	if order.Amount < 0 {
		bid.Remove(float64(-price))
	} else {
		ask.Remove(float64(price))
	}
	// send matched
	matchingInfoChan <- matchingInfo
}

// run in single thread,  需确证可撮合的订单进入
func matchingOrder(side *SkipList, order *types.KernelOrder, isAsk bool) {
	wg := sync.WaitGroup{}

	// GTC order
	if order.TimeInForce == types.GTC {
	Loop:
		for b := side.Front(); b != nil; b = b.Next() {
			bucket := b.Value().(*priceBucket)
			bucketListHead := bucket.l.Front().Value.(*types.KernelOrder)
			// check price

			if isAsk && bucketListHead.Price < order.Price {
				break Loop
			} else if !isAsk && bucketListHead.Price > order.Price {
				break Loop
			}
			// check whether enough amount of order in a bucket
			if (isAsk && bucket.Left <= -order.Left) || (!isAsk && bucket.Left >= -order.Left) {
				// 可异步清空价价格篮子
				order.Left += bucket.Left
				order.FilledTotal -= bucket.Left * bucketListHead.Price
				wg.Add(1)
				go clearBucket(bucket, order, &wg)
			} else {
				if order.Left == 0 {
					//order.Status = types.CLOSED
					break Loop
				}
				// 匹配剩余订单
				matchingInfo := &matchingInfo{
					makerOrders: nil, // avoid reallocate
					takerOrder:  *order,
				}
				makerOrders := make([]*types.KernelOrder, 0, bucket.l.Len())
				for v := bucket.l.Back(); v != nil; v = v.Prev() {
					matchedOrder := v.Value.(*types.KernelOrder)
					unixNano := time.Now().UnixNano()
					if (isAsk && matchedOrder.Left <= -order.Left) || (!isAsk && matchedOrder.Left >= -order.Left) {
						// 全部吃完
						order.Left += matchedOrder.Left

						matchedOrder.FilledTotal = matchedOrder.Amount * matchedOrder.Price
						matchedOrder.Left = 0
						matchedOrder.Status = types.CLOSED
						matchedOrder.UpdateTime = unixNano
						makerOrders = append(makerOrders, matchedOrder)

						bucket.l.Remove(v)
					} else {
						// 吃了完了还有剩余
						order.Left = 0

						matchedOrder.FilledTotal = -order.Left * matchedOrder.Price
						matchedOrder.Left += order.Left
						matchedOrder.UpdateTime = unixNano
						makerOrders = append(makerOrders, matchedOrder)
					}
					if order.Left == 0 {
						order.UpdateTime = unixNano
						order.Status = types.CLOSED
						// send matched
						matchingInfoChan <- matchingInfo
						break Loop
					}
				}
			}
		}
		// Loop end

		// 还有剩余的不能成交, 插入卖单队列
		if order.Left != 0 {
			insertCheckedOrder(order)
		}
	}

	// 等待异步处理完成
	wg.Wait()
	// 必须等待异步处理完成后, 更新卖一价格
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
