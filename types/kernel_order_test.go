package types

import (
	"container/list"
	"testing"
	"unsafe"
)

func TestOrderSize(t *testing.T) {
	kernelOrder := KernelOrder{
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
	}
	println(unsafe.Sizeof(kernelOrder))
}

func TestList(t *testing.T) {
	l := list.List{}
	l.PushBack(1)
	println(l.Len())
}
