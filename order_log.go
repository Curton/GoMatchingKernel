/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package exchangeKernel

import (
	"bytes"
	"encoding/gob"
	"exchangeKernel/types"
	"os"
	"strconv"
	"time"
)

var f *os.File

func writeOrderLog(kernelOder *types.KernelOrder) bool {

	if f == nil {
		var err error
		f, err = os.OpenFile(strconv.FormatInt(time.Now().Unix(), 10)+".log",
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
