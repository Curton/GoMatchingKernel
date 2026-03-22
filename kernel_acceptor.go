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

	paused := false
	for {
		if paused {
			select {
			case <-s.kernel.ctx.Done():
				return
			case <-kernel.pauseChan:
				paused = false
			}
		} else {
			select {
			case <-s.kernel.ctx.Done():
				return
			case <-kernel.pauseChan:
				paused = true
			case order := <-orderChan:
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
				kernelOrder.KernelOrderID = (uint64R >> (16 - 1)) | s.serverMask

				if saveOrderLog && numArgs == 0 {
					if !writeOrderLog(s.f, s.acceptorDescription, order) {
						log.Panicln("Error in writing order log.")
					}
				}

				if order.Amount == 0 {
					order.Status = types.CANCELLED
					orderReceivedChan <- order
					kernel.cancelOrder(order)
					continue
				}

				orderReceivedChan <- &kernelOrder

				if kernelOrder.Type == types.LIMIT {
					if kernelOrder.Amount > 0 {
						if kernelOrder.Price < kernel.ask1Price {
							kernel.insertUnmatchedOrder(&kernelOrder)
						} else {
							kernel.matchingOrder(kernel.ask, &kernelOrder, false)
						}
					} else {
						if kernelOrder.Price > kernel.bid1Price {
							kernel.insertUnmatchedOrder(&kernelOrder)
						} else {
							kernel.matchingOrder(kernel.bid, &kernelOrder, true)
						}
					}
				} else {
					panic("Unsuported OrderType")
				}
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
