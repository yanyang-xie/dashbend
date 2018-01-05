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
	reqResultChan           chan *model.ReqestResult
	reqResultDataCollection chan *ResultDataCollection

	mutexLock *sync.RWMutex

	deltaResultData *ResultDataCollection
	totalResultData *ResultDataCollection
}

func NewResultCollector(reqestChan chan *model.ReqestResult, reqResultDataCollection chan *ResultDataCollection) *ResultCollector {
	mutex := &sync.RWMutex{}

	summarizedResult := NewResultDataCollection(false)
	deltaResult := NewResultDataCollection(true)

	return &ResultCollector{reqestChan, reqResultDataCollection, mutex, deltaResult, summarizedResult}
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
		case <-ticker.C:
			//send request result data to reporter
			r.reqResultDataCollection <- r.deltaResultData.clone()
			r.reqResultDataCollection <- r.totalResultData.clone()

			r.deltaResultData = NewResultDataCollection(true)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

type ResultDataCollection struct {
	TotalCount      int64
	TotalErrorCount int64
	TotalTimeCount  int64

	StatusCodeCountMap *map[int]int
	TimeMetricList     *[]*TimeMetric

	IsDelta bool
}

func (r *ResultDataCollection) String() string {
	if r.IsDelta {
		return fmt.Sprintf("Delta:      Total request: %10d, Total error: %v, Avg request time:%4d, StatusCodeCountMap: %v, TimeMetricList: %v", r.TotalCount, r.TotalErrorCount, r.TotalTimeCount/r.TotalCount, *r.StatusCodeCountMap, *r.TimeMetricList)
	} else {
		return fmt.Sprintf("Summarized: Total request: %10d, Total error: %v, Avg request time:%4d, StatusCodeCountMap: %v, TimeMetricList: %v", r.TotalCount, r.TotalErrorCount, r.TotalTimeCount/r.TotalCount, *r.StatusCodeCountMap, *r.TimeMetricList)
	}
}

func (r *ResultDataCollection) clone() *ResultDataCollection {
	dc := &ResultDataCollection{}
	dc.IsDelta = r.IsDelta
	dc.TotalCount = r.TotalCount
	dc.TotalTimeCount = r.TotalTimeCount
	dc.TotalErrorCount = r.TotalErrorCount

	statusCodeCountMap := make(map[int]int, 0)
	for k, v := range *r.StatusCodeCountMap {
		statusCodeCountMap[k] = v
	}
	dc.StatusCodeCountMap = &statusCodeCountMap

	metricsList := make([]*TimeMetric, 0)
	for _, metric := range *r.TimeMetricList {
		metricsList = append(metricsList, metric)
	}
	dc.TimeMetricList = &metricsList

	return dc
}

func NewResultDataCollection(isDelta bool) *ResultDataCollection {
	statusCodeCountMap := make(map[int]int, 0)
	timeMetricList := generateTimeMetricList(requestTimeMetrics)

	return &ResultDataCollection{0, 0, 0, &statusCodeCountMap, &timeMetricList, isDelta}
}

func (r *ResultDataCollection) count(reqResult *model.ReqestResult, lock *sync.RWMutex) {
	logrus.Debugf("Receive request result: %v", reqResult)
	lock.RLock()

	//total count
	r.TotalCount += 1
	r.TotalTimeCount += reqResult.RequestTime

	//error count
	if reqResult.IsError {
		r.TotalErrorCount += 1
	}

	//error sorter
	statusCodeCountMap := *r.StatusCodeCountMap
	_, exist := statusCodeCountMap[reqResult.ResponseCode]
	if !exist {
		statusCodeCountMap[reqResult.ResponseCode] = 0
	}
	statusCodeCountMap[reqResult.ResponseCode] += 1

	//time requestTimeMetrics
	for _, timeMetric := range *r.TimeMetricList {
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
