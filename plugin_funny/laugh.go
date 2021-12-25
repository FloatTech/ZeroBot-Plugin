package funny

import (
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

var (
	engine = control.Register("funny", &control.Options{
		DisableOnDefault: false,
		Help: "讲个笑话\n" +
			"- 讲个笑话[@xxx]|讲个笑话[qq号]\n",
	})
	limit = rate.NewManager(time.Minute, 20)
	db    = &sql.Sqlite{DBPath: dbfile}
)

func init() {
	engine.OnPrefix("讲个笑话").SetBlock(true).FirstPriority().Handle(func(ctx *zero.Ctx) {
		if !limit.Load(ctx.Event.GroupID).Acquire() {
			return
		}
		// 获取名字
		name := ctx.State["args"].(string)
		if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
			qq, _ := strconv.ParseInt(ctx.Event.Message[1].Data["qq"], 10, 64)
			name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("card").String()
			if name == "" {
				name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false).Get("nickname").String()
			}
		} else if name == "" {
			name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, ctx.Event.UserID, false).Get("card").String()
			if name == "" {
				name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, ctx.Event.UserID, false).Get("nickname").String()
			}
		}
		var j joke
		err := db.Pick("jokes", &j)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(strings.ReplaceAll(j.Text, "%name", name)))
	})
}
