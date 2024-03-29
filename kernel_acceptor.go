/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package ker

import (
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/Curton/GoMatchingKernel/types"
)

// scheduler struct represents the scheduler used in the trading system.
type scheduler struct {
	kernel              *kernel
	redoKernel          *kernel
	newOrderChan        chan *types.KernelOrder  // new orders are sending to the channel
	redoOrderChan       chan *types.KernelOrder  // redo orders are sending to the channel
	orderReceivedChan   chan *types.KernelOrder  // get order received confirmation
	internalRequestChan chan internalRequestCode // reserve
	serverId            uint64
	serverMask          uint64
	acceptorDescription string
	r                   *rand.Rand
	f                   *[1]*os.File // kernelOrder logger file
}

type redoKernelStatus uint8

const (
	DISABLED redoKernelStatus = iota
	WAITTING_START
	STARTED
	WAITTING_STOP
	STOPPED
)

const (
	REDO_KERNEL = 1
)

// internalRequestCode represents a code for internal requests.
type internalRequestCode uint16

// orderAcceptor should be run in a goroutine. It processes new orders and checks if they are limit or market orders.
// classify limit orders and market orders
func (s *scheduler) orderAcceptor(kernelFlag ...int) {

	numArgs := len(kernelFlag)
	if numArgs > 1 {
		panic("too many kernelFlag arguments")
	}

	var orderChan chan *types.KernelOrder
	var kernel *kernel
	var orderReceivedChan chan *types.KernelOrder

	if numArgs == 1 {
		if kernelFlag[0] == REDO_KERNEL {
			orderChan = s.redoOrderChan
			kernel = s.redoKernel
			// Dummy orderReceivedChan for redo kernel
			orderReceivedChan = make(chan *types.KernelOrder, 100)
			go func() {
				for {
					<-orderReceivedChan
				}
			}()

		} else {
			panic("Unknown kernelFlag")
		}

	} else {
		orderChan = s.newOrderChan
		kernel = s.kernel
		orderReceivedChan = s.orderReceivedChan
	}

	for {
		paused := false
		if paused {
			<-kernel.pauseChan
			paused = false
		} else {
			select {
			// kernel stop
			case <-s.kernel.ctx.Done():
				return
				// kernel pause
			case <-kernel.pauseChan:
				paused = true
			case order := <-orderChan:
				// Add validation here
				if math.Abs(float64(order.Left)) > math.Abs(float64(order.Amount)) && (order.Amount != 0) {
					log.Println("Invalid order: Left exceeds Amount")
					continue
				}
				if (order.Left < 0 && order.Amount > 0) || (order.Left > 0 && order.Amount < 0) {
					log.Println("Invalid order: Left and Amount have different signs")
					continue
				}
				kernelOrder := *order
				kernelOrder.CreateTime = time.Now().UnixNano()
				uint64R := uint64(s.r.Int63())
				kernelOrder.KernelOrderID = (uint64R >> (16 - 1)) | s.serverMask // use the first 16 bits as server Id

				// write log if saveOrderLog is true and kernelFlag not set
				if saveOrderLog && numArgs == 0 {
					if !writeOrderLog(s.f, s.acceptorDescription, order) {
						log.Panicln("Error in writing order log.")
					}
				}

				// cancel order if amount is zero
				if order.Amount == 0 {
					order.Status = types.CANCELLED
					// accept order signal
					orderReceivedChan <- order
					kernel.cancelOrder(order)
					continue
				}

				// accept order signal
				orderReceivedChan <- &kernelOrder

				if kernelOrder.Type == types.LIMIT {
					// limit order
					if kernelOrder.Amount > 0 {
						// bid order
						if kernelOrder.Price < kernel.ask1Price {
							// cannot match, insert immediatly
							kernel.insertUnmatchedOrder(&kernelOrder)
						} else {
							kernel.matchingOrder(kernel.ask, &kernelOrder, false)
						}
					} else {
						// ask order
						if kernelOrder.Price > kernel.bid1Price {
							// cannot match, insert immediatly
							kernel.insertUnmatchedOrder(&kernelOrder)
						} else {
							kernel.matchingOrder(kernel.bid, &kernelOrder, true)
						}
					}
				} else {
					panic("Unsuported OrderType")
				}
				// else if kernelOrder.Type == types.MARKET {
				// 	// market price order
				// 	// FOK: FillOrKill, fill either completely or none, Only `IOC` and `FOK` are supported when `kernelOrder.Type`=`MARKET`
				// 	// TODO: `IOC` or `FOK`?
				// 	kernelOrder.TimeInForce = types.FOK
				// 	if kernelOrder.Amount > 0 {
				// 		// bid, buy
				// 		if kernel.ask1Price != math.MaxInt64 {
				// 			kernelOrder.Price = int64(float64(kernel.ask1Price) * marketPriceOffset)
				// 			kernel.matchingOrder(kernel.ask, &kernelOrder, false)
				// 		} else {
				// 			kernelOrder.Price = 0
				// 			kernelOrder.Status = types.CANCELLED
				// 			kernel.matchedInfoChan <- &matchedInfo{
				// 				takerOrder: kernelOrder,
				// 			}
				// 		}
				// 	} else {
				// 		// ask, sell
				// 		if kernel.bid1Price != math.MinInt64 {
				// 			kernelOrder.Price = int64(float64(kernel.ask1Price) * marketPriceOffset)
				// 			kernel.matchingOrder(kernel.bid, &kernelOrder, true)
				// 		} else {
				// 			kernelOrder.Price = 0
				// 			kernelOrder.Status = types.CANCELLED
				// 			kernel.matchedInfoChan <- &matchedInfo{
				// 				takerOrder: kernelOrder,
				// 			}
				// 		}
				// 	}
				// }

			case rq := <-s.internalRequestChan:
				_ = rq
			}
		}

	}

}

func initAcceptor(serverId uint64, acceptorDescription string) *scheduler {
	return &scheduler{
		kernel:              newKernel(),
		newOrderChan:        make(chan *types.KernelOrder, 1),
		orderReceivedChan:   make(chan *types.KernelOrder),
		serverId:            serverId,
		serverMask:          serverId << (64 - 16 - 1),
		r:                   rand.New(rand.NewSource(time.Now().UnixNano())),
		acceptorDescription: acceptorDescription,
		f:                   &[1]*os.File{nil},
	}
}

func (s *scheduler) startRedoKernel() {
	s.redoKernel = newKernel()
	s.redoOrderChan = make(chan *types.KernelOrder)
	go s.orderAcceptor(REDO_KERNEL)
	s.redoKernel.startDummyMatchedInfoChan()
	// redo orders from log file
	go orderLogReader(s)
}

func (s *scheduler) startDummyOrderReceivedChan() {
	go func() {
		for {
			<-s.orderReceivedChan
		}
	}()
}
