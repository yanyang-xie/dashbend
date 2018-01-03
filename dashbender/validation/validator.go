package validation

import "sync"

//todo VEX Response validation

//record reponse validation result
type RespValidationResult struct {
	mutexLock     *sync.RWMutex

	totalCount int64
	errorCount int64

	errorDetailList []string //设置一个最大长度1000,再多不存了

}

//这里应该是抽查, 而不是全部都检查. 从配置文件中读取抽查的百分比

//这里和reporter哪里都是定期把统计数据传递给reporterChannel