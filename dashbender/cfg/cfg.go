package cfg

import (
	"github.com/astaxie/beego/config"
	"fmt"
	"flag"
	"os"
	"github.com/Sirupsen/logrus"
)

var (
	ConfigFile = "config.ini"
)

type (
	logConf struct{
		LogLevel logrus.Level
		LogFilePath string
	}

	benchmarkConf struct{
		BRate int
	}

	httpRequestConf struct{
		UrlFile string
		Timeout int
		RetryCount int
		RetryDelay int
	}
)

var (
	LogConf = &logConf{}
	BenchmarkConf = &benchmarkConf{}
	HttpRequestConf = &httpRequestConf{}
)

func init(){
	fmt.Printf(">>>Start to init CFG\n")

	flag.String(ConfigFile, "config.ini", "Configuration file for benchmark test, default is config.ini")
	flag.Parse()

	//http://blog.csdn.net/u013210620/article/details/78574930
	conf, err := config.NewConfig("ini", ConfigFile)
	if err != nil {
		fmt.Printf("Failed to load conf from file %v, err: %v", ConfigFile, err)
		os.Exit(0)
	}

	//init
	BenchmarkConf.BRate = conf.DefaultInt("benchmark::rate", 10)

	HttpRequestConf.UrlFile = conf.DefaultString("http::urlFile", "urls.txt")
	HttpRequestConf.Timeout = conf.DefaultInt("http::timeout", 6)
	HttpRequestConf.RetryCount = conf.DefaultInt("http::retryCount", 0)
	HttpRequestConf.RetryDelay = conf.DefaultInt("http::retryDelay", 1)

	LogConf.LogFilePath = conf.DefaultString("logs::logPath", "benchmark.log")
	level, err := logrus.ParseLevel(conf.DefaultString("logs::logLevel", "info"))
	if err != nil{
		LogConf.LogLevel = logrus.InfoLevel
	}else{
		LogConf.LogLevel = level
	}
}