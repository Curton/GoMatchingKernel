/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	_ "fmt"
	"github.com/stretchr/testify/assert"
	_ "math"
	"math/rand"
	_ "math/rand"
	"sync"
	_ "sync"
	"testing"
	"time"
)

func Test_insertPriceCheckedOrder_WithSamePrice(t *testing.T) {
	k := NewKernel()
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

//GOMAXPROCS=1 go test -bench=. -run=none -benchtime=1s -benchmem
//Benchmark_insertPriceCheckedOrder        6766042               187 ns/op              48 B/op          1 allocs/op
func Benchmark_insertPriceCheckedOrder(b *testing.B) {
	b.ReportAllocs()
	k := NewKernel()
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
	k := NewKernel()
	testSize := 1_000_000
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
}

//// 买单(bid)列表只有一个订单, 卖单(ask)匹配到一个同价格且同数量订单, 匹配完成后ask/bid全空
//func Test_matchingAskOrder_MatchOneAndComplete(t *testing.T) {
//	// bid
//	order := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        100,
//		Price:         200,
//		Left:          100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// ask
//	order2 := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -100,
//		Price:         200,
//		Left:          -100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//
//	i := 0
//	for info := range matchingInfoChan {
//		forCheck1 := types.KernelOrder{
//			KernelOrderID: info.makerOrders[0].KernelOrderID,
//			CreateTime:    info.makerOrders[0].CreateTime,
//			UpdateTime:    info.makerOrders[0].UpdateTime,
//			Amount:        100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		forCheck2 := types.KernelOrder{
//			KernelOrderID: info.takerOrder.KernelOrderID,
//			CreateTime:    info.takerOrder.CreateTime,
//			UpdateTime:    info.takerOrder.UpdateTime,
//			Amount:        -100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   -20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		assert.Equal(t, forCheck1, info.makerOrders[0])
//		assert.Equal(t, forCheck2, info.takerOrder)
//		i++
//		if i == 1 {
//			break
//		}
//	}
//	time.Sleep(100 * time.Millisecond)
//	assert.Equal(t, int64(math.MaxInt64), ask1Price)
//	assert.Equal(t, int64(math.MinInt64), bid1Price)
//	assert.Equal(t, 0, ask.Length)
//	assert.Equal(t, 0, bid.Length)
//	close(orderChan)
//}
//
//// 卖单(ask)列表只有一个订单, 买单(bid)匹配到一个同价格且同数量订单, 匹配完成后ask/bid全空
//func Test_matchingBidOrder_MatchOneAndComplete(t *testing.T) {
//	// ask
//	order := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -100,
//		Price:         200,
//		Left:          -100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// bid
//	order2 := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        100,
//		Price:         200,
//		Left:          100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//
//	i := 0
//	for info := range matchingInfoChan {
//		// taker, bid
//		forCheck1 := types.KernelOrder{
//			KernelOrderID: 1,
//			CreateTime:    info.takerOrder.CreateTime,
//			UpdateTime:    info.takerOrder.UpdateTime,
//			Amount:        100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		// maker, ask
//		forCheck2 := types.KernelOrder{
//			KernelOrderID: 0,
//			CreateTime:    info.makerOrders[0].CreateTime,
//			UpdateTime:    info.makerOrders[0].UpdateTime,
//			Amount:        -100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   -20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		assert.Equal(t, forCheck1, info.takerOrder)
//		assert.Equal(t, forCheck2, info.makerOrders[0])
//		i++
//		if i == 1 {
//			break
//		}
//	}
//	time.Sleep(100 * time.Millisecond)
//	assert.Equal(t, int64(math.MaxInt64), ask1Price)
//	assert.Equal(t, int64(math.MinInt64), bid1Price)
//	assert.Equal(t, 0, ask.Length)
//	assert.Equal(t, 0, bid.Length)
//	close(orderChan)
//}
//
//// 买单(bid)列表只有一个订单, 卖单(ask)匹配到一个更高价格且同数量订单, 匹配完成后ask/bid全空
//func Test_matchingAskOrder_MatchOneAndComplete2(t *testing.T) {
//	order := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        100,
//		Price:         200,
//		Left:          100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	order2 := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -100,
//		Price:         100,
//		Left:          -100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//
//	i := 0
//	for info := range matchingInfoChan {
//		forCheck1 := types.KernelOrder{
//			KernelOrderID: 0,
//			CreateTime:    info.makerOrders[0].CreateTime,
//			UpdateTime:    info.makerOrders[0].UpdateTime,
//			Amount:        100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		forCheck2 := types.KernelOrder{
//			KernelOrderID: 1,
//			CreateTime:    info.takerOrder.CreateTime,
//			UpdateTime:    info.takerOrder.UpdateTime,
//			Amount:        -100,
//			Price:         100,
//			Left:          0,
//			FilledTotal:   -20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		assert.Equal(t, forCheck1, info.makerOrders[0])
//		assert.Equal(t, forCheck2, info.takerOrder)
//		i++
//		if i == 1 {
//			break
//		}
//	}
//	time.Sleep(100 * time.Millisecond)
//	assert.Equal(t, int64(math.MaxInt64), ask1Price)
//	assert.Equal(t, int64(math.MinInt64), bid1Price)
//	assert.Equal(t, 0, ask.Length)
//	assert.Equal(t, 0, bid.Length)
//	close(orderChan)
//}
//
//// 卖单(ask)列表只有一个订单, 买单(bid)匹配到一个更高价格且同数量订单, 匹配完成后ask/bid全空
//func Test_matchingBidOrder_MatchOneAndComplete2(t *testing.T) {
//	// ask
//	order := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -100,
//		Price:         200,
//		Left:          -100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// bid
//	order2 := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        100,
//		Price:         300,
//		Left:          100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//
//	i := 0
//	for info := range matchingInfoChan {
//		forCheck1 := types.KernelOrder{
//			KernelOrderID: 1,
//			CreateTime:    info.takerOrder.CreateTime,
//			UpdateTime:    info.takerOrder.UpdateTime,
//			Amount:        100,
//			Price:         300,
//			Left:          0,
//			FilledTotal:   20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		forCheck2 := types.KernelOrder{
//			KernelOrderID: 0,
//			CreateTime:    info.makerOrders[0].CreateTime,
//			UpdateTime:    info.makerOrders[0].UpdateTime,
//			Amount:        -100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   -20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		assert.Equal(t, forCheck1, info.takerOrder)
//		assert.Equal(t, forCheck2, info.makerOrders[0])
//		i++
//		if i == 1 {
//			break
//		}
//	}
//	time.Sleep(100 * time.Millisecond)
//	assert.Equal(t, int64(math.MaxInt64), ask1Price)
//	assert.Equal(t, int64(math.MinInt64), bid1Price)
//	assert.Equal(t, 0, ask.Length)
//	assert.Equal(t, 0, bid.Length)
//	close(orderChan)
//}
//
//// 买单(bid)列表只有一个订单, 卖单(ask)匹配到一个同价格且同但数量不足的订单, 匹配完成后bid全空, ask剩余部分创建一个新挂单
//func Test_matchingAskOrder_MatchOneButIncomplete(t *testing.T) {
//	// bid
//	order := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        100,
//		Price:         200,
//		Left:          100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// ask
//	order2 := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -1000,
//		Price:         199,
//		Left:          -1000,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//
//	i := 0
//	for info := range matchingInfoChan {
//		// bid
//		forCheck1 := types.KernelOrder{
//			KernelOrderID: 0,
//			CreateTime:    info.makerOrders[0].CreateTime,
//			UpdateTime:    info.makerOrders[0].UpdateTime,
//			Amount:        100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		// ask
//		forCheck2 := types.KernelOrder{
//			KernelOrderID: 1,
//			CreateTime:    info.takerOrder.CreateTime,
//			UpdateTime:    info.takerOrder.UpdateTime,
//			Amount:        -1000,
//			Price:         199,
//			Left:          -900,
//			FilledTotal:   -20000,
//			Status:        types.OPEN,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		assert.Equal(t, forCheck1, info.makerOrders[0])
//		assert.Equal(t, forCheck2, info.takerOrder)
//		i++
//		if i == 1 {
//			break
//		}
//	}
//	time.Sleep(100 * time.Millisecond)
//	assert.Equal(t, int64(199), ask1Price)
//	assert.Equal(t, int64(math.MinInt64), bid1Price)
//	assert.Equal(t, 1, ask.Length)
//	assert.Equal(t, 0, bid.Length)
//	bucket := ask.Front().value.(*priceBucket)
//	kernelOrder := bucket.l.Back().Value.(*types.KernelOrder)
//	assert.Equal(t, int64(-900), kernelOrder.Left)
//	assert.Equal(t, int64(-20000), kernelOrder.FilledTotal)
//	close(orderChan)
//}
//
//// 卖单(ask)列表只有一个订单, 买单(bid)匹配到一个同价格且同但数量不足的订单, 匹配完成后ask全空, bid剩余部分创建一个新挂单
//func Test_matchingDidOrder_MatchOneButIncomplete2(t *testing.T) {
//	// ask
//	order := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -100,
//		Price:         200,
//		Left:          -100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// bid
//	order2 := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        1000,
//		Price:         300,
//		Left:          1000,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//
//	i := 0
//	for info := range matchingInfoChan {
//		// taker, bid
//		forCheck1 := types.KernelOrder{
//			KernelOrderID: 1,
//			CreateTime:    info.takerOrder.CreateTime,
//			UpdateTime:    info.takerOrder.UpdateTime,
//			Amount:        1000,
//			Price:         300,
//			Left:          900,
//			FilledTotal:   20000,
//			Status:        types.OPEN,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		// maker, ask
//		forCheck2 := types.KernelOrder{
//			KernelOrderID: 0,
//			CreateTime:    info.makerOrders[0].CreateTime,
//			UpdateTime:    info.makerOrders[0].UpdateTime,
//			Amount:        -100,
//			Price:         200,
//			Left:          0,
//			FilledTotal:   -20000,
//			Status:        types.CLOSED,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		assert.Equal(t, forCheck1, info.takerOrder)
//		assert.Equal(t, forCheck2, info.makerOrders[0])
//		i++
//		if i == 1 {
//			break
//		}
//	}
//	time.Sleep(100 * time.Millisecond)
//	assert.Equal(t, int64(math.MaxInt64), ask1Price)
//	assert.Equal(t, int64(300), bid1Price)
//	assert.Equal(t, 0, ask.Length)
//	assert.Equal(t, 1, bid.Length)
//	bucket := bid.Front().value.(*priceBucket)
//	kernelOrder := bucket.l.Back().Value.(*types.KernelOrder)
//	assert.Equal(t, int64(900), kernelOrder.Left)
//	assert.Equal(t, int64(20000), kernelOrder.FilledTotal)
//	close(orderChan)
//}
//
//// 买单(bid)列表有多个订单, 卖单(ask)匹配到刚好匹配完所有订单, 匹配完成后bid全空, ask剩余部分创建一个新挂单
//func Test_matchingAskOrder_MatchMultipleComplete(t *testing.T) {
//	// bid
//	order := &types.KernelOrder{
//		KernelOrderID: 1,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        100,
//		Price:         200,
//		Left:          100,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// bid2
//	order2 := &types.KernelOrder{
//		KernelOrderID: 2,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        150,
//		Price:         200,
//		Left:          150,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// bid3
//	order3 := &types.KernelOrder{
//		KernelOrderID: 3,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        110,
//		Price:         199,
//		Left:          110,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//	// ask
//	order4 := &types.KernelOrder{
//		KernelOrderID: 4,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        -400,
//		Price:         198,
//		Left:          -400,
//		FilledTotal:   0,
//		Status:        0,
//		Type:          0,
//		TimeInForce:   0,
//		Id:            0,
//	}
//
//	go startOrderAcceptor()
//	orderChan <- order
//	orderChan <- order2
//	orderChan <- order3
//	orderChan <- order4
//
//	go func() {
//		time.Sleep(100 * time.Millisecond)
//		assert.Equal(t, int64(198), ask1Price)
//		assert.Equal(t, int64(math.MinInt64), bid1Price)
//		assert.Equal(t, 1, ask.Length)
//		assert.Equal(t, 0, bid.Length)
//		close(orderChan)
//		close(matchingInfoChan)
//	}()
//
//	//i := 0
//	for info := range matchingInfoChan {
//		fmt.Println(info.takerOrder)
//		fmt.Println("------------------")
//		for i := range info.makerOrders {
//			fmt.Println(info.makerOrders[i])
//		}
//		fmt.Println("##################")
//
//		//forCheck1 := types.KernelOrder{
//		//	KernelOrderID: 0,
//		//	CreateTime:    info.makerOrders[0].CreateTime,
//		//	UpdateTime:    info.makerOrders[0].UpdateTime,
//		//	Amount:        100,
//		//	Price:         200,
//		//	Left:          0,
//		//	FilledTotal:   20000,
//		//	Status:        types.CLOSED,
//		//	Type:          0,
//		//	TimeInForce:   0,
//		//	Id:            0,
//		//}
//		//forCheck2 := types.KernelOrder{
//		//	KernelOrderID: 1,
//		//	CreateTime:    info.takerOrder.CreateTime,
//		//	UpdateTime:    info.takerOrder.UpdateTime,
//		//	Amount:        -100,
//		//	Price:         200,
//		//	Left:          0,
//		//	FilledTotal:   -20000,
//		//	Status:        types.CLOSED,
//		//	Type:          0,
//		//	TimeInForce:   0,
//		//	Id:            0,
//		//}
//		//assert.Equal(t, forCheck1, info.makerOrders[0])
//		//assert.Equal(t, forCheck2, info.takerOrder)
//		//i++
//		//if i == 1 {
//		//	break
//		//}
//	}
//
//}
//
//// 买单(bid)列表有多个(2_000_000)订单, 卖单(ask)匹配到刚好匹配完所有订单, 匹配完成后bid全空, ask剩余部分创建一个新挂单
//// test clearBucket
//func Test_matchingAskOrder_MatchMultipleComplete2(t *testing.T) {
//	testSize := 2_000_000
//	//testSize := 100_000
//	asks := make([]*types.KernelOrder, 0, testSize)
//	//bids := make([]*types.KernelOrder, 0, testSize)
//	var askSize int64 = 0
//	//var bidSize int64 = 0
//	r := rand.New(rand.NewSource(time.Now().UnixNano()))
//	for i := 0; i < testSize; i++ {
//		// 1 - 1000, price
//		i2 := r.Int63n(int64(1000)) + 1
//		//i2 := int64(i + 1)
//		// 1 - 100
//		i3 := r.Int63n(int64(100)) + 1
//		order := &types.KernelOrder{
//			KernelOrderID: 0,
//			CreateTime:    0,
//			UpdateTime:    0,
//			Amount:        -i3,
//			Price:         i2,
//			Left:          -i3,
//			FilledTotal:   0,
//			Status:        0,
//			Type:          0,
//			TimeInForce:   0,
//			Id:            0,
//		}
//		asks = append(asks, order)
//		askSize -= i3
//	}
//
//	go startOrderAcceptor()
//
//	takerVolumeMap := make(map[uint64]int64)
//	makerVolumeMap := make(map[uint64]int64)
//	go func() {
//		for {
//			info := <-matchingInfoChan
//			var checkSum int64 = 0
//			for _, v := range info.matchedSizeMap {
//				assert.NotEqual(t, int64(0), v)
//				checkSum += v
//			}
//			assert.Equal(t, int64(0), checkSum)
//
//			//fmt.Println(*info)
//
//			// taker
//			takerOrder := info.takerOrder
//			i, ok := takerVolumeMap[takerOrder.KernelOrderID]
//			if ok {
//				if i < 0 && i > takerOrder.Amount-takerOrder.Left {
//					takerVolumeMap[takerOrder.KernelOrderID] = takerOrder.Amount - takerOrder.Left
//				} else if i > 0 && i < takerOrder.Amount-takerOrder.Left {
//					takerVolumeMap[takerOrder.KernelOrderID] = takerOrder.Amount - takerOrder.Left
//				}
//			} else {
//				takerVolumeMap[takerOrder.KernelOrderID] = takerOrder.Amount - takerOrder.Left
//			}
//
//			makerOrders := info.makerOrders
//			for i2 := range makerOrders {
//				mapV, ok := makerVolumeMap[makerOrders[i2].KernelOrderID]
//				if ok {
//					if mapV < 0 && mapV > makerOrders[i2].Amount-makerOrders[i2].Left {
//						makerVolumeMap[makerOrders[i2].KernelOrderID] = makerOrders[i2].Amount - makerOrders[i2].Left
//					} else if mapV > 0 && mapV < makerOrders[i2].Amount-makerOrders[i2].Left {
//						makerVolumeMap[makerOrders[i2].KernelOrderID] = makerOrders[i2].Amount - makerOrders[i2].Left
//					}
//				} else {
//					makerVolumeMap[makerOrders[i2].KernelOrderID] = makerOrders[i2].Amount - makerOrders[i2].Left
//				}
//			}
//
//		}
//	}()
//
//	for i := range asks {
//		orderChan <- asks[i]
//	}
//
//	// bid, 一个大单吃完
//	order2 := &types.KernelOrder{
//		KernelOrderID: 0,
//		CreateTime:    0,
//		UpdateTime:    0,
//		Amount:        math.MaxInt64,
//		//Amount:        200,
//		Price: 500000,
//		Left:  math.MaxInt64,
//		//Left:          200,
//		FilledTotal: 0,
//		Status:      0,
//		Type:        0,
//		TimeInForce: 0,
//		Id:          0,
//	}
//	orderChan <- order2
//	for ask1Price != math.MaxInt64 {
//		time.Sleep(time.Millisecond * 100)
//		//println(ask1Price)
//	}
//	assert.Equal(t, int64(math.MaxInt64), ask1Price)
//	assert.Equal(t, int64(500000), bid1Price)
//	assert.Equal(t, 0, ask.Length)
//	assert.Equal(t, 1, bid.Length)
//	assert.Equal(t, math.MaxInt64+askSize, bid.Front().value.(*priceBucket).Left)
//	var takerSum int64
//	for _, v := range takerVolumeMap {
//		takerSum += v
//	}
//	var makerSum int64
//	for _, v := range makerVolumeMap {
//		makerSum += v
//	}
//	assert.Equal(t, takerSum, -askSize)
//	assert.Equal(t, takerSum, -makerSum)
//	assert.Equal(t, makerSum, askSize)
//	//fmt.Println("takerSum : ", takerSum)
//	//fmt.Println("makerSum : ", makerSum)
//	//fmt.Println("askSize : ", askSize)
//}
//
// 100_000个随机订单, 50_000(bid) & 50_000(ask)
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

	// 等待订单处理完毕, 开始审计
	for b := range done {
		if b == true {
			// wait all matching finished
			println(8*testSize, "orders done in ", (time.Now().UnixNano()-start)/(1000*1000), " ms")
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
}
