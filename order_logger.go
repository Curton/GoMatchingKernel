/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package ker

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"encoding/gob"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Curton/GoMatchingKernel/types"
)

// writeOrderLog writes new orders to a log file. It creates a new file if one doesn't exist.
// It returns a bool indicating success or failure.
func writeOrderLog(f *[1]*os.File, acceptorDescription string, kernelOrder *types.KernelOrder) bool {
	// Check if the file is nil. If it is, create the directory and the file.
	if f[0] == nil {
		// Create the directory if it doesn't exist.
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

	// Write the binary representation of the order to the file.
	if _, err := f[0].Write(getOrderBinary(kernelOrder)); err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}

// getBytes returns the byte slice representation of the order.
func getBytes(order *types.KernelOrder) []byte {
	// Create a new buffer and an encoder that writes to the buffer.
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	// Encode the order to the buffer.
	err := enc.Encode(order)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// bytesToKernelOrder converts a byte slice to a KernelOrder.
func bytesToKernelOrder(b []byte) *types.KernelOrder {
	// Create a new buffer and write the bytes to it.
	buf := new(bytes.Buffer)
	buf.Write(b)

	// Create a new decoder that reads from the buffer.
	dec := gob.NewDecoder(buf)

	// Decode the bytes from the buffer to a KernelOrder.
	order := &types.KernelOrder{}
	err := dec.Decode(order)
	if err != nil {
		log.Println(err.Error())
	}
	return order
}

// kernelOrderListToBytes converts a list of KernelOrders to a byte slice.
func kernelOrderListToBytes(list *list.List) []byte {
	// Create a new buffer and an encoder that writes to the buffer.
	buf := new(bytes.Buffer)

	// Convert the list to a slice.
	slice := make([]types.KernelOrder, 0, list.Len())
	for i := list.Front(); i != nil; i = i.Next() {
		slice = append(slice, *i.Value.(*types.KernelOrder))
	}

	// Encode the slice to the buffer.
	enc := gob.NewEncoder(buf)
	err := enc.Encode(slice)
	if err != nil {
		log.Println(err.Error())
	}

	return buf.Bytes()
}

// readListFromBytes converts a byte slice to a list of KernelOrders.
func readListFromBytes(b []byte) *list.List {
	// Create a new buffer and write the bytes to it.
	var buf bytes.Buffer
	buf.Write(b)

	// Create a new decoder that reads from the buffer.
	dec := gob.NewDecoder(&buf)

	// Decode the bytes from the buffer to a slice.
	var slice []types.KernelOrder
	err := dec.Decode(&slice)
	if err != nil {
		log.Println(err.Error())
	}

	// Convert the slice to a list.
	l := list.New()
	for i := range slice {
		l.PushBack(&slice[i])
	}
	return l
}

// getOrderBinary returns the binary representation of the order.
func getOrderBinary(order *types.KernelOrder) []byte {
	// Create a new buffer.
	buf := new(bytes.Buffer)

	// Write the binary representation of the order to the buffer.
	err := binary.Write(buf, binary.LittleEndian, order)
	if err != nil {
		log.Println("binary.Write failed:", err)
	}

	// Return the bytes of the buffer.
	return buf.Bytes()
}

// readOrderBinary converts a binary representation of an order to a KernelOrder.
func readOrderBinary(b []byte) *types.KernelOrder {
	// Create a new KernelOrder.
	order := new(types.KernelOrder)

	// Create a new reader with the bytes.
	buf := bytes.NewReader(b)

	// Read the binary data into the KernelOrder.
	err := binary.Read(buf, binary.LittleEndian, order)
	if err != nil {
		log.Println("binary.Read failed:", err)
	}

	// Return the KernelOrder.
	return order
}

// orderLogReader reads orders from a file to the redo order channel.
func orderLogReader(s *scheduler) {
	// Loop until we can open the file properly.
	for s.f[0] == nil {
		log.Println("Failed to read file in orderLogReader : FD is <nil>, retry after ", redoSnapshotInterval, " second.")
		time.Sleep(redoSnapshotInterval)
	}

	// Build a sample order to determine the size of each order in the file.
	sample := getOrderBinary(&types.KernelOrder{})
	size := len(sample)

	// Temporary buffer for reading orders.
	tmp := make([]byte, size)

	// Offset for reading
	var off int64 = 0
	var lastKernelOrder *types.KernelOrder

	// Loop reading orders from the file.
	for {
		_, err := s.f[0].ReadAt(tmp, off)
		if err != nil {
			if err == io.EOF {
				// time.Sleep(time.Second)

				// pause redokernel
				s.redoKernel.Pause()
				// ensure matching work done
				time.Sleep(time.Microsecond)
				st := time.Now().UnixNano()
				s.redoKernel.takeSnapshot("redo", lastKernelOrder)
				et := time.Now().UnixNano()
				s.redoKernel.Resume()
				log.Println("orderLogReader() :redo snapshot finished in ", (et-st)/(1000*1000), " ms")
				time.Sleep(redoSnapshotInterval)
				continue
			}
			log.Println(err.Error())
			time.Sleep(time.Second)
			continue
		}

		// Increment the offset for the next read.
		off += int64(size)

		// Convert the bytes back to a KernelOrder.
		o := readOrderBinary(tmp)

		// Send the KernelOrder to the redo order channel.
		s.redoOrderChan <- o
		lastKernelOrder = o
	}
}
