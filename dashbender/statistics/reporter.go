package statistics

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/validation"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"time"
)

var reportData *ReportData

type Reporter struct {
	reqResultDataChan        chan *ResultDataCollection
	validationResultDataChan chan *validation.RespValidationData
}

func NewReportor(reqResultChan chan *ResultDataCollection, validationResultChan chan *validation.RespValidationData) *Reporter {
	return &Reporter{reqResultChan, validationResultChan}
}

func (r *Reporter) Start(ctx context.Context) {
	logrus.Infof("Start Result Reporter  ...")
	defer logrus.Infof("Result Reporter has stopped")

	reportData = &ReportData{}
	for {
		select {
		case reqResult := <-r.reqResultDataChan:
			if reqResult.IsDelta {
				reportData.DeltaResultData = reqResult
				reportData.DeltaTimeTaken = int64(60)
			} else {
				reportData.TotalResultData = reqResult
				reportData.TotalTimeTaken = getTimeTake(cfg.BenchmarkStartTime)
			}
			logrus.Infof("Request Counter: %v", reqResult)
		case validationResultData := <-r.validationResultDataChan:
			if validationResultData.IsDelta {
				reportData.DeltaValidationResultData = validationResultData
			} else {
				reportData.TotalValidationResultData = validationResultData
			}
			logrus.Infof("Validation counter: %v", validationResultData)
		case <-ctx.Done():
			return
		}
	}
}

type ReportData struct {
	//time taken from benchmart start (seconds)
	DeltaTimeTaken int64
	TotalTimeTaken int64

	//request result
	DeltaResultData *ResultDataCollection
	TotalResultData *ResultDataCollection

	//validation result
	DeltaValidationResultData *validation.RespValidationData
	TotalValidationResultData *validation.RespValidationData
}

func (r *ReportData) String() string {
	return fmt.Sprintf("%v, %v, %v, %v", r.TotalResultData, r.TotalValidationResultData, r.DeltaResultData, r.DeltaValidationResultData)
}

func ReportHandler(w http.ResponseWriter, r *http.Request) {
	jsonReportByte, err := json.Marshal(reportData)
	if err != nil {
		logrus.Errorf("Fetch report failed. Error: %v", err.Error())
		http.Error(w, "Fetch report failed.", http.StatusInternalServerError)
	} else {
		jsonReport := string(jsonReportByte)
		logrus.Debugf("Fetch report: %v", jsonReport)
		fmt.Fprintf(w, "%v", jsonReport)
	}

}

func getTimeTake(startTime time.Time) int64 {
	return int64(time.Now().Sub(startTime) / time.Second)
}
