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
	ReqResultDataCollection chan *ReqResultDataCollection

	MutexLock *sync.RWMutex

	DeltaResultData *ReqResultDataCollection
	TotalResultData *ReqResultDataCollection
}

func NewResultCollector(reqestChan chan *model.ReqestResult, reqResultDataCollection chan *ReqResultDataCollection) *ResultCollector {
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

type ReqResultDataCollection struct {
	TotalReqCount      int64
	TotalReqErrorCount int64
	TotalReqTimeCount  int64

	ReqStatusCodeCountMap *map[int]int
	ReqTimeMetricList     *[]*TimeMetric

	IsDelta bool
}

func (r *ReqResultDataCollection) String() string {
	if r.IsDelta {
		return fmt.Sprintf("Delta:      Total request: %10d, Total error: %v, Avg request time:%4d, ReqStatusCodeCountMap: %v, ReqTimeMetricList: %v", r.TotalReqCount, r.TotalReqErrorCount, r.TotalReqTimeCount/r.TotalReqCount, *r.ReqStatusCodeCountMap, *r.ReqTimeMetricList)
	} else {
		return fmt.Sprintf("Summarized: Total request: %10d, Total error: %v, Avg request time:%4d, ReqStatusCodeCountMap: %v, ReqTimeMetricList: %v", r.TotalReqCount, r.TotalReqErrorCount, r.TotalReqTimeCount/r.TotalReqCount, *r.ReqStatusCodeCountMap, *r.ReqTimeMetricList)
	}
}

func (r *ReqResultDataCollection) clone() *ReqResultDataCollection {
	dc := &ReqResultDataCollection{}
	dc.IsDelta = r.IsDelta
	dc.TotalReqCount = r.TotalReqCount
	dc.TotalReqTimeCount = r.TotalReqTimeCount
	dc.TotalReqErrorCount = r.TotalReqErrorCount

	statusCodeCountMap := make(map[int]int, 0)
	for k, v := range *r.ReqStatusCodeCountMap {
		statusCodeCountMap[k] = v
	}
	dc.ReqStatusCodeCountMap = &statusCodeCountMap

	metricsList := make([]*TimeMetric, 0)
	for _, metric := range *r.ReqTimeMetricList {
		metricsList = append(metricsList, metric)
	}
	dc.ReqTimeMetricList = &metricsList

	return dc
}

func NewResultDataCollection(isDelta bool) *ReqResultDataCollection {
	statusCodeCountMap := make(map[int]int, 0)
	timeMetricList := generateTimeMetricList(requestTimeMetrics)

	return &ReqResultDataCollection{0, 0, 0, &statusCodeCountMap, &timeMetricList, isDelta}
}

func (r *ReqResultDataCollection) count(reqResult *model.ReqestResult, lock *sync.RWMutex) {
	logrus.Debugf("Receive request result: %v", reqResult)
	lock.RLock()

	//total Count
	r.TotalReqCount += 1
	r.TotalReqTimeCount += reqResult.RequestTime

	//error Count
	if reqResult.IsError {
		r.TotalReqErrorCount += 1
	}

	//error sorter
	statusCodeCountMap := *r.ReqStatusCodeCountMap
	_, exist := statusCodeCountMap[reqResult.ResponseCode]
	if !exist {
		statusCodeCountMap[reqResult.ResponseCode] = 0
	}
	statusCodeCountMap[reqResult.ResponseCode] += 1

	//time requestTimeMetrics
	for _, timeMetric := range *r.ReqTimeMetricList {
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
