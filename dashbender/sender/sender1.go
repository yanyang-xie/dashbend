package sender

import (
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const workers int = 5

var client *http.Client

//HTTPSender represents http load tester
type HTTPSender struct {
	URL           string
	RatePerSecond int
	Duration      int
}

type reqResult struct {
	reqTime time.Duration
	isError bool
}

//Result contains requests statistics
type Result struct {
	AvgReqTime   float64
	TotalReqSent int64
	NumOfErrors  int64
}

func init() {
	tr := &http.Transport{
		MaxIdleConns:        10000,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 10000,
	}
	client = &http.Client{Transport: tr}
}

//Start starts load testing
func (s HTTPSender) Start() Result {
	timePerRequest := time.Duration(int64(time.Second) / int64(s.RatePerSecond))
	totalProDuration := time.Duration(int64(s.Duration) * int64(time.Second))
	out := make(chan reqResult)
	in := make(chan interface{})
	result := make(chan Result)
	done := make(chan interface{})

	go countResponses(out, result, done)

	for i := 0; i < workers; i++ {
		go startWorker(s, in, out)
	}

	progStart := time.Now()
	for {
		start := time.Now()

		select {
		case in <- nil:
		default:
			go startWorker(s, in, out) //start new worker if  busy
			in <- nil
		}
		duration := time.Since(start)
		if duration < timePerRequest {
			time.Sleep(timePerRequest - duration)
		}
		progDuration := time.Since(progStart)
		if progDuration >= totalProDuration {
			close(in)
			done <- nil
			return <-result
		}
	}
}

func startWorker(s HTTPSender, in <-chan interface{}, out chan<- reqResult) {
	for range in {
		sendRequest(s, out)
	}
}

func sendRequest(s HTTPSender, ch chan<- reqResult) {
	result := reqResult{}
	result.isError = true

	start := time.Now()
	resp, err := client.Get(s.URL)
	if err != nil {
		ch <- result
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(ioutil.Discard, resp.Body)

	if err != nil {
		ch <- result
		return
	}
	result.reqTime = time.Since(start)
	result.isError = false
	ch <- result
}

func countResponses(ch chan reqResult, result chan Result, done chan interface{}) {
	var reqNum int64
	var errorsCount int64
	var totalDur time.Duration
	for {
		select {
		case response := <-ch:
			if !response.isError {
				reqNum++
				totalDur += response.reqTime
			} else {
				errorsCount++
			}
		case <-done:
			result <- Result{TotalReqSent: reqNum,
				NumOfErrors: errorsCount,
				AvgReqTime:  float64(totalDur/time.Second) / float64(reqNum)}
			break
		}
	}
}
