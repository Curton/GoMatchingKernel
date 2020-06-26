/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package exchangeKernel

import (
	"container/list"
	"exchangeKernel/skiplist"
	"exchangeKernel/types"
	"math"
)

var (
	ask             = skiplist.New() // thread safe
	bid             = skiplist.New() // thread safe
	ask1Price int64 = math.MaxInt64
	bid1Price int64 = math.MinInt64
)

type matchingInfo struct {
	makerOrders []*types.KernelOrder
	takerOrder  *types.KernelOrder
}

//type makerInfo struct {
//	makerKernelId int64
//	filledSize int64
//}

// new order at the head of the list, old order at the tail of the list
type priceBucket struct {
	l    list.List
	size int64
}

//type orderBook struct {
//	ask skiplist.SkipList
//	bid skiplist.SkipList
//}
//
//type orderBookItem struct {
//	Price int64
//	Size  int64
//}

// after price checked
func insertPriceCheckedOrder(order *types.KernelOrder) bool {
	if order.Price >= ask1Price {
		get := ask.Get(float64(order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			return true
		} else {
			l := list.List{}
			l.PushFront(order)
			ask.Set(float64(order.Price), &priceBucket{
				l:    l,
				size: order.Amount,
			})
			return true
		}
	}

	if order.Price <= bid1Price {
		get := bid.Get(float64(-order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
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
			size: order.Amount,
		})
		return true
	}

	// first bid order
	if order.Amount > 0 {
		l := list.List{}
		l.PushFront(order)
		bid.Set(float64(-order.Price), &priceBucket{
			l:    l,
			size: order.Amount,
		})
		return true
	}
	return false
}

// after price checked, size < 0
func matchingAskOrder(order *types.KernelOrder) *matchingInfo {

	for b := bid.Front(); b != nil; b = b.Next() {

	}

	return nil
}

// after price checked, size > 0
func matchingBidOrder(order *types.KernelOrder) *matchingInfo {
	return nil
}

func takeSnapshot() {

}
