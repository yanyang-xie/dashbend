package producer

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"github.com/Sirupsen/logrus"
	"time"
	"math/rand"
)

type Producer struct {
	reqestChan chan *model.ReqestModel
	rate       int
	urls       []string
}

func NewProducer(reqestChan chan *model.ReqestModel, urls []string) *Producer {
	return &Producer{reqestChan: reqestChan, rate: cfg.BenchmarkConf.BRate, urls:urls}
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
		url := p.urls[rand.Intn(len(p.urls))]

		req := model.NewReqestModel(url, "GET")
		p.reqestChan <- req
		logrus.Debugf("Put request into RequestModel channel. %v", req)
	}
}
