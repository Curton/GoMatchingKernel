/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:51
 */

package exchangeKernel

import (
	"container/list"
	"exchangeKernel/types"
	"fmt"
	"testing"
)

// GOMAXPROCS=1 go test -bench=BenchmarkWrite -run=none -benchtime=1s -benchmem
func BenchmarkWrite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		writeOrderLog(nil, "test", &types.KernelOrder{
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

func TestWrite(t *testing.T) {
	for i := 0; i < 10000; i++ {
		if writeOrderLog(nil, "test", &types.KernelOrder{
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

	l1.PushFront(types.KernelOrder{
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
	l1.PushFront(types.KernelOrder{
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
	l1.PushFront(types.KernelOrder{
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
	l1.PushFront(types.KernelOrder{
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

	l1Bytes := writeListAsBytes(l1)
	fmt.Println(l1Bytes)

	l2 := readListFromBytes(l1Bytes)

	l2Bytes := writeListAsBytes(l2)
	fmt.Println(l2Bytes)

}

func BenchmarkWriteList(b *testing.B) {
	l := list.New()
	for i := 0; i < b.N; i++ {
		l.PushBack(types.KernelOrder{
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
	writeListAsBytes(l)
}
