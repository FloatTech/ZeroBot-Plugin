package dice

import (
	"strconv"
	"strings"

	fcext "github.com/FloatTech/floatbox/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex(`^[.。]jrrp`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			jrrp := fcext.RandSenderPerDayN(ctx.Event.UserID, 100) + 1
			var j strjrrp
			err := db.Find("strjrrp", &j, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
			if err == nil {
				ctx.SendGroupMessage(ctx.Event.GroupID, customjrrp(ctx, j.Strjrrp))
			} else {
				ctx.SendChain(message.At(uid), message.Text("阁下今日的人品值为", jrrp, "呢~"))
			}
		})
	engine.OnRegex(`^设置jrrp([\s\S]*)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			j := &strjrrp{
				GrpID:   ctx.Event.GroupID,
				Strjrrp: ctx.State["regex_matched"].([]string)[1],
			}
			err := db.Insert("strjrrp", j)
			if err == nil {
				ctx.SendChain(message.Text("记住啦!"))
			} else {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
		})
}

// customjrrp 自定义jrrp
func customjrrp(ctx *zero.Ctx, strjrrp string) string {
	uid := strconv.FormatInt(ctx.Event.UserID, 10)
	at := "[CQ:at,qq=" + uid + "]"
	jrrp := fcext.RandSenderPerDayN(ctx.Event.UserID, 100) + 1
	jrrps := strconv.Itoa(jrrp)
	str := strings.ReplaceAll(strjrrp, "{jrrp}", jrrps)
	str = strings.ReplaceAll(str, "{at}", at)
	return str
}
