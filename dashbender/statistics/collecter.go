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

	deltaResultData *ResultDataCollection
	totalResultData *ResultDataCollection
}

func NewResultCollector(reqestChan chan *model.ReqestResult) *ResultCollector {
	mutex := &sync.RWMutex{}

	summarizedResult := NewResultDataCollection(false)
	deltaResult := NewResultDataCollection(false)

	return &ResultCollector{reqestChan, mutex, deltaResult, summarizedResult}
}

func (r *ResultCollector) Start(ctx context.Context) {
	logrus.Infof("Start Result Collector  ...")
	defer logrus.Infof("Result Collector has stopped")

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case reqResult := <-r.reqResultChan:
			r.deltaResultData.count(reqResult, r.mutexLock)
			r.totalResultData.count(reqResult, r.mutexLock)
		case <- ticker.C:
			logrus.Infof("Counter:%v", r.deltaResultData)
			logrus.Infof("Counter:%v", r.totalResultData)

			//@todo 把统计数据传递给reporter模块
			logrus.Debugf("Clear delta statistics data.")
			r.deltaResultData = NewResultDataCollection(true)

		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

type ResultDataCollection struct {
	totalCount      int64
	totalErrorCount int64
	totalTimeCount  int64

	statusCodeCountMap *map[int]int
	timeMetricList     *[]*TimeMetric

	isDelta bool
}

func (r *ResultDataCollection) String() string {
	if r.isDelta{
		return fmt.Sprintf("Delta:      Total request: %v, Total error: %v, Avg request time:%v, statusCodeCountMap: %v, timeMetricList: %v", r.totalCount, r.totalErrorCount, r.totalTimeCount/r.totalCount, *r.statusCodeCountMap, *r.timeMetricList)
	}else{
		return fmt.Sprintf("Summarized: Total request: %v, Total error: %v, Avg request time:%v, statusCodeCountMap: %v, timeMetricList: %v", r.totalCount, r.totalErrorCount, r.totalTimeCount/r.totalCount, *r.statusCodeCountMap, *r.timeMetricList)
	}
}

func (r *ResultDataCollection) clone() *ResultDataCollection{
	dc :=  &ResultDataCollection{}
	dc.isDelta = r.isDelta
	dc.totalCount = r.totalCount
	dc.totalTimeCount = r.totalTimeCount
	dc.totalErrorCount = r.totalErrorCount

	statusCodeCountMap := make(map[int]int, 0)
	for k,v := range *r.statusCodeCountMap {
		statusCodeCountMap[k] = v
	}
	dc.statusCodeCountMap = &statusCodeCountMap

	metricsList := make([]*TimeMetric, 0)
	for _, metric := range *r.timeMetricList{
		metricsList = append(metricsList, metric)
	}
	dc.timeMetricList = &metricsList

	return dc
}

func NewResultDataCollection(isDelta bool) *ResultDataCollection{
	statusCodeCountMap := make(map[int]int, 0)
	timeMetricList := generateTimeMetricList(requestTimeMetrics)

	return &ResultDataCollection{0, 0, 0, &statusCodeCountMap, &timeMetricList, isDelta}
}

func (r *ResultDataCollection) count(reqResult *model.ReqestResult, lock *sync.RWMutex) {
	logrus.Debugf("Receive request result: %v", reqResult)
	lock.RLock()

	//total count
	r.totalCount += 1
	r.totalTimeCount += reqResult.RequestTime

	//error count
	if reqResult.IsError {
		r.totalErrorCount += 1
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

	lock.RUnlock()
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