/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package ker

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Curton/GoMatchingKernel/types"
)

func Test_insertPriceCheckedOrder_WithSamePrice(t *testing.T) {
	k := newKernel()
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
	k.insertCheckedOrder(&order)
	k.insertCheckedOrder(&order)
	k.insertCheckedOrder(&order)
	k.insertCheckedOrder(&order2)
	k.insertCheckedOrder(&order2)
	k.insertCheckedOrder(&order2)

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

// GOMAXPROCS=1 go test -bench=. -run=none -benchtime=1s -benchmem
// Benchmark_insertPriceCheckedOrder        6766042               187 ns/op              48 B/op          1 allocs/op
func Benchmark_insertPriceCheckedOrder(b *testing.B) {
	b.ReportAllocs()
	k := newKernel()
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
			Id:            0,
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
			Id:            0,
		}
		bids = append(bids, order)
		bidSize += i3
	}
	b.ResetTimer()
	for i := range asks {
		k.insertCheckedOrder(asks[i])
		k.insertCheckedOrder(bids[i])
	}
}

func Test_insertPriceCheckedOrder_WithRandomPrice(t *testing.T) {
	k := newKernel()
	testSize := 1_000_000
	asks := make([]*types.KernelOrder, 0, testSize)
	bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	var bidSize int64 = 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {

		i2 := r.Int63n(int64(2000)) + 1 + 2000
		i3 := r.Int63n(int64(1000)) + 1
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
			Id:            0,
		}
		asks = append(asks, order)
		askSize -= i3
	}
	for i := 0; i < testSize; i++ {
		i2 := r.Int63n(int64(2000)) + 1
		i3 := r.Int63n(int64(1000)) + 1
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
			Id:            0,
		}
		bids = append(bids, order)
		bidSize += i3
	}

	for i := range asks {
		k.insertCheckedOrder(asks[i])
		k.insertCheckedOrder(bids[i])
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

// There is only one order in the buy order (bid) list, and a sell order (ask) matches an order of the same price and the same quantity.
// After the matching is completed, the ask/bid is completely empty.
func Test_matchingAskOrder_MatchOneAndComplete(t *testing.T) {
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
		Id:            0,
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
		Id:            0,
	}

	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()

	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
		forCheck1 := types.KernelOrder{
			KernelOrderID: info.makerOrders[0].KernelOrderID,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        100,
			Price:         200,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		forCheck2 := types.KernelOrder{
			KernelOrderID: info.takerOrder.KernelOrderID,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		assert.Equal(t, forCheck1, info.makerOrders[0])
		assert.Equal(t, forCheck2, info.takerOrder)
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

// There is only one order in the sell order (ask) list, and a buy order (bid) matches an order of the same price and the same quantity.
// After the matching is completed, the ask/bid is completely empty.
func Test_matchingBidOrder_MatchOneAndComplete(t *testing.T) {
	acceptor := initAcceptor(1, "test")
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
		Id:            0,
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
		Id:            0,
	}

	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()
	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
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
			Id:            0,
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
			Id:            0,
		}
		forCheck1.KernelOrderID = info.takerOrder.KernelOrderID
		forCheck2.KernelOrderID = info.makerOrders[0].KernelOrderID
		assert.Equal(t, forCheck1, info.takerOrder)
		assert.Equal(t, forCheck2, info.makerOrders[0])
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

// There is only one order in the buy order (bid) list, and a sell order (ask) matches an order at a higher price and the same quantity.
// After the matching is completed, the ask/bid is completely empty.
func Test_matchingAskOrder_MatchOneAndComplete2(t *testing.T) {
	acceptor := initAcceptor(1, "test")
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
		Id:            0,
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
		Id:            0,
	}

	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()
	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
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
			Id:            0,
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
			Id:            0,
		}
		forCheck2.KernelOrderID = info.takerOrder.KernelOrderID
		forCheck1.KernelOrderID = info.makerOrders[0].KernelOrderID
		assert.Equal(t, forCheck1, info.makerOrders[0])
		assert.Equal(t, forCheck2, info.takerOrder)
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

// There is only one order in the sell order (ask) list, and a buy order (bid) matches an order at a higher price and the same quantity.
// After the matching is completed, the ask/bid is completely empty.
func Test_matchingBidOrder_MatchOneAndComplete2(t *testing.T) {
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
		Id:            0,
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
		Id:            0,
	}

	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()
	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
		forCheck1 := types.KernelOrder{
			KernelOrderID: info.takerOrder.KernelOrderID,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        100,
			Price:         300,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		forCheck2 := types.KernelOrder{
			KernelOrderID: info.makerOrders[0].KernelOrderID,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}

		assert.Equal(t, forCheck1, info.takerOrder)
		assert.Equal(t, forCheck2, info.makerOrders[0])
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

// There is only one order in the buy order (bid) list, and a sell order (ask) matches an order at the same price but with insufficient quantity.
// After the matching is completed, the bid is completely empty, and the remaining part of the ask creates a new pending order.
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
		Id:            0,
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
		Id:            0,
	}

	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()
	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
		// bid
		forCheck1 := types.KernelOrder{
			KernelOrderID: info.makerOrders[0].KernelOrderID,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        100,
			Price:         200,
			Left:          0,
			FilledTotal:   20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		// ask
		forCheck2 := types.KernelOrder{
			KernelOrderID: info.takerOrder.KernelOrderID,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        -1000,
			Price:         199,
			Left:          -900,
			FilledTotal:   -20000,
			Status:        types.OPEN,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		assert.Equal(t, forCheck1, info.makerOrders[0])
		assert.Equal(t, forCheck2, info.takerOrder)
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(199), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
	bucket := acceptor.kernel.ask.Front().value.(*priceBucket)
	kernelOrder := bucket.l.Back().Value.(*types.KernelOrder)
	assert.Equal(t, int64(-900), kernelOrder.Left)
	assert.Equal(t, int64(-20000), kernelOrder.FilledTotal)
}

// There is only one order in the sell order (ask) list, and a buy order (bid) matches an order at the same price but with insufficient quantity.
// After the matching is completed, the ask is completely empty, and the remaining part of the bid creates a new pending order.
func Test_matchingBidOrder_MatchOneButIncomplete2(t *testing.T) {
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
		Id:            0,
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
		Id:            0,
	}

	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()
	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
		// taker, bid
		forCheck1 := types.KernelOrder{
			KernelOrderID: info.takerOrder.KernelOrderID,
			CreateTime:    info.takerOrder.CreateTime,
			UpdateTime:    info.takerOrder.UpdateTime,
			Amount:        1000,
			Price:         300,
			Left:          900,
			FilledTotal:   20000,
			Status:        types.OPEN,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		// maker, ask
		forCheck2 := types.KernelOrder{
			KernelOrderID: info.makerOrders[0].KernelOrderID,
			CreateTime:    info.makerOrders[0].CreateTime,
			UpdateTime:    info.makerOrders[0].UpdateTime,
			Amount:        -100,
			Price:         200,
			Left:          0,
			FilledTotal:   -20000,
			Status:        types.CLOSED,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}
		assert.Equal(t, forCheck1, info.takerOrder)
		assert.Equal(t, forCheck2, info.makerOrders[0])
		i++
		if i == 1 {
			break
		}
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(300), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)
	bucket := acceptor.kernel.bid.Front().value.(*priceBucket)
	kernelOrder := bucket.l.Back().Value.(*types.KernelOrder)
	assert.Equal(t, int64(900), kernelOrder.Left)
	assert.Equal(t, int64(20000), kernelOrder.FilledTotal)
}

// There are multiple orders in the buy order (bid) list, and the sell order (ask) matches all orders.
// After the matching is completed, the bid is completely empty, and the remaining part of the ask creates a new pending order.
func Test_matchingAskOrder_MatchMultipleComplete(t *testing.T) {
	// bid
	order := &types.KernelOrder{
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
		Id:            0,
	}
	// bid2
	order2 := &types.KernelOrder{
		KernelOrderID: 2,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        150,
		Price:         200,
		Left:          150,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            0,
	}
	// bid3
	order3 := &types.KernelOrder{
		KernelOrderID: 3,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        110,
		Price:         199,
		Left:          110,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            0,
	}
	// ask
	order4 := &types.KernelOrder{
		KernelOrderID: 4,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -400,
		Price:         198,
		Left:          -400,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            0,
	}

	acceptor := initAcceptor(1, "test")

	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()
	acceptor.kernel.startDummyMatchedInfoChan()
	acceptor.newOrderChan <- order
	acceptor.newOrderChan <- order2
	acceptor.newOrderChan <- order3
	acceptor.newOrderChan <- order4

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(198), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
	left := acceptor.kernel.ask.Front().value.(*priceBucket).Left
	assert.Equal(t, int64(-40), left)
}

// There are multiple (200,000) orders in the buy order (bid) list, and the sell order (ask) matches exactly all orders.
// After the matching is completed, the bid is completely empty, and the remaining part of the ask creates a new pending order.
// test clearBucket
func Test_matchingAskOrder_MatchMultipleComplete2(t *testing.T) {
	testSize := 200_000
	asks := make([]*types.KernelOrder, 0, testSize)
	//bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	//var bidSize int64 = 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {
		// 1 - 1000, price
		i2 := r.Int63n(int64(1000)) + 1
		//i2 := int64(i + 1)
		// 1 - 100
		i3 := r.Int63n(int64(100)) + 1
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
			Id:            0,
		}
		asks = append(asks, order)
		askSize -= i3
	}
	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()
	acceptor.startDummyOrderConfirmedChan()

	takerVolumeMap := make(map[uint64]int64)
	makerVolumeMap := make(map[uint64]int64)
	go func() {
		for {
			info := <-acceptor.kernel.matchedInfoChan
			var checkSum int64 = 0
			for _, v := range info.matchedSizeMap {
				assert.NotEqual(t, int64(0), v)
				checkSum += v
			}
			assert.Equal(t, int64(0), checkSum)

			//fmt.Println(*info)

			// taker
			takerOrder := info.takerOrder
			i, ok := takerVolumeMap[takerOrder.KernelOrderID]
			if ok {
				if i < 0 && i > takerOrder.Amount-takerOrder.Left {
					takerVolumeMap[takerOrder.KernelOrderID] = takerOrder.Amount - takerOrder.Left
				} else if i > 0 && i < takerOrder.Amount-takerOrder.Left {
					takerVolumeMap[takerOrder.KernelOrderID] = takerOrder.Amount - takerOrder.Left
				}
			} else {
				takerVolumeMap[takerOrder.KernelOrderID] = takerOrder.Amount - takerOrder.Left
			}

			makerOrders := info.makerOrders
			for i2 := range makerOrders {
				mapV, ok := makerVolumeMap[makerOrders[i2].KernelOrderID]
				if ok {
					if mapV < 0 && mapV > makerOrders[i2].Amount-makerOrders[i2].Left {
						makerVolumeMap[makerOrders[i2].KernelOrderID] = makerOrders[i2].Amount - makerOrders[i2].Left
					} else if mapV > 0 && mapV < makerOrders[i2].Amount-makerOrders[i2].Left {
						makerVolumeMap[makerOrders[i2].KernelOrderID] = makerOrders[i2].Amount - makerOrders[i2].Left
					}
				} else {
					makerVolumeMap[makerOrders[i2].KernelOrderID] = makerOrders[i2].Amount - makerOrders[i2].Left
				}
			}

		}
	}()

	for i := range asks {
		acceptor.newOrderChan <- asks[i]
	}

	// bid, consuming all pending orders
	order2 := &types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        math.MaxInt64,
		//Amount:        200,
		Price: 500000,
		Left:  math.MaxInt64,
		//Left:          200,
		FilledTotal: 0,
		Status:      0,
		Type:        0,
		TimeInForce: 0,
		Id:          0,
	}
	acceptor.newOrderChan <- order2
	for acceptor.kernel.ask1Price != math.MaxInt64 {
		time.Sleep(time.Millisecond * 100)
		//println(ask1Price)
	}
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(500000), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)
	assert.Equal(t, math.MaxInt64+askSize, acceptor.kernel.bid.Front().value.(*priceBucket).Left)
	var takerSum int64
	for _, v := range takerVolumeMap {
		takerSum += v
	}
	var makerSum int64
	for _, v := range makerVolumeMap {
		makerSum += v
	}
	assert.Equal(t, takerSum, -askSize)
	assert.Equal(t, takerSum, -makerSum)
	assert.Equal(t, makerSum, askSize)
}

// 100,000 random orders, 50,000 buy orders (bid) & 50,000 sell orders (ask).
func Test_matchingOrders_withRandomPriceAndSize(t *testing.T) {
	testSize := 12_500 // * 8
	asks := make([]*types.KernelOrder, 0, testSize)
	bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	var bidSize int64 = 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {
		// 1 - 1000
		i2 := r.Int63n(int64(1000)) + 1
		// 1 - 100
		i3 := r.Int63n(int64(100)) + 1
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
			Id:            0,
		}
		asks = append(asks, order)
		askSize -= i3
	}
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {
		i2 := r.Int63n(int64(1000)) + 1
		i3 := r.Int63n(int64(100)) + 1
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
			Id:            0,
		}
		bids = append(bids, order)
		bidSize += i3
	}

	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()

	//takerOrderSizeMap := make(map[uint64]int64)
	//makerOrderSizeMap := make(map[uint64]int64)
	orderVolumeMap := make(map[uint64]int64)
	//mux0 := sync.RWMutex{}
	//mux1 := sync.RWMutex{}

	go func() {
		for {
			info := <-acceptor.kernel.matchedInfoChan
			order := info.takerOrder
			i, ok := orderVolumeMap[order.KernelOrderID]
			if ok {
				if i < 0 && i > order.Amount-order.Left {
					orderVolumeMap[order.KernelOrderID] = order.Amount - order.Left
				} else if i > 0 && i < order.Amount-order.Left {
					orderVolumeMap[order.KernelOrderID] = order.Amount - order.Left
				}
			} else {
				orderVolumeMap[order.KernelOrderID] = order.Amount - order.Left
			}

			orders := info.makerOrders
			for i2 := range orders {
				i3, ok := orderVolumeMap[orders[i2].KernelOrderID]
				if ok {
					if i3 < 0 && i3 > orders[i2].Amount-orders[i2].Left {
						orderVolumeMap[orders[i2].KernelOrderID] = orders[i2].Amount - orders[i2].Left
					} else if i3 > 0 && i3 < orders[i2].Amount-orders[i2].Left {
						orderVolumeMap[orders[i2].KernelOrderID] = orders[i2].Amount - orders[i2].Left
					}
				} else {
					orderVolumeMap[orders[i2].KernelOrderID] = orders[i2].Amount - orders[i2].Left
				}
			}
			var checkSum int64 = 0
			for _, v := range info.matchedSizeMap {
				assert.NotEqual(t, int64(0), v)
				checkSum += v
			}
			assert.Equal(t, int64(0), checkSum)
		}
	}()

	go func() {
		for {
			<-acceptor.orderReceivedChan
		}
	}()

	acceptor.enableRedoKernel()

	done := make(chan bool)
	start := time.Now().UnixNano()
	go func() {
		for i := range asks {
			acceptor.newOrderChan <- asks[i]
			acceptor.newOrderChan <- bids[i]

			acceptor.newOrderChan <- bids[i]
			acceptor.newOrderChan <- asks[i]

			acceptor.newOrderChan <- bids[i]
			acceptor.newOrderChan <- asks[i]

			acceptor.newOrderChan <- asks[i]
			acceptor.newOrderChan <- bids[i]
			if i == testSize-1 {
				done <- true
			}
		}
	}()

	// Wait for the order to be processed, then start the audit
	for b := range done {
		if b == true {
			// wait all matching finished
			matching_time := (time.Now().UnixNano() - start) / (1000 * 1000)
			println(8*testSize, "orders matching finished in ", matching_time, " ms, ", int64(8*testSize)/matching_time, " ops. per second")
			// wait for the order to be processed
			time.Sleep(time.Second)
			break
		}
	}

	wg := sync.WaitGroup{}
	var askLeftCalFromBucketLeft int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := acceptor.kernel.ask.Front(); bucket != nil; bucket = bucket.Next() {
			askLeftCalFromBucketLeft += bucket.value.(*priceBucket).Left
		}
	}()
	var askLeftCalFromList int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := acceptor.kernel.ask.Front(); bucket != nil; bucket = bucket.Next() {
			l := bucket.value.(*priceBucket).l
			for i := l.Front(); i != nil; i = i.Next() {
				askLeftCalFromList += i.Value.(*types.KernelOrder).Left
			}
		}
	}()

	var bidLeftCalFromBucketLeft int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := acceptor.kernel.bid.Front(); bucket != nil; bucket = bucket.Next() {
			bidLeftCalFromBucketLeft += bucket.value.(*priceBucket).Left
		}
	}()

	var bidLeftCalFromList int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := acceptor.kernel.bid.Front(); bucket != nil; bucket = bucket.Next() {
			l := bucket.value.(*priceBucket).l
			for i := l.Front(); i != nil; i = i.Next() {
				bidLeftCalFromList += i.Value.(*types.KernelOrder).Left
			}
		}
	}()

	wg.Wait()

	assert.Equal(t, askLeftCalFromBucketLeft, askLeftCalFromList)
	assert.Equal(t, bidLeftCalFromBucketLeft, bidLeftCalFromList)

	var askSum int64 = 0
	var bidSum int64 = 0
	for _, v := range orderVolumeMap {
		if v < 0 {
			askSum += v
		} else {
			bidSum += v
		}
	}
	assert.Equal(t, askSum, -bidSum)
	//st := time.Now().UnixNano()
	//acceptor.kernel.takeSnapshot("test")
	//et := time.Now().UnixNano()
	//fmt.Println("snapshot finished in ", (et-st)/(1000*1000), " ms")
	time.Sleep(3 * time.Second)
	//log.Println(*acceptor.kernel.fullDepth())
	// fmt.Println(acceptor.kernel.ask1Price)
	// fmt.Println(acceptor.kernel.bid1Price)
	// fmt.Println(acceptor.redoKernel.ask1Price)
	// fmt.Println(acceptor.redoKernel.bid1Price)
	assert.Equal(t, acceptor.kernel.ask1Price, acceptor.redoKernel.ask1Price)
	assert.Equal(t, acceptor.kernel.bid1Price, acceptor.redoKernel.bid1Price)
}

func TestRestoreKernel(t *testing.T) {
	baseDir := kernelSnapshotPath + "redo/"
	dir, err := os.ReadDir(baseDir)
	if err != nil {
		fmt.Println("Err in TestRestoreKernel", err.Error())
	}
	info := dir[len(dir)-1]
	st := time.Now().UnixNano()
	k, b := restoreKernel(baseDir + info.Name() + "/")
	et := time.Now().UnixNano()
	fmt.Println("Restore finished in ", (et-st)/(1000*1000), " ms")
	assert.Equal(t, true, b)
	_ = k
}

func Test_kernel_cancelOrder(t *testing.T) {
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
		Id:            0,
	}
	// ask
	order2 := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    0,
		UpdateTime:    0,
		Amount:        -100,
		Price:         201,
		Left:          -100,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            0,
	}

	acceptor := initAcceptor(1, "test")
	go acceptor.startOrderAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	go func() {
		acceptor.newOrderChan <- order
		acceptor.newOrderChan <- order2
	}()

	ids := make([]*types.KernelOrder, 0, 2)
	for i := 0; i < 2; i++ {
		v := <-acceptor.orderReceivedChan
		ids = append(ids, v)
	}

	go func() {
		for {
			<-acceptor.orderReceivedChan
		}
	}()

	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	for i := range ids {
		ids[i].Amount = 0
		acceptor.newOrderChan <- ids[i]
	}
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
}

// TODO
func Test_partialMatchWithMultipleAsks(t *testing.T) {
	// initialize orders and the acceptor here

	// send the bid order to the acceptor

	// iterate over multiple ask orders and send them to the acceptor

	// check that the bid order is partially filled and that its status is correct

	// check that each ask order is either fully filled or partially filled

	// check that the order book is updated correctly
}

func Test_partialMatchWithMultipleBids(t *testing.T) {
	// initialize orders and the acceptor here

	// send the ask order to the acceptor

	// iterate over multiple bid orders and send them to the acceptor

	// check that the ask order is partially filled and that its status is correct

	// check that each bid order is either fully filled or partially filled

	// check that the order book is updated correctly
}

func Test_invalidOrder(t *testing.T) {
	// initialize an invalid order and the acceptor here

	// send the invalid order to the acceptor

	// check that an error is returned and that the order is not added to the order book
}
