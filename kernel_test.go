/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package ker

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Curton/GoMatchingKernel/types"
)

func Test_kernel_cancelOrder_NotFound(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(200, 100)
	acceptor.newOrderChan <- ask
	time.Sleep(10 * time.Millisecond)

	orderNotInBook := &types.KernelOrder{
		KernelOrderID: 999999,
		CreateTime:    time.Now().UnixNano(),
		UpdateTime:    time.Now().UnixNano(),
		Amount:        100,
		Price:         300,
		Left:          100,
		Status:        types.OPEN,
		Type:          types.LIMIT,
	}
	acceptor.kernel.cancelOrder(orderNotInBook)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(200), acceptor.kernel.ask1Price)
}

func Test_kernel_cancelOrder_AskSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(200, 100)
	acceptor.newOrderChan <- ask

	receivedOrder := <-acceptor.orderReceivedChan
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(200), acceptor.kernel.ask1Price)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)

	cancelledOrder := &types.KernelOrder{
		KernelOrderID: receivedOrder.KernelOrderID,
		Price:         receivedOrder.Price,
		Amount:        receivedOrder.Amount,
		Left:          receivedOrder.Left,
	}
	acceptor.kernel.cancelOrder(cancelledOrder)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
}

func Test_kernel_cancelOrder_BidSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	acceptor.newOrderChan <- bid

	receivedOrder := <-acceptor.orderReceivedChan
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(200), acceptor.kernel.bid1Price)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	cancelledOrder := &types.KernelOrder{
		KernelOrderID: receivedOrder.KernelOrderID,
		Price:         receivedOrder.Price,
		Amount:        receivedOrder.Amount,
		Left:          receivedOrder.Left,
	}
	acceptor.kernel.cancelOrder(cancelledOrder)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

func Test_kernel_cancelOrder_PartialFill(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	acceptor.newOrderChan <- bid
	<-acceptor.orderReceivedChan
	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(199, 50)
	acceptor.newOrderChan <- ask
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	bidToCancel := acceptor.kernel.bid.Front().value.(*priceBucket).l.Front().Value.(*types.KernelOrder)
	acceptor.kernel.cancelOrder(bidToCancel)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
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
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	ask := newTestAskOrder(201, 100)

	go func() {
		acceptor.newOrderChan <- bid
		acceptor.newOrderChan <- ask
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

func Test_fullDepth(t *testing.T) {
	acceptor := newTestAcceptor()
	go acceptor.orderAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()
	acceptor.kernel.fullDepth()
}

func Test_orderAcceptor_ZeroAmount(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	acceptor.newOrderChan <- bid

	receivedBid := <-acceptor.orderReceivedChan
	time.Sleep(10 * time.Millisecond)

	zeroAmountOrder := &types.KernelOrder{
		KernelOrderID: receivedBid.KernelOrderID,
		CreateTime:    receivedBid.CreateTime,
		UpdateTime:    receivedBid.UpdateTime,
		Amount:        0,
		Price:         receivedBid.Price,
		Left:          0,
		Status:        types.OPEN,
		Type:          types.LIMIT,
	}
	acceptor.newOrderChan <- zeroAmountOrder
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

func Test_orderAcceptor_UnknownKernelFlag(t *testing.T) {
	assert.Panics(t, func() {
		acceptor := initAcceptor(1, "test")
		acceptor.orderAcceptor(999)
	})
}

func Test_insertUnmatchedOrder_AskNewPrice(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(200, 100)
	acceptor.kernel.insertUnmatchedOrder(ask)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(200), acceptor.kernel.ask1Price)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
}

func Test_insertUnmatchedOrder_BidNewPrice(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	acceptor.kernel.insertUnmatchedOrder(bid)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(200), acceptor.kernel.bid1Price)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)
}

func Test_insertUnmatchedOrder_AskExistingPrice(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(200, 100)
	acceptor.kernel.insertUnmatchedOrder(ask1)
	time.Sleep(10 * time.Millisecond)

	ask2 := newTestAskOrder(200, 50)
	acceptor.kernel.insertUnmatchedOrder(ask2)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(200), acceptor.kernel.ask1Price)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	bucket := acceptor.kernel.ask.Front().value.(*priceBucket)
	assert.Equal(t, int64(-150), bucket.Left)
}

func Test_insertUnmatchedOrder_BidExistingPrice(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid1 := newTestBidOrder(200, 100)
	acceptor.kernel.insertUnmatchedOrder(bid1)
	time.Sleep(10 * time.Millisecond)

	bid2 := newTestBidOrder(200, 50)
	acceptor.kernel.insertUnmatchedOrder(bid2)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(200), acceptor.kernel.bid1Price)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)
	bucket := acceptor.kernel.bid.Front().value.(*priceBucket)
	assert.Equal(t, int64(150), bucket.Left)
}
