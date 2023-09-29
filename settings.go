/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package ker

import "time"

var (
	kernelOrderLogPath   = "./kernelorder_log/"
	kernelSnapshotPath   = "./orderbook_snapshot/"
	saveOrderLog         = true
	redoSnapshotInterval = time.Second
	marketPriceOffset    = 1.1
)

type settings struct {
	kernelOrderLogPath   string
	kernelSnapshotPath   string
	saveOrderLog         bool
	redoSnapshotInterval uint64
	marketPriceOffset    float64
}