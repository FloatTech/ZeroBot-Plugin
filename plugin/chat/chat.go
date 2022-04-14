// Package chat 对话插件
package chat

import (
	"math/rand"
	"strconv"
	"time"

	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	poke   = rate.NewManager[int64](time.Minute*5, 8) // 戳一戳
	engine = control.Register("chat", &control.Options{
		DisableOnDefault: false,
		Help:             "chat\n- [BOT名字]\n- [戳一戳BOT]\n- 空调开\n- 空调关\n- 群温度\n- 设置温度[正整数]",
	})
)

func init() { // 插件主体
	// 被喊名字
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					nickname + "在此，有何贵干~",
					"(っ●ω●)っ在~",
					"这里是" + nickname + "(っ●ω●)っ",
					nickname + "不在呢~",
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
				// ctx.SendChain(message.Text("请不要戳", nickname, " >_<"))
				// pokereply(ctx, nickname)
				ctx.SendChain(message.Text(
					[]string{
						"请不要戳" + nickname + " >_<",
						"喂(#`O′) 戳" + nickname + "干嘛!",
						"别戳了…痒……",
						"呜…别戳了…",
						"别戳了！",
						"喵~",
						"…把手拿开",
						"有笨蛋在戳我，我不说是谁",
						"达咩呦，达咩达咩",
						"好怪..你不要过来啊啊啊啊啊",
						"嗯。嗯~嗯？嗯！",
						"咿呀！吓我一跳~",
						"别戳啦！",
						"你一定是变态对吧？！！",
						"你再戳我要生气了！",
						"好怪..你不要过来啊啊啊啊啊",
						"我好像瞌睡了",
						"可恶啊...性御旺盛的大人真是讨厌..",
						"不...不行的啦！",
						"好啦..今天就满足你吧~",
						"我家也没什么值钱的了，唯一能拿得出手的也就是我了",
						"你干嘛！",
						"变态变态变态变态！！！",
						"只能..一点点..哦?",
					}[rand.Intn(24)],
				))

			case poke.Load(ctx.Event.GroupID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				// ctx.SendChain(message.Text("喂(#`O′) 戳", nickname, "干嘛！"))
				// pokereply(ctx, nickname)
				ctx.SendChain(message.Text(
					[]string{
						"请不要戳" + nickname + " >_<",
						"喂(#`O′) 戳" + nickname + "干嘛!",
						"别戳了…痒……",
						"呜…别戳了…",
						"别戳了！",
						"喵~",
						"…把手拿开",
						"戳回去<( ￣^￣)",
						"有笨蛋在戳我，我不说是谁",
						"达咩呦，达咩达咩",
						"好怪..你不要过来啊啊啊啊啊",
						"嗯。嗯~嗯？嗯！",
						"咿呀！吓我一跳~",
						"别戳啦！",
						"你一定是变态对吧？！！",
						"你再戳我要生气了！",
						"好怪..你不要过来啊啊啊啊啊",
						"我好像瞌睡了",
						"可恶啊...性御旺盛的大人真是讨厌..",
						"不...不行的啦！",
						"好啦..今天就满足你吧~",
						"我家也没什么值钱的了，唯一能拿得出手的也就是我了",
						"你干嘛！",
						"变态变态变态变态！！！",
						"只能..一点点..哦?",
					}[rand.Intn(25)],
				))

				ctx.Send(message.Poke(ctx.Event.UserID))
			default:
				// 频繁触发，不回复
			}
		})
	// 群空调
	var AirConditTemp = map[int64]int{}
	var AirConditSwitch = map[int64]bool{}
	engine.OnFullMatch("空调开").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = true
			ctx.SendChain(message.Text("❄️哔~"))
		})
	engine.OnFullMatch("空调关").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = false
			delete(AirConditTemp, ctx.Event.GroupID)
			ctx.SendChain(message.Text("💤哔~"))
		})
	engine.OnRegex(`设置温度(\d+)`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if _, exist := AirConditTemp[ctx.Event.GroupID]; !exist {
				AirConditTemp[ctx.Event.GroupID] = 26
			}
			if AirConditSwitch[ctx.Event.GroupID] {
				temp := ctx.State["regex_matched"].([]string)[1]
				AirConditTemp[ctx.Event.GroupID], _ = strconv.Atoi(temp)
				ctx.SendChain(message.Text(
					"❄️风速中", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			} else {
				ctx.SendChain(message.Text(
					"💤", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			}
		})
	engine.OnFullMatch(`群温度`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if _, exist := AirConditTemp[ctx.Event.GroupID]; !exist {
				AirConditTemp[ctx.Event.GroupID] = 26
			}
			if AirConditSwitch[ctx.Event.GroupID] {
				ctx.SendChain(message.Text(
					"❄️风速中", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			} else {
				ctx.SendChain(message.Text(
					"💤", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			}
		})
}

/*
func pokereply(ctx *zero.Ctx, nickname string) {
	ctx.SendChain(message.Text(
		[]string{
			"请不要戳" + nickname + " >_<",
			"喂(#`O′) 戳" + nickname + "干嘛!",
			"别戳了…痒……",
			"呜…别戳了…",
			"别戳了！",
			"喵~",
			"…把手拿开",
			"戳回去<( ￣^￣)",
			"有笨蛋在戳我，我不说是谁",
		}[rand.Intn(9)],
	))
}
*/
