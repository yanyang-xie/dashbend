package sender

import (
	"context"
	"time"
	"github.com/Sirupsen/logrus"
	"net/http"
	"dashbend/dashbender/model"
)

type Sender struct{
	reqChan chan *model.ReqestModel

	reqResultChan chan *model.ReqestResult
	respValidationChan chan *http.Response
}

func NewSender(reqChan chan *model.ReqestModel, reqResultChan chan *model.ReqestResult, respValidationChan chan *http.Response) *Sender{
	return &Sender{
		reqChan,
		reqResultChan,
		respValidationChan,
	}
}

func (s *Sender) Start(ctx context.Context){
	logrus.Infof("Start sender  ...")
	for {
		select {
		case requestModel := <-s.reqChan:
			go s.send(requestModel)
		case <-ctx.Done():
			return
		}
	}

	defer logrus.Infof("Sender has stopped")
}

func (s *Sender) send(req *model.ReqestModel){
	logrus.Debugf("Do send: %v", req)

	switch req.Method {
	case "post":
		s.post(req)
	default :
		s.get(req)
	}
}

func (s *Sender) get(req *model.ReqestModel){
	result := model.NewReqestResult()

	start := time.Now()
	resp, err := client.Get(req.URL)

	r := model.NewReqestResult()
	if err != nil {
		r.IsError = true
		r.ResponseErrorMessage = err.Error()
		r.ResponseCode = resp.StatusCode

		s.reqResultChan <- result
		return
	}else{
		r.RequestTime = getTimeTake(start)
		s.reqResultChan <- result
	}

	defer resp.Body.Close()
	s.respValidationChan <- resp
}

func (s *Sender) post(req *model.ReqestModel){

}


/**
//http://blog.csdn.net/shengzhu1/article/details/67633075
func fetch(url string, ch chan<- string) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}
	nbytes, err := io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close() // don't leak resources
	if err != nil {
		ch <- fmt.Sprintf("while reading %s: %v", url, err)
		return
	}
	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("%.2fs  %7d  %s", secs, nbytes, url)
}
*/

func getTimeTake(startTime time.Time) int64 {
	return int64(time.Now().Sub(startTime) / 1000000)
}

