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
type priceBucket struct {
	l    list.List
	size int64
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
			bucket.size += order.Left
			return true
		} else {
			l := list.List{}
			l.PushFront(order)
			ask.Set(float64(order.Price), &priceBucket{
				l:    l,
				size: order.Left,
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
			bucket.size += order.Left
			return true
		} else {
			l := list.List{}
			l.PushFront(order)
			bid.Set(float64(-order.Price), &priceBucket{
				l:    l,
				size: order.Left,
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

	//// first ask order
	//if order.Amount < 0 {
	//	l := list.List{}
	//	l.PushFront(order)
	//	ask.Set(float64(order.Price), &priceBucket{
	//		l:    l,
	//		size: order.Left,
	//	})
	//	ask1Price = order.Price
	//	return true
	//}
	//
	//// first bid order
	//if order.Amount > 0 {
	//	l := list.List{}
	//	l.PushFront(order)
	//	bid.Set(float64(-order.Price), &priceBucket{
	//		l:    l,
	//		size: order.Left,
	//	})
	//	bid1Price = order.Price
	//	return true
	//}
	return false
}

// after price checked, size < 0, 撮合卖单, 匹配对应买单
func matchingAskOrder(order *types.KernelOrder) {
	wg := sync.WaitGroup{}

	// GTC order
	if order.TimeInForce == types.GTC {
	Loop:
		for b := bid.Front(); b != nil; b = b.Next() {
			bucket := b.Value().(*priceBucket)
			bucketListHead := bucket.l.Front().Value.(*types.KernelOrder)
			// check price
			if bucketListHead.Price < order.Price {
				break Loop
			}
			// check whether enough amount of order in a bucket
			if bucket.size <= -order.Amount {
				// 可异步清空价价格篮子
				order.Left += bucket.size
				order.FilledTotal -= bucket.size * bucketListHead.Price
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
					if matchedOrder.Left <= -order.Left {
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
	if bid.Length != 0 {
		bucket := bid.Front().value.(*priceBucket)
		price := bucket.l.Front().Value.(*types.KernelOrder).Price
		bid1Price = price
	} else {
		bid1Price = math.MinInt64
	}
}

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

// after price checked, size > 0, 撮合买单, 匹配对应卖单
func matchingBidOrder(order *types.KernelOrder) *matchingInfo {
	return nil
}

func takeSnapshot() {
	// will block ask & bid list
	// should not call frequently
}
