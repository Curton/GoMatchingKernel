package types

type OrderStatus uint8

type OrderType uint8

type TimeInForce uint8

const (
	/* 
	A limit order is an order to buy or sell a security at a specific price or better.
	A buy limit order can only be executed at the limit price or lower, and a sell limit order can only be executed at the limit price or higher. 
	This gives the trader control over the price at which the trade is executed, but it does not guarantee that the order will be filled.
	*/
	LIMIT OrderType = iota 
	/* 
	A market order is an order to buy or sell a security at the best available price in the current market. 
	It is widely used because it guarantees that the order will be executed, but it does not guarantee the execution price. 
	A market order generally will execute at or near the current bid (for a sell order) or ask (for a buy order) price. 
	However, it is possible for lack of depth in the order book, causing market orders to execute at a price that is significantly different from the current bid or ask price.
	*/
	MARKET
)

const (
	OPEN OrderStatus = iota
	CLOSED
	CANCELLED
)

const (
	GTC TimeInForce = iota  /* "Good Till Cancelled", the order remains active until the user cancels it */
	FOK             		/* "Fill or Kill", an order to buy or sell that must be executed immediately in its entirety; otherwise, the entire order will be cancelled (not partially filled) */
	POC						/* "Post-Only-Order","Pending Or Cancelled", an order that only executes if it will add liquidity to the market. */
	IOC 					/* "Immediate or Cancel", this order fills as much as it can immediately, and cancels any remaining part of the order */
)

// '1' is represent as '1,000,000,000' in Price
const (
	ONE int64 = 1_000_000_000
)

// 72 bytes
type KernelOrder struct {
	// Exchange Kernel KernelOrder ID
	KernelOrderID uint64 `json:"kernel_order_id,omitempty"`
	// KernelOrder creation time
	CreateTime int64 `json:"create_time,omitempty"`
	// KernelOrder last modification time
	UpdateTime int64 `json:"update_time,omitempty"`
	// Trade amount
	Amount int64 `json:"amount"`
	// KernelOrder price, 1,000,000,000 -> 1
	Price int64 `json:"price"`
	// Amount left to fill
	Left int64 `json:"left,omitempty"`
	// Total filled in quote currency
	FilledTotal int64 `json:"filled_total,omitempty"`
	// Order ID
	Id uint64 `json:"id,omitempty"`
	// KernelOrder status  - `open`: to be filled - `closed`: filled - `cancelled`: cancelled
	Status OrderStatus `json:"status,omitempty"`
	// KernelOrder type. limit - limit order
	Type OrderType `json:"type,omitempty"`
	// Time in force  - gtc: GoodTillCancelled - ioc: ImmediateOrCancelled, taker only - poc: PendingOrCancelled, reduce only
	TimeInForce TimeInForce `json:"time_in_force,omitempty"`
}
