/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"log"
	"math/rand"
	"os"
	"time"
)

type scheduler struct {
	kernel               *kernel
	redoKernel           *kernel
	newOrderChan         chan *types.KernelOrder
	redoOrderChan        chan *types.KernelOrder
	orderConfirmedChan   chan *types.KernelOrder
	requestQuotationChan chan bool
	serverId             uint64
	serverMask           uint64
	r                    *rand.Rand
	acceptorDescription  string
	f                    *[1]*os.File // kernelOrder logger file
}

// should run in go routine
// 判断 限价单 与 市价单
func (s *scheduler) startOrderAcceptor() {
	for {
		select {
		case recv := <-s.newOrderChan:
			kernelOrder := *recv
			kernelOrder.CreateTime = time.Now().UnixNano()
			uint64R := uint64(s.r.Int63())
			kernelOrder.KernelOrderID = (uint64R >> (16 - 1)) | s.serverMask // use the first 16 bits as server Id
			// write log
			if saveOrderLog {
				if writeOrderLog(s.f, s.acceptorDescription, recv) != true {
					log.Panicln("Err in write order log.")
				}
			}
			// cancel order
			if recv.Amount == 0 {
				recv.Status = types.CANCELLED
				// accept order signal
				s.orderConfirmedChan <- recv
				s.kernel.cancelOrder(recv)
				continue
			}

			// accept order signal
			tmp := kernelOrder
			s.orderConfirmedChan <- &tmp

			if kernelOrder.Type == types.LIMIT {
				// limit order, 限价单
				if kernelOrder.Amount > 0 {
					// bid order
					if kernelOrder.Price < s.kernel.ask1Price {
						// 不能成交,直接插入
						s.kernel.insertCheckedOrder(&kernelOrder)
					} else {
						s.kernel.matchingOrder(s.kernel.ask, &kernelOrder, false)
					}
				} else {
					// ask order
					if kernelOrder.Price > s.kernel.bid1Price {
						// 不能成交,直接插入
						s.kernel.insertCheckedOrder(&kernelOrder)
					} else {
						s.kernel.matchingOrder(s.kernel.bid, &kernelOrder, true)
					}
				}
			} else if kernelOrder.Type == types.MARKET {
				// 市价单
				// todo
			}

		case rq := <-s.requestQuotationChan:
			_ = rq
		}
	}

}

func (s *scheduler) startRedoOrderAcceptor() {
	for recv := range s.redoOrderChan {
		// cancel order
		if recv.Amount == 0 {
			recv.Status = types.CANCELLED
			s.redoKernel.cancelOrder(recv)
			continue
		}

		if recv.Type == types.LIMIT {
			// limit order, 限价单
			if recv.Amount > 0 {
				// bid order
				if recv.Price < s.redoKernel.ask1Price {
					// 不能成交,直接插入
					s.redoKernel.insertCheckedOrder(recv)
				} else {
					s.redoKernel.matchingOrder(s.redoKernel.ask, recv, false)
				}
			} else {
				// ask order
				if recv.Price > s.redoKernel.bid1Price {
					// 不能成交,直接插入
					s.redoKernel.insertCheckedOrder(recv)
				} else {
					s.redoKernel.matchingOrder(s.redoKernel.bid, recv, true)
				}
			}
		} else if recv.Type == types.MARKET {
			// 市价单
			// todo
		}
	}
}

func initAcceptor(serverId uint64, acceptorDescription string) *scheduler {
	return &scheduler{
		kernel:              newKernel(),
		newOrderChan:        make(chan *types.KernelOrder, 1),
		orderConfirmedChan:  make(chan *types.KernelOrder),
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
			<-s.orderConfirmedChan
		}
	}()
}
