// Package wenben 文本链接
package wenben

import (
	"fmt"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	tianqi = "https://xiaobai.klizi.cn/API/other/weather_1.php?data=&msg=%v" // api地址
	pinyin = "http://ovooa.com/API/pinyin/api.php?type=text&msg=%v"
)

func init() { // 主函数
	en := control.Register("tianqi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "文本命令大全\n" +
			"- 天气查询：地名 + 天气" +
			"- 拼音查询：文字 + 拼音",
	})
	en.OnSuffix("天气").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["args"].(string)
			es, err := web.GetData(fmt.Sprintf(tianqi, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}
			km := string(es)
			ctx.SendChain(message.Text(str+"天气如下:\n", km))
		})
	en.OnSuffix("拼音").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["args"].(string)
			es, err := web.GetData(fmt.Sprintf(pinyin, str)) // 将网站返回结果赋值
			if err != nil {
				ctx.SendChain(message.Text("出现错误捏：", err))
			}
			km := string(es)
			ctx.SendChain(message.Text(str+"的拼音为：", km))
		})
}
