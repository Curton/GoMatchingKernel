/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package exchangeKernel

import (
	"bytes"
	"encoding/binary"
	"exchangeKernel/types"
	"os"
	"strconv"
	"time"
)

func WriteOrderLog(kernelOder *types.KernelOrder) bool {

	f, err := os.OpenFile(strconv.FormatInt(time.Now().Unix(), 10)+".log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}

	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.LittleEndian, kernelOder); err != nil {
		return false
	}

	if _, err := f.Write(buf.Bytes()); err != nil {
		return false
	}

	if err := f.Sync(); err != nil {
		return false
	}
	if err := f.Close(); err != nil {
		return false
	}
	return true
}

var f *os.File

var data = []byte("qwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmq\nqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmqwertyuiopasdfghjklzxcvbnmq\n")

func WriteLog() bool {

	if f == nil {
		var err error
		f, err = os.OpenFile(strconv.FormatInt(time.Now().Unix(), 10)+".log",
			os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
		if err != nil {
			return false
		}
	}

	if _, err := f.Write(data); err != nil {
		return false
	}

	return true
}
