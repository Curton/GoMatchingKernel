/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:51
 */

package ker

import (
	"container/list"
	_ "fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"os"
	"testing"

	"github.com/Curton/GoMatchingKernel/types"
)

// GOMAXPROCS=1 go test -bench=BenchmarkWrite -run=none -benchtime=1s -benchmem
func BenchmarkWriteOrderLog(b *testing.B) {
	var f = &[1]*os.File{nil}
	for i := 0; i < b.N; i++ {
		writeOrderLog(f, "test", &types.KernelOrder{
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
		})
	}
}

func TestWriteOrderLog(t *testing.T) {
	var f = &[1]*os.File{nil}
	for i := 0; i < 10_000; i++ {
		if writeOrderLog(f, "test", &types.KernelOrder{
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

func TestWriteAndReadList(t *testing.T) {
	l1 := list.New()
	//l2 := list.New()

	l1.PushFront(&types.KernelOrder{
		KernelOrderID: uint64(1),
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
	})
	l1.PushFront(&types.KernelOrder{
		KernelOrderID: uint64(2),
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
	})
	l1.PushFront(&types.KernelOrder{
		KernelOrderID: uint64(3),
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
	})
	l1.PushFront(&types.KernelOrder{
		KernelOrderID: uint64(4),
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
	})

	l1Bytes := kernelOrderListToBytes(l1)
	//fmt.Println(l1Bytes)

	l2 := readListFromBytes(l1Bytes)

	l2Bytes := kernelOrderListToBytes(l2)
	//fmt.Println(l2Bytes)

	assert.Equal(t, l1Bytes, l2Bytes)

}

func BenchmarkWriteList(b *testing.B) {
	l := list.New()
	for i := 0; i < b.N; i++ {
		l.PushBack(&types.KernelOrder{
			KernelOrderID: uint64(i),
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
		})
	}
	b.ResetTimer()
	kernelOrderListToBytes(l)
}

func Test_bytesToKernelOrder(t *testing.T) {
	bytes := getBytes(&types.KernelOrder{
		KernelOrderID: 123,
		CreateTime:    2,
		UpdateTime:    3,
		Amount:        4,
		Price:         5,
		Left:          6,
		FilledTotal:   7,
		Status:        8,
		Type:          9,
		TimeInForce:   10,
		Id:            11,
	})

	order := bytesToKernelOrder(bytes)
	_ = order
	// fmt.Println(*order)
}

func Test_getBytes(t *testing.T) {
	bytes := getBytes(&types.KernelOrder{
		KernelOrderID: math.MaxUint64,
		CreateTime:    math.MinInt64,
		UpdateTime:    math.MinInt64,
		Amount:        math.MinInt64,
		Price:         math.MinInt64,
		Left:          math.MinInt64,
		FilledTotal:   math.MinInt64,
		Status:        math.MaxUint8,
		Type:          math.MaxUint8,
		TimeInForce:   math.MaxUint8,
		Id:            math.MaxUint64,
	})
	assert.Equal(t, len(bytes), 256)
	assert.Equal(t, cap(bytes), 352)
	// fmt.Println(len(bytes))
	// fmt.Println(cap(bytes))
}

func Test_getBinary(t *testing.T) {
	order := &types.KernelOrder{
		KernelOrderID: math.MaxUint64,
		CreateTime:    math.MinInt64,
		UpdateTime:    math.MinInt64,
		Amount:        math.MinInt64,
		Price:         math.MinInt64,
		Left:          math.MinInt64,
		FilledTotal:   math.MinInt64,
		Status:        math.MaxUint8,
		Type:          math.MaxUint8,
		TimeInForce:   math.MaxUint8,
		Id:            math.MaxUint64,
	}
	bytes := getOrderBinary(order)
	order2 := &types.KernelOrder{}
	bytes2 := getOrderBinary(order2)
	assert.Equal(t, true, *readOrderBinary(bytes) == *order)
	assert.Equal(t, true, *readOrderBinary(bytes2) == *order2)
}
