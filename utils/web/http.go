// Package web 网络处理相关
package web

import (
	"errors"
	"io"
	"net/http"
)

// ReqWith 使用自定义请求头获取数据
func ReqWith(url string, method string, referer string, ua string) (data []byte, err error) {
	client := &http.Client{}
	// 提交请求
	var request *http.Request
	request, err = http.NewRequest(method, url, nil)
	if err == nil {
		// 增加header选项
		request.Header.Add("Referer", referer)
		request.Header.Add("User-Agent", ua)
		var response *http.Response
		response, err = client.Do(request)
		if err == nil {
			data, err = io.ReadAll(response.Body)
			response.Body.Close()
		}
	}
	return
}

// GetData 获取数据
func GetData(url string) (data []byte, err error) {
	var response *http.Response
	response, err = http.Get(url)
	if err == nil {
		if response.ContentLength <= 0 {
			err = errors.New("web.GetData: empty body")
			response.Body.Close()
			return
		}
		data, err = io.ReadAll(response.Body)
		response.Body.Close()
	}
	return
}
