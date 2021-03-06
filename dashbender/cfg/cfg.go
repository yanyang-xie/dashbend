package cfg

import (
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/config"
	"os"
	"time"
)

type (
	logConf struct {
		LogLevel    logrus.Level
		LogFileDir  string
		LogFileName string
	}

	benchmarkConf struct {
		WarmUpMinute int
		BRate        int
	}

	httpRequestConf struct {
		UrlFile    string
		Timeout    int
		RetryCount int
		RetryDelay int
	}

	reportConf struct {
		ListenPort int
	}

	validationConf struct {
		Percent   float64
		ThreadNum int
	}
)

var (
	ConfigFile string

	LogConf         = &logConf{}
	BenchmarkConf   = &benchmarkConf{}
	HttpRequestConf = &httpRequestConf{}
	ReportConf      = &reportConf{}
	ValidationConf  = &validationConf{}

	BenchmarkStartTime = time.Now()
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
	BenchmarkConf.BRate = conf.DefaultInt("benchmark::rate", 2)
	BenchmarkConf.WarmUpMinute = conf.DefaultInt("benchmark::warmup_minute", 0)

	//init http request conf
	HttpRequestConf.UrlFile = conf.DefaultString("http::url_file", "urls.txt")
	HttpRequestConf.Timeout = conf.DefaultInt("http::timeout", 6)
	HttpRequestConf.RetryCount = conf.DefaultInt("http::retry_count", 0)
	HttpRequestConf.RetryDelay = conf.DefaultInt("http::retry_delay", 1)

	//init response validation conf
	ValidationConf.Percent = conf.DefaultFloat("validation::Percent", 1.0)
	ValidationConf.ThreadNum = conf.DefaultInt("validation::thread_num", 100)

	//init log conf
	LogConf.LogFileDir = conf.DefaultString("logs::log_file_dir", "")
	LogConf.LogFileName = conf.DefaultString("logs::log_file_name", "benchmark.log")
	level, err := logrus.ParseLevel(conf.DefaultString("logs::log_level", "info"))
	if err != nil {
		LogConf.LogLevel = logrus.InfoLevel
	} else {
		LogConf.LogLevel = level
	}

	//init report
	ReportConf.ListenPort = conf.DefaultInt("report::listen_port", 9000)
}

//@todo read config from DB, and then update config. 应该单独产生一个db的类, 为这里和周期性sync db config做准备
func updateConfFromDB() {

}
