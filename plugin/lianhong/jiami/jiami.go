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
	jiami1 = "http://ovooa.com/API/sho_u/?msg=%v"          // 加密api地址
	jiami2 = "http://ovooa.com/API/sho_u/?format=1&msg=%v" // 解密api地址

)

type nmd struct { // struct解析格式大概是
	Code int    `json:"code"`
	Text string `json:"text"`
	Data struct {
		Message string `json:"Message"`
	} `json:"data"`
}

func init() { // 主函数
	en := control.Register("jiami", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "兽语加解密\n" +
			"- 兽语加密xxx\n- 兽语解密xxx",
	})
	en.OnRegex(`^兽语加密\s*(.+)$`).SetBlock(true). // 正值输入
							Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			// es := base14.EncodeString(str)
			es, err := web.GetData(fmt.Sprintf(jiami1, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}

			// es1 := fmt.Sprintf("%s", es)
			var r nmd                    // r数组
			err = json.Unmarshal(es, &r) // 填api返回结果，struct地址
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}

			/*if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}*/

			ctx.SendChain(message.Text(r.Data.Message)) // 输出提取后的结果
		})

	en.OnRegex(`^兽语解密\s*(.+)$`).SetBlock(true). // 正值输入
							Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			// es := base14.EncodeString(str)
			es, err := web.GetData(fmt.Sprintf(jiami2, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}
			var n nmd                    // r数组
			err = json.Unmarshal(es, &n) // 填api返回结果，struct地址
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}
			ctx.SendChain(message.Text(n.Data.Message)) // 输出提取后的结果
		})
}

/*en.OnRegex(`^解密\s*([一-踀]+[㴁-㴆]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := base14.DecodeString(str)

			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}*/

/* Package jiami 兽语加密专版
package jiami

import (
	"fmt"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	tiangouURL = "http://ovooa.com/API/sho_u/?msg=%v"
)

func init() {
	engine := control.Register("jiami", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "舔狗日记\n- 舔狗123456",
	})
	engine.OnRegex(`^兽语加密\s?(\d+)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		regexMatched := ctx.State["regex_matched"].([]string)
		// ctx.SendChain(message.Text(regexMatched[1]))
		data, err := web.GetData(fmt.Sprintf(tiangouURL, regexMatched[1]))
		if err != nil {
			ctx.SendChain(message.Text("Error:", err))
		}
		ctx.SendChain(message.Text(data))
	})
}
*/
