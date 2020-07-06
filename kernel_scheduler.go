/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"math/rand"
	"time"
)

//var (
//	orderChan = make(chan *types.KernelOrder)
//	//writeLogChan        = make(chan *types.KernelOrder)
//	//writeLogConfirmChan = make(chan bool)
//	//orderBuff = make([]*types.KernelOrder, 0, 8)
//	serverId   uint16 = 1
//	serverMask        = uint64(serverId) << (64 - 16 - 1)
//	r                 = rand.New(rand.NewSource(time.Now().UnixNano()))
//)

type scheduler struct {
	kernel              *kernel
	newOrderChan        chan *types.KernelOrder
	serverId            uint64
	serverMask          uint64
	r                   *rand.Rand
	acceptorDescription string
}

// run in go routine
// 判断 限价单 与 市价单
func (s *scheduler) startOrderAcceptor() {
	//go logAcceptor()
	for recv := range s.newOrderChan {
		kernelOrder := *recv
		kernelOrder.CreateTime = time.Now().UnixNano()
		uint64R := uint64(s.r.Int63())
		kernelOrder.KernelOrderID = (uint64R >> (16 - 1)) | s.serverMask // use the first 16 bits as server Id
		// write log
		writeOrderLog(&kernelOrder)
		//writeLogChan <- &kernelOrder
		//<- writeLogConfirmChan
		if kernelOrder.Type == types.LIMIT {
			// 限价单
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
	}
}

func initAcceptor(serverId uint64, acceptorDescription string) *scheduler {
	return &scheduler{
		kernel:              NewKernel(),
		newOrderChan:        make(chan *types.KernelOrder),
		serverId:            serverId,
		serverMask:          uint64(serverId) << (64 - 16 - 1),
		r:                   rand.New(rand.NewSource(time.Now().UnixNano())),
		acceptorDescription: acceptorDescription,
	}
}

//
//func logAcceptor() {
//	for order := range writeLogChan {
//		writeOrderLog(order)
//		writeLogConfirmChan <- true
//	}
//}
