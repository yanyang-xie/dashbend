package producer

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"github.com/Sirupsen/logrus"
	"math/rand"
	"time"
)

type Producer struct {
	reqestChan chan *model.ReqestModel
	urls       []string

	rate         int
	warmupMinute int
}

func NewProducer(reqestChan chan *model.ReqestModel, urls []string) *Producer {
	return &Producer{reqestChan: reqestChan, urls: urls, rate: cfg.BenchmarkConf.BRate, warmupMinute: cfg.BenchmarkConf.WarmUpMinute}
}

func (p *Producer) updateRate(rate int) {
	p.rate = rate
}

func (p *Producer) Start(ctx context.Context) {
	logrus.Infof("Start Producer  ...")
	defer logrus.Infof("Producer has stopped")

	//do warm up
	p.ProduceWarmUpRequests()

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			p.ProduceRequests(p.rate)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (p *Producer) ProduceWarmUpRequests() {
	warmupSteps := generateWarmUpSteps(p.rate, p.warmupMinute)
	if warmupSteps == nil {
		return
	}

	logrus.Infof("Start to generate warm up request. Warmup: %v minutes, Rate: %v", p.warmupMinute, warmupSteps)
	for _, step := range warmupSteps {
		sum := 0
		for sum < 60 {
			p.ProduceRequests(step)
			time.Sleep(1 * time.Second)
			sum += 1
		}

	}
}

func (p *Producer) ProduceRequests(n int) {
	var i int
	for i = 0; i < n; i++ {
		url := p.urls[rand.Intn(len(p.urls))]

		req := model.NewReqestModel(url, "GET")
		p.reqestChan <- req
		logrus.Debugf("Put request into RequestModel channel. %v", req)
	}
}

func generateWarmUpSteps(rate, warmupMinute int) []int {
	if warmupMinute <= 0 {
		return nil
	}

	warmupList := make([]int, 0)

	step := rate / warmupMinute
	if rate%warmupMinute != 0 {
		step += 1
	}

	var i int
	for i = 0; i < warmupMinute; i++ {
		n := ((i + 1) * step)
		if n <= rate {
			warmupList = append(warmupList, n)
		} else {
			warmupList = append(warmupList, rate)
		}
	}

	return warmupList
}
