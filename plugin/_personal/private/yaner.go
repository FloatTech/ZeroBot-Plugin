package base

import (
	"math/rand"
	"time"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const serviceName = "base"

var poke = rate.NewManager[int64](time.Minute*5, 6) // 戳一戳

func init() {
	engine := control.Register(serviceName, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "基础指令",
		Help:             "- @bot醒醒\n- @bot备份代码\n- @bot上传代码\n- @bot检查更新",
		OnDisable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("宝↗生↘永↗梦↘！！！！"))
		},
	})
	// 被喊名字
	engine.OnKeywordGroup([]string{"醒醒"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("啊？啊。啊，抱歉，刚刚不小心打瞌睡了=w="))
		})
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					nickname + "在窥屏哦",
					"我在听",
					"请问找" + nickname + "有什么事吗",
					"？怎么了",
				}[rand.Intn(4)],
			))
		})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if !poke.Load(ctx.Event.GroupID).AcquireN(1) {
				return // 最多戳6次
			}
			nickname := zero.BotConfig.NickName[0]
			switch rand.Intn(7) {
			case 1:
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText("哼！（打手）"))
				ctx.SendChain(message.Poke(ctx.Event.UserID))
			default:
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText(
					"哼！",
					"（打手）",
					nickname+"的脸不是拿来捏的！",
					nickname+"要生气了哦",
					"?",
				))
			}
		})
	engine.OnKeywordGroup([]string{"好吗", "行不行", "能不能", "可不可以"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			if rand.Intn(4) == 0 {
				nickname := zero.BotConfig.NickName[0]
				if rand.Intn(2) == 0 {
					ctx.SendChain(message.Text(nickname + "..." + nickname + "觉得不行"))
				} else {
					ctx.SendChain(message.Text(nickname + "..." + nickname + "觉得可以！"))
				}
			}
		})
}

func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}
