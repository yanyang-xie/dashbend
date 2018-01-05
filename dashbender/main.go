package main

import (
	"context"
	"dashbend/dashbender/cfg"
	"dashbend/dashbender/model"
	"dashbend/dashbender/producer"
	"dashbend/dashbender/sender"
	"dashbend/dashbender/statistics"
	"dashbend/dashbender/util"
	"dashbend/dashbender/validation"
	"fmt"
	"github.com/Sirupsen/logrus"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func initLogger() *os.File {
	logFileDir := cfg.LogConf.LogFileDir
	if logFileDir != "" {
		if !strings.HasPrefix(logFileDir, string(os.PathSeparator)) {
			fmt.Printf("Error create log dir: %v. Log dir must be started with '/'", logFileDir)
			os.Exit(0)
		} else {
			if !strings.HasSuffix(logFileDir, string(os.PathSeparator)) {
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

func initURLs(urlFile string) []string {
	_, err := os.Stat(urlFile)
	if err != nil {
		fmt.Printf("Failed to check url file:%v. Error: %v", urlFile, err)
		os.Exit(0)
	}

	return util.File2lines(urlFile)
}

func startReportServer() {
	logrus.Infof("Start Report Service[:%v]...", cfg.ReportConf.ListenPort)
	http.HandleFunc("/report", statistics.ReportHandler)

	listen := fmt.Sprintf(":%v", cfg.ReportConf.ListenPort)
	err := http.ListenAndServe(listen, nil)
	if err != nil {
		fmt.Printf("Error start report service. Error: %v", err)
		os.Exit(0)
	}
}

func main() {
	fmt.Printf("Start Benchmark test....")
	runtime.GOMAXPROCS(runtime.NumCPU())

	f := initLogger()
	defer f.Close()

	ctx, cancel := context.WithCancel(context.Background())

	reqChannel := make(chan *model.ReqestModel, 10000)
	reqResultChan := make(chan *model.ReqestResult, 10000)
	respValidationChan := make(chan *model.RespValidationModel, 10000)

	reqResultDataChan := make(chan *statistics.ResultDataCollection, 10000)
	validationResultDataChan := make(chan *validation.RespValidationData, 10000)

	producer := producer.NewProducer(reqChannel, initURLs(cfg.HttpRequestConf.UrlFile))
	go producer.Start(ctx)

	sender := sender.NewSender(reqChannel, reqResultChan, respValidationChan)
	go sender.Start(ctx)

	collector := statistics.NewResultCollector(reqResultChan, reqResultDataChan)
	go collector.Start(ctx)

	validator := validation.NewResultValidator(respValidationChan, validationResultDataChan)
	go validator.Start(ctx)

	reporter := statistics.NewReportor(reqResultDataChan, validationResultDataChan)
	go reporter.Start(ctx)

	startReportServer()
	cancel()
}
