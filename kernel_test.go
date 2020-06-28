/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"github.com/stretchr/testify/assert"
	"math"
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
	insertCheckedOrder(&order)
	insertCheckedOrder(&order)
	insertCheckedOrder(&order)
	insertCheckedOrder(&order2)
	insertCheckedOrder(&order2)
	insertCheckedOrder(&order2)
	value := ask.Front().Value()
	bucket := value.(*priceBucket)
	assert.Equal(t, int64(-30), bucket.Left)
	assert.Equal(t, 3, bucket.l.Len())
	assert.Equal(t, 1, ask.Length)

	value2 := bid.Front().Value()
	bucket2 := value2.(*priceBucket)
	assert.Equal(t, int64(60), bucket2.Left)
	assert.Equal(t, 3, bucket2.l.Len())
	assert.Equal(t, 1, bid.Length)
}

// GOMAXPROCS=1 go test -bench=. -run=none -benchtime=1s -benchmem
// Benchmark_insertPriceCheckedOrder        6766042               187 ns/op              48 B/op          1 allocs/op
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
		insertCheckedOrder(asks[i])
		insertCheckedOrder(bids[i])
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
		insertCheckedOrder(asks[i])
		insertCheckedOrder(bids[i])
	}

	var checkAskSize int64 = 0
	var checkBidSize int64 = 0
	for i := ask.Front(); i != nil; i = i.Next() {
		bucket := i.Value().(*priceBucket)
		checkAskSize += bucket.Left
	}
	for i := bid.Front(); i != nil; i = i.Next() {
		bucket := i.Value().(*priceBucket)
		checkBidSize += bucket.Left
	}
	assert.Equal(t, askSize, checkAskSize)
	assert.Equal(t, bidSize, checkBidSize)
}

// 买单(bid)列表只有一个订单, 卖单(ask)匹配到一个同价格且同数量订单, 匹配完成后ask/bid全空
func Test_matchingAskOrder_MatchOneAndComplete0(t *testing.T) {
	// bid
	order := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        100,
		Price:         200,
		Left:          100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	// ask
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -100,
		Price:         200,
		Left:          -100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}

	go newOrderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
		forCheck1 := types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        100,
			Price:         200,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		forCheck2 := types.KernelOrder{
			KernelOrderID: 1,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		assert.Equal(t, forCheck1, info.makerOrders[0])
		assert.Equal(t, forCheck2, info.takerOrder)
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), ask1Price)
	assert.Equal(t, int64(math.MinInt64), bid1Price)
	assert.Equal(t, 0, ask.Length)
	assert.Equal(t, 0, bid.Length)
	close(orderChan)
}

// 卖单(ask)列表只有一个订单, 买单(bid)匹配到一个同价格且同数量订单, 匹配完成后ask/bid全空
func Test_matchingAskOrder_MatchOneAndComplete1(t *testing.T) {
	// ask
	order := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -100,
		Price:         200,
		Left:          -100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	// bid
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        100,
		Price:         200,
		Left:          100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}

	go newOrderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
		// taker, bid
		forCheck1 := types.KernelOrder{
			KernelOrderID: 1,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        100,
			Price:         200,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		// maker, ask
		forCheck2 := types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		assert.Equal(t, forCheck1, info.takerOrder)
		assert.Equal(t, forCheck2, info.makerOrders[0])
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), ask1Price)
	assert.Equal(t, int64(math.MinInt64), bid1Price)
	assert.Equal(t, 0, ask.Length)
	assert.Equal(t, 0, bid.Length)
	close(orderChan)
}

// 买单(bid)列表只有一个订单, 卖单(ask)匹配到一个更高价格且同数量订单, 匹配完成后ask/bid全空
func Test_matchingAskOrder_MatchOneAndComplete2(t *testing.T) {
	order := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        100,
		Price:         200,
		Left:          100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -100,
		Price:         100,
		Left:          -100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}

	go newOrderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
		forCheck1 := types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        100,
			Price:         200,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		forCheck2 := types.KernelOrder{
			KernelOrderID: 1,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        -100,
			Price:         100,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		assert.Equal(t, forCheck1, info.makerOrders[0])
		assert.Equal(t, forCheck2, info.takerOrder)
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), ask1Price)
	assert.Equal(t, int64(math.MinInt64), bid1Price)
	assert.Equal(t, 0, ask.Length)
	assert.Equal(t, 0, bid.Length)
	close(orderChan)
}

// 卖单(ask)列表只有一个订单, 买单(bid)匹配到一个更高价格且同数量订单, 匹配完成后ask/bid全空
func Test_matchingAskOrder_MatchOneAndComplete3(t *testing.T) {
	// ask
	order := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -100,
		Price:         200,
		Left:          -100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	// bid
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        100,
		Price:         300,
		Left:          100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}

	go newOrderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
		forCheck1 := types.KernelOrder{
			KernelOrderID: 1,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        100,
			Price:         300,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		forCheck2 := types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		assert.Equal(t, forCheck1, info.takerOrder)
		assert.Equal(t, forCheck2, info.makerOrders[0])
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), ask1Price)
	assert.Equal(t, int64(math.MinInt64), bid1Price)
	assert.Equal(t, 0, ask.Length)
	assert.Equal(t, 0, bid.Length)
	close(orderChan)
}

// 买单(bid)列表只有一个订单, 卖单(ask)匹配到一个同价格且同但数量不足的订单, 匹配完成后bid全空, ask剩余部分创建一个新挂单
func Test_matchingAskOrder_MatchOneButIncomplete(t *testing.T) {
	// bid
	order := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        100,
		Price:         200,
		Left:          100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	// ask
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -1000,
		Price:         199,
		Left:          -1000,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}

	go newOrderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
		// bid
		forCheck1 := types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        100,
			Price:         200,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		// ask
		forCheck2 := types.KernelOrder{
			KernelOrderID: 1,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        -1000,
			Price:         199,
			Left:          -900,
			FilledTotal:   -20000,
			Status:        types.OPEN,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		assert.Equal(t, forCheck1, info.makerOrders[0])
		assert.Equal(t, forCheck2, info.takerOrder)
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(199), ask1Price)
	assert.Equal(t, int64(math.MinInt64), bid1Price)
	assert.Equal(t, 1, ask.Length)
	assert.Equal(t, 0, bid.Length)
	bucket := ask.Front().value.(*priceBucket)
	kernelOrder := bucket.l.Back().Value.(*types.KernelOrder)
	assert.Equal(t, int64(-900), kernelOrder.Left)
	assert.Equal(t, int64(-20000), kernelOrder.FilledTotal)
	close(orderChan)
}

// 卖单(ask)列表只有一个订单, 买单(bid)匹配到一个同价格且同但数量不足的订单, 匹配完成后ask全空, bid剩余部分创建一个新挂单
func Test_matchingAskOrder_MatchOneButIncomplete2(t *testing.T) {
	// ask
	order := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -100,
		Price:         200,
		Left:          -100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	// bid
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        1000,
		Price:         300,
		Left:          1000,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}

	go newOrderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
		// taker, bid
		forCheck1 := types.KernelOrder{
			KernelOrderID: 1,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        1000,
			Price:         300,
			Left:          900,
			FilledTotal:   20000,
			Status:        types.OPEN,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		// maker, ask
		forCheck2 := types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            "",
		}
		assert.Equal(t, forCheck1, info.takerOrder)
		assert.Equal(t, forCheck2, info.makerOrders[0])
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), ask1Price)
	assert.Equal(t, int64(300), bid1Price)
	assert.Equal(t, 0, ask.Length)
	assert.Equal(t, 1, bid.Length)
	bucket := bid.Front().value.(*priceBucket)
	kernelOrder := bucket.l.Back().Value.(*types.KernelOrder)
	assert.Equal(t, int64(900), kernelOrder.Left)
	assert.Equal(t, int64(20000), kernelOrder.FilledTotal)
	close(orderChan)
}
