package validation

import (
	"context"
	"dashbend/dashbender/model"
	"github.com/Sirupsen/logrus"
	"sync"
	"time"
)

//todo VEX Response validation
// 也要和collector一样, summarized的

//最多100个线程去做validation. 抽查的百分比应该是在collector处进行
type ResultValidator struct {
	respValidationChan   chan *model.RespValidationModel
	validationResultChan chan *RespValidationData

	mutexLock *sync.RWMutex

	deltaValidationData *RespValidationData
	totalValidationData *RespValidationData
}

//record reponse validation result
type RespValidationData struct {
	TotalCount int64
	ErrorCount int64

	ErrorDetailList []string //设置一个最大长度1000,再多不存了
	IsDelta         bool
}

func NewResultValidator(respValidationChan chan *model.RespValidationModel, validationResultChan chan *RespValidationData) *ResultValidator {
	mutex := &sync.RWMutex{}
	deltaValidationData := &RespValidationData{}
	totalValidationData := &RespValidationData{}

	return &ResultValidator{respValidationChan, validationResultChan, mutex, deltaValidationData, totalValidationData}
}

func (r *ResultValidator) Start(ctx context.Context) {
	logrus.Infof("Start Response Validator  ...")
	defer logrus.Infof("Response Validator has stopped")

	ticker := time.NewTicker(1 * time.Minute)
	for {
		select {
		case respModel := <-r.respValidationChan:
			logrus.Debugf("Received %v", respModel)
			//do validation and then count
			//这里只能有固定的个线程去做validation, 不能影响主线程
		case <-ticker.C:
			//send validation result data to reporter

			/**
			r.reqResultDataCollection <- r.totalResultData
			r.reqResultDataCollection <- r.deltaResultData

			logrus.Infof("Counter:%v", r.deltaResultData)
			logrus.Infof("Counter:%v", r.totalResultData)
			logrus.Debugf("Clear delta statistics data.")
			r.deltaResultData = NewResultDataCollection(true)
			*/
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

//这里应该是抽查, 而不是全部都检查. 从配置文件中读取抽查的百分比. 抽查的百分比应该是在collector处进行

//这里和reporter哪里都是定期把统计数据传递给reporterChannel
