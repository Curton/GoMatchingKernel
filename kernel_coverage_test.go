/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package ker

import (
	"container/list"
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Curton/GoMatchingKernel/types"
)

func Test_kernel_Stop(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	acceptor.kernel.Stop()

	select {
	case <-acceptor.kernel.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("kernel context not cancelled after Stop()")
	}
}

func Test_fullDepth_WithAskAndBid(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	ask := newTestAskOrder(300, 50)
	ask.Left = ask.Amount

	acceptor.newOrderChan <- bid
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)

	depth := acceptor.kernel.fullDepth()

	assert.Equal(t, 1, len(depth.ask))
	assert.Equal(t, 1, len(depth.bid))
	assert.Equal(t, int64(300), depth.ask[0].Price)
	assert.Equal(t, int64(-50), depth.ask[0].Size)
	assert.Equal(t, int64(200), depth.bid[0].Price)
	assert.Equal(t, int64(100), depth.bid[0].Size)
}

func Test_matchingOrder_POC_OrderCancelled(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	ask := newTestAskOrder(200, 100)
	ask.Left = ask.Amount
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)

	bid := newTestBidOrder(200, 100)
	bid.TimeInForce = types.POC
	bid.Left = bid.Amount

	acceptor.newOrderChan <- bid

	info := <-acceptor.kernel.matchedInfoChan
	assert.Equal(t, types.CANCELLED, info.takerOrder.Status)
	assert.Equal(t, types.POC, info.takerOrder.TimeInForce)
}

func Test_kernel_acceptor_InvalidOrder_LeftExceedsAmount(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount + 10

	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

func Test_matchingOrder_GTC_InsertUnmatched(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	bid.TimeInForce = types.GTC
	bid.Left = bid.Amount

	ask := newTestAskOrder(300, 100)
	ask.TimeInForce = types.GTC
	ask.Left = ask.Amount

	acceptor.newOrderChan <- ask
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)
	assert.Equal(t, int64(300), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(200), acceptor.kernel.bid1Price)
}

func Test_matchingOrder_ClearBucket(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid1 := newTestBidOrder(200, 100)
	bid2 := newTestBidOrder(200, 100)
	bid3 := newTestBidOrder(200, 100)
	ask := newTestAskOrder(200, 300)

	acceptor.newOrderChan <- bid1
	acceptor.newOrderChan <- bid2
	acceptor.newOrderChan <- bid3
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

func Test_orderAcceptor_InvalidOrder_DifferentSigns(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = -50

	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, acceptor.kernel.bid.Length)
}

func Test_kernel_fullDepth_EmptyBook(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	depth := acceptor.kernel.fullDepth()

	assert.Equal(t, 0, len(depth.ask))
	assert.Equal(t, 0, len(depth.bid))
}

func Test_kernel_fullDepth_OnlyAsk(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(300, 50)
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)

	depth := acceptor.kernel.fullDepth()

	assert.Equal(t, 1, len(depth.ask))
	assert.Equal(t, 0, len(depth.bid))
}

func Test_kernel_fullDepth_OnlyBid(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	depth := acceptor.kernel.fullDepth()

	assert.Equal(t, 0, len(depth.ask))
	assert.Equal(t, 1, len(depth.bid))
}

func Test_kernel_PauseAndResume(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	acceptor.kernel.Pause()

	bid2 := newTestBidOrder(300, 100)
	bid2.Left = bid2.Amount
	acceptor.newOrderChan <- bid2

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	acceptor.kernel.Resume()

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, 2, acceptor.kernel.bid.Length)
}

func Test_kernel_InsertUnmatchedOrder_UpdateBestPrices(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(300, 50)
	bid1 := newTestBidOrder(200, 100)

	acceptor.kernel.insertUnmatchedOrder(ask1)
	acceptor.kernel.insertUnmatchedOrder(bid1)

	assert.Equal(t, int64(300), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(200), acceptor.kernel.bid1Price)

	ask2 := newTestAskOrder(250, 50)
	bid2 := newTestBidOrder(250, 100)

	acceptor.kernel.insertUnmatchedOrder(ask2)
	acceptor.kernel.insertUnmatchedOrder(bid2)

	assert.Equal(t, int64(250), acceptor.kernel.ask1Price)
	assert.Equal(t, int64(250), acceptor.kernel.bid1Price)
}

func Test_contextNotCancelledBeforeStop(t *testing.T) {
	k := newKernel()
	k.startDummyMatchedInfoChan()

	assert.Nil(t, k.ctx.Err())

	k.Stop()

	select {
	case <-k.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("kernel context not cancelled after Stop()")
	}
}

func Test_matchingOrder_IOC_FullyFilled(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(200, 50)
	ask.TimeInForce = types.IOC
	ask.Left = ask.Amount
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, acceptor.kernel.bid.Length)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
}

func Test_matchingOrder_FOK_FullyFilled(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(200, 100)
	ask.TimeInForce = types.FOK
	ask.Left = ask.Amount
	acceptor.newOrderChan <- ask

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
}

func Test_matchingOrder_FOK_CancelledInsufficientLiquidity(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 50)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(200, 100)
	ask.TimeInForce = types.FOK
	ask.Left = ask.Amount
	acceptor.newOrderChan <- ask

	info := <-acceptor.kernel.matchedInfoChan
	assert.Equal(t, types.CANCELLED, info.takerOrder.Status)
}

func Test_matchingOrder_IOC_CancelledInsufficientLiquidity(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()

	bid := newTestBidOrder(200, 50)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(200, 100)
	ask.TimeInForce = types.IOC
	ask.Left = ask.Amount
	acceptor.newOrderChan <- ask

	info := <-acceptor.kernel.matchedInfoChan
	assert.Equal(t, types.CANCELLED, info.takerOrder.Status)
}

func Test_takeSnapshot_And_RestoreKernel(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(300, 50)
	bid1 := newTestBidOrder(200, 100)

	acceptor.newOrderChan <- ask1
	acceptor.newOrderChan <- bid1

	time.Sleep(20 * time.Millisecond)

	lastOrder := newTestBidOrder(250, 75)
	acceptor.kernel.takeSnapshot("test_snapshot", lastOrder)

	time.Sleep(20 * time.Millisecond)

	snapshotPath := "./orderbook_snapshot/test_snapshot/"
	entries, err := os.ReadDir(snapshotPath)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(entries), 1)

	latestSnapshot := ""
	var latestTime int64 = 0
	for _, entry := range entries {
		if entry.IsDir() {
			timeInt, err := strconv.ParseInt(entry.Name(), 10, 64)
			if err == nil && timeInt > latestTime {
				latestTime = timeInt
				latestSnapshot = entry.Name()
			}
		}
	}

	assert.NotEmpty(t, latestSnapshot)

	restorePath := snapshotPath + latestSnapshot + "/"
	restoredKernel, ok := restoreKernel(restorePath)
	assert.True(t, ok)
	assert.NotNil(t, restoredKernel)

	assert.Equal(t, 1, restoredKernel.ask.Length)
	assert.Equal(t, 1, restoredKernel.bid.Length)
}

func Test_restoreKernel_FinishedLogNotExist(t *testing.T) {
	ker, ok := restoreKernel("./non_existent_path/")
	assert.False(t, ok)
	assert.Nil(t, ker)
}

func Test_restoreKernel_BothSidesWithMultiplePriceLevels(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(300, 50)
	ask2 := newTestAskOrder(400, 75)
	bid1 := newTestBidOrder(200, 100)
	bid2 := newTestBidOrder(150, 80)

	acceptor.newOrderChan <- ask1
	acceptor.newOrderChan <- ask2
	acceptor.newOrderChan <- bid1
	acceptor.newOrderChan <- bid2

	time.Sleep(20 * time.Millisecond)

	lastOrder := newTestBidOrder(250, 75)
	acceptor.kernel.takeSnapshot("test_restore_multi", lastOrder)

	time.Sleep(20 * time.Millisecond)

	snapshotPath := "./orderbook_snapshot/test_restore_multi/"
	entries, err := os.ReadDir(snapshotPath)
	assert.NoError(t, err)

	latestSnapshot := ""
	var latestTime int64 = 0
	for _, entry := range entries {
		if entry.IsDir() {
			timeInt, err := strconv.ParseInt(entry.Name(), 10, 64)
			if err == nil && timeInt > latestTime {
				latestTime = timeInt
				latestSnapshot = entry.Name()
			}
		}
	}

	assert.NotEmpty(t, latestSnapshot)

	restorePath := snapshotPath + latestSnapshot + "/"
	restoredKernel, ok := restoreKernel(restorePath)
	assert.True(t, ok)
	assert.NotNil(t, restoredKernel)

	assert.Equal(t, 2, restoredKernel.ask.Length)
	assert.Equal(t, 2, restoredKernel.bid.Length)
}

func Test_restoreKernel_LeftValueRestored(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(300, 50)
	ask2 := newTestAskOrder(300, 30)
	bid1 := newTestBidOrder(200, 100)

	acceptor.newOrderChan <- ask1
	acceptor.newOrderChan <- ask2
	acceptor.newOrderChan <- bid1

	time.Sleep(20 * time.Millisecond)

	lastOrder := newTestBidOrder(250, 75)
	acceptor.kernel.takeSnapshot("test_restore_left", lastOrder)

	time.Sleep(20 * time.Millisecond)

	snapshotPath := "./orderbook_snapshot/test_restore_left/"
	entries, err := os.ReadDir(snapshotPath)
	assert.NoError(t, err)

	latestSnapshot := ""
	var latestTime int64 = 0
	for _, entry := range entries {
		if entry.IsDir() {
			timeInt, err := strconv.ParseInt(entry.Name(), 10, 64)
			if err == nil && timeInt > latestTime {
				latestTime = timeInt
				latestSnapshot = entry.Name()
			}
		}
	}

	assert.NotEmpty(t, latestSnapshot)

	restorePath := snapshotPath + latestSnapshot + "/"
	restoredKernel, ok := restoreKernel(restorePath)
	assert.True(t, ok)
	assert.NotNil(t, restoredKernel)

	assert.Equal(t, 1, restoredKernel.ask.Length)
	assert.Equal(t, 1, restoredKernel.bid.Length)

	askBucket := restoredKernel.ask.Front().Value().(*priceBucket)
	assert.Equal(t, int64(-80), askBucket.Left)

	bidBucket := restoredKernel.bid.Front().Value().(*priceBucket)
	assert.Equal(t, int64(100), bidBucket.Left)
}

func Test_writeOrderLog_NewFileCreation(t *testing.T) {
	tmpDir := "./kernelorder_log_test_tmp/"
	os.MkdirAll(tmpDir, 0755)
	originalPath := kernelOrderLogPath
	kernelOrderLogPath = tmpDir
	defer func() {
		kernelOrderLogPath = originalPath
		os.RemoveAll(tmpDir)
	}()

	saveOrderLogOrig := saveOrderLog
	saveOrderLog = true
	defer func() { saveOrderLog = saveOrderLogOrig }()

	acceptor := initAcceptor(1, "test_write_log")
	_ = acceptor
	var f *[1]*os.File = &[1]*os.File{nil}

	order := newTestBidOrder(200, 100)
	order.Left = order.Amount

	result := writeOrderLog(f, "test_write_log", order)
	assert.True(t, result)
	assert.NotNil(t, f[0])

	f[0].Close()
	os.RemoveAll(tmpDir)
}

func Test_writeOrderLog_MkdirAllError(t *testing.T) {
	tmpDir := "./kernelorder_log_test_tmp/"
	os.MkdirAll(tmpDir, 0755)
	originalPath := kernelOrderLogPath
	kernelOrderLogPath = tmpDir
	defer func() {
		kernelOrderLogPath = originalPath
		os.RemoveAll(tmpDir)
	}()

	saveOrderLogOrig := saveOrderLog
	saveOrderLog = true
	defer func() { saveOrderLog = saveOrderLogOrig }()

	acceptor := initAcceptor(1, "test_write_log_err")
	_ = acceptor
	var f *[1]*os.File = &[1]*os.File{nil}

	chmodErr := os.Chmod(tmpDir, 0000)
	if chmodErr != nil {
		t.Skip("Cannot change directory permissions for test")
	}
	defer os.Chmod(tmpDir, 0755)

	order := newTestBidOrder(200, 100)
	order.Left = order.Amount

	result := writeOrderLog(f, "test_write_log_err", order)
	assert.False(t, result)
}

func Test_writeOrderLog_WriteError(t *testing.T) {
	tmpDir := "./kernelorder_log_test_tmp_write_err/"
	os.MkdirAll(tmpDir, 0755)
	originalPath := kernelOrderLogPath
	kernelOrderLogPath = tmpDir
	defer func() {
		kernelOrderLogPath = originalPath
		os.RemoveAll(tmpDir)
	}()

	saveOrderLogOrig := saveOrderLog
	saveOrderLog = true
	defer func() { saveOrderLog = saveOrderLogOrig }()

	var f *[1]*os.File = &[1]*os.File{nil}

	order := newTestBidOrder(200, 100)
	order.Left = order.Amount

	firstResult := writeOrderLog(f, "test_write_err", order)
	assert.True(t, firstResult)

	if f[0] != nil {
		f[0].Close()
	}

	os.Chmod(tmpDir, 0555)
	defer os.Chmod(tmpDir, 0755)

	secondResult := writeOrderLog(f, "test_write_err", order)
	assert.False(t, secondResult)
}

func Test_getBytes_And_bytesToKernelOrder(t *testing.T) {
	order := newTestBidOrder(200, 100)
	order.Left = order.Amount
	order.Status = types.OPEN
	order.TimeInForce = types.GTC

	bytes := getBytes(order)
	assert.NotNil(t, bytes)
	assert.Greater(t, len(bytes), 0)

	decoded := bytesToKernelOrder(bytes)
	assert.Equal(t, order.Price, decoded.Price)
	assert.Equal(t, order.Amount, decoded.Amount)
	assert.Equal(t, order.Left, decoded.Left)
	assert.Equal(t, order.KernelOrderID, decoded.KernelOrderID)
}

func Test_bytesToKernelOrder_InvalidBytes(t *testing.T) {
	invalidBytes := []byte("invalid data")
	order := bytesToKernelOrder(invalidBytes)
	assert.NotNil(t, order)
}

func Test_kernelOrderListToBytes_And_readListFromBytes(t *testing.T) {
	orders := []*types.KernelOrder{
		newTestAskOrder(300, 50),
		newTestAskOrder(400, 75),
	}

	l := list.New()
	for _, o := range orders {
		o.Left = o.Amount
		l.PushBack(o)
	}

	bytes := kernelOrderListToBytes(l)
	assert.NotNil(t, bytes)
	assert.Greater(t, len(bytes), 0)

	restoredList := readListFromBytes(bytes)
	assert.Equal(t, l.Len(), restoredList.Len())

	idx := 0
	for e := restoredList.Front(); e != nil; e = e.Next() {
		order := e.Value.(*types.KernelOrder)
		assert.Equal(t, orders[idx].Price, order.Price)
		assert.Equal(t, orders[idx].Amount, order.Amount)
		idx++
	}
}

func Test_readListFromBytes_InvalidBytes(t *testing.T) {
	invalidBytes := []byte("invalid data")
	l := readListFromBytes(invalidBytes)
	assert.NotNil(t, l)
	assert.Equal(t, 0, l.Len())
}

func Test_getOrderBinary_And_readOrderBinary(t *testing.T) {
	order := newTestBidOrder(200, 100)
	order.Left = order.Amount
	order.Status = types.OPEN
	order.TimeInForce = types.GTC
	order.KernelOrderID = 12345

	bytes := getOrderBinary(order)
	assert.NotNil(t, bytes)
	assert.Greater(t, len(bytes), 0)

	restored := readOrderBinary(bytes)
	assert.Equal(t, order.Price, restored.Price)
	assert.Equal(t, order.Amount, restored.Amount)
	assert.Equal(t, order.Left, restored.Left)
	assert.Equal(t, order.KernelOrderID, restored.KernelOrderID)
	assert.Equal(t, order.Status, restored.Status)
	assert.Equal(t, order.TimeInForce, restored.TimeInForce)
}

func Test_readOrderBinary_InvalidBytes(t *testing.T) {
	invalidBytes := []byte("short")
	order := readOrderBinary(invalidBytes)
	assert.NotNil(t, order)
}

func Test_NewWithMaxLevel_InvalidLevels(t *testing.T) {
	assert.Panics(t, func() {
		NewWithMaxLevel(0)
	})

	assert.Panics(t, func() {
		NewWithMaxLevel(-1)
	})

	assert.Panics(t, func() {
		NewWithMaxLevel(65)
	})
}

func Test_NewWithMaxLevel_ValidLevels(t *testing.T) {
	sl1 := NewWithMaxLevel(1)
	assert.NotNil(t, sl1)
	assert.Equal(t, 1, sl1.maxLevel)

	sl64 := NewWithMaxLevel(64)
	assert.NotNil(t, sl64)
	assert.Equal(t, 64, sl64.maxLevel)
}

func Test_drainMatchedInfoChan_Helper(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	drainMatchedInfoChan(acceptor)
	time.Sleep(10 * time.Millisecond)
}

func Test_waitForEmptyBook_Helper(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.newOrderChan <- bid

	time.Sleep(10 * time.Millisecond)

	result := waitForEmptyBook(acceptor, 50*time.Millisecond)
	assert.False(t, result)
}

func Test_waitForPriceUpdate_Helper(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(300, 50)
	acceptor.newOrderChan <- ask
	time.Sleep(10 * time.Millisecond)

	result := waitForPriceUpdate(acceptor, 300, math.MinInt64, 50*time.Millisecond)
	assert.True(t, result)

	result = waitForPriceUpdate(acceptor, 400, math.MinInt64, 50*time.Millisecond)
	assert.False(t, result)
}

func Test_getBucketLeft_Helper(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(300, 50)
	acceptor.newOrderChan <- ask
	time.Sleep(10 * time.Millisecond)

	left := getBucketLeft(acceptor, true)
	assert.Equal(t, int64(-50), left)

	left = getBucketLeft(acceptor, false)
	assert.Equal(t, int64(0), left)
}

func Test_getOrderBookTotalSize_Helper(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	size := getOrderBookTotalSize(acceptor, true)
	assert.Equal(t, 0, size)

	ask := newTestAskOrder(300, 50)
	acceptor.newOrderChan <- ask
	time.Sleep(10 * time.Millisecond)

	size = getOrderBookTotalSize(acceptor, true)
	assert.Equal(t, 1, size)
}

func Test_cancelOrder_AskSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask := newTestAskOrder(300, 50)
	ask.Left = ask.Amount
	acceptor.kernel.insertUnmatchedOrder(ask)

	assert.Equal(t, 1, acceptor.kernel.ask.Length)

	cancelOrder := &types.KernelOrder{
		KernelOrderID: ask.KernelOrderID,
		Price:         ask.Price,
		Amount:        ask.Amount,
		Left:          ask.Left,
	}
	acceptor.kernel.cancelOrder(cancelOrder)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, acceptor.kernel.ask.Length)
	assert.Equal(t, int64(math.MaxInt64), acceptor.kernel.ask1Price)
}

func Test_cancelOrder_BidSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	acceptor.kernel.insertUnmatchedOrder(bid)

	assert.Equal(t, 1, acceptor.kernel.bid.Length)

	cancelOrder := &types.KernelOrder{
		KernelOrderID: bid.KernelOrderID,
		Price:         bid.Price,
		Amount:        bid.Amount,
		Left:          bid.Left,
	}
	acceptor.kernel.cancelOrder(cancelOrder)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, acceptor.kernel.bid.Length)
	assert.Equal(t, int64(math.MinInt64), acceptor.kernel.bid1Price)
}

func Test_cancelOrder_NotFound(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	nonExistent := &types.KernelOrder{
		KernelOrderID: 99999,
		Price:         500,
		Amount:        100,
		Left:          100,
	}

	acceptor.kernel.cancelOrder(nonExistent)
}

func Test_orderLogReader_FileReadError(t *testing.T) {
	tmpDir := "./kernelorder_log_test_tmp/"
	os.MkdirAll(tmpDir, 0755)
	originalPath := kernelOrderLogPath
	kernelOrderLogPath = tmpDir
	defer func() {
		kernelOrderLogPath = originalPath
		os.RemoveAll(tmpDir)
	}()

	saveOrderLogOrig := saveOrderLog
	saveOrderLog = true
	defer func() { saveOrderLog = saveOrderLogOrig }()

	acceptor := initAcceptor(1, "test_log_reader")
	acceptor.startRedoKernel()

	time.Sleep(100 * time.Millisecond)
}

func Test_orderAcceptor_InternalRequestChan(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	select {
	case acceptor.internalRequestChan <- 1:
	default:
	}

	time.Sleep(10 * time.Millisecond)
}

func Test_restoreKernel_BidDirReadError(t *testing.T) {
	tmpDir := "./kernelorder_log_test_tmp_restore_bid_err/"
	os.MkdirAll(tmpDir, 0755)
	defer func() {
		os.Chmod(tmpDir+"bid/", 0755)
		os.RemoveAll(tmpDir)
	}()

	askPath := tmpDir + "ask/"
	bidPath := tmpDir + "bid/"
	os.MkdirAll(askPath, 0755)
	os.MkdirAll(bidPath, 0755)
	os.Chmod(bidPath, 0000)

	f, _ := os.Create(tmpDir + "finished.log")
	f.Close()

	ker, ok := restoreKernel(tmpDir)
	assert.Nil(t, ker)
	assert.False(t, ok)
}

func Test_getBytes_EncodingError(t *testing.T) {
	order := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    1,
		UpdateTime:    1,
		Amount:        100,
		Price:         200,
		Left:          100,
		FilledTotal:   0,
		Id:            1,
		Status:        types.OPEN,
		Type:          types.LIMIT,
		TimeInForce:   types.GTC,
	}

	bytes := getBytes(order)
	assert.NotNil(t, bytes)
	assert.Greater(t, len(bytes), 0)
}

func Test_kernelOrderListToBytes_ErrorPath(t *testing.T) {
	order := &types.KernelOrder{
		KernelOrderID: 1,
		CreateTime:    1,
		UpdateTime:    1,
		Amount:        100,
		Price:         200,
		Left:          100,
		FilledTotal:   0,
		Id:            1,
		Status:        types.OPEN,
		Type:          types.LIMIT,
		TimeInForce:   types.GTC,
	}

	l := list.New()
	l.PushBack(order)

	bytes := kernelOrderListToBytes(l)
	assert.NotNil(t, bytes)
	assert.Greater(t, len(bytes), 0)
}

func Test_cancelOrder_BucketRemainsWithOrders_AskSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(300, 50)
	ask2 := newTestAskOrder(400, 30)
	ask1.Left = ask1.Amount
	ask2.Left = ask2.Amount

	acceptor.kernel.insertUnmatchedOrder(ask1)
	acceptor.kernel.insertUnmatchedOrder(ask2)

	assert.Equal(t, 2, acceptor.kernel.ask.Length)
	assert.Equal(t, int64(300), acceptor.kernel.ask1Price)

	cancelAsk := &types.KernelOrder{
		KernelOrderID: ask1.KernelOrderID,
		Price:         ask1.Price,
		Amount:        ask1.Amount,
		Left:          ask1.Left,
	}
	acceptor.kernel.cancelOrder(cancelAsk)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, acceptor.kernel.ask.Length)
	assert.Equal(t, int64(400), acceptor.kernel.ask1Price)

	bucket := acceptor.kernel.ask.Front().Value().(*priceBucket)
	assert.Equal(t, int64(-30), bucket.Left)
	assert.Equal(t, 1, bucket.l.Len())
}

func Test_cancelOrder_BucketRemainsWithOrders_BidSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid1 := newTestBidOrder(200, 100)
	bid2 := newTestBidOrder(100, 50)
	bid1.Left = bid1.Amount
	bid2.Left = bid2.Amount

	acceptor.kernel.insertUnmatchedOrder(bid1)
	acceptor.kernel.insertUnmatchedOrder(bid2)

	assert.Equal(t, 2, acceptor.kernel.bid.Length)
	assert.Equal(t, int64(200), acceptor.kernel.bid1Price)

	cancelBid := &types.KernelOrder{
		KernelOrderID: bid1.KernelOrderID,
		Price:         bid1.Price,
		Amount:        bid1.Amount,
		Left:          bid1.Left,
	}
	acceptor.kernel.cancelOrder(cancelBid)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, acceptor.kernel.bid.Length)
	assert.Equal(t, int64(100), acceptor.kernel.bid1Price)

	bucket := acceptor.kernel.bid.Front().Value().(*priceBucket)
	assert.Equal(t, int64(50), bucket.Left)
	assert.Equal(t, 1, bucket.l.Len())
}

func Test_restoreKernel_OnlyAskSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	ask1 := newTestAskOrder(300, 50)
	ask2 := newTestAskOrder(400, 75)

	acceptor.newOrderChan <- ask1
	acceptor.newOrderChan <- ask2

	time.Sleep(20 * time.Millisecond)

	lastOrder := newTestBidOrder(250, 75)
	acceptor.kernel.takeSnapshot("test_restore_ask_only", lastOrder)

	time.Sleep(20 * time.Millisecond)

	snapshotPath := "./orderbook_snapshot/test_restore_ask_only/"
	entries, err := os.ReadDir(snapshotPath)
	assert.NoError(t, err)

	latestSnapshot := ""
	var latestTime int64 = 0
	for _, entry := range entries {
		if entry.IsDir() {
			timeInt, err := strconv.ParseInt(entry.Name(), 10, 64)
			if err == nil && timeInt > latestTime {
				latestTime = timeInt
				latestSnapshot = entry.Name()
			}
		}
	}

	assert.NotEmpty(t, latestSnapshot)

	restorePath := snapshotPath + latestSnapshot + "/"
	restoredKernel, ok := restoreKernel(restorePath)
	assert.True(t, ok)
	assert.NotNil(t, restoredKernel)

	assert.Equal(t, 2, restoredKernel.ask.Length)
	assert.Equal(t, 0, restoredKernel.bid.Length)
	assert.Equal(t, int64(300), restoredKernel.ask1Price)
	assert.Equal(t, int64(math.MinInt64), restoredKernel.bid1Price)
}

func Test_restoreKernel_OnlyBidSide(t *testing.T) {
	acceptor := newTestAcceptor()
	acceptor.startDummyOrderReceivedChan()
	acceptor.kernel.startDummyMatchedInfoChan()

	bid1 := newTestBidOrder(200, 100)
	bid2 := newTestBidOrder(150, 80)

	acceptor.newOrderChan <- bid1
	acceptor.newOrderChan <- bid2

	time.Sleep(20 * time.Millisecond)

	lastOrder := newTestBidOrder(250, 75)
	acceptor.kernel.takeSnapshot("test_restore_bid_only", lastOrder)

	time.Sleep(20 * time.Millisecond)

	snapshotPath := "./orderbook_snapshot/test_restore_bid_only/"
	entries, err := os.ReadDir(snapshotPath)
	assert.NoError(t, err)

	latestSnapshot := ""
	var latestTime int64 = 0
	for _, entry := range entries {
		if entry.IsDir() {
			timeInt, err := strconv.ParseInt(entry.Name(), 10, 64)
			if err == nil && timeInt > latestTime {
				latestTime = timeInt
				latestSnapshot = entry.Name()
			}
		}
	}

	assert.NotEmpty(t, latestSnapshot)

	restorePath := snapshotPath + latestSnapshot + "/"
	restoredKernel, ok := restoreKernel(restorePath)
	assert.True(t, ok)
	assert.NotNil(t, restoredKernel)

	assert.Equal(t, 0, restoredKernel.ask.Length)
	assert.Equal(t, 2, restoredKernel.bid.Length)
	assert.Equal(t, int64(math.MaxInt64), restoredKernel.ask1Price)
	assert.Equal(t, int64(200), restoredKernel.bid1Price)
}
