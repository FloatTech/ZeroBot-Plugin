// Package funny 冷笑话
package funny

import (
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	sql "github.com/FloatTech/sqlite"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"

	"github.com/FloatTech/zbputils/control/order"
)

var db = &sql.Sqlite{DBPath: dbfile}

func init() {
	control.Register("funny", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "讲个笑话\n" +
			"- 讲个笑话[@xxx] | 讲个笑话[qq号]",
	}).OnPrefix("讲个笑话").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
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
