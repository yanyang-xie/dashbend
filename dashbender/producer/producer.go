package producer

import (
	"context"
	"github.com/Sirupsen/logrus"
	"dashbend/dashbender/model"
	"time"
)

type Producer struct{
	reqestChan chan *model.ReqestModel
	rate   int
}

func (s *Producer) NewProducer(reqestChan chan *model.ReqestModel) *Producer{
	//@todo rate应该从配置文件中读取
	return &Producer{reqestChan:reqestChan, rate:10}
}

func (p *Producer) updateRate(rate int){
	p.rate = rate
}

func (p *Producer) Start(ctx context.Context) {
	logrus.Infof("Start Producer  ...")
	defer logrus.Infof("Producer has stopped")

	//@todo 读取配置文件读取的ticker单元. 默认值是1秒
	ticker := time.NewTicker(1 * time.Second)
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

func (p *Producer) ProduceRequests(){
	var i int
	for i=0; i<p.rate; i++{
		req := model.NewReqestModel("http://www.baidu.com", "get")
		p.reqestChan <- req
		logrus.Debugf("Put request into RequestModel channel. %v", req)
	}
}


