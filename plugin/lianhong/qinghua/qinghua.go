// Package qinghua 文字返回获取集合
package qinghua

import (
	"fmt"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("qinghua", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "全部语言指令\n" +
			"- 每日情话\n" +
			"- 每日鸡汤\n" +
			"- 绕口令\n",
	})
	engine.OnFullMatch("每日情话").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://xiaobai.klizi.cn/API/other/wtqh.php")
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			km := fmt.Sprintf("%s", data)
			ctx.SendChain(message.Text(km))
		})
	engine.OnFullMatch("每日鸡汤").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://api.btstu.cn/yan/api.php?charset=utf-8&encode=text")
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			km := string(data)
			ctx.SendChain(message.Text(km))
		})
	engine.OnFullMatch("绕口令").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://ovooa.com/API/rao/api.php?type=text")
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			km := string(data)
			ctx.SendChain(message.Text(km))
		})
}
