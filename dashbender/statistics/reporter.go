package statistics

import (
	"net/http"
	"fmt"
	"github.com/Sirupsen/logrus"
)

//@todo 定义两个json


//@todo 通过api把report传递出去
func ReportHandler(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Get report......")
	fmt.Fprintf(w, "this is report")
}