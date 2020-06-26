package exchangeKernel

import (
	"container/list"
	"exchangeKernel/skiplist"
	"exchangeKernel/types"
	"math"
)

var (
	Ask                  = skiplist.New()
	Bid                  = skiplist.New()
	CurrentAskLow  int64 = math.MinInt64
	CurrentBidHigh int64 = math.MaxInt64
)

// new order at the head of the list, old order at the tail of the list
type samePriceBucket struct {
	l    list.List
	size int64
}

//type orderBook struct {
//	Ask skiplist.SkipList
//	Bid skiplist.SkipList
//}
//
//type orderBookItem struct {
//	Price int64
//	Size  int64
//}

func insertOrder(order *types.KernelOrder) bool {
	if order.Price >= CurrentAskLow {
		get := Ask.Get(float64(order.Price))
		if get != nil {
			bucket := get.Value().(*samePriceBucket)
			bucket.l.PushFront(order)
			return true
		} else {
			l := list.List{}
			l.PushFront(order)
			Ask.Set(float64(order.Price), &samePriceBucket{
				l:    l,
				size: order.Amount,
			})
		}

	}

	if order.Price <= CurrentBidHigh {
		bucket := Bid.Get(float64(order.Price)).Value().(samePriceBucket)
		bucket.l.PushFront(order)
		return true
	}

	return false
}
