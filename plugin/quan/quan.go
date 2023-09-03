// Package quan  qq权重查询
package quan

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	quan = "http://tfapi.top/API/qqqz.php?type=json&qq=" // api
)

type result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Qz   string `json:"qz"`
}

func init() { // 主函数
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "QQ权重查询",
		Help: "权重查询方式\n" +
			"- 权重查询+@xxx" +
			"- 权重查询+QQ号(可以不写，默认本人)",
	})
	en.OnRegex(`^权重查询\s*(\[CQ:at,qq=)?(\d+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		str := ctx.State["regex_matched"].([]string)[2] // 获取uid
		if str == "" {                                  // user
			str = strconv.FormatInt(ctx.Event.UserID, 10)
		}
		es, err := web.GetData(quan + str) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏: ", err))
			return
		}
		var data result
		err = json.Unmarshal(es, &data)
		if err != nil {
			ctx.SendChain(message.Text("解析json错误: ", err))
			return
		}
		var msg strings.Builder
		msg.WriteString("查询账号: ")
		msg.WriteString(str)
		msg.WriteString("\n")
		msg.WriteString("查询状态: ")
		msg.WriteString(data.Msg)
		msg.WriteString("\n您的权重为: ")
		msg.WriteString(data.Qz)
		ctx.SendChain(message.Text(msg.String())) // 输出结果
	})
}
