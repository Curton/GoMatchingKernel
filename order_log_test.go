/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:51
 */

package exchangeKernel

import (
	"exchangeKernel/types"
	"testing"
)

// GOMAXPROCS=1 go test -bench=BenchmarkWrite -run=none -benchtime=1s -benchmem
func BenchmarkWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WriteOrderLog(nil)
	}
}

func TestWrite(t *testing.T) {
	for i := 0; i < 10000; i++ {
		if WriteOrderLog(&types.KernelOrder{
			KernelOrderID: 0,
			CreateTime:    0,
			UpdateTime:    0,
			Amount:        0,
			Price:         0,
			Left:          0,
			FilledTotal:   0,
			Status:        0,
			Type:          0,
			TimeInForce:   0,
			Id:            0,
		}) == false {
			panic("")
		}
	}
}
