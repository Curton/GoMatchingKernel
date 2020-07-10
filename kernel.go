/**
 * @author Covey Liu, covey@liukedun.com
 * @date 2020/7/6 11:24
 */

package exchangeKernel

import (
	"container/list"
	"exchangeKernel/types"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

type kernel struct {
	ask             *SkipList
	bid             *SkipList
	ask1Price       int64
	bid1Price       int64
	matchedInfoChan chan *matchedInfo
	errorInfoChan   chan *KernelErr
	ask1PriceMux    sync.Mutex
	bid1PriceMux    sync.Mutex
}

type matchedInfo struct {
	makerOrders    []types.KernelOrder
	matchedSizeMap map[uint64]int64
	takerOrder     types.KernelOrder
}

type KernelErr error

type Kernel struct {
}

// new order at the head of the list, old order at the tail of the list
// a.k.a price level
type priceBucket struct {
	l    *list.List
	Left int64
}

type orderBook struct {
	ask []orderBookItem
	bid []orderBookItem
}

type orderBookItem struct {
	Price int64
	Size  int64
}

// should stop kernel before calling this func
func (k *kernel) takeSnapshot(description string) {
	uTime := time.Now().Unix()
	uTimeFmt := strconv.FormatInt(uTime, 10)
	askBasePath := kernelSnapshotPath + description + "/" + uTimeFmt + "/ask/"
	bidBasePath := kernelSnapshotPath + description + "/" + uTimeFmt + "/bid/"
	//syscall.Umask(0)
	_ = os.MkdirAll(askBasePath, 0755)
	_ = os.MkdirAll(bidBasePath, 0755)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := k.ask.Front(); bucket != nil; bucket = bucket.Next() {
			pb := bucket.value.(*priceBucket)
			order := pb.l.Front().Value.(*types.KernelOrder)
			price := order.Price
			path := askBasePath + strconv.FormatInt(price, 10) + ".list"
			f, err := os.OpenFile(path, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err.Error())
			}
			bytes := kernelOrderListToBytes(pb.l)
			_, err = f.Write(bytes)
			if err != nil {
				panic(err.Error())
			}
			err = f.Close()
			if err != nil {
				panic(err.Error())
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for bucket := k.bid.Front(); bucket != nil; bucket = bucket.Next() {
			pb := bucket.value.(*priceBucket)
			order := pb.l.Front().Value.(*types.KernelOrder)
			price := order.Price
			path := bidBasePath + strconv.FormatInt(price, 10) + ".list"
			f, _ := os.OpenFile(path, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0644)
			bytes := kernelOrderListToBytes(pb.l)
			_, _ = f.Write(bytes)
			_ = f.Close()
		}
	}()

	wg.Wait()
	f, _ := os.OpenFile(kernelSnapshotPath+description+"/"+uTimeFmt+"/finished.log", os.O_EXCL|os.O_CREATE|os.O_WRONLY|os.O_SYNC, 0644)
	_, _ = f.WriteString("If you see this file, it means snapshot is completed.")
	_ = f.Close()
}

// should sync call
func (k *kernel) cancelOrder(order *types.KernelOrder) {
	get := k.ask.Get(float64(order.Price))
	if get != nil {
		bucket := get.Value().(*priceBucket)
		for i := bucket.l.Front(); i != nil; i = i.Next() {
			kernelOrder := i.Value.(*types.KernelOrder)
			if kernelOrder.KernelOrderID == order.KernelOrderID {
				bucket.Left -= kernelOrder.Left
				bucket.l.Remove(i)
				break
			}
		}
		// remove bucket if empty
		if bucket.Left == 0 {
			k.ask.Remove(float64(order.Price))
		}
		// reset ask1Price
		if k.ask.Length == 0 {
			k.ask1Price = math.MaxInt64
		} else {
			k.ask1Price = k.ask.Front().value.(*priceBucket).l.Front().Value.(*types.KernelOrder).Price
		}
		return
	}

	get2 := k.bid.Get(float64(-order.Price))
	if get2 != nil {
		bucket := get2.Value().(*priceBucket)
		for i := bucket.l.Front(); i != nil; i = i.Next() {
			kernelOrder := i.Value.(*types.KernelOrder)
			if kernelOrder.KernelOrderID == order.KernelOrderID {
				bucket.Left -= kernelOrder.Left
				bucket.l.Remove(i)
				break
			}
		}
		// remove bucket if empty
		if bucket.Left == 0 {
			k.bid.Remove(float64(-order.Price))
		}
		// reset bid1Price
		if k.bid.Length == 0 {
			k.bid1Price = math.MinInt64
		} else {
			k.bid1Price = k.bid.Front().value.(*priceBucket).l.Front().Value.(*types.KernelOrder).Price
		}
		return
	}
	log.Println("cancel err, can't find order : ", order.KernelOrderID)
}

// after price & amount checked, 附条件异步, (不同价格)的(不可成交)订单可以同时插入, 以上两个条件需同时满足
func (k *kernel) insertCheckedOrder(order *types.KernelOrder) bool {
	if order.Amount < 0 {
		get := k.ask.Get(float64(order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			bucket.Left += order.Left
			return true
		} else {
			l := list.New()
			l.PushFront(order)
			k.ask.Set(float64(order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL 减少锁开销
			if k.ask1Price == math.MaxInt64 {
				k.ask1PriceMux.Lock()
				if k.ask1Price == math.MaxInt64 {
					k.ask1Price = order.Price
				} else if k.ask1Price > order.Price {
					k.ask1Price = order.Price
				}
				k.ask1PriceMux.Unlock()
			} else if k.ask1Price > order.Price {
				k.ask1PriceMux.Lock()
				if k.ask1Price > order.Price {
					k.ask1Price = order.Price
				}
				k.ask1PriceMux.Unlock()
			}
			return true
		}
	} else {
		get := k.bid.Get(float64(-order.Price))
		if get != nil {
			bucket := get.Value().(*priceBucket)
			bucket.l.PushFront(order)
			bucket.Left += order.Left
			return true
		} else {
			l := list.New()
			l.PushFront(order)
			k.bid.Set(float64(-order.Price), &priceBucket{
				l:    l,
				Left: order.Left,
			})
			// DCL 减少锁开销
			if k.bid1Price == math.MinInt64 {
				k.bid1PriceMux.Lock()
				if k.bid1Price == math.MinInt64 {
					k.bid1Price = order.Price
				} else if k.bid1Price < order.Price {
					k.bid1Price = order.Price
				}
				k.bid1PriceMux.Unlock()
			} else if k.bid1Price < order.Price {
				k.bid1PriceMux.Lock()
				if k.bid1Price < order.Price {
					k.bid1Price = order.Price
				}
				k.bid1PriceMux.Unlock()
			}
			return true
		}
	}
}

// clear a price level/bucket
func (k *kernel) clearBucket(e *Element, takerOrder types.KernelOrder, wg *sync.WaitGroup, took int64) {
	defer wg.Done()
	bucket := e.Value().(*priceBucket)
	matchingInfo := &matchedInfo{
		makerOrders:    nil,
		takerOrder:     takerOrder,
		matchedSizeMap: make(map[uint64]int64),
	}
	matchingInfo.matchedSizeMap[takerOrder.KernelOrderID] = -took
	makerOrders := make([]types.KernelOrder, 0, bucket.l.Len())
	element := bucket.l.Back()
	for v := element; v != nil; v = v.Prev() {
		matchedOrder := v.Value.(*types.KernelOrder)
		matchedOrder.FilledTotal += matchedOrder.Left * matchedOrder.Price
		matchingInfo.matchedSizeMap[matchedOrder.KernelOrderID] = matchedOrder.Left
		matchedOrder.Left = 0
		matchedOrder.Status = types.CLOSED
		matchedOrder.UpdateTime = takerOrder.UpdateTime
		makerOrders = append(makerOrders, *matchedOrder)
	}
	matchingInfo.makerOrders = makerOrders
	k.matchedInfoChan <- matchingInfo
}

// run in single thread,  需确证可撮合的订单进入
func (k *kernel) matchingOrder(side *SkipList, takerOrder *types.KernelOrder, isAsk bool) {
	wg := sync.WaitGroup{}
	removeBucketKeyList := list.New()
	// POC
	if takerOrder.TimeInForce == types.POC {
		// cancel
		takerOrder.UpdateTime = time.Now().UnixNano()
		takerOrder.Status = types.CANCELLED
		k.matchedInfoChan <- &matchedInfo{
			makerOrders:    nil,
			matchedSizeMap: nil,
			takerOrder:     *takerOrder,
		}
		return
	}
	// IOC : Immediate Or Cancel
	if takerOrder.TimeInForce == types.IOC {
		var priceMatchedLeft int64
		for skipListElement := side.Front(); skipListElement != nil; skipListElement = skipListElement.Next() {
			bucket := skipListElement.Value().(*priceBucket)
			bucketListHead := bucket.l.Front().Value.(*types.KernelOrder)
			// check price
			if (isAsk && bucketListHead.Price < takerOrder.Price) || (!isAsk && bucketListHead.Price > takerOrder.Price) {
				break
			}
			priceMatchedLeft += bucket.Left
		}
		// not enough orders left
		if (isAsk && takerOrder.Left < priceMatchedLeft) || (!isAsk && takerOrder.Left > priceMatchedLeft) {
			// cancel
			takerOrder.UpdateTime = time.Now().UnixNano()
			takerOrder.Status = types.CANCELLED
			k.matchedInfoChan <- &matchedInfo{
				makerOrders:    nil,
				matchedSizeMap: nil,
				takerOrder:     *takerOrder,
			}
			return
		}
	}

	// GTC takerOrder
	if takerOrder.TimeInForce == types.GTC {
	Loop:
		for skipListElement := side.Front(); skipListElement != nil; skipListElement = skipListElement.Next() {
			bucket := skipListElement.Value().(*priceBucket)
			bucketListHead := bucket.l.Front().Value.(*types.KernelOrder)
			// check price
			if (isAsk && bucketListHead.Price < takerOrder.Price) || (!isAsk && bucketListHead.Price > takerOrder.Price) {
				break Loop
			}
			// check if enough amount of left order in a bucket
			if (isAsk && bucket.Left <= -takerOrder.Left) || (!isAsk && bucket.Left >= -takerOrder.Left) {
				// async clear price bucket
				takerOrder.Left += bucket.Left
				takerOrder.FilledTotal -= bucket.Left * bucketListHead.Price
				if takerOrder.Left == 0 {
					takerOrder.Status = types.CLOSED
				}
				takerOrder.UpdateTime = time.Now().UnixNano()
				removeBucketKeyList.PushBack(skipListElement.key)
				wg.Add(1)
				go k.clearBucket(skipListElement, *takerOrder, &wg, bucket.Left)
			} else { // matching remaining order
				if takerOrder.Left == 0 {
					break Loop
				}
				matchingInfo := &matchedInfo{
					makerOrders:    make([]types.KernelOrder, 0, bucket.l.Len()),
					matchedSizeMap: make(map[uint64]int64),
				}
				matchingInfo.matchedSizeMap[takerOrder.KernelOrderID] = takerOrder.Left

				for listElement := bucket.l.Back(); ; /*  listElement != nil */ /* listElement = listElement.Prev() */ {
					matchedOrder := listElement.Value.(*types.KernelOrder)
					unixNano := time.Now().UnixNano()
					if (isAsk && matchedOrder.Left <= -takerOrder.Left) || (!isAsk && matchedOrder.Left >= -takerOrder.Left) {
						// clear matched maker Order
						bucket.Left -= matchedOrder.Left
						takerOrder.Left += matchedOrder.Left
						matchedOrder.FilledTotal += matchedOrder.Left * matchedOrder.Price
						takerOrder.FilledTotal -= matchedOrder.Left * matchedOrder.Price
						matchingInfo.matchedSizeMap[matchedOrder.KernelOrderID] = matchedOrder.Left
						matchedOrder.Left = 0
						matchedOrder.Status = types.CLOSED
						matchedOrder.UpdateTime = unixNano
						matchingInfo.makerOrders = append(matchingInfo.makerOrders, *matchedOrder)
						rm := listElement
						listElement = listElement.Prev()
						bucket.l.Remove(rm)
					} else {
						// 吃了完了matchedOrder还有剩余
						matchedOrder.FilledTotal -= takerOrder.Left * matchedOrder.Price
						takerOrder.FilledTotal += takerOrder.Left * matchedOrder.Price
						matchingInfo.matchedSizeMap[matchedOrder.KernelOrderID] = -takerOrder.Left
						matchedOrder.Left += takerOrder.Left
						bucket.Left += takerOrder.Left
						takerOrder.Left = 0
						matchedOrder.UpdateTime = unixNano
						matchingInfo.makerOrders = append(matchingInfo.makerOrders, *matchedOrder)
					}
					if takerOrder.Left == 0 {
						takerOrder.UpdateTime = unixNano
						takerOrder.Status = types.CLOSED
						matchingInfo.takerOrder = *takerOrder
						// send matched info.
						k.matchedInfoChan <- matchingInfo
						break Loop
					}
				}
			}
		}
		// Loop end

		// 还有剩余的不能成交, 插入卖单队列
		if takerOrder.Left != 0 {
			k.insertCheckedOrder(takerOrder)
		}
	}

	// remove cleared bucket
	for e := removeBucketKeyList.Front(); e != nil; e = e.Next() {
		side.Remove(e.Value.(float64))
	}

	// 等待异步处理完成
	wg.Wait()

	// 必须等待异步处理完成后, 更新买一/卖一价格
	if side.Length != 0 {
		bucket := side.Front().value.(*priceBucket)
		kernelOrder := bucket.l.Front().Value.(*types.KernelOrder)
		price := kernelOrder.Price
		if isAsk {
			k.bid1Price = price
		} else {
			k.ask1Price = price
		}
	} else {
		if isAsk {
			k.bid1Price = math.MinInt64
		} else {
			k.ask1Price = math.MaxInt64
		}
	}
}

func newKernel() *kernel {
	return &kernel{
		ask:             NewSkipList(),
		bid:             NewSkipList(),
		ask1Price:       math.MaxInt64,
		bid1Price:       math.MinInt64,
		matchedInfoChan: make(chan *matchedInfo),
		errorInfoChan:   make(chan *KernelErr),
		ask1PriceMux:    sync.Mutex{},
		bid1PriceMux:    sync.Mutex{},
	}
}

func restoreKernel(path string) (*kernel, bool) {
	k := &kernel{
		ask:             NewSkipList(),
		bid:             NewSkipList(),
		ask1Price:       math.MaxInt64,
		bid1Price:       math.MinInt64,
		matchedInfoChan: make(chan *matchedInfo),
		errorInfoChan:   make(chan *KernelErr),
		ask1PriceMux:    sync.Mutex{},
		bid1PriceMux:    sync.Mutex{},
	}

	_, err := os.Stat(path + "finished.log")
	if os.IsNotExist(err) {
		log.Println("check file: finished.log not found")
		return nil, false
	}

	wg := &sync.WaitGroup{}

	askDir, err := ioutil.ReadDir(path + "ask/")
	if err != nil {
		log.Println(err.Error())
		return nil, false
	}

	for i := range askDir {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			bytes, err := ioutil.ReadFile(path + "ask/" + askDir[i].Name())
			//fmt.Println(askDir[i].Name())
			if err != nil {
				log.Println(err.Error())
			}
			l := readListFromBytes(bytes)
			price := l.Front().Value.(*types.KernelOrder).Price
			var left int64
			for j := l.Front(); j != nil; j = j.Next() {
				left += j.Value.(*types.KernelOrder).Left
			}
			k.ask.Set(float64(price), &priceBucket{
				l:    l,
				Left: 0,
			})
		}()

	}

	bidDir, err := ioutil.ReadDir(path + "bid/")
	if err != nil {
		log.Println(err.Error())
		return nil, false
	}

	for i := range bidDir {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			bytes, err := ioutil.ReadFile(path + "bid/" + bidDir[i].Name())
			if err != nil {
				log.Println(err.Error())
			}
			l := readListFromBytes(bytes)
			price := l.Front().Value.(*types.KernelOrder).Price
			var left int64
			for j := l.Front(); j != nil; j = j.Next() {
				left += j.Value.(*types.KernelOrder).Left
			}
			k.bid.Set(float64(price), &priceBucket{
				l:    l,
				Left: 0,
			})
		}()
	}

	wg.Wait()

	if k.ask.Length != 0 {
		k.ask1Price = k.ask.Front().value.(*priceBucket).l.Front().Value.(*types.KernelOrder).Price
	}

	if k.bid.Length != 0 {
		k.bid1Price = k.bid.Front().value.(*priceBucket).l.Front().Value.(*types.KernelOrder).Price
	}

	return k, true
}
