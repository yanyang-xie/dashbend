package producer

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"github.com/Sirupsen/logrus"
	"time"
)

type Producer struct {
	reqestChan chan *model.ReqestModel
	rate       int
}

func NewProducer(reqestChan chan *model.ReqestModel) *Producer {
	return &Producer{reqestChan: reqestChan, rate: cfg.BenchmarkConf.BRate}
}

func (p *Producer) updateRate(rate int) {
	p.rate = rate
}

func (p *Producer) Start(ctx context.Context) {
	logrus.Infof("Start Producer  ...")
	defer logrus.Infof("Producer has stopped")

	//do warm up
	p.ProduceWarmUpRequests()

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			p.ProduceRequests()
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

//todo: warm up process
func (p *Producer) ProduceWarmUpRequests() {
	logrus.Infof("Start to generate warm up request...")
}

func (p *Producer) ProduceRequests() {
	var i int
	for i = 0; i < p.rate; i++ {
		req := model.NewReqestModel("http://www.baidu.com", "GET")
		p.reqestChan <- req
		logrus.Debugf("Put request into RequestModel channel. %v", req)
	}
}
