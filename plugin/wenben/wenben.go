// Package wenben 文本链接
package wenben

import (
	"encoding/json"
	"fmt"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	tianqi = "https://xiaobai.klizi.cn/API/other/weather_1.php?data=&msg=%v" // api地址
	pinyin = "http://ovooa.com/API/pinyin/api.php?type=text&msg=%v"
	url    = "https://v1.hitokoto.cn/?c=a&c=b&c=c&c=d&c=h&c=i" // 动漫 漫画 游戏 文学 影视 诗词
)

type RspData struct {
	Id         int    `json:"ID"`
	Uuid       string `json:"UUID"`
	Hitokoto   string `json:"hitokoto"`
	Type       string `json:"type"`
	From       string `json:"from"`
	FromWho    string `json:"from_who"`
	Creator    string `json:"creator"`
	CreatorUid int    `json:"creator_UID"`
	Reviewer   int    `json:"reviewer"`
	CommitFrom string `json:"commit_from"`
	CreatedAt  string `json:"created_at"`
	Length     int    `json:"length"`
}

func init() { // 主函数
	en := control.Register("tianqi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "文本命令大全\n" +
			"- 天气查询：xxx天气" +
			"- 拼音查询：xxx拼音" +
			"- 每日一言" +
			"- 每日鸡汤" +
			"- 每日情话" +
			"- 绕口令",
	})
	en.OnFullMatch("每日一言").SetBlock(true).Handle(handle) //每日一言
	en.OnSuffix("天气").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["args"].(string)
			es, err := web.GetData(fmt.Sprintf(tianqi, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
				return
			}
			ctx.SendChain(message.Text(str+"天气如下:\n", helper.BytesToString(es)))
		})
	en.OnSuffix("拼音").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["args"].(string)
			es, err := web.GetData(fmt.Sprintf(pinyin, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
				return
			}
			ctx.SendChain(message.Text(str+"的拼音为：", helper.BytesToString(es)))
		})
	en.OnFullMatch("每日情话").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://xiaobai.klizi.cn/API/other/wtqh.php")
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			km := fmt.Sprintf("%s", data)
			ctx.SendChain(message.Text(km))
		})
	en.OnFullMatch("每日鸡汤").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://api.btstu.cn/yan/api.php?charset=utf-8&encode=text")
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			ctx.SendChain(message.Text(helper.BytesToString(data)))
		})
	en.OnFullMatch("绕口令").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://ovooa.com/API/rao/api.php?type=text")
			if err != nil {
				ctx.SendChain(message.Text("获取失败惹", err))
				return
			}
			ctx.SendChain(message.Text(helper.BytesToString(data)))
		})
}
func handle(ctx *zero.Ctx) {
	var rsp RspData
	data, err := web.GetData(url)
	if err != nil {
		ctx.SendChain(message.Text("Err:", err))
		return
	}
	err = json.Unmarshal(data, &rsp)
	if err != nil {
		ctx.SendChain(message.Text("出现错误捏：", err))
		return
	}
	msg := ""
	msg += rsp.Hitokoto + "\n出自：" + rsp.From + "\n"
	if len(rsp.FromWho) != 0 {
		msg += "作者：" + rsp.FromWho
	}
	ctx.SendChain(message.Text(msg))
}
