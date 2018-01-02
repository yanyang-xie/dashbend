package cfg

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/config"
	"os"
)

type (
	logConf struct {
		LogLevel    logrus.Level
		LogFilePath string
	}

	benchmarkConf struct {
		BRate int
	}

	httpRequestConf struct {
		UrlFile    string
		Timeout    int
		RetryCount int
		RetryDelay int
	}
)

var (
	ConfigFile string

	LogConf         = &logConf{}
	BenchmarkConf   = &benchmarkConf{}
	HttpRequestConf = &httpRequestConf{}
)

func init() {
	flag.StringVar(&ConfigFile, "configFile", "config.ini", "Configuration file for benchmark test, default is config.ini")
	flag.Parse()

	loadConfFromINI()
	updateConfFromDB()
}

func loadConfFromINI() {
	conf, err := config.NewConfig("ini", ConfigFile)
	if err != nil {
		fmt.Printf("Failed to load conf from file %v, err: %v", ConfigFile, err)
		os.Exit(0)
	}

	//init benchmark conf
	BenchmarkConf.BRate = conf.DefaultInt("benchmark::rate", 10)

	//init http request conf
	HttpRequestConf.UrlFile = conf.DefaultString("http::urlFile", "urls.txt")
	HttpRequestConf.Timeout = conf.DefaultInt("http::timeout", 6)
	HttpRequestConf.RetryCount = conf.DefaultInt("http::retryCount", 0)
	HttpRequestConf.RetryDelay = conf.DefaultInt("http::retryDelay", 1)

	//init log conf
	LogConf.LogFilePath = conf.DefaultString("logs::logPath", "benchmark.log")
	level, err := logrus.ParseLevel(conf.DefaultString("logs::logLevel", "info"))
	if err != nil {
		LogConf.LogLevel = logrus.InfoLevel
	} else {
		LogConf.LogLevel = level
	}
}

//@todo read config from DB, and then update config. 应该单独产生一个db的类, 为这里和周期性sync db config做准备
func updateConfFromDB() {

}
