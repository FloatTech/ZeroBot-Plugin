package web

import (
	"io"
	"net/http"
)

func ReqWith(url string, method string, referer string, ua string) (data []byte, err error) {
	client := &http.Client{}
	// 提交请求
	var reqest *http.Request
	reqest, err = http.NewRequest(method, url, nil)
	if err == nil {
		// 增加header选项
		reqest.Header.Add("Referer", referer)
		reqest.Header.Add("User-Agent", ua)
		var response *http.Response
		response, err = client.Do(reqest)
		if err == nil {
			data, err = io.ReadAll(response.Body)
			response.Body.Close()
		}
	}
	return
}
