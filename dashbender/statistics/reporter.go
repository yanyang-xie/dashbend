package statistics

import (
	"context"
	"dashbend/dashbender/validation"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
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
				reportData.deltaResultData = reqResult
			} else {
				reportData.totalResultData = reqResult
			}
			logrus.Infof("Request Counter: %v", reqResult)

		case validationResultData := <-r.validationResultDataChan:
			if validationResultData.IsDelta {
				reportData.deltavValidationResultData = validationResultData
			} else {
				reportData.totalValidationResultData = validationResultData
			}

			logrus.Infof("Validation counter: %v", reportData)
		case <-ctx.Done():
			return
		}
	}
}

type ReportData struct {
	//request result
	deltaResultData *ResultDataCollection
	totalResultData *ResultDataCollection

	//validation result
	deltavValidationResultData *validation.RespValidationData
	totalValidationResultData  *validation.RespValidationData
}

//@todo 通过api把report传递出去
func ReportHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Get report......")
	fmt.Fprintf(w, "this is report")
}
