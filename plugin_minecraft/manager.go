// Package minecraft MCSManager
package minecraft

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

// 此功能实现依赖MCSManager项目对服务器的管理api，mc服务器如果没有在该管理平台部署此功能无效
// 项目地址: https://github.com/Suwings/MCSManager
// 项目的api文档: https://github.com/Suwings/MCSManager/wiki/API-Documentation

const api = "http://your.addr:23333/api/start_server/%s/?apikey=apikey"

func init() {
	engine := control.Register("minecraft", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "minecraft\n" +
			"- /mcstart xxx\n" +
			"- /mcstop xxx\n" +
			"- /mclist servername\n" +
			"- https://github.com/Suwings/MCSManager",
	})
	engine.OnCommand("mcstart").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			ctx.SendChain(message.Text("开启服务器: ", model.Args, "....."))
			result := start(model.Args)
			ctx.SendChain(message.Text(result))
		})
	engine.OnCommand("mcstop").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			ctx.SendChain(message.Text("开启服务器: ", model.Args, "....."))
			result := stop(model.Args)
			ctx.SendChain(message.Text(result))
		})

	// 这里填对应mc服务器的登录地址
	servers["ftbi"] = "115.28.186.22:25710"
	servers["ges"] = "115.28.186.22:25701"

	engine.OnCommand("mclist").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			// 支持多个服务器
			gesjson := infoapi(servers[model.Args])
			var str = gesjson.Players.List
			cs := strings.Join(str, "\n")
			ctx.SendChain(message.Text(
				"服务器名字: ", gesjson.Motd.Raw[0], "\n",
				"在线人数: ", gesjson.Players.Online, "/", gesjson.Players.Max, "\n",
				"以下为玩家名字: ", "\n", cs,
			))
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
