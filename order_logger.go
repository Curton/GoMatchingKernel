/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package exchangeKernel

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"exchangeKernel/types"
	"fmt"
	"os"
	"strconv"
	"time"
)

func writeOrderLog(f *os.File, acceptorDescription string, kernelOder *types.KernelOrder) bool {

	if f == nil {
		var err error
		f, err = os.OpenFile(kernelOrderLogPath+acceptorDescription+"_"+strconv.FormatInt(time.Now().Unix(), 10)+".log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
		if err != nil {
			return false
		}
	}

	if _, err := f.Write(getBytes(kernelOder)); err != nil {
		return false
	}

	return true
}

func getBytes(order *types.KernelOrder) []byte {
	var buffer bytes.Buffer        // Stand-in for a buffer connection
	enc := gob.NewEncoder(&buffer) // Will write to buffer.
	err := enc.Encode(order)
	if err != nil {
		return nil
	}
	return buffer.Bytes()
}

func RecoverSince() {

}

func writeListAsBytes(list *list.List) []byte {
	var buf bytes.Buffer
	slice := make([]types.KernelOrder, 0, list.Len())
	for i := list.Front(); i != nil; i = i.Next() {
		slice = append(slice, i.Value.(types.KernelOrder))
	}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(slice)
	if err != nil {
		fmt.Println(err.Error())
	}
	return buf.Bytes()
}

func readListFromBytes(b []byte) *list.List {
	var buf bytes.Buffer
	buf.Write(b)
	dec := gob.NewDecoder(&buf)
	var slice []types.KernelOrder
	err := dec.Decode(&slice)
	if err != nil {
		fmt.Println(err.Error())
	}
	l := list.New()
	if slice != nil {
		for i := range slice {
			l.PushBack(slice[i])
		}
	}
	return l
}
