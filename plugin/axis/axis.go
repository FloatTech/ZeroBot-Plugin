package axis

import (
	"math/rand"
	"os"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/process"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

var (
	ygorules = []string{
		"一，村规：",
		"1.对方回合主要阶段最多发一次打断（包括手坑），进入战阶之后发什么都可以。",
		"2.禁止一次到位的打断（大宇宙，魔封，滑板，虚无等，鹰身女妖的吹雪，古遗物死镰等只能自己回合使用）",
		"3.禁止OTK，FTK",
		"\n二，比赛规则：",
		"1.参赛卡组要发出来让大家都看一下，然后投票选出是否可以参赛",
		"2.其他规则遵循比赛内容和本群村规",
		"\n三，暗黑决斗：",
		"1.双方指定对方一张卡，以灵魂作为赌约，进行三局两胜制决斗。",
		"2.输的一方将自己的灵魂封印到对方指定的卡，以后与对方决斗时禁止使用被封印的卡。",
	}
	ygorule = strings.Join(ygorules, "\n")
	zoomr   = []string{
		"好耶，我来学习牌技！快来这个房间吧ヾ(≧▽≦*)o",
		"打牌！房间已经给你们开好了哦~",
		"运气也是一种实力！来房间进行闪光抽卡吧！决斗者",
	}
	zooms = []string{
		"TM0#为所欲为",
		"TM0#WRGP",
		"TM0#阿克西斯",
	}
)

var (
	poke = rate.NewManager[int64](time.Minute*5, 11) // 戳一戳
)

func init() { // 插件主体
	engine := control.Register("axis", &control.Options{
		DisableOnDefault: false,
		Help:             "本插件为阿克西斯作者柳煜自建词库\n",
		OnEnable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("阿克西斯闪亮登场！锵↘锵↗~"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("宝↗生↘永↗梦↘！！！！"))
		},
		PrivateDataFolder: "axis",
	})
	//重启
	engine.OnFullMatchGroup([]string{"重启", "restart", "kill"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			os.Exit(0)
		})
	// 撤回最后的发言
	engine.OnMessage(func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		if msg[0].Type != "reply" {
			return false
		}
		for _, elem := range msg {
			if elem.Type == "text" {
				text := elem.Data["text"]
				text = strings.ReplaceAll(text, " ", "")
				text = strings.ReplaceAll(text, "\r", "")
				text = strings.ReplaceAll(text, "\n", "")
				if text == "多嘴" {
					return true
				}
			}
		}
		return false
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.Message[1].Data["qq"] != "" {
			ctx.SendChain(message.Text("9494，要像阿克西斯一样乖乖的才行哟~"))
		} else {
			ctx.SendChain(message.Text("呜呜呜呜"))
		}
		ctx.DeleteMessage(message.NewMessageIDFromString(ctx.Event.Message[0].Data["id"]))
	})
	// 被喊名字
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					nickname + "在窥屏哦",
					"我在听",
					"请问找" + nickname + "有什么事吗",
					nickname + "在忙呢",
				}[rand.Intn(4)],
			))
		})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			switch {
			case poke.Load(ctx.Event.GroupID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText(
					nickname+"很想和你打牌，可是我打不过你",
					"有本事你和改造君大哥打牌",
					nickname+"只能在终端打牌，请问你可以into the Vrains吗？",
					"你能教"+nickname+"怎么组杂技吗，我不会ಠಿ_ಠ",
					nickname+"的卡组还没有组好。。。",
				))
			case poke.Load(ctx.Event.GroupID).AcquireN(1):
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText("你的卡组和"+nickname+"一样是垃圾卡组吗", "喂！你再拍，我卡组都要皱了！"))
				ctx.SendChain(message.Poke(ctx.Event.UserID))
			default:
				// 频繁触发，不回复
			}
		})
	//石头剪刀布
	engine.OnFullMatchGroup([]string{"剪刀", "石头", "布"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.Event.Message[0].Data["text"]
			var rand_str = []string{"剪刀", "石头", "布"}
			result := rand_str[rand.Intn(len(rand_str))]
			ctx.SendChain(message.Text(result))
			process.SleepAbout1sTo2s()
			if rand.Intn(3) != 1 {
				if txt == result {
					ctx.SendChain(randText(
						"再来！",
						"想法一样吗",
						"。。。。。。",
						"平局吗，你的实力我认可了！",
					))
				} else if (txt == "剪刀" && result == "石头") || (txt == "石头" && result == "布") || (txt == "布" && result == "剪刀") {
					ctx.SendChain(randText(
						"阿这(´･ω･`)?",
						"阿克西斯的猜拳有这么厉害吗(＃°Д°)",
						"唉？我居然赢了！（＾∀＾●）ﾉｼ",
						"哼恩！~我赢了！",
					))
				} else {
					ctx.SendChain(randText(
						"再来！",
						"我好菜啊ಠಿ_ಠ",
						ctx.CardOrNickName(ctx.Event.UserID)+"tql",
						"岂可索！(。>︿<)_θ",
					))
				}
			}
		})
	engine.OnFullMatchGroup([]string{"早安", "早哇", "早上好", "ohayo", "哦哈哟", "お早う", "早好", "早", "早早早"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			process.SleepAbout1sTo2s()
			switch {
			case now < 6: // 凌晨
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"起个床都要这么卷的吗",
					"虽然但是，现在六点都没到，再睡一会吧",
					"zzz...早上好..zzz....",
					"你一定只是起床尿尿而已是吧，不然怎么起这么早",
					"偷偷起床去练抽卡吗？可是城市里没有熊给你拿来覆盖啊！",
				))
			case now >= 6 && now < 9:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"啊......早上好...(哈欠)",
					"早，一定要记得吃决斗饭团哟~\n今天你也要一飞冲天啊！",
					"老婆，早~",
					"早",
					"我这边天还是黑的咧，你那边天气怎么样",
					"早~又是新的决带笑的一天呢",
				))
			case now >= 9 && now < 18:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"emmm，昨天打牌打了这么晚，现在肯定还很累吧",
					"emmm，昨天组卡打了这么晚，现在肯定还很累吧",
					"守夜冠军起床了？",
				))
			case now >= 18 && now < 24:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"早个啥？哼唧！我都准备洗洗睡了！",
					"不是...你看看几点了，哼！",
					"晚上好哇",
				))
			}
		})
	engine.OnFullMatchGroup([]string{"中午好", "午安", "午好"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now().Hour()
			if now > 11 && now < 15 { // 中午
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"你知道吗，你午睡的时候，你的卡片精灵会在旁边偷偷陪着你的噢",
					"午觉要好好睡哦，不然卡片精灵会生气的！",
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
					"晚安，守夜人",
					"你是起床去尿尿了吗",
					"老婆，你真棒，晚安~",
				))
			case now >= 6 && now < 11:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"你可猝死算了吧！",
					"？啊这",
					"亲，这边建议赶快去睡觉呢~~~",
					"不可忍不可忍不可忍！！为何这还不猝死！！",
				))
			case now >= 11 && now < 15:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"午安w",
					"午觉要好好睡哦，不然卡片精灵会生气的！",
					"午安，决斗者",
					"赶紧睡吧，狗命要紧",
				))
			case now >= 15 && now < 19:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"难不成？？晚上不想睡觉？？现在休息",
					"就......挺离谱的...现在睡觉",
					"现在还是白天哦，睡觉还太早了",
				))
			case now >= 19 && now < 24:
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText(
					"你知道吗，你睡觉的时候，你的卡片精灵会在旁边偷偷陪着你的噢",
					"晚上要好好睡哦，不然卡片精灵会生气的！",
					"......(打瞌睡)",
					"呼...呼...已经睡着了哦~...呼......",
					"...嘿嘿嘿，抽卡...通招...通....zzzzz....",
				))
			}
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
				ctx.SendChain(randImage("AZ.png", "AZ1.jpg", "AZ2.jpg", "AZ3.png"))
			}
		})
	engine.OnKeywordGroup([]string{"我好了"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			if rand.Intn(2) == 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), randText("不许好！", "憋回去！"))
			} else {
				ctx.SendChain(randImage("AZ3.png", "BS.jpg"))
			}
		})
	engine.OnFullMatchGroup([]string{"？", "?", "¿"}, isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			switch rand.Intn(5) {
			case 0:
				ctx.SendChain(randText("?", "？", "嗯？", "(。´・ω・)ん?", "ん？"))
			case 1, 2:
				ctx.SendChain(randImage("AZ.png", "AZ1.jpg", "AZ2.jpg", "AZ3.png"))
			}
		})
	engine.OnKeyword("离谱", isAtriSleeping).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			switch rand.Intn(5) {
			case 0:
				ctx.SendChain(randText("?", "？", "嗯？", "(。´・ω・)ん?", "ん？"))
			case 1, 2:
				ctx.SendChain(randImage("AZ.png", "AZ1.jpg", "AZ3.gng", "WY.jpg"))
			}
		})
	engine.OnKeywordGroup([]string{"草你妈", "操你妈", "脑瘫", "废柴", "fw", "five", "废物", "战斗", "爬", "爪巴", "sb", "SB", "傻B"}, isAtriSleeping, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randImage("FN.jpg", "BS.jpg", "AZ.png"))
		})
	engine.OnKeywordGroup([]string{"好耶", "nice", "成功", "完成"}, isAtriSleeping, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(randImage("OK.jpg", "OK2.jpg", "OK3.gif"))
		})

	// 软件
	engine.OnFullMatchGroup([]string{"/软件", ".软件"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(
			message.Text("下载链接：https://ygo233.com/download"))
	})
	// 先行卡
	engine.OnFullMatchGroup([]string{"/先行卡", ".先行卡", "先行卡"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("先行卡链接：https://ygo233.com/pre"))
	})
	// 村规
	engine.OnFullMatchGroup([]string{"/村规", ".村规", "村规", "群规", "暗黑决斗"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		ctx.SendChain(message.Text(ygorule))
	})
	// 房间
	engine.OnFullMatchGroup([]string{"/房间", ".房间", "房间", "开房"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		ctx.SendChain(message.Text(zoomr[rand.Intn(len(zoomr))]))
		//process.SleepAbout1sTo2s()
		ctx.SendChain(message.Text(zooms[rand.Intn(len(zooms))]))
	})
	engine.OnFullMatchGroup([]string{"/双打", ".双打", "双打"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if ctx.Event.GroupID != 979031435 {
			return
		}
		ctx.SendChain(message.Text(zoomr[rand.Intn(len(zoomr))]))
		//process.SleepAbout1sTo2s()
		ctx.SendChain(message.Text("T," + zooms[rand.Intn(len(zooms))]))
	})

}

func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}

func randImage(file ...string) message.MessageSegment {
	return message.Image("file:///C:/Users/17254/Desktop/阿克西斯/ZeroBot-Plugin-master/data/axis/" + file[rand.Intn(len(file))])
}

// isAtriSleeping 凌晨0点到6点，ATRI 在睡觉，不回应任何请求
func isAtriSleeping(ctx *zero.Ctx) bool {
	if now := time.Now().Hour(); now >= 1 && now < 6 {
		return false
	}
	return true
}
