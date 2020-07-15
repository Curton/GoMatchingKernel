/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package exchangeKernel

import "time"

var (
	kernelOrderLogPath   = "./kernelorder_log/"
	kernelSnapshotPath   = "./orderbook_snapshot/"
	saveOrderLog         = true
	redoSnapshotInterval = time.Second
	marketOrderOffset    = 1.1
)
