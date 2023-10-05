# Order Matching Engine
[![Go](https://github.com/Curton/GoMatchingKernel/actions/workflows/go.yml/badge.svg)](https://github.com/Curton/GoMatchingKernel/actions/workflows/go.yml)  

- A simple order matching engine written in Go.  
- A study project.  
- `kernel` acting the core of an Order Matching Engine.  
- `kernel` is designed to efficiently match buy and sell orders in a financial exchanges.  

## Features
* Order matching: The kernel uses skiplist for both buy and sell orders to efficiently match orders based on price and quantity.
* Multiple Order types: Supports different order types including Good-Till-Cancelled (GTC), Immediate-Or-Cancel (IOC), Fill-Or-Kill (FOK) and Post-Only-Order/Pending-Or-Cancelled (POC)
* Concurrency: Handle multiple orders simultaneously.
* Order cancellation: Orders can be cancelled as per the conditions defined.
* Snapshot feature: The engine can take a snapshot of the current state of the order book for recovery or analysis.
* Redo `kernel` for fast recovery or error correction. 
* Simple WAL is used for data integrity
* 
## Usage
This codebase is intended to be used as a library in a trading engine. 
To use it, you would need to import it into your project and create a new instance of the kernel. 
You can then use the methods provided by the kernel to interact with the order book.
```go
// insert orders
// create bid order
bidOrder := &types.KernelOrder{
    Amount:        100,
    Price:         200,
    Left:          100,
}
// create ask order, ask order represent in negative value
askOrder := &types.KernelOrder{
    Amount:        -100,
    Price:         201,
    Left:          -100,
}

// init a new order acceptor with serverId and acceptorDescription
acceptor := initAcceptor(1, "test")

// start order acceptor
go acceptor.orderAcceptor()

// ignore matching info nofitied by the kernel
acceptor.kernel.startDummyMatchedInfoChan()

// submit orders to acceptor
go func() {
    acceptor.newOrderChan <- bidOrder
    acceptor.newOrderChan <- askOrder
}()

// collect the details of received orders
ids := make([]*types.KernelOrder, 0, 2)
for i := 0; i < 2; i++ {
    v := <-acceptor.orderReceivedChan
    ids = append(ids, v)
}

// cancel order
// make sure not block the orderReceivedChan
go func() {
    for {
        // ignore order received info
        <-acceptor.orderReceivedChan
    }
}()

// set Amount to 0 to cancel order
for i := range ids {
    ids[i].Amount = 0
    acceptor.newOrderChan <- ids[i]
}

// all other use cases are in the kernel_test.go
```

## TDDO
- more test cases, many thanks to who can help us implement the TODO test cases
- code coverage
- monitoring
- kernel data visualization

## Core Code Structure
The main concepts and functionsare:

* kernel: The core struct representing the order matching engine.
* matchedInfo: Represents information about matched orders.
* priceBucket: Represents a price level in the order book.
* takeSnapshot(): Takes a snapshot of the current state of the order book.
* cancelOrder(): Cancels a given order.
* insertCheckedOrder(): Inserts an order into the order book after checking for validity.
* clearBucket(): Clears a price level in the order book.
* matchingOrder(): Matches incoming orders with the orders in the order book.
* restoreKernel(): Restores an instance of the order matching engine from a snapshot.

### Acceptor
Used to accept and schedule orders.
* Order Acceptance: The function accepts new orders, assigns an order ID, logs the order if enabled, and checks if the orders are limit or market orders.
* Order Processing: Processes the orders based on their type - limit or market, and routes them to the appropriate matching function in the kernel.
* Concurrent Processing: handle multiple orders simultaneously.
* Redo Log: The acceptor can also process a redo log to recover the state of the kernel in case of a engine failure.

The main types and functions are:

* scheduler: The main struct representing the acceptor. It contains channels for handling new orders, redo orders, received orders, and internal requests. It also contains an  instance of the kernel and the redo kernel.
* startRedoOrderAcceptor(): This function also runs in a goroutine and processes redo orders. It's almost identical to startOrderAcceptor() but operates on the redo kernel.
* startDummyOrderConfirmedChan(): This function starts a goroutine that drains the orderReceivedChan. This function is used for testing and should not be used in production.

## Implementation Details

### Kernel
* `ask` and `bid`: These are `*SkipList` types representing the sell (ask) and buy (bid) order books respectively.
* `ask1Price` and `bid1Price`: These `int64` types represent the current top (best) prices on the respective order books.
* `matchedInfoChan` and `errorInfoChan`: These channels are used for inter-goroutine communication, sending matched order information and potential errors respectively.
* `ask1PriceMux` and `bid1PriceMux`: These sync.Mutex types are used to provide safe concurrent access to `ask1Price` and `bid1Price` respectively.

### SkipList
* Two skip list is used to maintain the order book in a way that allows for efficient matching of orders.

### priceBucket
A priceBucket is represents a price level in the order book:
* `l`: A list of orders at this price level.
* `Left`: The total amount of the orders left at this price level.

### orderBookItem
An orderBookItem represents an item in the order book:
* `Price`: The price of the order.
* `Size`: The size (quantity) of the order.

### matchedInfo
The matchedInfo represents information about matched orders. 
* `makerOrders`: A slice of `KernelOrder` that were on the order book and have been matched with the taker order.
* `matchedSizeMap`: A map tracking the size that has been matched for each order.
* `takerOrder`: The `KernelOrder` that came to the order book and was matched with the maker orders.

### redoKernel
The `redoKernel` is an instance of the `kernel` type. It's used to process redo orders. 
Redo orders are a type of order that the engine has processed before but needs to process again, for fast recovery or error correction. 
Enabling the `redoKernel` will start the `RedoOrderAcceptor`, which reads and processes redo orders from the `redoOrderChan` channel.
The main difference between the `kernel` and `redoKernel` is that the `redoKernel` only processes redo orders while the `kernel` processes new orders.  

### Orders log
Orders log is a crucial part of any trading engine, serving as a crucial tool for audit, debugging and in some cases, recovery.  

Features:

- Order Logging: The engine logs all incoming orders with details including order ID, type, price, quantity, and more.
- Binary Encoding: The orders are encoded into binary format before being written to the log file to save space and improve performance.
- Log Reading: The logger can read orders from a log file and convert them back into orders.
- Redo Log Reading: The logger can read redo logs and send them to a redo kernel for processing.


## License
Released under the [MIT License](LICENSE).
