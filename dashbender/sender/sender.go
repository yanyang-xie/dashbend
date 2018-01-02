package sender

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"github.com/Sirupsen/logrus"
	"net/http"
	"time"
	"fmt"
	"io/ioutil"
	"strings"
)

type Sender struct {
	client             *http.Client
	reqChan            chan *model.ReqestModel
	reqResultChan      chan *model.ReqestResult
	respValidationChan chan *model.RespValidationModel
}

func NewSender(reqChan chan *model.ReqestModel, reqResultChan chan *model.ReqestResult, respValidationChan chan *model.RespValidationModel) *Sender {
	tr := &http.Transport{
		MaxIdleConns:        10000,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 10000,
	}

	//&http.DefaultClient
	client = &http.Client{Transport: tr, Timeout: time.Duration(cfg.HttpRequestConf.Timeout) * time.Second}
	return &Sender{
		client,
		reqChan,
		reqResultChan,
		respValidationChan,
	}
}

func (s *Sender) Start(ctx context.Context) {
	logrus.Infof("Start sender  ...")
	defer logrus.Infof("Sender has stopped")

	for {
		select {
		case requestModel := <-s.reqChan:
			go s.send(requestModel)
		case <-ctx.Done():
			return
		}
	}


}

func (s *Sender) send(req *model.ReqestModel) {
	logrus.Debugf("Do send: %v", req)

	switch req.Method {
	case "post":
		req.Method = "POST"
	default:
		req.Method = "GET"
	}

	s.sendRequest(req)
}

func (s *Sender) sendRequest(reqModel *model.ReqestModel) {
	logrus.Debugf("Send request: %v", reqModel)
	reqResult := model.NewReqestResult()

	req, err := http.NewRequest(strings.ToUpper(reqModel.Method), reqModel.URL, strings.NewReader(reqModel.Body))
	if err != nil {
		logrus.Errorf("Failed to create request object. Err: %v", err.Error())
		reqResult.IsError = true
		reqResult.ResponseErrorMessage = fmt.Sprintf("Failed to create request object. Err: %v", err.Error())
		s.reqResultChan <- reqResult
		return
	}

	for key, value := range reqModel.Headers {
		req.Header.Add(key, value)
	}

	start := time.Now()
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		reqResult.IsError = true
		reqResult.ResponseErrorMessage = err.Error()
		reqResult.ResponseCode = resp.StatusCode

		s.reqResultChan <- reqResult
		return
	} else {
		reqResult.RequestTime = getTimeTake(start)
		if resp.StatusCode == http.StatusOK {
			//read response body
			respBodyBytes, err2 := ioutil.ReadAll(resp.Body)
			respBodyString := string(respBodyBytes)

			if err2 != nil {
				reqResult.IsError = true
				reqResult.ResponseErrorMessage = fmt.Sprintf("Failed to read respose body. Err: %v", err2)
			}
			s.reqResultChan <- reqResult

			s.respValidationChan <- model.NewRespValidationModel(resp.Request, respBodyString)
			logrus.Debugf("ResponseBody: %v", respBodyString)
		} else {
			reqResult.IsError = true
			reqResult.ResponseErrorMessage = resp.Status
			reqResult.ResponseCode = resp.StatusCode
			s.reqResultChan <- reqResult
		}
	}
}

func getTimeTake(startTime time.Time) int64 {
	return int64(time.Now().Sub(startTime) / 1000000)
}
