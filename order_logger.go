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
		var err error
		var f2 *os.File
		f2, err = os.OpenFile(kernelOrderLogPath+acceptorDescription+"_"+strconv.FormatInt(time.Now().Unix(), 10)+".log",
			os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_SYNC, 0644)
		if err != nil {
			return false
		}
		f[0] = f2
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

	if _, err := f[0].Write(getBytes(kernelOrder)); err != nil {
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

func bytesToKernelOrder(b []byte) *types.KernelOrder {
	var buf bytes.Buffer
	buf.Write(b)
	dec := gob.NewDecoder(&buf)
	order := &types.KernelOrder{}
	err := dec.Decode(order)
	if err != nil {
		log.Println(err.Error())
	}
	return order
}

func kernelOrderListToBytes(list *list.List) []byte {
	var buf bytes.Buffer
	slice := make([]types.KernelOrder, 0, list.Len())
	for i := list.Front(); i != nil; i = i.Next() {
		slice = append(slice, *i.Value.(*types.KernelOrder))
	}
	enc := gob.NewEncoder(&buf)
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

//func orderLogReader(f *[1]*os.File) {
//	var off int64 = 0
//	bs := int64(kernelOrderByteSize)
//	for {
//		tmp := make([]byte, 0, kernelOrderByteSize)
//		n, _ := f[0].ReadAt(tmp, off)
//
//		if n != kernelOrderByteSize {
//			time.Sleep(time.Second)
//			stat, _ := f[0].Stat()
//			//fmt.Println(err.Error())
//			fmt.Println("kernelOrderByteSize: ",kernelOrderByteSize)
//			fmt.Println(stat.Name())
//			fmt.Println(stat.Size())
//			fmt.Println(stat.Mode())
//			fmt.Println(stat.ModTime())
//			fmt.Println(stat.Sys())
//			fmt.Println("off: ", off)
//			fmt.Println("size: ", n)
//			continue
//		}
//
//		off += bs
//
//		fmt.Println("orderLogReader : ", bytesToKernelOrder(tmp))
//
//	}
//}
