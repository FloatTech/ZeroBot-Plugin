// Package minecraft MCSManager
package minecraft

import (
	"fmt"
	"io/ioutil"
	"net/http"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

// 此功能实现依赖MCSManager项目对服务器的管理api，mc服务器如果没有在该管理平台部署此功能无效
// 项目地址: https://github.com/Suwings/MCSManager
// 项目的api文档: https://github.com/Suwings/MCSManager/wiki/API-Documentation

const api = "http://your.addr:23333/api/start_server/%s/?apikey=apikey"

var engine = control.Register("minecraft", &control.Options{
	DisableOnDefault: false,
	Help: "minecraft\n" +
		"- /mcstart xxx\n" +
		"- /mcstop xxx\n" +
		"- /mclist servername\n" +
		"- https://github.com/Suwings/MCSManager",
})

func init() {
	engine.OnCommand("mcstart").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			ctx.SendChain(message.Text("开启服务器: ", model.Args, "....."))
			result := start(model.Args)
			ctx.Send(result)
		})
	engine.OnCommand("mcstop").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			ctx.SendChain(message.Text("开启服务器: ", model.Args, "....."))
			result := stop(model.Args)
			ctx.Send(result)
		})
}

// 开启服务器的api请求
func start(name string) string {
	url := fmt.Sprintf(api, name)
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
	url := fmt.Sprintf(api, name)
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
