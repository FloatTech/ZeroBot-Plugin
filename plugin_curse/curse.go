package curse

import (
	"regexp"
	"strconv"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/web"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

const (
	qqReg    = `\d+`
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
	engine.OnPrefix("骂他").SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			ctx.SendChain(message.Text("请稍后重试0x0..."))
			return
		}
		var uid int64
		var text string
		reg := regexp.MustCompile(qqReg)
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			uid, _ = strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
		} else if reg.MatchString(ctx.Event.RawMessage) {
			result := reg.FindAllString(ctx.Event.RawMessage, -1)
			uid, _ = strconv.ParseInt(result[0], 10, 64)
		} else if uid == 0 {
			uid = ctx.Event.UserID
		}
		data, err := web.ReqWith(curseURL, "GET", "", ua)
		if err != nil {
			log.Println("err为:", err)
		}
		text = helper.BytesToString(data)
		ctx.SendChain(message.At(uid), message.Text(text))
	})
}
