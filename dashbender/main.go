package main

import (
	"os"
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"dashbend/dashbender/cfg"
	"runtime"
	"dashbend/dashbender/model"
	"net/http"
	"dashbend/dashbender/sender"
)

func initLogger() *os.File{
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
	formatter.TimestampFormat = "1981-12-15 05:24:00"
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)

	logrus.Debugf("this is first log")

	return f
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	f := initLogger()
	defer f.Close()

	ctx, cancel := context.WithCancel(context.Background())

	reqChannel := make(chan *model.ReqestModel, 0)
	reqResultChan := make(chan *model.ReqestResult, 0)
	respValidationChan := make(chan *http.Response, 0)

	sender := sender.NewSender(reqChannel,reqResultChan,respValidationChan)
	sender.Start(ctx)

	cancel()
}