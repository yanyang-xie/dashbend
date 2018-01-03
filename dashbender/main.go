package main

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"dashbend/dashbender/producer"
	"dashbend/dashbender/sender"
	"fmt"
	"github.com/Sirupsen/logrus"
	"os"
	"runtime"
	"dashbend/dashbender/statistics"
	"net/http"
)

func initLogger() *os.File {
	//@todo logfile的文件夹不存在的时候怎么处理?

	f, err := os.OpenFile(cfg.LogConf.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(0)
	}

	logrus.SetOutput(f)
	//logrus.SetOutput(os.Stdout)
	logrus.SetLevel(cfg.LogConf.LogLevel)
	formatter := new(logrus.TextFormatter)
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)

	return f
}

func startReportServer(){
	logrus.Infof("Start Report Service...")
	http.HandleFunc("/report", statistics.ReportHandler)
	http.ListenAndServe(":9000", nil)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	f := initLogger()
	defer f.Close()

	ctx, cancel := context.WithCancel(context.Background())

	reqChannel := make(chan *model.ReqestModel, 10000)
	reqResultChan := make(chan *model.ReqestResult, 10000)
	respValidationChan := make(chan *model.RespValidationModel, 10000)

	producer := producer.NewProducer(reqChannel)
	go producer.Start(ctx)

	sender := sender.NewSender(reqChannel, reqResultChan, respValidationChan)
	go sender.Start(ctx)

	collector := statistics.NewResultCollector(reqResultChan)
	go collector.Start(ctx)

	startReportServer()
	cancel()
}
