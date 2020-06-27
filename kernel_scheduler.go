/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package exchangeKernel

import "exchangeKernel/types"

func acceptNewOrder(order *types.KernelOrder) bool {
	// 订单获得 KernelOrderId 表示被撮合核心接受
	return true
}
