// Package minecraft MCSManager
package minecraft

import (
	"fmt"
	"io/ioutil"
	"net/http"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 此功能实现依赖MCSManager项目对服务器的管理api，mc服务器如果没有在该管理平台部署此功能无效
// 项目地址: https://github.com/Suwings/MCSManager
// 项目的api文档: https://github.com/Suwings/MCSManager/wiki/API-Documentation

func init() {
	zero.OnRegex(`^/start (.*)$`).
		Handle(func(ctx *zero.Ctx) {
			name := ctx.State["regex_matched"].([]string)[1]
			ctx.SendChain(message.Text("开启服务器: ", name, "....."))
			result := start(name)
			ctx.Send(result)
		})
}

func init() {
	zero.OnRegex(`^/stop (.*)$`).
		Handle(func(ctx *zero.Ctx) {
			name := ctx.State["regex_matched"].([]string)[1]
			ctx.SendChain(message.Text("关闭服务器: ", name, "....."))
			result := stop(name)
			ctx.Send(result)
		})
}

// 开启服务器的api请求
func start(name string) string {
	url := fmt.Sprintf("http://your.addr:23333/api/start_server/%s/?apikey=apikey", name)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	return string(body)
}

// 关闭服务器的api请求
func stop(name string) string {
	url := fmt.Sprintf("http://your.addr:23333/api/stop_server/%s/?apikey=apikey", name)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}
	return string(body)
}
