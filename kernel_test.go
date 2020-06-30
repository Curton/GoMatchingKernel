/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"sync"
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

	go orderAcceptor()
	orderChan <- order
	orderChan <- order2

	i := 0
	for info := range matchingInfoChan {
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
			Id:            "",
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
func Test_matchingBidOrder_MatchOneAndComplete(t *testing.T) {
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

	go orderAcceptor()
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

	go orderAcceptor()
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

	go orderAcceptor()
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

	go orderAcceptor()
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
func Test_matchingDidOrder_MatchOneButIncomplete2(t *testing.T) {
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

	go orderAcceptor()
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

// 买单(bid)列表有多个订单, 卖单(ask)匹配到刚好匹配完所有订单, 匹配完成后bid全空, ask剩余部分创建一个新挂单
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
		Id:            "",
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
		Id:            "",
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
		Id:            "",
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
		Id:            "",
	}

	go orderAcceptor()
	orderChan <- order
	orderChan <- order2
	orderChan <- order3
	orderChan <- order4

	go func() {
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, int64(198), ask1Price)
		assert.Equal(t, int64(math.MinInt64), bid1Price)
		assert.Equal(t, 1, ask.Length)
		assert.Equal(t, 0, bid.Length)
		close(orderChan)
		close(matchingInfoChan)
	}()

	//i := 0
	for info := range matchingInfoChan {
		fmt.Println(info.takerOrder)
		fmt.Println("------------------")
		for i := range info.makerOrders {
			fmt.Println(info.makerOrders[i])
		}
		fmt.Println("##################")

		//forCheck1 := types.KernelOrder{
		//	KernelOrderID: 0,
		//	CreateTime:    info.makerOrders[0].CreateTime,
		//	UpdateTime:    info.makerOrders[0].UpdateTime,
		//	Amount:        100,
		//	Price:         200,
		//	Left:          0,
		//	FilledTotal:   20000,
		//	Status:        types.CLOSED,
		//	Type:          0,
		//	TimeInForce:   0,
		//	Id:            "",
		//}
		//forCheck2 := types.KernelOrder{
		//	KernelOrderID: 1,
		//	CreateTime:    info.takerOrder.CreateTime,
		//	UpdateTime:    info.takerOrder.UpdateTime,
		//	Amount:        -100,
		//	Price:         200,
		//	Left:          0,
		//	FilledTotal:   -20000,
		//	Status:        types.CLOSED,
		//	Type:          0,
		//	TimeInForce:   0,
		//	Id:            "",
		//}
		//assert.Equal(t, forCheck1, info.makerOrders[0])
		//assert.Equal(t, forCheck2, info.takerOrder)
		//i++
		//if i == 1 {
		//	break
		//}
	}

}

// 买单(bid)列表有多个(2_000_000)订单, 卖单(ask)匹配到刚好匹配完所有订单, 匹配完成后bid全空, ask剩余部分创建一个新挂单
func Test_matchingAskOrder_MatchMultipleComplete2(t *testing.T) {
	testSize := 2_000_000
	asks := make([]*types.KernelOrder, 0, testSize)
	//bids := make([]*types.KernelOrder, 0, testSize)
	var askSize int64 = 0
	//var bidSize int64 = 0
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
			Id:            "",
		}
		asks = append(asks, order)
		askSize -= i3
	}
	//for i := 0; i < testSize; i++ {
	//	i2 := r.Int63n(int64(100)) + 99
	//	i3 := r.Int63n(int64(100))
	//	order := &types.KernelOrder{
	//		KernelOrderID: 0,
	//		CreateTime:    0,
	//		UpdateTime:    0,
	//		Amount:        i3,
	//		Price:         i2,
	//		Left:          i3,
	//		FilledTotal:   0,
	//		Status:        0,
	//		Type:          0,
	//		TimeInForce:   0,
	//		Id:            "",
	//	}
	//	bids = append(bids, order)
	//	bidSize += i3
	//}

	go orderAcceptor()

	go func() {
		for {
			<-matchingInfoChan
		}
	}()

	//go func() {
	//	for info := range matchingInfoChan {
	//		fmt.Println("taker: ", info.takerOrder)
	//		fmt.Println("------------------")
	//		for i := range info.makerOrders {
	//			fmt.Println("maker: ",info.makerOrders[i])
	//		}
	//		fmt.Println("##################")
	//	}
	//}()

	for i := range asks {
		orderChan <- asks[i]
	}

	// bid, 一个大单吃完
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
		Id:          "",
	}
	orderChan <- order2
	time.Sleep(time.Millisecond)
	for ask1Price != math.MaxInt64 {
		time.Sleep(time.Millisecond)
		//println(ask1Price)
	}
	assert.Equal(t, int64(math.MaxInt64), ask1Price)
	assert.Equal(t, int64(500000), bid1Price)
	assert.Equal(t, 0, ask.Length)
	assert.Equal(t, 1, bid.Length)
	assert.Equal(t, math.MaxInt64+askSize, bid.Front().value.(*priceBucket).Left)
}

// 1_000_000个随机订单, 500_000(bid) & 500_000(ask)
func Test_matchingOrders_withRandomPriceAndSize(t *testing.T) {
	testSize := 5
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
			Id:            "",
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
			Id:            "",
		}
		bids = append(bids, order)
		bidSize += i3
	}

	go orderAcceptor()

	takerOrderSizeMap := make(map[uint64]int64)
	makerOrderSizeMap := make(map[uint64]int64)
	mux0 := sync.RWMutex{}
	mux1 := sync.RWMutex{}
	wg0 := sync.WaitGroup{}

	go func() {
		for {
			info := <-matchingInfoChan
			wg0.Add(1)
			go func() {
				defer wg0.Done()
				order := info.takerOrder
				mux0.RLock()
				i, ok := takerOrderSizeMap[order.KernelOrderID]
				mux0.RUnlock()
				if ok {
					if i < 0 && i < order.Left {
						mux0.Lock()
						takerOrderSizeMap[order.KernelOrderID] = order.Amount - order.Left
						mux0.Unlock()
					} else if i > 0 && i > order.Left {
						mux0.Lock()
						takerOrderSizeMap[order.KernelOrderID] = order.Amount - order.Left
						mux0.Unlock()
					}
				} else {
					mux0.Lock()
					takerOrderSizeMap[order.KernelOrderID] = order.Amount - order.Left
					mux0.Unlock()
				}

				orders := info.makerOrders
				for i2 := range orders {
					mux1.RLock()
					i3, ok := makerOrderSizeMap[orders[i2].KernelOrderID]
					mux1.RUnlock()
					if ok {
						if i3 < 0 && i3 < orders[i2].Left {
							mux1.Lock()
							makerOrderSizeMap[orders[i2].KernelOrderID] = orders[i2].Amount - orders[i2].Left
							mux1.Unlock()
						} else if i3 > 0 && i3 > orders[i2].Left {
							mux1.Lock()
							makerOrderSizeMap[orders[i2].KernelOrderID] = orders[i2].Amount - orders[i2].Left
							mux1.Unlock()
						}
					} else {
						mux1.Lock()
						makerOrderSizeMap[orders[i2].KernelOrderID] = orders[i2].Amount - orders[i2].Left
						mux1.Unlock()
					}
				}
				fmt.Println(info)
			}()
		}
	}()

	done := make(chan bool)
	start := time.Now().UnixNano()
	go func() {
		for i := range asks {
			orderChan <- asks[i]
			orderChan <- bids[i]

			orderChan <- bids[i]
			orderChan <- asks[i]

			orderChan <- bids[i]
			orderChan <- asks[i]

			orderChan <- asks[i]
			orderChan <- bids[i]
			if i == testSize-1 {
				done <- true
			}
		}
	}()

	for b := range done {
		if b == true {
			// wait all matching finished
			time.Sleep(time.Millisecond)
			println(8*testSize, "orders done in ", (time.Now().UnixNano()-start)/(1000*1000), " ms")
			break
		}
	}

	wg := sync.WaitGroup{}
	var askLeftCalFromBucketLeft int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := ask.Front(); bucket != nil; bucket = bucket.Next() {
			askLeftCalFromBucketLeft += bucket.value.(*priceBucket).Left
		}
	}()
	var askLeftCalFromList int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := ask.Front(); bucket != nil; bucket = bucket.Next() {
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
		for bucket := bid.Front(); bucket != nil; bucket = bucket.Next() {
			bidLeftCalFromBucketLeft += bucket.value.(*priceBucket).Left
		}
	}()

	var bidLeftCalFromList int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := bid.Front(); bucket != nil; bucket = bucket.Next() {
			l := bucket.value.(*priceBucket).l
			for i := l.Front(); i != nil; i = i.Next() {
				bidLeftCalFromList += i.Value.(*types.KernelOrder).Left
			}
		}
	}()

	wg0.Wait()
	wg.Wait()

	assert.Equal(t, askLeftCalFromBucketLeft, askLeftCalFromList)
	assert.Equal(t, bidLeftCalFromBucketLeft, bidLeftCalFromList)
	fmt.Println("----------")
	//fmt.Println(4 * askSize)
	//fmt.Println(4 * bidSize)
	fmt.Println("----------")
	//fmt.Println(bidLeftCalFromBucketLeft)
	//fmt.Println("----------")
	//fmt.Println(-(4*askSize + askLeftCalFromBucketLeft))
	//fmt.Println(4*bidSize - bidLeftCalFromBucketLeft)
	fmt.Println("----------")
	var takerSum int64 = 0
	for k, v := range takerOrderSizeMap {
		fmt.Println(k, " : ", v)
		takerSum += v
	}
	fmt.Println("takerSum : ", takerSum)

	var makerSum int64 = 0
	for k, v := range makerOrderSizeMap {
		fmt.Println(k, " : ", v)
		takerSum += v
	}
	fmt.Println("makerSum : ", makerSum)
}
