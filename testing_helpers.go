/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package ker

import (
	"math"
	"math/rand"
	"time"

	"github.com/Curton/GoMatchingKernel/types"
)

func newTestOrder(amount, price int64) *types.KernelOrder {
	now := time.Now().UnixNano()
	return &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    now,
		UpdateTime:    now,
		Amount:        amount,
		Price:         price,
		Left:          amount,
		FilledTotal:   0,
		Status:        types.OPEN,
		Type:          types.LIMIT,
		TimeInForce:   types.GTC,
		Id:            0,
	}
}

func newTestAskOrder(price, amount int64) *types.KernelOrder {
	order := newTestOrder(-amount, price)
	order.Left = -amount
	return order
}

func newTestBidOrder(price, amount int64) *types.KernelOrder {
	return newTestOrder(amount, price)
}

func newTestAcceptor() *scheduler {
	acceptor := initAcceptor(1, "test")
	go acceptor.orderAcceptor()
	return acceptor
}

func drainMatchedInfoChan(acceptor *scheduler) {
	go func() {
		for {
			select {
			case <-acceptor.kernel.matchedInfoChan:
			default:
				return
			}
		}
	}()
}

func generateTestOrders(count int, priceRange, amountRange int64, asAsk bool) []*types.KernelOrder {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	orders := make([]*types.KernelOrder, count)
	for i := 0; i < count; i++ {
		price := r.Int63n(priceRange) + 1
		amount := r.Int63n(amountRange) + 1
		if asAsk {
			orders[i] = newTestAskOrder(price, amount)
		} else {
			orders[i] = newTestBidOrder(price, amount)
		}
	}
	return orders
}

func waitForEmptyBook(acceptor *scheduler, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if acceptor.kernel.ask.Length == 0 && acceptor.kernel.bid.Length == 0 {
			return true
		}
		time.Sleep(time.Millisecond * 10)
	}
	return false
}

func waitForPriceUpdate(acceptor *scheduler, expectedAskPrice, expectedBidPrice int64, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if acceptor.kernel.ask1Price == expectedAskPrice && acceptor.kernel.bid1Price == expectedBidPrice {
			return true
		}
		time.Sleep(time.Millisecond * 10)
	}
	return false
}

func getBucketLeft(acceptor *scheduler, isAsk bool) int64 {
	var list *SkipList
	if isAsk {
		list = acceptor.kernel.ask
	} else {
		list = acceptor.kernel.bid
	}
	if list.Front() == nil {
		return 0
	}
	return list.Front().Value().(*priceBucket).Left
}

func getOrderBookTotalSize(acceptor *scheduler, isAsk bool) int {
	var list *SkipList
	if isAsk {
		list = acceptor.kernel.ask
	} else {
		list = acceptor.kernel.bid
	}
	return list.Length
}

const (
	testAsk1PriceEmpty = math.MaxInt64
	testBid1PriceEmpty = math.MinInt64
)
