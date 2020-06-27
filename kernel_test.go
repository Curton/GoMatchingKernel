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
	// insert ask & bid order to empty 'ask' & 'bid'
	nano := time.Now().UnixNano()
	order := types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    nano,
		UpdateTime:    nano,
		Amount:        -10,
		Price:         100,
		Left:          -10,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	order2 := types.KernelOrder{
		KernelOrderID: 0,
		CreateTime:    nano,
		UpdateTime:    nano,
		Amount:        20,
		Price:         99,
		Left:          20,
		FilledTotal:   0,
		Status:        0,
		Type:          0,
		TimeInForce:   0,
		Id:            "",
	}
	insertPriceCheckedOrder(&order)
	insertPriceCheckedOrder(&order)
	insertPriceCheckedOrder(&order)
	insertPriceCheckedOrder(&order2)
	insertPriceCheckedOrder(&order2)
	insertPriceCheckedOrder(&order2)
	value := ask.Front().Value()
	bucket := value.(*priceBucket)
	println(bucket.size)
	println(bucket.l.Len())
	println(ask.Length)
	println()
	value2 := bid.Front().Value()
	bucket2 := value2.(*priceBucket)
	println(bucket2.size)
	println(bucket2.l.Len())
	println(bid.Length)
}

func TestBreak(t *testing.T) {
	sum := 0
EXIT:
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			sum++
			break EXIT
		}
	}
	println(sum)
}
