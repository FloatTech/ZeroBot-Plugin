package bilibili

import (
	"net/http"
	"strconv"
)

// humanNum 格式化人数
func humanNum(res int) string {
	if res/10000 != 0 {
		return strconv.FormatFloat(float64(res)/10000, 'f', 2, 64) + "万"
	}
	return strconv.Itoa(res)
}

// getrealurl 获取跳转后的链接
func getrealurl(url string) (realurl string, err error) {
	data, err := http.Head(url)
	if err != nil {
		return
	}
	_ = data.Body.Close()
	realurl = data.Request.URL.String()
	return
}
