/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:20
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"testing"
	"time"
)

func Test_insertPriceCheckedOrder1(t *testing.T) {
	// insert ask order to empty ask
	order := types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    time.Now().UnixNano(),
		UpdateTime:    0,
		Amount:        -10,
		Price:         100,
		Left:          -10,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	insertPriceCheckedOrder(&order)
}
