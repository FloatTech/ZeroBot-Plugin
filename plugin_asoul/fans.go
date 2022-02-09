package asoul

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	var (
		diana  = fansapi("672328094")
		ava    = fansapi("672346917")
		eileen = fansapi("672342685")
		bella  = fansapi("672353429")
		carol  = fansapi("351609538")
		asoul  = fansapi("703007996")
	)
	// 指令触发查询 asoul 粉丝数据
	engine.OnFullMatch("粉丝信息").
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				diana.Uname, "  ", "粉丝：", diana.Follower, "  +", diana.Rise, "\n",
				ava.Uname, "  ", "粉丝：", ava.Follower, "  +", ava.Rise, "\n",
				eileen.Uname, "  ", "粉丝：", eileen.Follower, "  +", eileen.Rise, "\n",
				bella.Uname, "  ", "粉丝：", bella.Follower, "  +", bella.Rise, "\n",
				carol.Uname, "  ", "粉丝：", carol.Follower, "  +", carol.Rise, "\n",
				asoul.Uname, "  ", "粉丝：", carol.Follower, "  +", carol.Rise, "\n",
			))
		})
}
