/*
Package atri 本文件基于 https://github.com/Kyomotoi/ATRI
为 Golang 移植版，语料、素材均来自上述项目
本项目遵守 AGPL v3 协议进行开源

年糕魔改版
*/
package atri

import (
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

const (
	// 服务名
	servicename = "atri"
	// 所有媒体的 本地 文件
	res = "file:///C:\\Users\\SGK2\\OneDrive\\Media\\Bot\\ATRI\\"
)

func init() { // 插件主体
	engine := control.Register(servicename, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "本插件基于 ATRI ，为 Golang 移植版\n此版本已被年糕深度魔改\n" +
			"- 烤年糕\n- 糕糕贴贴 | 贴贴 | 糕糕! | Rua | \n" +
			"- 抱抱糕糕 | 糕糕抱抱\n" +
			"- 早安 | 早哇 | 早上好 | ohayo | 哦哈哟 | 早啊 | 早好 | 早\n" +
			"- 中午好 | 午安 | 午好\n- 晚安 | 明天见\n" +
			"- 晚好 | 晚上好\n" +
			"- 我好了\n" +
			"- ？ | ? | ¿\n" +
			"- 离谱\n" +
			"- (糕糕)答应我\n" +
			"- TEST",
		OnEnable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("嗯...嗯..?"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("Zzz……Zzz……"))
		},
	})

	engine.OnFullMatchGroup([]string{"烤年糕", "炸年糕"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		switch rand.Intn(2) {
		case 0:
			ctx.SendChain(randText("年糕糕不是食物!", "不是食物...不能吃!"))
		}
	})

	engine.OnFullMatchGroup([]string{"糕糕贴贴", "贴贴", "贴贴!", "贴贴！", "糕糕!", "糕糕！"}, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), randText("<3", "贴贴!", "贴贴~", "w\n贴贴~"))
	})

	engine.OnFullMatchGroup([]string{"Rua", "rua", "Rua!", "rua!"}, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), randText("嗯?", "唔呣.", "(晃)", "唔呣唔呣."))
	})

	engine.OnFullMatchGroup([]string{"抱抱糕糕", "糕糕抱抱"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Text("抱."))
	})

	//	engine.OnKeywordGroup([]string{"草你妈", "操你妈", "脑瘫", "废柴", "fw", "five", "废物", "战斗", "爬", "爪巴", "sb", "SB", "傻B"}, isAtriSleeping, zero.OnlyToMe).SetBlock(true).
	//		Handle(func(ctx *zero.Ctx) {
	//			process.SleepAbout1sTo2s()
	//			ctx.SendChain(randImage("FN.jpg", "WQ.jpg", "WQ1.jpg"))
	//		})

	engine.OnFullMatchGroup([]string{
		"早安", "早安!", "早安！", "早安~", "早哇", "早哇!", "早哇！", "早哇~", "早上好", "早上好!", "早上好！", "早上好~", "ohayo", "哦哈哟", "早啊", "早啊!", "早啊！", "早啊~", "早", "早~", "早!", "早！",
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		now := time.Now().Hour()
		process.SleepAbout1sTo2s()
		switch {
		case now < 5: // 凌晨
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"这么早??",
				"才几点啊, 再休息会吧.",
				"嗯? 不会是没睡吧..?",
				"真的... 很早呢.",
				"(看时间\n还很早! 还能再睡一会...",
				"太早了吧?",
			))
		case now >= 5 && now < 9: // 早上
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"早.",
				"早~",
				"早!",
				"早啊.",
				"早啊~",
				"早啊!",
				"嗯, 早上好.",
				"早哦.",
				"早上好.",
				"哦哈哟~",
				"Morning~",
				"早饭吃什么?",
				"打算吃早饭不?",
				"早饭吃了吗?",
				"吃早饭了吗?",
			))
		case now >= 9 && now < 18: // 上午至下午
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"故意的吧? ...绝对是故意的吧?",
				"昨晚熬夜了?",
				"这可不早啊.",
				"不早啦. 这都什么时候了!",
				"这都什么时候了...",
				"欸? 看看时间啊喂.",
				"真的是刚醒吗?",
				"别熬夜啦.",
				"健康作息从我做起... 早点睡.",
				"..?",
				"? ...",
				"..?!",
				"欸?",
				"嗯?",
			))
		case now >= 18 && now < 24: // 晚上至半夜
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"你这... 是不打算休息了?",
				"不会真的是刚醒吧?!",
				"什么阴间作息啊!",
			))
		}
	})

	engine.OnFullMatchGroup([]string{"中午好", "午安", "午好"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		now := time.Now().Hour()
		if now > 11 && now < 15 { // 中午
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"午安~",
				"中午好, 嗯.",
				"中午好.",
				"睡个午觉?",
			))
		}
	})

	engine.OnFullMatchGroup([]string{"下午好"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		now := time.Now().Hour()
		if now > 12 && now < 17 { // 下午
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"嗯, 下午好.",
				"下午好啊.",
				"嗯哼.",
				"下午好~",
				"中饭吃了吗?",
				"有吃中饭吗?",
				"打算吃晚饭吗?",
				"打算今晚吃什么?",
			))
		}
	})

	engine.OnFullMatchGroup([]string{"晚安", "明天见"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		now := time.Now().Hour()
		process.SleepAbout1sTo2s()
		switch {
		case now < 6: // 凌晨
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"嗯, 早点休息.",
				"晚上好, 不要熬太晚哦.",
				"别熬夜了睡觉吧!",
				"是时候睡觉了!",
				"嗯.",
				"呼噜...",
			))
		case now >= 6 && now < 11: // 早上
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"这时候晚安?",
				"?啊这.",
				"睡觉吧...",
				"该休息了, 小心头发掉光!",
			))
		case now >= 11 && now < 15: // 中午至下午
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"是午觉吗?",
				"睡个午觉?",
			))
		case now >= 15 && now < 19: // 下午至黄昏
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"这有点早吧?",
				"有点早吧?",
				"晚点再来打招呼吧!",
				"现在还是白天欸, 没到晚上呢.",
			))
		case now >= 19 && now < 24: // 晚上至半夜
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"晚安~",
				"......呼噜",
				"晚安w",
				"嗯, 明天见.",
				"明天见啦~",
				"做个好梦!",
				"晚安好梦.",
				"晚安!",
				"嗯, 早点睡吧.",
			))
		}
	})

	engine.OnFullMatchGroup([]string{"晚上好", "晚好"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		now := time.Now().Hour()
		if now >= 18 && now <= 24 {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"嗯哼, 晚上好.",
				"晚上好♪",
				"晚上好呀.",
				"晚好~",
				"晚上好, 夜之城! (划去)",
				"晚~上好~",
			))
		}
	})

	//engine.OnKeywordGroup([]string{"好吗", "是吗", "行不行", "能不能", "可不可以"}).SetBlock(true).
	//	Handle(func(ctx *zero.Ctx) {
	//		process.SleepAbout1sTo2s()
	//		if rand.Intn(2) == 0 {
	//			ctx.SendChain(randImage("YES.png", "NO.jpg"))
	//		}
	//	})
	//engine.OnKeywordGroup([]string{"啊这"}).SetBlock(true).
	//	Handle(func(ctx *zero.Ctx) {
	//		process.SleepAbout1sTo2s()
	//		if rand.Intn(2) == 0 {
	//			ctx.SendChain(randImage("AZ.jpg", "AZ1.jpg"))
	//		}
	//	})

	engine.OnKeywordGroup([]string{"我好了"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), randText("不许好!", "憋回去!", "什么东西让糕糕也看看?", "看什么这么好?", "太快了太快了!"))
	})

	engine.OnFullMatchGroup([]string{"？", "?", "¿"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		switch rand.Intn(5) {
		case 1, 2:
			ctx.SendChain(randText("?", "？", "嗯？", "什么?", "欸?"))
		case 0:
			ctx.SendChain(randImage("WH1.jpg", "WH5.jpg", "WH6.jpg", "WH7.jpg", "WH8.jpg", "WH9.jpg", "WH10.jpg"))
		}
	})

	engine.OnKeyword("离谱").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		switch rand.Intn(5) {
		case 0:
			ctx.SendChain(randText("?", "？", "嗯？", "什么?", "欸?"))
		case 1, 2:
			ctx.SendChain(randImage("WH2.jpg", "WH3.jpg", "WH4.jpg"))
		}
	})

	engine.OnKeyword("答应我", zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), randText("甚么?!"))
	})

	engine.OnFullMatchGroup([]string{"年糕涩涩", "年糕涩涩!", "糕糕涩涩", "糕糕涩涩!", "涩涩", "涩涩!"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		switch rand.Intn(5) {
		case 0, 1:
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText("owo", "=w=", "涩涩!", "好哦~", "=v=", "omo", "ovo"))
		case 2, 3:
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randImage("H1.jpg", "H2.jpg", "H3.jpg", "H4.jpg", "H5.jpg", "H6.jpg", "H7.jpg", "H8.jpg", "H9.jpg"))
		case 4:
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randImage("HH1.jpg", "HH2.jpg", "HH3.jpg", "HH4.jpg"))
		}
	})

	engine.OnFullMatchGroup([]string{"催我起床", "起床", "起床!", "起床!!", "起床！", "起床！！", "起床..", "起床..."}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
			"真的要起床了吗?!",
			"快点快点.",
			"喂喂喂, 别说了不做啊!",
			"氢氢敲醒沉睡的心灵(物理)",
			"(戳)",
			"(踢一脚)",
		))
	})

	engine.OnFullMatch("TEST").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(randImage("HH1.jpg"))
	})

}

//随机文字
func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}

//随机图像(从设置的路径尝试匹配文件名)
func randImage(file ...string) message.MessageSegment {
	return message.Image(res + file[rand.Intn(len(file))])
}

// func randRecord(file ...string) message.MessageSegment {
// 	return message.Record(res + file[rand.Intn(len(file))])
// }
