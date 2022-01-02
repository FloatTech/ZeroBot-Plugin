// Package funny 冷笑话
package funny

import (
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/ctxext"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

var (
	engine = control.Register("funny", &control.Options{
		DisableOnDefault: false,
		Help: "讲个笑话\n" +
			"- 讲个笑话[@xxx]|讲个笑话[qq号]",
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
		name := ctxext.NickName(ctx)
		var j joke
		err := db.Pick("jokes", &j)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(strings.ReplaceAll(j.Text, "%name", name)))
	})
}
