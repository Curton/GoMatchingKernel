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
	"io"
	"log"
	"os"
	"strconv"
	"time"
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
	}

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
	for i := range slice {
		l.PushBack(&slice[i])
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

// read orders from file to redo order
func orderLogReader(s *scheduler) {
	for s.f[0] == nil {
		log.Println("Err in orderLogReader : FD is <nil>")
		time.Sleep(time.Second)
	}
	sample := getOrderBinary(&types.KernelOrder{})
	size := len(sample)
	tmp := make([]byte, size)
	var off int64 = 0
	var lastKernelOrder *types.KernelOrder
	for {
		_, err := s.f[0].ReadAt(tmp, off)
		if err != nil {
			if err == io.EOF {
				// stop kernel
				time.Sleep(time.Second)
				st := time.Now().UnixNano()
				s.redoKernel.takeSnapshot("redo", lastKernelOrder)
				et := time.Now().UnixNano()
				log.Println("redo snapshot finished in ", (et-st)/(1000*1000), " ms")
				time.Sleep(redoSnapshotInterval)
				continue
			}
			log.Println(err.Error())
			time.Sleep(time.Second)
			continue
		}
		off += int64(size)
		o := readOrderBinary(tmp)
		s.redoOrderChan <- o
		lastKernelOrder = o
	}
}
