/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package ker

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Curton/GoMatchingKernel/types"
)

func Test_insertPriceCheckedOrder_WithSamePrice(t *testing.T) {
	k := newKernel()
	nano := time.Now().UnixNano()
	order := types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    nano,
		UpdateTime:    nano,
		Amount:        -70,
		Price:         100,
		Left:          -10,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            0,
	}
	order2 := types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    nano,
		UpdateTime:    nano,
		Amount:        50,
		Price:         99,
		Left:          20,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            0,
	}
	k.insertUnmatchedOrder(&order)
	k.insertUnmatchedOrder(&order)
	k.insertUnmatchedOrder(&order)
	k.insertUnmatchedOrder(&order2)
	k.insertUnmatchedOrder(&order2)
	k.insertUnmatchedOrder(&order2)

	value := k.ask.Front().Value()
	bucket := value.(*priceBucket)
	assert.Equal(t, int64(-30), bucket.Left)
	assert.Equal(t, 3, bucket.l.Len())
	assert.Equal(t, 1, k.ask.Length)

	value2 := k.bid.Front().Value()
	bucket2 := value2.(*priceBucket)
	assert.Equal(t, int64(60), bucket2.Left)
	assert.Equal(t, 3, bucket2.l.Len())
	assert.Equal(t, 1, k.bid.Length)
}

func Test_insertPriceCheckedOrder_WithRandomPrice(t *testing.T) {
	k := newKernel()
	testSize := 1_000_000
	asks := make([]*types.KernelOrder, 0, testSize)
	bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	var bidSize int64 = 0

	for i := 0; i < testSize; i++ {
		price := int64(i%2000) + 2001
		amount := int64(i%1000) + 1
		asks = append(asks, newTestAskOrder(price, amount))
		askSize -= amount
	}
	for i := 0; i < testSize; i++ {
		price := int64(i%2000) + 1
		amount := int64(i%1000) + 1
		bids = append(bids, newTestBidOrder(price, amount))
		bidSize += amount
	}

	for i := range asks {
		k.insertUnmatchedOrder(asks[i])
		k.insertUnmatchedOrder(bids[i])
	}

	var checkAskSize int64 = 0
	var checkBidSize int64 = 0
	for i := k.ask.Front(); i != nil; i = i.Next() {
		bucket := i.Value().(*priceBucket)
		checkAskSize += bucket.Left
	}
	for i := k.bid.Front(); i != nil; i = i.Next() {
		bucket := i.Value().(*priceBucket)
		checkBidSize += bucket.Left
	}
	assert.Equal(t, askSize, checkAskSize)
	assert.Equal(t, bidSize, checkBidSize)

	st := time.Now().UnixNano()
	k.takeSnapshot("test_insert", bids[len(bids)-1])
	et := time.Now().UnixNano()
	fmt.Println("Snapshot finished in ", (et-st)/(1000*1000), " ms")
}

func Test_insertOrder_SamePriceDifferentAmounts(t *testing.T) {
	k := newKernel()

	ask1 := newTestAskOrder(100, 50)
	ask2 := newTestAskOrder(100, 30)
	ask3 := newTestAskOrder(100, 20)

	k.insertUnmatchedOrder(ask1)
	k.insertUnmatchedOrder(ask2)
	k.insertUnmatchedOrder(ask3)

	assert.Equal(t, 1, k.ask.Length)
	bucket := k.ask.Front().Value().(*priceBucket)
	assert.Equal(t, int64(-100), bucket.Left)
	assert.Equal(t, 3, bucket.l.Len())
}

func Test_insertOrder_AscendingPrices(t *testing.T) {
	k := newKernel()

	for price := int64(100); price <= 110; price++ {
		k.insertUnmatchedOrder(newTestAskOrder(price, 10))
	}

	assert.Equal(t, 11, k.ask.Length)
}

func Test_insertOrder_DescendingPrices(t *testing.T) {
	k := newKernel()

	for price := int64(110); price >= 100; price-- {
		k.insertUnmatchedOrder(newTestAskOrder(price, 10))
	}

	assert.Equal(t, 11, k.ask.Length)
}
