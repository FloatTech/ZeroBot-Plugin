// Package jiami 兽语加密与解密
package jiami

import (
	"encoding/json"
	"fmt"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	jiami1 = "http://ovooa.caonm.net/API/sho_u/?msg=%v"          // 加密api地址
	jiami2 = "http://ovooa.caonm.net/API/sho_u/?format=1&msg=%v" // 解密api地址

)

type nmd struct { // struct解析格式大概是
	Data struct {
		Message string
	} `json:"data"`
}

func init() { // 主函数
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "兽语加解密",
		Help: "兽语加解密\n" +
			"- 兽语加密xxx\n- 兽语解密xxx",
	})
	en.OnRegex(`^兽语加密\s*(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		str := ctx.State["regex_matched"].([]string)[1]
		es, err := web.GetData(fmt.Sprintf(jiami1, str)) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		var r nmd                    // r数组
		err = json.Unmarshal(es, &r) // 填api返回结果，struct地址
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		ctx.SendChain(message.Text(r.Data.Message)) // 输出提取后的结果
	})

	en.OnRegex(`^兽语解密\s*(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		str := ctx.State["regex_matched"].([]string)[1]
		es, err := web.GetData(fmt.Sprintf(jiami2, str)) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		var n nmd                    // r数组
		err = json.Unmarshal(es, &n) // 填api返回结果，struct地址
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		ctx.SendChain(message.Text(n.Data.Message)) // 输出提取后的结果
	})
}
