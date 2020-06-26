/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/6/26 19:05
 */

package exchangeKernel

import (
	"fmt"
	"testing"
)

// we only use int64 in exchange kernel

func BenchmarkFP64(b *testing.B) {
	var sum float64 = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum += 3
	}
	b.StopTimer()
	fmt.Println(sum)
}

func BenchmarkInt64(b *testing.B) {
	var sum int64 = 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum += 3
	}
	b.StopTimer()
	fmt.Println(sum)
}

//BenchmarkFP64
//3
//300
//30000
//3e+06
//3e+08
//3e+09
//BenchmarkFP64-16     	1000000000	         0.975 ns/op
//BenchmarkInt64
//3
//300
//30000
//3000000
//300000000
//3000000000
//BenchmarkInt64-16    	1000000000	         0.487 ns/op
//PASS
