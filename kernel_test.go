/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func Test_insertPriceCheckedOrder_WithSamePrice(t *testing.T) {
	// insert ask & bid order to empty 'ask' & 'bid'
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
		Id:            "",
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
		Id:            "",
	}
	insertPriceCheckedOrder(&order)
	insertPriceCheckedOrder(&order)
	insertPriceCheckedOrder(&order)
	insertPriceCheckedOrder(&order2)
	insertPriceCheckedOrder(&order2)
	insertPriceCheckedOrder(&order2)
	value := ask.Front().Value()
	bucket := value.(*priceBucket)
	assert.Equal(t, int64(-30), bucket.size)
	assert.Equal(t, 3, bucket.l.Len())
	assert.Equal(t, 1, ask.Length)

	value2 := bid.Front().Value()
	bucket2 := value2.(*priceBucket)
	assert.Equal(t, int64(60), bucket2.size)
	assert.Equal(t, 3, bucket2.l.Len())
	assert.Equal(t, 1, bid.Length)
}

// GOMAXPROCS=1 go test -bench=. -run=none -benchtime=1s -benchmem
// Benchmark_insertPriceCheckedOrder        7015816               235 ns/op              48 B/op          1 allocs/op
func Benchmark_insertPriceCheckedOrder(b *testing.B) {
	testSize := 1
	if b.N > 1 {
		testSize = b.N / 2
	}
	asks := make([]*types.KernelOrder, 0, testSize)
	bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	var bidSize int64 = 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {
		i2 := r.Int63n(int64(100)) + 200
		i3 := r.Int63n(int64(100))
		order := &types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    0,
			UpdateTime:    0,
			Amount:        -i3,
			Price:         i2,
			Left:          -i3,
			FilledTotal:   0,
			Status:        0,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		asks = append(asks, order)
		askSize -= i3
	}
	for i := 0; i < testSize; i++ {
		i2 := r.Int63n(int64(100)) + 99
		i3 := r.Int63n(int64(100))
		order := &types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    0,
			UpdateTime:    0,
			Amount:        i3,
			Price:         i2,
			Left:          i3,
			FilledTotal:   0,
			Status:        0,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		bids = append(bids, order)
		bidSize += i3
	}
	b.ResetTimer()
	for i := range asks {
		insertPriceCheckedOrder(asks[i])
		insertPriceCheckedOrder(bids[i])
	}
}

func Test_insertPriceCheckedOrder_WithRandomPrice(t *testing.T) {
	testSize := 100_000
	asks := make([]*types.KernelOrder, 0, testSize)
	bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	var bidSize int64 = 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {

		i2 := r.Int63n(int64(100)) + 200
		i3 := r.Int63n(int64(100))
		order := &types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    0,
			UpdateTime:    0,
			Amount:        -i3,
			Price:         i2,
			Left:          -i3,
			FilledTotal:   0,
			Status:        0,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		asks = append(asks, order)
		askSize -= i3
	}
	for i := 0; i < testSize; i++ {
		i2 := r.Int63n(int64(100)) + 99
		i3 := r.Int63n(int64(100))
		order := &types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    0,
			UpdateTime:    0,
			Amount:        i3,
			Price:         i2,
			Left:          i3,
			FilledTotal:   0,
			Status:        0,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		bids = append(bids, order)
		bidSize += i3
	}

	for i := range asks {
		insertPriceCheckedOrder(asks[i])
		insertPriceCheckedOrder(bids[i])
	}

	var checkAskSize int64 = 0
	var checkBidSize int64 = 0
	for i := ask.Front(); i != nil; i = i.Next() {
		bucket := i.Value().(*priceBucket)
		checkAskSize += bucket.size
	}
	for i := bid.Front(); i != nil; i = i.Next() {
		bucket := i.Value().(*priceBucket)
		checkBidSize += bucket.size
	}
	assert.Equal(t, askSize, checkAskSize)
	assert.Equal(t, bidSize, checkBidSize)
}

// 匹配到一个订单, 刚好完成
func Test_matchingAskOrder_MatchFirstComplete(t *testing.T) {

}

func TestRand(t *testing.T) {
	for i := 0; i < 100; i++ {
		println(rand.Int63n(int64(20)))
	}
}
