// Package groupwife 群老婆
package groupwife

import (
	"sort"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("groupwife", &control.Options{
		DisableOnDefault: true,
		Help: "群老婆\n" +
			"-哪个群友是我老婆",
	})
	engine.OnFullMatchGroup([]string{"哪个群友是我老婆", "哪位群友是我老婆", "今天谁是我老婆"}, zero.OnlyGroup).SetBlock(false).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			list := ctx.CallAction("get_group_member_list", zero.Params{
				"group_id": ctx.Event.GroupID,
				"no_cache": false,
			}).Data
			temp := list.Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			uid := ctx.Event.UserID
			who := temp[ctxext.RandSenderPerDayN(ctx, len(temp))]
			wifename := who.Get("card").Str
			wifeid := who.Get("user_id").Int()
			if wifename == "" {
				wifename = who.Get("nickname").Str
			}
			ctx.SendChain(message.At(uid),
				message.Text("今天你的群友老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(wifeid, 10)+"&s=640").Add("ceche", 0),
				message.Text("\n【"+wifename+"】("+strconv.FormatInt(wifeid, 10)+")哒！"))
		})
}
