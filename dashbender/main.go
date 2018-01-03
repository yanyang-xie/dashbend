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
	"strings"
)

func initLogger() *os.File {
	logFileDir := cfg.LogConf.LogFileDir
	if logFileDir != ""{
		if !strings.HasPrefix(logFileDir, string(os.PathSeparator)){
			fmt.Printf("Error create log dir: %v. Log dir must be started with '/'", logFileDir)
			os.Exit(0)
		}else{
			if !strings.HasSuffix(logFileDir, string(os.PathSeparator)){
				logFileDir += string(os.PathSeparator)
			}
			os.MkdirAll(logFileDir, 0777)
		}
	}

	logFile := logFileDir + cfg.LogConf.LogFileName
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("Error opening log file: %v, Error: %v", logFile, err)
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
