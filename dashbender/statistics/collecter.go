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
	Min   int64
	Max   int64
	Count int64

	lock *sync.RWMutex
}

func (t *TimeMetric) String() string {
	return fmt.Sprintf("%v-%v:%v", t.Min, t.Max, t.Count)
}

func (t *TimeMetric) increment(cTime int64) bool {
	if t.Min < cTime && cTime <= t.Max {
		t.lock.RLock()
		t.Count += 1
		t.lock.RUnlock()
		return true
	} else {
		return false
	}
}

type ResultCollector struct {
	ReqResultChan           chan *model.ReqestResult
	ReqResultDataCollection chan *ResultDataCollection

	MutexLock *sync.RWMutex

	DeltaResultData *ResultDataCollection
	TotalResultData *ResultDataCollection
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
		case reqResult := <-r.ReqResultChan:
			r.DeltaResultData.count(reqResult, r.MutexLock)
			r.TotalResultData.count(reqResult, r.MutexLock)
		case <-ticker.C:
			//send request result data to reporter
			r.ReqResultDataCollection <- r.DeltaResultData.clone()
			r.ReqResultDataCollection <- r.TotalResultData.clone()

			r.DeltaResultData = NewResultDataCollection(true)
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

	//total Count
	r.TotalCount += 1
	r.TotalTimeCount += reqResult.RequestTime

	//error Count
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
		timeMetric.Min = metrics[i-1]
		timeMetric.Max = m
		timeMetric.lock = &sync.RWMutex{}

		metricsList = append(metricsList, timeMetric)
	}
	return metricsList
}
