package web

import (
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
)

// IsSupportIPv6 检查本机是否支持 ipv6
var IsSupportIPv6 = func() bool {
	resp, err := http.Get("http://v6.ipv6-test.com/json/widgetdata.php?callback=?")
	if err != nil {
		logrus.Infoln("[web] 本机不支持ipv6")
		return false
	}
	_, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	logrus.Infoln("[web] 本机支持ipv6")
	return true
}()
