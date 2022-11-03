// Package baike 百度百科
package baike

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
	api = "http://ovooa.com/API/bdbk/?Msg=%v" // api地址

)

type baka struct {
	Code int    `json:"code"`
	Text string `json:"text"`
	Data struct {
		Msg   string `json:"Msg"`
		Info  string `json:"info"`
		Image string `json:"image"`
		URL   string `json:"url"`
	} `json:"data"`
}

func init() { // 主函数
	en := control.Register("baike", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "夸克百科\n" +
			"- 百科+[关键字]",
	})
	en.OnRegex(`^百科\s*(.+)$`).SetBlock(true). // 正值输入
							Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			// es := base14.EncodeString(str)
			es, err := web.GetData(fmt.Sprintf(api, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}

			// es1 := fmt.Sprintf("%s", es)
			var r baka                   // r数组
			err = json.Unmarshal(es, &r) // 填api返回结果，struct地址
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}
			ctx.SendChain(message.Text(r.Data.Info+"\n详情查看:", r.Data.URL)) // 输出提取后的结果
		})
}
