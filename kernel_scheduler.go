/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package exchangeKernel

import (
	"exchangeKernel/types"
)

var (
	orderChan = make(chan *types.KernelOrder)
	//orderBuff = make([]*types.KernelOrder, 0, 8)
)

// run in go routine
// 判断 限价单 与 市价单
func orderAcceptor() {
	for kernelOrder := range orderChan {
		if kernelOrder.Type == types.LIMIT {
			// 限价单
			if kernelOrder.Amount > 0 {
				// bid order
				if kernelOrder.Price < ask1Price {
					// 不能成交,直接插入
					insertCheckedOrder(kernelOrder)
				} else {
					matchingOrder(ask, kernelOrder, false)
				}
			} else {
				// ask order
				if kernelOrder.Price > bid1Price {
					// 不能成交,直接插入
					insertCheckedOrder(kernelOrder)
				} else {
					matchingOrder(bid, kernelOrder, true)
				}
			}
		} else if kernelOrder.Type == types.MARKET {
			// 市价单
			// todo
		}
	}
}
