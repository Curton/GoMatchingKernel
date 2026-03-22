# Order Matching Engine

[![Go](https://github.com/Curton/GoMatchingKernel/actions/workflows/go.yml/badge.svg)](https://github.com/Curton/GoMatchingKernel/actions/workflows/go.yml)

A high-performance order matching engine for financial exchanges, built with Go 1.21 as a study project.

## Features

- **Order Matching**: SkipList-based matching for efficient price-level ordering
- **Order Types**: GTC, IOC, FOK, and POC (Post-Only/Pending-Or-Cancelled)
- **Concurrency**: Simultaneous order processing with goroutines and channels
- **Order Cancellation**: Full support for order revocation
- **Snapshots**: Order book state capture for recovery and analysis
- **WAL**: Write-Ahead Logging for data integrity and fast recovery
- **Redo Processing**: Error correction through redo log replay

## Quick Start

```go
// Create orders (negative Amount represents ask orders)
bidOrder := &types.KernelOrder{
    Amount: 100,
    Price:  200,
    Left:   100,
}
askOrder := &types.KernelOrder{
    Amount: -100,
    Price:  201,
    Left:   -100,
}

// Initialize acceptor
acceptor := initAcceptor(1, "test")
go acceptor.orderAcceptor()
acceptor.kernel.startDummyMatchedInfoChan()

// Submit orders
go func() {
    acceptor.newOrderChan <- bidOrder
    acceptor.newOrderChan <- askOrder
}()
```

See `kernel_test.go` for complete usage examples.

## Architecture

### Core Components

| Component | Description |
|-----------|-------------|
| `kernel` | Core matching engine with ask/bid SkipLists |
| `priceBucket` | Price level container using `container/list` |
| `matchedInfo` | Match result with maker orders and taker order |
| `scheduler` | Order acceptor and scheduler |

### Data Structures

**SkipList**: Maintains order books sorted by price for efficient matching

**priceBucket**: 
- `l`: List of orders at this price level
- `Left`: Total remaining amount

**KernelOrder**: 72-byte order struct with fixed-size fields for binary serialization

### Channels

- `matchedInfoChan`: Matched order notifications
- `errorInfoChan`: Error reporting
- `newOrderChan`: Incoming orders
- `orderReceivedChan`: Order receipt confirmations

## Testing

```bash
go test -v ./...
GOMAXPROCS=1 go test -bench=. -run=none -benchtime=1s -benchmem
./cover.sh  # HTML coverage report
```

## TODO

- Monitoring integration
- Kernel data visualization

## License

MIT License