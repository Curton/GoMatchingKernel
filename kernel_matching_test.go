/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package ker

import (
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Curton/GoMatchingKernel/types"
)

func Test_matchingAskOrder_MatchOneAndComplete(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	ask := newTestAskOrder(200, 100)

	acceptor.newOrderChan <- bid
	acceptor.newOrderChan <- ask

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

func Test_matchingBidOrder_MatchOneAndComplete(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	ask := newTestAskOrder(200, 100)
	bid := newTestBidOrder(200, 100)

	acceptor.newOrderChan <- ask
	acceptor.newOrderChan <- bid

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
		forCheck1 := types.KernelOrder{
			KernelOrderID: info.takerOrder.KernelOrderID,
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
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

func Test_matchingAskOrder_MatchOneAndComplete2(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	ask := newTestAskOrder(100, 100)

	acceptor.newOrderChan <- bid
	acceptor.newOrderChan <- ask

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
			Price:         100,
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

func Test_matchingBidOrder_MatchOneAndComplete2(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	ask := newTestAskOrder(200, 100)
	bid := newTestBidOrder(300, 100)

	acceptor.newOrderChan <- ask
	acceptor.newOrderChan <- bid

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

func Test_matchingAskOrder_MatchOneButIncomplete(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	ask := newTestAskOrder(199, 1000)

	acceptor.newOrderChan <- bid
	acceptor.newOrderChan <- ask

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

func Test_matchingBidOrder_MatchOneButIncomplete2(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	ask := newTestAskOrder(200, 100)
	bid := newTestBidOrder(300, 1000)

	acceptor.newOrderChan <- ask
	acceptor.newOrderChan <- bid

	i := 0
	for info := range acceptor.kernel.matchedInfoChan {
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

func Test_matchingAskOrder_MatchMultipleComplete(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid1 := newTestBidOrder(200, 100)
	bid2 := newTestBidOrder(200, 150)
	bid3 := newTestBidOrder(199, 110)
	ask := newTestAskOrder(198, 400)

	acceptor.newOrderChan <- bid1
	acceptor.newOrderChan <- bid2
	acceptor.newOrderChan <- bid3
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(198), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
	left := acceptor.kernel.ask.Front().value.(*priceBucket).Left
	assert.Equal(t, int64(-40), left)
}

func Test_matchingAskOrder_MatchMultipleComplete2(t *testing.T) {
	testSize := 200_000
	asks := generateTestOrders(testSize, 1000, 100, true)

	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

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

	var askSize int64 = 0
	for _, a := range asks {
		askSize += a.Left
	}

	bid := newTestBidOrder(500000, math.MaxInt64)
	acceptor.newOrderChan <- bid

	for acceptor.kernel.ask1Price != math.MaxInt64 {
		time.Sleep(time.Millisecond * 100)
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

func Test_matchingOrders_withRandomPriceAndSize(t *testing.T) {
	testSize := 12_500
	asks := generateTestOrders(testSize, 1000, 100, true)
	bids := generateTestOrders(testSize, 1000, 100, false)

	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	orderVolumeMap := make(map[uint64]int64)

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

	acceptor.startRedoKernel()

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

	for b := range done {
		if b == true {
			matching_time := (time.Now().UnixNano() - start) / (1000 * 1000)
			println(8*testSize, "orders matching finished in ", matching_time, " ms, ", (int64(8*testSize)/matching_time)*1000, " ops. per second")
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
	time.Sleep(3 * time.Second)
	assert.Equal(t, acceptor.kernel.ask1Price, acceptor.redoKernel.ask1Price)
	assert.Equal(t, acceptor.kernel.bid1Price, acceptor.redoKernel.bid1Price)
}

func Benchmark_insertPriceCheckedOrder(b *testing.B) {
	b.ReportAllocs()
	k := newKernel()
	testSize := 1
	if b.N > 1 {
		testSize = b.N / 2
	}
	asks := make([]*types.KernelOrder, 0, testSize)
	bids := make([]*types.KernelOrder, 0, testSize)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < testSize; i++ {
		price := r.Int63n(100) + 200
		amount := r.Int63n(100)
		asks = append(asks, newTestAskOrder(price, amount))
	}
	for i := 0; i < testSize; i++ {
		price := r.Int63n(100) + 99
		amount := r.Int63n(100)
		bids = append(bids, newTestBidOrder(price, amount))
	}
	b.ResetTimer()
	for i := range asks {
		k.insertUnmatchedOrder(asks[i])
		k.insertUnmatchedOrder(bids[i])
	}
}
