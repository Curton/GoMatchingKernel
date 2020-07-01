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

var (
	orderChan = make(chan *types.KernelOrder)
	//orderBuff = make([]*types.KernelOrder, 0, 8)
	serverId   uint16 = 1
	serverMask        = uint64(serverId) << (64 - 16 - 1)
	r                 = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// run in go routine
// 判断 限价单 与 市价单
func orderAcceptor() {
	for recv := range orderChan {
		kernelOrder := *recv
		kernelOrder.CreateTime = time.Now().UnixNano()
		uint64R := uint64(r.Int63())
		kernelOrder.KernelOrderID = (uint64R >> (16 - 1)) | serverMask // use the first 16 bits as server Id
		if kernelOrder.Type == types.LIMIT {
			// 限价单
			if kernelOrder.Amount > 0 {
				// bid order
				if kernelOrder.Price < ask1Price {
					// 不能成交,直接插入
					insertCheckedOrder(&kernelOrder)
				} else {
					matchingOrder(ask, &kernelOrder, false)
				}
			} else {
				// ask order
				if kernelOrder.Price > bid1Price {
					// 不能成交,直接插入
					insertCheckedOrder(&kernelOrder)
				} else {
					matchingOrder(bid, &kernelOrder, true)
				}
			}
		} else if kernelOrder.Type == types.MARKET {
			// 市价单
			// todo
		}
	}
}
