/*
Package atri 本文件基于 https://github.com/Kyomotoi/ATRI
为 Golang 移植版，语料、素材均来自上述项目
本项目遵守 AGPL v3 协议进行开源
*/
package atri

import (
	"math/rand"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/process"
)

const (
	// 服务名
	servicename = "atri"
	// ATRI 表情的 codechina 镜像
	res = "https://gitcode.net/u011570312/zbpdata/-/raw/main/Atri/"
)

func init() { // 插件主体
	engine := control.Register(servicename, &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "本插件基于 ATRI ，为 Golang 移植版\n" +
			"- ATRI醒醒\n- ATRI睡吧\n- 萝卜子\n- 喜欢 | 爱你 | 爱 | suki | daisuki | すき | 好き | 贴贴 | 老婆 | 亲一个 | mua\n" +
			"- 草你妈 | 操你妈 | 脑瘫 | 废柴 | fw | 废物 | 战斗 | 爬 | 爪巴 | sb | SB | 傻B\n- 早安 | 早哇 | 早上好 | ohayo | 哦哈哟 | お早う | 早好 | 早 | 早早早\n" +
			"- 中午好 | 午安 | 午好\n- 晚安 | oyasuminasai | おやすみなさい | 晚好 | 晚上好\n- 高性能 | 太棒了 | すごい | sugoi | 斯国一 | よかった\n" +
			"- 没事 | 没关系 | 大丈夫 | 还好 | 不要紧 | 没出大问题 | 没伤到哪\n- 好吗 | 是吗 | 行不行 | 能不能 | 可不可以\n- 啊这\n- 我好了\n- ？ | ? | ¿\n" +
			"- 离谱\n- 答应我",
		OnEnable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("嗯呜呜……夏生先生……？"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("Zzz……Zzz……"))
		},
	})
	engine.OnFullMatch("蛇", isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			switch rand.Intn(2) {
			case 0:
				ctx.SendChain(randText("【蛇】在哦", "【蛇】盯上你了哦", "是想来找我玩吗~小白鼠？"))
			case 1:
				ctx.SendChain(randText("抓住你了哦~小白鼠~"))
			}
		})
	engine.OnFullMatchGroup([]string{"蛇蛇", "蛇~蛇~", "梅比乌斯"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randText("【蛇】在哦", "【蛇】盯上你了哦", "是想来找我玩吗~小白鼠？", "抓住你了哦~小白鼠~"))
		})
	engine.OnFullMatchGroup([]string{"喜欢", "爱你", "爱", "suki", "daisuki", "すき", "好き", "贴贴", "老婆", "亲一个", "mua"}, isAtriSleeping, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randText(
				"是要隔着衣服贴，还是从领口伸进去贴呀?小~白~鼠~",
				"小~白~鼠~？",
				"贴这么近，是对我有什么想法吗？小白鼠？",
				"来吧小白鼠，牵起我的手，加入这进化的路途吧~",
				"可以哟小白鼠，来和我做点有意思的事吧~",
				"看来我们都很闲呢，要去我的实验室里坐坐吗~？",
				"这是...表白吗？真是意外呢，我的小白鼠~",
				"你是喜欢我这副躯体呢？还是...（笑~",
				"想让我也喜欢你？你知道该怎么做~ ",
			))
		})
	engine.OnKeywordGroup([]string{"草你妈", "操你妈", "脑瘫", "废柴", "fw", "five", "废物", "战斗", "爬", "爪巴", "sb", "SB", "傻B"}, isAtriSleeping, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randText(
				"既然说出了这样的话~ 那你应该已经做好觉悟了吧，呵呵呵~",
				"做好准备哦，小白鼠~接下来，可是会很痛的~",
				"把你做成标本，怎么样~",
				"呵呵呵~ 可不要~逃走哦~！",
				"哎呀，生命可真是脆弱呢~ 你觉得呢？我的小白鼠~？",
			))
		})
	engine.OnFullMatchGroup([]string{"早安", "早哇", "早上好", "ohayo", "哦哈哟", "お早う", "早好", "早", "早早早"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			process.SleepAbout1sTo2s()
			switch {
			case now < 6: // 凌晨
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"zzzz......",
					"zzzzzzzz......",
					"克莱因...来..zzz..帮人家..zzz..",
					"如果是要找梅比乌斯博士的话...博士还在休息",
					"有什么我可以帮忙的吗",
				))
			case now >= 6 && now < 12:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"啊......早上好...克莱因(哈欠)",
					"唔...哈啊啊~~~克莱因？......不是啊~",
					"早上好......无聊的早晨呢~陪我玩玩吧，小白鼠？",
					"早上好...睡觉？博士的工作...还没有做完，我还能...工作...",
				))
			case now >= 12 && now < 18:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"现在可不是早上好的时间哦~ ",
					"难道你昨天晚上做了什么吗？我的小白鼠~？",
					"繁衍，也是生命延续的一种形式...没有？呵呵~",
					"这个时间...小白鼠~？来陪我做点有意思的事吧~",
				))
			case now >= 18 && now < 24:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"即使是【蛇】...这个时间也该睡觉了呢~",
					"啊，早上...哦不对，晚上好",
					"早上好？难不成，小白鼠~ 你是昼伏夜出？",
				))
			}
		})
	engine.OnFullMatchGroup([]string{"中午好", "午安", "午好"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			if now > 11 && now < 15 { // 中午
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"午安哦~ 我的小白鼠~ ",
					"午安，小白鼠，做个好梦哦~ 呵呵~",
				))
			}
		})
	engine.OnFullMatchGroup([]string{"晚安", "oyasuminasai", "おやすみなさい", "晚好", "晚上好"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			process.SleepAbout1sTo2s()
			switch {
			case now < 6: // 凌晨
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"zzzz......",
					"zzzzzzzz......",
					"梅比乌斯博士已经休息了，有什么事情找我就行...",
					"不早了舰长，请注意休息...不然会影响实验结果",
				))
			case now >= 6 && now < 11:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"晚安？是我睡过头了吗？还是小白鼠你睡过头了呢~",
					"晚上好？难不成，小白鼠~ 你是昼伏夜出吗？呵呵~",
					"【蛇】要冬眠了哦~ 呵呵~",
				))
			case now >= 11 && now < 19:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"纠正，应该是午安……舰长",
					"这个时间...小白鼠~？来陪我做点有意思的事吧~",
				))
			case now >= 19 && now < 24:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"晚安，我的小白鼠，做个好梦~",
					"呵呵~ 小白鼠~ 明天见~",
					"小白鼠~猜猜我会不会趁你睡着的时候………… 呵呵~这就怕了吗~",
					"克莱因还需要继续完成博士的工作，舰长请先去休息",
				))
			}
		})
	engine.OnKeywordGroup([]string{"高性能", "太棒了", "すごい", "sugoi", "斯国一", "よかった"}, isAtriSleeping, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randText(
				"太棒了？你是说哪里呢？我的小白鼠~",
				"觉得自己运气不错？要不要去我的实验室里试试呀~",
			))
		})
	engine.OnKeywordGroup([]string{"好吗", "是吗", "行不行", "能不能", "可不可以"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			if rand.Intn(2) == 0 {
				ctx.SendChain(randImage("YES.png", "NO.jpg"))
			}
		})
	engine.OnKeywordGroup([]string{"啊这"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			if rand.Intn(2) == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"是什么让你觉得无语了呢？说来听听吧，小白鼠~",
					"发生什么了吗？我可爱的小白鼠~",
				))
			}
		})
	engine.OnKeywordGroup([]string{"我好了"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
				"你好了？你是说哪里呢？我的小白鼠~",
				"觉得自己运气不错？要不要去我的实验室里试试呀~",
			))
		})
	engine.OnFullMatchGroup([]string{"？", "?", "¿"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			switch rand.Intn(5) {
			case 0:
				ctx.SendChain(randText("?", "？", "嗯？"))
			case 1, 2:
				ctx.SendChain(randText("有什么问题吗？小白鼠~", "ん？"))
			}
		})
	engine.OnKeyword("离谱", isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			switch rand.Intn(5) {
			case 0:
				ctx.SendChain(randText("?", "？", "嗯？", "有什么问题吗？小白鼠~", "ん？"))
			case 1, 2:
				ctx.SendChain(randText("需要一点小小的帮助吗？呵呵~", "需要蛇来帮你一下吗？我可爱的小白鼠~"))
			}
		})
	engine.OnKeyword("答应我", isAtriSleeping, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randText("和蛇立约定？小心会被吞掉哦~"))
		})
}

func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}

func randImage(file ...string) message.MessageSegment {
	return message.Image(res + file[rand.Intn(len(file))])
}

func randRecord(file ...string) message.MessageSegment {
	return message.Record(res + file[rand.Intn(len(file))])
}

// isAtriSleeping 凌晨0点到6点，ATRI 在睡觉，不回应任何请求
func isAtriSleeping(ctx *zero.Ctx) bool {
	if now := time.Now().Hour(); now >= 1 && now < 6 {
		return false
	}
	return true
}
