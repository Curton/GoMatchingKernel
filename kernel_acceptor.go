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
	newOrderChan        chan *types.KernelOrder
	redoOrderChan       chan *types.KernelOrder
	orderReceivedChan   chan *types.KernelOrder
	internalRequestChan chan internalRequestCode
	serverId            uint64
	serverMask          uint64
	acceptorDescription string
	r                   *rand.Rand
	f                   *[1]*os.File // kernelOrder logger file
}

// internalRequestCode represents a code for internal requests.
type internalRequestCode uint16

// startOrderAcceptor should be run in a goroutine. It processes new orders and checks if they are limit or market orders.
// classify limit orders and market orders
func (s *scheduler) startOrderAcceptor() {
	for {
		select {
		case recv := <-s.newOrderChan:
			kernelOrder := *recv
			kernelOrder.CreateTime = time.Now().UnixNano()
			uint64R := uint64(s.r.Int63())
			kernelOrder.KernelOrderID = (uint64R >> (16 - 1)) | s.serverMask // use the first 16 bits as server Id

			// write log if saveOrderLog is true
			if saveOrderLog {
				if !writeOrderLog(s.f, s.acceptorDescription, recv) {
					log.Panicln("Error in writing order log.")
				}
			}

			// cancel order if amount is zero
			if recv.Amount == 0 {
				recv.Status = types.CANCELLED
				// accept order signal
				s.orderReceivedChan <- recv
				s.kernel.cancelOrder(recv)
				continue
			}

			// accept order signal
			tmp := kernelOrder
			s.orderReceivedChan <- &tmp

			if kernelOrder.Type == types.LIMIT {
				// limit order
				if kernelOrder.Amount > 0 {
					// bid order
					if kernelOrder.Price < s.kernel.ask1Price {
						// annot match, insert immediatly
						s.kernel.insertCheckedOrder(&kernelOrder)
					} else {
						s.kernel.matchingOrder(s.kernel.ask, &kernelOrder, false)
					}
				} else {
					// ask order
					if kernelOrder.Price > s.kernel.bid1Price {
						// cannot match, insert immediatly
						s.kernel.insertCheckedOrder(&kernelOrder)
					} else {
						s.kernel.matchingOrder(s.kernel.bid, &kernelOrder, true)
					}
				}
			} else if kernelOrder.Type == types.MARKET {
				// market price order
				// FOK: FillOrKill, fill either completely or none, Only `IOC` and `FOK` are supported when `kernelOrder.Type`=`MARKET`
				// TODO: `IOC` or `FOK`?
				kernelOrder.TimeInForce = types.FOK
				if kernelOrder.Amount > 0 {
					// bid, buy
					if s.kernel.ask1Price != math.MaxInt64 {
						kernelOrder.Price = int64(float64(s.kernel.ask1Price) * marketPriceOffset)
						s.kernel.matchingOrder(s.kernel.ask, &kernelOrder, false)
					} else {
						kernelOrder.Price = 0
						kernelOrder.Status = types.CANCELLED
						s.kernel.matchedInfoChan <- &matchedInfo{
							takerOrder: kernelOrder,
						}
					}
				} else {
					// ask, sell
					if s.kernel.bid1Price != math.MinInt64 {
						kernelOrder.Price = int64(float64(s.kernel.ask1Price) * marketPriceOffset)
						s.kernel.matchingOrder(s.kernel.bid, &kernelOrder, true)
					} else {
						kernelOrder.Price = 0
						kernelOrder.Status = types.CANCELLED
						s.kernel.matchedInfoChan <- &matchedInfo{
							takerOrder: kernelOrder,
						}
					}
				}
			}

		case rq := <-s.internalRequestChan:
			_ = rq
		}
	}

}

// todo: reuse startOrderAcceptor
func (s *scheduler) startRedoOrderAcceptor() {
	for recv := range s.redoOrderChan {
		// cancel order
		if recv.Amount == 0 {
			recv.Status = types.CANCELLED
			s.redoKernel.cancelOrder(recv)
			continue
		}

		if recv.Type == types.LIMIT {
			// limit order
			if recv.Amount > 0 {
				// bid order
				if recv.Price < s.redoKernel.ask1Price {
					// cannot match, insert immediatly
					s.redoKernel.insertCheckedOrder(recv)
				} else {
					s.redoKernel.matchingOrder(s.redoKernel.ask, recv, false)
				}
			} else {
				// ask order
				if recv.Price > s.redoKernel.bid1Price {
					// cannot match, insert immediatly
					s.redoKernel.insertCheckedOrder(recv)
				} else {
					s.redoKernel.matchingOrder(s.redoKernel.bid, recv, true)
				}
			}
		} else if recv.Type == types.MARKET {
			// market price order
			recv.TimeInForce = types.FOK
			if recv.Amount > 0 {
				// bid, buy
				if s.kernel.ask1Price != math.MaxInt64 {
					recv.Price = int64(float64(s.kernel.ask1Price) * marketPriceOffset)
					s.kernel.matchingOrder(s.kernel.ask, recv, false)
				} else {
					recv.Price = 0
					recv.Status = types.CANCELLED
				}
			} else {
				// ask, sell
				if s.kernel.bid1Price != math.MinInt64 {
					recv.Price = int64(float64(s.kernel.ask1Price) * marketPriceOffset)
					s.kernel.matchingOrder(s.kernel.bid, recv, true)
				} else {
					recv.Price = 0
					recv.Status = types.CANCELLED
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

func (s *scheduler) enableRedoKernel() {
	s.redoKernel = newKernel()
	s.redoOrderChan = make(chan *types.KernelOrder)
	go s.startRedoOrderAcceptor()
	s.redoKernel.startDummyMatchedInfoChan()
	go orderLogReader(s)
}

func (s *scheduler) startDummyOrderConfirmedChan() {
	go func() {
		for {
			<-s.orderReceivedChan
		}
	}()
}
