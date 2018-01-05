package validation

import (
	"context"
	"dashbend/dashbender/model"
	"github.com/Sirupsen/logrus"
	"sync"
	"time"
	"dashbend/dashbender/cfg"
	"fmt"
)

type ResultValidator struct {
	//receive all the response from sender
	respValidationChan   chan *model.RespValidationModel

	//store validation result
	validationResultChan chan *RespValidationData

	//only response to be validation can be stored in actualRespValidationChan. Determined by validation.percent
	actualRespValidationChan chan *model.RespValidationModel

	mutexLock *sync.RWMutex
	deltaValidationData *RespValidationData
	totalValidationData *RespValidationData

	respReceivedCount int64
	checkPoint int64
}

func NewResultValidator(respValidationChan chan *model.RespValidationModel, validationResultChan chan *RespValidationData) *ResultValidator {
	mutex := &sync.RWMutex{}
	deltaValidationData := NewRespValidationData(true)
	totalValidationData := NewRespValidationData(false)

	// using a channel of fixed channel to make the used resource of validation can be hold.
	actualRespValidationChan := make(chan *model.RespValidationModel, cfg.ValidationConf.ThreadNum)
	checkPoint := int64(float64(1)/cfg.ValidationConf.Percent)

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
		case respModel := <-r.respValidationChan:
			logrus.Debugf("Received %v", respModel)
			r.addToActualValidation(respModel)
		case <-ticker.C:
			//send validation result data to reporter.
			r.validationResultChan <- r.deltaValidationData.clone()
			r.validationResultChan <- r.totalValidationData.clone()

			//new delta validation data recorder
			r.deltaValidationData = NewRespValidationData(true)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (r *ResultValidator) addToActualValidation(respVM *model.RespValidationModel){
	r.mutexLock.RLock()
	r.respReceivedCount += 1
	r.mutexLock.RUnlock()

	if r.respReceivedCount % r.checkPoint == 0 && len(r.actualRespValidationChan) < cap(r.actualRespValidationChan){
		r.actualRespValidationChan <- respVM
	}
}

func (r *ResultValidator) startValidationWoker(ctx context.Context){
	for {
		select {
		case respModel := <-r.actualRespValidationChan:
			r.mutexLock.Lock()
			r.deltaValidationData.TotalCount += 1
			r.totalValidationData.TotalCount += 1
			r.mutexLock.Unlock()

			r.doValidation(respModel)
		case <-ctx.Done():
			return
		}
	}

}

func (r *ResultValidator) doValidation(respModel *model.RespValidationModel){
	// @todo if error happend, count the error, and put error messages into r.ErrorDetailList
	logrus.Infof("Do validation. Received:%v, checkponit: %v", r.respReceivedCount, r.checkPoint)
	logrus.Debugf("Do validation. Model:%v", respModel)
}

//record reponse validation result
type RespValidationData struct {
	TotalCount int64
	ErrorCount int64

	ErrorDetailList []string //设置一个最大长度1000,再多不存了
	IsDelta         bool
}

func NewRespValidationData(isDelta bool) *RespValidationData{
	return &RespValidationData{0,0, make([]string, 0), isDelta}
}

func (r *RespValidationData) String() string{
	if r.IsDelta{
		return fmt.Sprintf("Delta      validation: TotalCount:%10d, ErrorCount:%10d, Errors:%v", r.TotalCount, r.ErrorCount, r.ErrorDetailList)
	}else{
		return fmt.Sprintf("Summarized validation: TotalCount:%10d, ErrorCount:%10d, Errors:%v", r.TotalCount, r.ErrorCount, r.ErrorDetailList)
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
