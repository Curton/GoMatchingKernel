/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package exchangeKernel

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"encoding/gob"
	"exchangeKernel/types"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	CRLF2 = "\r\n\r\n"
)

var (
	cachedOrderMap = make(map[string]types.KernelOrder)
	lastTimeMap    = make(map[string]int64)
	CrlfBytes      = []byte(CRLF2)
)

func writeOrderLog(f *[1]*os.File, acceptorDescription string, kernelOrder *types.KernelOrder) bool {

	if f[0] == nil {
		if err := os.MkdirAll(kernelOrderLogPath, 0755); err != nil {
			log.Println(err.Error())
			return false
		}
		f2, err := os.OpenFile(kernelOrderLogPath+acceptorDescription+"_"+strconv.FormatInt(time.Now().Unix(), 10)+".log",
			os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
		if err != nil {
			log.Println(err.Error())
			return false
		}
		f[0] = f2
		log.Println("hhh")
	}

	//if kernelOrderLogCache {
	//	cachedOrder, ok := cachedOrderMap[acceptorDescription]
	//	if ok {
	//		b1 := getBytes(&cachedOrder)
	//		b2 := getBytes(kernelOrder)
	//		buf := make([]byte, 0, len(b1)+len(b2))
	//		buf = append(buf, b1...)
	//		buf = append(buf, b2...)
	//		if _, err := f[0].Write(buf); err != nil {
	//			return false
	//		}
	//		delete(cachedOrderMap, acceptorDescription)
	//	} else {
	//		i, ok := lastTimeMap[acceptorDescription]
	//		if ok {
	//			if kernelOrder.CreateTime-i < 3_000_000 {
	//				lastTimeMap[acceptorDescription] = kernelOrder.CreateTime
	//				cachedOrderMap[acceptorDescription] = *kernelOrder
	//				return true
	//			} else {
	//				if _, err := f[0].Write(getBytes(kernelOrder)); err != nil {
	//					return false
	//				}
	//			}
	//		} else {
	//			lastTimeMap[acceptorDescription] = kernelOrder.CreateTime
	//			if _, err := f[0].Write(getBytes(kernelOrder)); err != nil {
	//				return false
	//			}
	//		}
	//	}
	//} else {
	//	if _, err := f[0].Write(getBytes(kernelOrder)); err != nil {
	//		return false
	//	}
	//}

	if _, err := f[0].Write(getOrderBinary(kernelOrder)); err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

func getBytes(order *types.KernelOrder) []byte {
	buf := new(bytes.Buffer)   // Stand-in for a buffer connection
	enc := gob.NewEncoder(buf) // Will write to buffer.
	err := enc.Encode(order)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func bytesToKernelOrder(b []byte) *types.KernelOrder {
	buf := new(bytes.Buffer)
	buf.Write(b)
	dec := gob.NewDecoder(buf)
	order := &types.KernelOrder{}
	err := dec.Decode(order)
	if err != nil {
		log.Println(err.Error())
	}
	return order
}

func kernelOrderListToBytes(list *list.List) []byte {
	buf := new(bytes.Buffer)
	slice := make([]types.KernelOrder, 0, list.Len())
	for i := list.Front(); i != nil; i = i.Next() {
		slice = append(slice, *i.Value.(*types.KernelOrder))
	}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(slice)
	if err != nil {
		log.Println(err.Error())
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
		log.Println(err.Error())
	}
	l := list.New()
	if slice != nil {
		for i := range slice {
			l.PushBack(&slice[i])
		}
	}
	return l
}

func getOrderBinary(order *types.KernelOrder) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, order)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func readOrderBinary(b []byte) *types.KernelOrder {
	order := new(types.KernelOrder)
	buf := bytes.NewReader(b)
	err := binary.Read(buf, binary.LittleEndian, order)
	if err != nil {
		log.Println("binary.Read failed:", err)
	}
	return order
}

func orderLogReader(f *[1]*os.File) {
	if f == nil || f[0] == nil {
		log.Println("Err in orderLogReader : FD is <nil>")
		return
	}
	sample := getOrderBinary(&types.KernelOrder{})
	size := len(sample)
	tmp := make([]byte, size)
	var off int64 = 0
	for {
		n, err := f[0].ReadAt(tmp, off)
		if err != nil {
			// todo
			log.Println(err.Error())
			log.Println("read: ", n)
			time.Sleep(time.Second)
			continue
		}
		off += (int64)(size)
		o := readOrderBinary(tmp)
		fmt.Println(o)
	}
}
