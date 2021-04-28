// 本文件基于 https://github.com/Kyomotoi/ATRI
// 为 Golang 移植版，语料、素材均来自上述项目
// 本项目遵守 AGPL v3 协议进行开源
package atri

import (
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	PRIO   = -1
	RES    = "https://raw.dihe.moe/Yiwen-Chan/ZeroBot-Plugin/master/atri/"
	ENABLE = true
)

func init() { // 插件主体
	zero.OnFullMatch("ATRI醒醒", zero.AdminPermission).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ENABLE = true
			ctx.SendChain(randText("嗯呜呜……夏生先生……？"))
		})
	zero.OnFullMatch("ATRI睡吧", zero.AdminPermission).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ENABLE = false
			ctx.SendChain(randText("Zzz……Zzz……"))
		})
	zero.OnFullMatch("萝卜子", AtriSwitch(), AtriSleep()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			switch rand.Intn(2) {
			case 0:
				ctx.SendChain(randText("萝卜子是对机器人的蔑称！", "是亚托莉......萝卜子可是对机器人的蔑称"))
			case 1:
				ctx.SendChain(randRecord("RocketPunch.amr"))
			}
		})
	zero.OnKeywordGroup([]string{"喜欢", "爱你", "爱", "suki", "daisuki", "すき", "好き", "贴贴", "老婆", "亲一个", "mua"}, AtriSwitch(), AtriSleep(), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			if rand.Intn(10) == 0 {
				ctx.SendChain(randImage("SUKI.jpg", "SUKI1.jpg", "SUKI2.png"))
			}
		})
	zero.OnKeywordGroup([]string{"草你妈", "操你妈", "脑瘫", "废柴", "fw", "five", "废物", "战斗", "爬", "爪巴", "sb", "SB", "傻B"}, AtriSwitch(), AtriSleep(), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO - 1).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randImage("FN.jpg", "WQ.jpg", "WQ1.jpg"))
		})
	zero.OnFullMatchGroup([]string{"早安", "早哇", "早上好", "ohayo", "哦哈哟", "お早う"}, AtriSwitch()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			switch {
			case now < 6: // 凌晨
				ctx.SendChain(randText(
					"zzzz......",
					"zzzzzzzz......",
					"zzz...好涩哦..zzz....",
					"别...不要..zzz..那..zzz..",
					"嘻嘻..zzz..呐~..zzzz..",
					"...zzz....哧溜哧溜....",
				))
			case now >= 6 && now < 9:
				ctx.SendChain(randText(
					"啊......早上好...(哈欠)",
					"唔......吧唧...早上...哈啊啊~~~\n早上好......",
					"早上好......",
					"早上好呜......呼啊啊~~~~",
					"啊......早上好。\n昨晚也很激情呢！",
					"吧唧吧唧......怎么了...已经早上了么...",
					"早上好！",
					"......看起来像是傍晚，其实已经早上了吗？",
					"早上好......欸~~~脸好近呢",
				))
			case now >= 9 && now < 18:
				ctx.SendChain(randText(
					"哼！这个点还早啥，昨晚干啥去了！？",
					"熬夜了对吧熬夜了对吧熬夜了对吧？？？！",
					"是不是熬夜是不是熬夜是不是熬夜？！",
				))
			case now >= 18 && now < 24:
				ctx.SendChain(randText(
					"早个啥？哼唧！我都准备洗洗睡了！",
					"不是...你看看几点了，哼！",
					"晚上好哇",
				))
			}
		})
	zero.OnFullMatchGroup([]string{"中午好", "午安"}, AtriSwitch()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			switch {
			case now < 6: // 凌晨
				ctx.SendChain(randText(
					"午安w",
					"午觉要好好睡哦，ATRI会陪伴在你身旁的w",
					"嗯哼哼~睡吧，就像平常一样安眠吧~o(≧▽≦)o",
					"睡你午觉去！哼唧！！",
				))
			}
		})
	zero.OnFullMatchGroup([]string{"晚安", "oyasuminasai", "おやすみなさい"}, AtriSwitch()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			switch {
			case now < 6: // 凌晨
				ctx.SendChain(randText(
					"zzzz......",
					"zzzzzzzz......",
					"zzz...好涩哦..zzz....",
					"别...不要..zzz..那..zzz..",
					"嘻嘻..zzz..呐~..zzzz..",
					"...zzz....哧溜哧溜....",
				))
			case now >= 6 && now < 11:
				ctx.SendChain(randText(
					"你可猝死算了吧！",
					"？啊这",
					"亲，这边建议赶快去睡觉呢~~~",
					"不可忍不可忍不可忍！！为何这还不猝死！！",
				))
			case now >= 11 && now < 15:
				ctx.SendChain(randText(
					"午安w",
					"午觉要好好睡哦，ATRI会陪伴在你身旁的w",
					"嗯哼哼~睡吧，就像平常一样安眠吧~o(≧▽≦)o",
					"睡你午觉去！哼唧！！",
				))
			case now >= 15 && now < 19:
				ctx.SendChain(randText(
					"难不成？？晚上不想睡觉？？现在休息",
					"就......挺离谱的...现在睡觉",
					"现在还是白天哦，睡觉还太早了",
				))
			case now >= 19 && now < 24:
				ctx.SendChain(randText(
					"嗯哼哼~睡吧，就像平常一样安眠吧~o(≧▽≦)o",
					"......(打瞌睡)",
					"呼...呼...已经睡着了哦~...呼......",
					"......我、我会在这守着你的，请务必好好睡着",
				))
			}
		})
	zero.OnKeywordGroup([]string{"高性能", "太棒了", "すごい", "sugoi", "斯国一", "よかった"}, AtriSwitch(), AtriSleep(), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText(
				"当然，我是高性能的嘛~！",
				"小事一桩，我是高性能的嘛",
				"怎么样？还是我比较高性能吧？",
				"哼哼！我果然是高性能的呢！",
				"因为我是高性能的嘛！嗯哼！",
				"因为我是高性能的呢！",
				"哎呀~，我可真是太高性能了",
				"正是，因为我是高性能的",
				"是的。我是高性能的嘛♪",
				"毕竟我可是高性能的！",
				"嘿嘿，我的高性能发挥出来啦♪",
				"我果然是很高性能的机器人吧！",
				"是吧！谁叫我这么高性能呢！哼哼！",
				"交给我吧，有高性能的我陪着呢",
				"呣......我的高性能，毫无遗憾地施展出来了......",
			))
		})
	zero.OnKeywordGroup([]string{"没事", "没关系", "大丈夫", "还好", "不要紧", "没出大问题", "没伤到哪"}, AtriSwitch(), AtriSleep(), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText(
				"当然，我是高性能的嘛~！",
				"没事没事，因为我是高性能的嘛！嗯哼！",
				"没事的，因为我是高性能的呢！",
				"正是，因为我是高性能的",
				"是的。我是高性能的嘛♪",
				"毕竟我可是高性能的！",
				"那种程度的事不算什么的。\n别看我这样，我可是高性能的",
				"没问题的，我可是高性能的",
			))
		})

	zero.OnKeywordGroup([]string{"好吗", "是吗", "行不行", "能不能", "可不可以"}, AtriSwitch(), AtriSleep()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			if rand.Intn(2) == 0 {
				ctx.SendChain(randImage("YES.png", "NO.jpg"))
			}
		})
	zero.OnFullMatchGroup([]string{"啊这"}, AtriSwitch(), AtriSleep()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			if rand.Intn(2) == 0 {
				ctx.SendChain(randImage("AZ.jpg", "AZ1.jpg"))
			}
		})
	zero.OnFullMatchGroup([]string{"我好了"}, AtriSwitch(), AtriSleep()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("不许好！", "憋回去！"))
		})
	zero.OnFullMatchGroup([]string{"？", "?", "¿"}, AtriSwitch(), AtriSleep()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			switch rand.Intn(5) {
			case 0:
				ctx.SendChain(randText("?", "？", "嗯？", "(。´・ω・)ん?", "ん？"))
			case 1, 2:
				ctx.SendChain(randImage("WH.jpg", "WH1.jpg", "WH2.jpg", "WH3.jpg"))
			}
		})
	zero.OnKeyword("离谱", AtriSwitch(), AtriSleep()).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			switch rand.Intn(5) {
			case 0:
				ctx.SendChain(randText("?", "？", "嗯？", "(。´・ω・)ん?", "ん？"))
			case 1, 2:
				ctx.SendChain(randImage("WH.jpg"))
			}
		})
	zero.OnKeyword("答应我", AtriSwitch(), AtriSleep(), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("我无法回应你的请求"))
		})
}

func randText(text ...string) message.MessageSegment {
	length := len(text)
	return message.Text(text[rand.Intn(length)])
}

func randImage(file ...string) message.MessageSegment {
	length := len(file)
	return message.Image(RES + file[rand.Intn(length)])
}

func randRecord(file ...string) message.MessageSegment {
	length := len(file)
	return message.Record(RES + file[rand.Intn(length)])
}

// AtriSwitch 控制 ATRI 的开关
func AtriSwitch() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		return ENABLE
	}
}

// AtriSleep 凌晨0点到6点，ATRI 在睡觉，不回应任何请求
func AtriSleep() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if now := time.Now().Hour(); now >= 1 && now < 6 {
			return false
		}
		return true
	}
}
