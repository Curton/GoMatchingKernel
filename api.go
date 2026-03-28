package ker

import (
	"github.com/Curton/GoMatchingKernel/types"
)

// SetSaveOrderLog controls whether orders are written to the WAL log file.
func SetSaveOrderLog(v bool) {
	saveOrderLog = v
}

// PriceLevel represents a single price level in the order book.
type PriceLevel struct {
	Price int64 `json:"price"`
	Size  int64 `json:"size"`
}

// OrderBookSnapshot is a snapshot of the current order book.
type OrderBookSnapshot struct {
	Asks []PriceLevel `json:"asks"`
	Bids []PriceLevel `json:"bids"`
}

// MatchResult holds the result of a matching event.
type MatchResult struct {
	TakerOrder    types.KernelOrder
	MakerOrders   []types.KernelOrder
	MatchedSizeMap map[uint64]int64
}

// MatchingEngine is the exported wrapper around the internal matching kernel.
type MatchingEngine struct {
	s             *scheduler
	matchResultCh chan MatchResult
}

// NewMatchingEngine creates a new matching engine instance.
func NewMatchingEngine(serverID uint64, desc string) *MatchingEngine {
	return &MatchingEngine{
		s:             initAcceptor(serverID, desc),
		matchResultCh: make(chan MatchResult, 256),
	}
}

// Start begins order processing. Must be called before submitting orders.
func (e *MatchingEngine) Start() {
	go e.s.orderAcceptor()
	e.s.startDummyOrderReceivedChan()

	// Bridge internal matchedInfoChan to exported matchResultCh
	go func() {
		for {
			mi, ok := <-e.s.kernel.matchedInfoChan
			if !ok {
				return
			}
			makerOrders := make([]types.KernelOrder, len(mi.makerOrders))
			copy(makerOrders, mi.makerOrders)
			sizeMap := make(map[uint64]int64, len(mi.matchedSizeMap))
			for k, v := range mi.matchedSizeMap {
				sizeMap[k] = v
			}
			e.matchResultCh <- MatchResult{
				TakerOrder:     mi.takerOrder,
				MakerOrders:    makerOrders,
				MatchedSizeMap: sizeMap,
			}
		}
	}()
}

// SubmitOrder sends an order into the matching engine.
func (e *MatchingEngine) SubmitOrder(order *types.KernelOrder) {
	e.s.newOrderChan <- order
}

// MatchedInfoChan returns a read-only channel of match results.
func (e *MatchingEngine) MatchedInfoChan() <-chan MatchResult {
	return e.matchResultCh
}

// OrderBook returns a snapshot of the current order book.
func (e *MatchingEngine) OrderBook() *OrderBookSnapshot {
	ob := e.s.kernel.fullDepth()
	asks := make([]PriceLevel, len(ob.ask))
	for i, item := range ob.ask {
		asks[i] = PriceLevel{Price: item.Price, Size: item.Size}
	}
	bids := make([]PriceLevel, len(ob.bid))
	for i, item := range ob.bid {
		bids[i] = PriceLevel{Price: item.Price, Size: item.Size}
	}
	return &OrderBookSnapshot{Asks: asks, Bids: bids}
}

// BestAsk returns the current best (lowest) ask price.
func (e *MatchingEngine) BestAsk() int64 {
	return e.s.kernel.ask1Price
}

// BestBid returns the current best (highest) bid price.
func (e *MatchingEngine) BestBid() int64 {
	return e.s.kernel.bid1Price
}

// AskLength returns the number of ask price levels.
func (e *MatchingEngine) AskLength() int {
	return e.s.kernel.ask.Length
}

// BidLength returns the number of bid price levels.
func (e *MatchingEngine) BidLength() int {
	return e.s.kernel.bid.Length
}

// Stop gracefully shuts down the matching engine.
func (e *MatchingEngine) Stop() {
	e.s.kernel.Stop()
	close(e.matchResultCh)
}
