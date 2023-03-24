// Package baidu 百度百科
package baidu

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
	api = "https://api.a20safe.com/api.php?api=21&key=7d06a110e9e20a684e02934549db1d3d&text=%s" // api地址
)

type result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Content string `json:"content"`
	} `json:"data"`
}

func init() { // 主函数
	en := control.Register("baidu", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "百度百科\n" +
			"- 百度/百科[关键字]",
	})
	en.OnRegex(`^百[度科]\s*(.+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		es, err := web.GetData(fmt.Sprintf(api, ctx.State["regex_matched"].([]string)[1])) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		var r result                 // r数组
		err = json.Unmarshal(es, &r) // 填api返回结果，struct地址
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		if r.Code == 0 && len(r.Data) > 0 {
			ctx.SendChain(message.Text(r.Data[0].Content)) // 输出提取后的结果
		} else {
			ctx.SendChain(message.Text("API访问错误"))
		}
	})
}
