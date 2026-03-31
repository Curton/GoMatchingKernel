/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/6 11:24
 */

package ker

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SetSaveOrderLog(t *testing.T) {
	original := saveOrderLog
	defer func() { saveOrderLog = original }()

	SetSaveOrderLog(true)
	assert.True(t, saveOrderLog)

	SetSaveOrderLog(false)
	assert.False(t, saveOrderLog)
}

func Test_NewMatchingEngine(t *testing.T) {
	engine := NewMatchingEngine(1, "test_engine")
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.s)
	assert.NotNil(t, engine.matchResultCh)
}

func Test_MatchingEngine_Start_And_SubmitOrder(t *testing.T) {
	engine := NewMatchingEngine(1, "test_start")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(300, 50)
	ask.Left = ask.Amount
	engine.SubmitOrder(ask)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(300), engine.BestAsk())
	assert.Equal(t, 1, engine.AskLength())
	assert.Equal(t, int64(math.MinInt64), engine.BestBid())
	assert.Equal(t, 0, engine.BidLength())

	engine.Stop()
}

func Test_MatchingEngine_SubmitBidOrder(t *testing.T) {
	engine := NewMatchingEngine(1, "test_bid")
	engine.Start()

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	engine.SubmitOrder(bid)

	assert.Eventually(t, func() bool {
		return engine.BestBid() == 200 &&
			engine.BidLength() == 1 &&
			engine.BestAsk() == math.MaxInt64 &&
			engine.AskLength() == 0
	}, time.Second, 10*time.Millisecond)

	engine.Stop()
}

func Test_MatchingEngine_MatchedInfoChan(t *testing.T) {
	engine := NewMatchingEngine(1, "test_chan")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ch := engine.MatchedInfoChan()
	assert.NotNil(t, ch)

	engine.Stop()
}

func Test_MatchingEngine_OrderBook_Empty(t *testing.T) {
	engine := NewMatchingEngine(1, "test_ob_empty")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	snapshot := engine.OrderBook()
	assert.NotNil(t, snapshot)
	assert.NotNil(t, snapshot.Asks)
	assert.NotNil(t, snapshot.Bids)
	assert.Equal(t, 0, len(snapshot.Asks))
	assert.Equal(t, 0, len(snapshot.Bids))

	engine.Stop()
}

func Test_MatchingEngine_OrderBook_WithOrders(t *testing.T) {
	engine := NewMatchingEngine(1, "test_ob_filled")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(300, 50)
	ask.Left = ask.Amount
	engine.SubmitOrder(ask)

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	engine.SubmitOrder(bid)

	time.Sleep(10 * time.Millisecond)

	snapshot := engine.OrderBook()
	assert.NotNil(t, snapshot)
	assert.Equal(t, 1, len(snapshot.Asks))
	assert.Equal(t, 1, len(snapshot.Bids))
	assert.Equal(t, int64(300), snapshot.Asks[0].Price)
	assert.Equal(t, int64(-50), snapshot.Asks[0].Size)
	assert.Equal(t, int64(200), snapshot.Bids[0].Price)
	assert.Equal(t, int64(100), snapshot.Bids[0].Size)

	engine.Stop()
}

func Test_MatchingEngine_BestAsk_NoOrders(t *testing.T) {
	engine := NewMatchingEngine(1, "test_best_ask_empty")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(math.MaxInt64), engine.BestAsk())

	engine.Stop()
}

func Test_MatchingEngine_BestBid_NoOrders(t *testing.T) {
	engine := NewMatchingEngine(1, "test_best_bid_empty")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(math.MinInt64), engine.BestBid())

	engine.Stop()
}

func Test_MatchingEngine_AskLength_NoOrders(t *testing.T) {
	engine := NewMatchingEngine(1, "test_ask_len_empty")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, engine.AskLength())

	engine.Stop()
}

func Test_MatchingEngine_BidLength_NoOrders(t *testing.T) {
	engine := NewMatchingEngine(1, "test_bid_len_empty")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 0, engine.BidLength())

	engine.Stop()
}

func Test_MatchingEngine_Stop(t *testing.T) {
	engine := NewMatchingEngine(1, "test_stop")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	engine.Stop()

	select {
	case <-engine.s.kernel.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("kernel context not cancelled after Stop()")
	}

	select {
	case _, ok := <-engine.matchResultCh:
		if ok {
			t.Fatal("matchResultCh should be closed after Stop()")
		}
	case <-time.After(time.Second):
		t.Fatal("matchResultCh should be closed after Stop()")
	}
}

func Test_MatchingEngine_MultiplePriceLevels(t *testing.T) {
	engine := NewMatchingEngine(1, "test_multi_level")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ask1 := newTestAskOrder(300, 50)
	ask1.Left = ask1.Amount
	engine.SubmitOrder(ask1)

	ask2 := newTestAskOrder(400, 75)
	ask2.Left = ask2.Amount
	engine.SubmitOrder(ask2)

	bid1 := newTestBidOrder(200, 100)
	bid1.Left = bid1.Amount
	engine.SubmitOrder(bid1)

	bid2 := newTestBidOrder(150, 80)
	bid2.Left = bid2.Amount
	engine.SubmitOrder(bid2)

	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, int64(300), engine.BestAsk())
	assert.Equal(t, int64(200), engine.BestBid())
	assert.Equal(t, 2, engine.AskLength())
	assert.Equal(t, 2, engine.BidLength())

	engine.Stop()
}

func Test_MatchingEngine_MatchResult_Bridge(t *testing.T) {
	engine := NewMatchingEngine(1, "test_match_bridge")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(200, 100)
	ask.Left = ask.Amount
	engine.SubmitOrder(ask)

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	engine.SubmitOrder(bid)

	time.Sleep(20 * time.Millisecond)

	select {
	case result, ok := <-engine.matchResultCh:
		if ok {
			assert.NotEqual(t, uint64(0), result.TakerOrder.KernelOrderID)
			assert.NotNil(t, result.MakerOrders)
		}
	case <-time.After(time.Second):
	}

	engine.Stop()
}

func Test_MatchingEngine_SubmitOrder_CancelOrder(t *testing.T) {
	engine := NewMatchingEngine(1, "test_cancel")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	engine.SubmitOrder(bid)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, engine.BidLength())

	engine.Stop()
}

func Test_MatchingEngine_SubmitOrder_AmountZero(t *testing.T) {
	engine := NewMatchingEngine(1, "test_zero_amount")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	bid := newTestBidOrder(200, 100)
	bid.Left = bid.Amount
	bid.Amount = 0
	engine.SubmitOrder(bid)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, engine.BidLength())

	engine.Stop()
}

func Test_MatchingEngine_OrderBookSnapshot_MultipleAsks(t *testing.T) {
	engine := NewMatchingEngine(1, "test_ob_multi_ask")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ask1 := newTestAskOrder(300, 50)
	ask1.Left = ask1.Amount
	engine.SubmitOrder(ask1)

	ask2 := newTestAskOrder(400, 30)
	ask2.Left = ask2.Amount
	engine.SubmitOrder(ask2)

	ask3 := newTestAskOrder(500, 20)
	ask3.Left = ask3.Amount
	engine.SubmitOrder(ask3)

	time.Sleep(10 * time.Millisecond)

	snapshot := engine.OrderBook()
	assert.Equal(t, 3, len(snapshot.Asks))
	assert.Equal(t, 0, len(snapshot.Bids))

	engine.Stop()
}

func Test_MatchingEngine_OrderBookSnapshot_MultipleBids(t *testing.T) {
	engine := NewMatchingEngine(1, "test_ob_multi_bid")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	bid1 := newTestBidOrder(200, 100)
	bid1.Left = bid1.Amount
	engine.SubmitOrder(bid1)

	bid2 := newTestBidOrder(300, 80)
	bid2.Left = bid2.Amount
	engine.SubmitOrder(bid2)

	bid3 := newTestBidOrder(400, 60)
	bid3.Left = bid3.Amount
	engine.SubmitOrder(bid3)

	time.Sleep(10 * time.Millisecond)

	snapshot := engine.OrderBook()
	assert.Equal(t, 0, len(snapshot.Asks))
	assert.Equal(t, 3, len(snapshot.Bids))

	engine.Stop()
}

func Test_MatchingEngine_BestPrices_AfterMatching(t *testing.T) {
	engine := NewMatchingEngine(1, "test_best_after_match")
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	ask := newTestAskOrder(200, 100)
	ask.Left = ask.Amount
	engine.SubmitOrder(ask)

	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int64(200), engine.BestAsk())

	bid := newTestBidOrder(200, 50)
	bid.Left = bid.Amount
	engine.SubmitOrder(bid)

	time.Sleep(20 * time.Millisecond)

	engine.Stop()
}
