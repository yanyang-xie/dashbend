package statistics

import (
	"context"
	"dashbend/dashbender/model"
	"fmt"
	"github.com/Sirupsen/logrus"
	"sync"
	"time"
)

//request time requestTimeMetrics
var requestTimeMetrics = []int64{0, 200, 500, 1000, 2000, 3000, 6000, 12000}

type TimeMetric struct {
	min   int64
	max   int64
	count int64

	lock *sync.RWMutex
}

func (t *TimeMetric) String() string {
	return fmt.Sprintf("%v-%v:%v", t.min, t.max, t.count)
}

func (t *TimeMetric) increment(cTime int64) bool {
	if t.min < cTime && cTime <= t.max {
		t.lock.RLock()
		t.count += 1
		t.lock.RUnlock()
		return true
	} else {
		return false
	}
}

type ResultCollector struct {
	reqResultChan chan *model.ReqestResult
	mutexLock     *sync.RWMutex

	totalCount     int64
	errorCount     int64
	totalTimeCount int64

	statusCodeCountMap *map[int]int
	timeMetricList     *[]*TimeMetric
}

func NewResultCollector(reqestChan chan *model.ReqestResult) *ResultCollector {
	mutex := &sync.RWMutex{}
	statusCodeCountMap := make(map[int]int, 0)
	timeMetricList := generateTimeMetricList(requestTimeMetrics)

	return &ResultCollector{reqestChan, mutex, 0, 0, 0, &statusCodeCountMap, &timeMetricList}
}

func (r *ResultCollector) String() string {
	return fmt.Sprintf("Total request: %v, Total error: %v, Avg request time:%v, statusCodeCountMap: %v, timeMetricList: %v", r.totalCount, r.errorCount, r.totalTimeCount/r.totalCount, *r.statusCodeCountMap, *r.timeMetricList)
}

func (r *ResultCollector) count(reqResult *model.ReqestResult) {
	logrus.Debugf("Receive request result: %v", reqResult)
	r.mutexLock.RLock()

	//total count
	r.totalCount += 1
	r.totalTimeCount += reqResult.RequestTime

	//error count
	if reqResult.IsError {
		r.errorCount += 1
	}

	//error sorter
	statusCodeCountMap := *r.statusCodeCountMap
	_, exist := statusCodeCountMap[reqResult.ResponseCode]
	if !exist {
		statusCodeCountMap[reqResult.ResponseCode] = 0
	}
	statusCodeCountMap[reqResult.ResponseCode] += 1

	//time requestTimeMetrics
	for _, timeMetric := range *r.timeMetricList {
		if timeMetric.increment(reqResult.RequestTime) {
			break
		}
	}

	r.mutexLock.RUnlock()
	//logrus.Debugf("Current counter: %v", r)
}

func (r *ResultCollector) Start(ctx context.Context) {
	logrus.Infof("Start Result Collector  ...")
	defer logrus.Infof("Result Collector has stopped")

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case reqResult := <-r.reqResultChan:
			go r.count(reqResult)
		case <- ticker.C:
			//export counter
			logrus.Infof("Counter:%v", r)
			//@todo 把统计数据传递给reporter模块
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func generateTimeMetricList(metrics []int64) []*TimeMetric {
	metricsList := make([]*TimeMetric, 0)
	for i, m := range metrics {
		if i == 0 {
			continue
		}

		timeMetric := &TimeMetric{}
		timeMetric.min = metrics[i-1]
		timeMetric.max = m
		timeMetric.lock = &sync.RWMutex{}

		metricsList = append(metricsList, timeMetric)
	}
	return metricsList
}