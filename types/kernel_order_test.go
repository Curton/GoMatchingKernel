package types

import (
	"container/list"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestKernelOrderSize(t *testing.T) {
	assert.Equal(t, 72, int(unsafe.Sizeof(KernelOrder{})))
}

func TestList(t *testing.T) {
	l := list.List{}
	l.PushBack(1)
	// println(l.Len())
}
