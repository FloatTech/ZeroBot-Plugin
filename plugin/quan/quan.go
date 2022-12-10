// Package quan  qq权重查询
package quan

import (
	"fmt"
	"strconv"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	quan = "http://tc.tfkapi.top/API/qqqz.php?qq=%v" // api
)

func init() { // 主函数
	en := control.Register("quan", &ctrl.Options[*zero.Ctx]{
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
		es, err := web.GetData(fmt.Sprintf(quan, str)) // 将网站返回结果赋值
		if err != nil {
			ctx.SendChain(message.Text("出现错误捏：", err))
			return
		}
		ctx.SendChain(message.Text(str, helper.BytesToString(es))) // 输出结果
	})
}
