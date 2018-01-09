package validation

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"fmt"
	"github.com/Sirupsen/logrus"
	"sync"
	"time"
)

type ResultValidator struct {
	//receive all the response from sender
	RespValidationChan chan *model.RespValidationModel

	//store validation result
	ValidationResultChan chan *RespValidationData

	//only response to be validation can be stored in ActualRespValidationChan. Determined by validation.percent
	ActualRespValidationChan chan *model.RespValidationModel

	MutexLock           *sync.RWMutex
	DeltaValidationData *RespValidationData
	TotalValidationData *RespValidationData

	RespReceivedCount int64
	CheckPoint        int64
}

func NewResultValidator(respValidationChan chan *model.RespValidationModel, validationResultChan chan *RespValidationData) *ResultValidator {
	mutex := &sync.RWMutex{}
	deltaValidationData := NewRespValidationData(true)
	totalValidationData := NewRespValidationData(false)

	// using a channel of fixed channel to make the used resource of validation can be hold.
	actualRespValidationChan := make(chan *model.RespValidationModel, cfg.ValidationConf.ThreadNum)
	checkPoint := int64(float64(1) / cfg.ValidationConf.Percent)

	return &ResultValidator{respValidationChan, validationResultChan, actualRespValidationChan, mutex, deltaValidationData, totalValidationData, 0, checkPoint}
}

func (r *ResultValidator) Start(ctx context.Context) {
	logrus.Infof("Start Response Validator  ...")
	defer logrus.Infof("Response Validator has stopped")

	for i := 0; i < cfg.ValidationConf.ThreadNum; i++ {
		go r.startValidationWoker(ctx)
	}

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case respModel := <-r.RespValidationChan:
			logrus.Debugf("Received %v", respModel)
			r.addToActualValidation(respModel)
		case <-ticker.C:
			//send validation result data to reporter.
			r.ValidationResultChan <- r.DeltaValidationData.clone()
			r.ValidationResultChan <- r.TotalValidationData.clone()

			//new delta validation data recorder
			r.DeltaValidationData = NewRespValidationData(true)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (r *ResultValidator) addToActualValidation(respVM *model.RespValidationModel) {
	r.MutexLock.RLock()
	r.RespReceivedCount += 1
	r.MutexLock.RUnlock()

	if r.RespReceivedCount%r.CheckPoint == 0 && len(r.ActualRespValidationChan) < cap(r.ActualRespValidationChan) {
		r.ActualRespValidationChan <- respVM
	}
}

func (r *ResultValidator) startValidationWoker(ctx context.Context) {
	for {
		select {
		case respModel := <-r.ActualRespValidationChan:
			r.MutexLock.Lock()
			r.DeltaValidationData.TotalCount += 1
			r.TotalValidationData.TotalCount += 1
			r.MutexLock.Unlock()

			r.doValidation(respModel)
		case <-ctx.Done():
			return
		}
	}

}

func (r *ResultValidator) doValidation(respModel *model.RespValidationModel) {
	// @todo if error happend, count the error, and put error messages into r.ErrorDetailList
	logrus.Infof("Do validation. Received:%v, checkponit: %v", r.RespReceivedCount, r.CheckPoint)
	logrus.Debugf("Do validation. Model:%v", respModel)
}

//record reponse validation result
type RespValidationData struct {
	TotalCount int64
	ErrorCount int64

	ErrorDetailList []string //设置一个最大长度1000,再多不存了
	IsDelta         bool
}

func NewRespValidationData(isDelta bool) *RespValidationData {
	return &RespValidationData{0, 0, make([]string, 0), isDelta}
}

func (r *RespValidationData) String() string {
	if r.IsDelta {
		return fmt.Sprintf("Delta      validation: TotalReqCount:%10d, ErrorCount:%10d, Errors:%v", r.TotalCount, r.ErrorCount, r.ErrorDetailList)
	} else {
		return fmt.Sprintf("Summarized validation: TotalReqCount:%10d, ErrorCount:%10d, Errors:%v", r.TotalCount, r.ErrorCount, r.ErrorDetailList)
	}
}

func (r *RespValidationData) clone() *RespValidationData {
	vd := &RespValidationData{}
	vd.IsDelta = r.IsDelta
	vd.TotalCount = r.TotalCount
	vd.ErrorCount = r.ErrorCount

	errorDetailList := make([]string, 0)
	for _, detail := range r.ErrorDetailList {
		errorDetailList = append(errorDetailList, detail)
	}
	vd.ErrorDetailList = errorDetailList

	return vd
}
