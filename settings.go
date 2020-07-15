/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/2 14:31
 */

package exchangeKernel

import "time"

var (
	kernelOrderLogPath   = "./kernelorder_log/"
	kernelSnapshotPath   = "./orderbook_snapshot/"
	kernelOrderLogCache  = true
	saveOrderLog         = true
	redoSnapshotInterval = time.Second
)
