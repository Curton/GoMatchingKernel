package types

type OrderStatus int8

type OrderType int8

type TimeInForce int8

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
	IOC
	POC
)

// '1' is represent as '1,000,000,000' in Price
const (
	One int64 = 1_000_000_000
)

// 80 bytes
type KernelOrder struct {
	// Exchange Kernel KernelOrder ID
	KernelOrderID int64 `json:"id,omitempty"`
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
	// KernelOrder status  - `open`: to be filled - `closed`: filled - `cancelled`: cancelled
	Status OrderStatus `json:"status,omitempty"`
	// KernelOrder type. limit - limit order
	Type OrderType `json:"type,omitempty"`
	// Time in force  - gtc: GoodTillCancelled - ioc: ImmediateOrCancelled, taker only - poc: PendingOrCancelled, reduce only
	TimeInForce TimeInForce `json:"time_in_force,omitempty"`
	// Order ID
	Id string `json:"id,omitempty"`
}
