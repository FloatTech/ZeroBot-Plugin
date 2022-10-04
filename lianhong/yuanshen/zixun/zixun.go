// Package yuanshenzixun 原神咨询
package yuanshenzixun

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
	api = "https://xiaobai.klizi.cn/API/other/yuanshen_zx.php?n=%v" // api地址

)

type mmmm struct {
	Title  string   `json:"title"`
	Text   string   `json:"text"`
	Images []string `json:"images"`
}

func init() { // 主函数
	en := control.Register("yuanshenzixun", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "原神咨询\n" +
			"- 原神咨询+[页数]",
	})
	en.OnRegex(`^原神咨询\s*(.+)$`).SetBlock(true). // 正值输入
							Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			// es := base14.EncodeString(str)
			es, err := web.GetData(fmt.Sprintf(api, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}

			// es1 := fmt.Sprintf("%s", es)
			var r mmmm                   // r数组
			err = json.Unmarshal(es, &r) // 填api返回结果，struct地址
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}
			ctx.SendChain(message.Text(r.Title, r.Text)) // 输出提取后的结果
			ctx.SendChain(message.Image(r.Images[1]))
		})
}
