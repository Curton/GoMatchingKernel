package types

type OrderStatus uint8

type OrderType uint8

type TimeInForce uint8

const (
	LIMIT OrderType = iota
	MARKET
)

const (
	OPEN OrderStatus = iota
	CLOSED
	CANCELLED
)

const (
	GTC TimeInForce = iota
	FOK             /* 无法全部立即成交就撤销 */
	POC
	IOC /* 无法立即成交(吃单)的部分就撤销 */
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
