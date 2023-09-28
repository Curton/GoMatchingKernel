# Exchange Matching Kernel

A simple exchange matching kernel written in Go.
A study project.
The kernel contains the core code for an Order Matching System.
The kernel is designed to efficiently match buy and sell orders in a financial exchange.

## Features
* Order matching: The system uses a skip list data structure for both buy and sell orders to efficiently match orders based on price and quantity.
* Order types: Supports different order types including Good-Till-Cancelled (GTC), Immediate-Or-Cancel (IOC), and Fill-Or-Kill (FOK).
* Concurrency: Makes use of Go's built-in concurrency features to handle multiple orders simultaneously.
* Order cancellation: Orders can be cancelled as per the conditions defined.
* Snapshot feature: The system can take a snapshot of the current state of the order book for recovery or analysis.
* Simple WAL is used for data integrity.

## Core Code Structure
The main types and functionsare:

* kernel: The core struct representing the order matching system.
* matchedInfo: Struct representing information about matched orders.
* priceBucket: Represents a price level in the order book.
* orderBookItem: Represents an item in the order book.
* fullDepth(): Returns a snapshot of the entire order book.
* takeSnapshot(): Takes a snapshot of the current state of the order book.
* cancelOrder(): Cancels a given order.
* insertCheckedOrder(): Inserts an order into the order book after checking for validity.
* clearBucket(): Clears a price level in the order book.
* matchingOrder(): Matches incoming orders with the orders in the order book.
* newKernel(): Creates a new instance of the order matching system.
* restoreKernel(): Restores an instance of the order matching system from a snapshot.

## Usage
This codebase is intended to be used as a library in a trading system. 
To use it, you would need to import it into your project and create a new instance of the kernel. 
You can then use the methods provided by the kernel to interact with the order book.
```go
kernel := newKernel()
// insert orders, cancel orders, etc.
// all other use cases are in the kernel_test.go
```
### Acceptor
Used to accept and schedule orders.
* Order Acceptance: The system accepts new orders, assigns an order ID, logs the order if enabled, and checks if the orders are limit or market orders.
* Order Processing: Processes the orders based on their type - limit or market, and routes them to the appropriate matching function in the kernel.
* Concurrent Processing: handle multiple orders simultaneously.
* Redo Log: The acceptor can also process a redo log to recover the state of the kernel in case of a system failure.

The main types and functions are:

* scheduler: The main struct representing the acceptor. It contains channels for handling new orders, redo orders, received orders, and internal requests. It also contains an  instance of the kernel and the redo kernel.
* startRedoOrderAcceptor(): This function also runs in a goroutine and processes redo orders. It's almost identical to startOrderAcceptor() but operates on the redo kernel.
* startDummyOrderConfirmedChan(): This function starts a goroutine that drains the orderReceivedChan. This function is used for testing and should not be used in production.
## Data structure
### Kernel
* ask and bid: These are *SkipList types representing the sell (ask) and buy (bid) order books respectively.
* ask1Price and bid1Price: These int64 types represent the current top (best) prices on the respective order books.
* matchedInfoChan and errorInfoChan: These channels are used for inter-goroutine communication, sending matched order information and potential errors respectively.
* ask1PriceMux and bid1PriceMux: These sync.Mutex types are used to provide safe concurrent access to ask1Price and bid1Price respectively.

### SkipList
* The skip list is used here to maintain the order book in a way that allows for efficient matching of orders.

### priceBucket
A priceBucket is represents a price level in the order book:
* l: A list of orders at this price level.
* Left: The total amount of the commodity left at this price level.

### orderBookItem
An orderBookItem represents an item in the order book:
* Price: The price of the order.
* Size: The size (quantity) of the order.

### matchedInfo
The matchedInfo represents information about matched orders. 
* makerOrders: A slice of KernelOrder that were on the order book and have been matched with the taker order.
* matchedSizeMap: A map tracking the size that has been matched for each order.
* takerOrder: The KernelOrder that came to the order book and was matched with the maker orders.
