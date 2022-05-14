// Package shadiao 来源于 https://shadiao.app/# 的接口
package shadiao

import (
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	chpURL          = "https://api.shadiao.app/chp"
	duURL           = "https://api.shadiao.app/du"
	pyqURL          = "https://api.shadiao.app/pyq"
	yduanziURL      = "http://www.yduanzi.com/duanzi/getduanzi"
	chayiURL        = "https://api.lovelive.tools/api/SweetNothings/Web/0"
	ganhaiURL       = "https://api.lovelive.tools/api/SweetNothings/Web/1"
	ergofabulousURL = "https://ergofabulous.org/luther/?"
	ua              = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	sdReferer       = "https://api.shadiao.app/"
	yduanziReferer  = "http://www.yduanzi.com/?utm_source=shadiao.app"
	loveliveReferer = "https://lovelive.tools/"
)

var (
	engine = control.Register("shadiao", &control.Options{
		DisableOnDefault: false,
		Help: "沙雕app\n" +
			"- 哄我\n- 渣我\n- 来碗绿茶\n- 发个朋友圈\n- 来碗毒鸡汤\n- 讲个段子\n- 马丁路德骂我\n",
	})
	sdMap = map[string]string{"哄我": chpURL, "来碗毒鸡汤": duURL, "发个朋友圈": pyqURL}
)

func init() {
	engine.OnFullMatchGroup([]string{"哄我", "来碗毒鸡汤", "发个朋友圈"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		requestURL := sdMap[ctx.State["matched"].(string)]
		data, err := web.RequestDataWith(web.NewDefaultClient(), requestURL, "GET", sdReferer, ua)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(gjson.GetBytes(data, "data.text").String()))
	})
}
