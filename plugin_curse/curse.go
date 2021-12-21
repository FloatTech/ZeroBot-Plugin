package curse

import (
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/math"
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"
)

const (
	ua       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	curseURL = "https://zuanbot.com/api.php?level=min&lang=zh_cn"
)

var (
	engine = control.Register("curse", &control.Options{
		DisableOnDefault: false,
		Help: "骂人\n" +
			"- 骂他[@xxx]|骂他[qq号]\n",
	})
	limit = rate.NewManager(time.Minute, 20)
)

func init() {
	engine.OnPrefix("骂我").SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			return
		}
		data, err := web.ReqWith(curseURL, "GET", "", ua)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.At(ctx.Event.UserID), message.Text(helper.BytesToString(data)))
	})
	engine.OnRegex(`^骂他.*?(\d+)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.GroupID).Acquire() {
				return
			}
			data, err := web.ReqWith(curseURL, "GET", "", ua)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.At(math.Str2Int64(ctx.State["regex_matched"].([]string)[1])), message.Text(helper.BytesToString(data)))
		})
}
