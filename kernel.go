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
	ask                    = New() // thread safe
	bid                    = New() // thread safe
	ask1Price        int64 = math.MaxInt64
	bid1Price        int64 = math.MinInt64
	matchingInfoChan       = make(chan *matchingInfo, 1000)
)

type matchingInfo struct {
	makerOrders []*types.KernelOrder
	takerOrder  *types.KernelOrder
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
		ask: make([]orderBookItem, ask.Length+100), // ensure do not reallocate array
		bid: make([]orderBookItem, bid.Length+100),
	}
	// todo
	return orderBook
}

// after price checked
func insertPriceCheckedOrder(order *types.KernelOrder) bool {
	if order.Price >= ask1Price {
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
			return true
		}
	}

	if order.Price <= bid1Price {
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
				size: order.Amount,
			})
			return true
		}
	}

	// first ask order
	if order.Amount < 0 {
		l := list.List{}
		l.PushFront(order)
		ask.Set(float64(order.Price), &priceBucket{
			l:    l,
			size: order.Left,
		})
		ask1Price = order.Price
		return true
	}

	// first bid order
	if order.Amount > 0 {
		l := list.List{}
		l.PushFront(order)
		bid.Set(float64(-order.Price), &priceBucket{
			l:    l,
			size: order.Left,
		})
		bid1Price = order.Price
		return true
	}
	return false
}

// after price checked, size < 0, 撮合卖单, 匹配对应买单
func matchingAskOrder(order *types.KernelOrder) {
	wg := &sync.WaitGroup{}

	// GTC order
	if order.TimeInForce == types.GTC {
	Loop:
		for b := bid.Front(); b != nil; b = b.Next() {
			bucket := b.Value().(*priceBucket)
			o := bucket.l.Front().Value.(*types.KernelOrder)
			// check price
			if o.Price < order.Price {
				break Loop
			}
			// check whether enough amount of order in a bucket
			if bucket.size <= -order.Amount {
				// 可异步清空价价格篮子
				order.Left += bucket.size
				go clearBucket(bucket, order, wg)
			} else {
				if order.Left == 0 {
					break Loop
				}
				// 匹配剩余订单
				matchingInfo := &matchingInfo{
					makerOrders: make([]*types.KernelOrder, bucket.l.Len()), // avoid reallocate
					takerOrder:  order,
				}
				makerOrders := matchingInfo.makerOrders
				for v := bucket.l.Back(); v != nil; v = v.Prev() {
					matchedOrder := v.Value.(*types.KernelOrder)
					if matchedOrder.Left <= -order.Left {
						// 全部吃完
						order.Left += matchedOrder.Left

						matchedOrder.FilledTotal = matchedOrder.Amount * matchedOrder.Price
						matchedOrder.Left = 0
						matchedOrder.Status = types.CLOSED
						matchedOrder.UpdateTime = time.Now().UnixNano()
						makerOrders = append(makerOrders, matchedOrder)

						bucket.l.Remove(v)
					} else {
						// 吃了完了还有剩余
						order.Left = 0

						matchedOrder.FilledTotal = -order.Left * matchedOrder.Price
						matchedOrder.Left += order.Left
						matchedOrder.UpdateTime = time.Now().UnixNano()
						makerOrders = append(makerOrders, matchedOrder)
					}
					if order.Left == 0 {
						// send matched
						matchingInfoChan <- matchingInfo
						break Loop
					}
				}
				// send matched
				matchingInfoChan <- matchingInfo
			}
		}
		// Loop end
		// 还有剩余的不能成交, 插入卖单队列
		if order.Left != 0 {
			insertPriceCheckedOrder(order)
		}
	}

	// 等待异步处理完成
	wg.Wait()
	// 必须等待异步处理完成后, 更新卖一价格
	bid.Front()
}

func clearBucket(bucket *priceBucket, order *types.KernelOrder, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	matchingInfo := &matchingInfo{
		makerOrders: make([]*types.KernelOrder, bucket.l.Len()),
		takerOrder:  order,
	}
	makerOrders := matchingInfo.makerOrders
	for v := bucket.l.Back(); v != nil; v = v.Prev() {
		matchedOrder := v.Value.(*types.KernelOrder)
		matchedOrder.FilledTotal = matchedOrder.Amount * matchedOrder.Price
		matchedOrder.Left = 0
		matchedOrder.Status = types.CLOSED
		matchedOrder.UpdateTime = time.Now().UnixNano()
		makerOrders = append(makerOrders, matchedOrder)
	}
	// remove bucket
	if order.Amount < 0 {
		bid.Remove(float64(-order.Price))
	} else {
		ask.Remove(float64(order.Price))
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
