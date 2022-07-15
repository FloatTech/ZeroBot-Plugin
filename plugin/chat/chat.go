// Package chat 对话插件
package chat

import (
	"math/rand"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	poke   = rate.NewManager[int64](time.Minute*5, 8) // 戳一戳
	engine = control.Register("chat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "chat\n- [BOT名字]\n- [戳一戳BOT]\n- 空调开\n- 空调关\n- 群温度\n- 设置温度[正整数]",
	})
)

func init() { // 插件主体
	// 被喊名字
	engine.OnFullMatch("【蛇】", zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			time.Sleep(time.Second * 1)
			ctx.SendChain(message.Text(
				[]string{
					"找人家有什么事吗？我可爱的小白鼠~？",
					"你是在找我吗？我可爱的小白鼠？",
					"呵呵~ 小白鼠，想和我来一起做些有趣的事情吗？",
					"我就是梅比乌斯~ 唯一的，真正的梅比乌斯~",
				}[rand.Intn(4)],
			))
		})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			switch {
			case poke.Load(ctx.Event.GroupID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText(
					"呵呵~ 有点痒呢~",
					"小白鼠~ 是想对人家做些什么吗？",
					"哎呀呀，我可爱的小白鼠~ 想去我的实验室坐坐吗？",
				))
			case poke.Load(ctx.Event.GroupID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(randText(
					"又怎么了？小白鼠~ 没什么事的话，不要来烦我",
					"怎么，小白鼠，你已经这么闲了吗？",
				))
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
			ctx.SendChain(randText(
				"（❄️哔~）小白鼠~ 把温度调低一点哦~",
				"（❄️哔~）蛇在低温的时候会冬眠~ 想试试吗，小白鼠？",
			))
		})
	engine.OnFullMatch("空调关").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = false
			delete(AirConditTemp, ctx.Event.GroupID)
			ctx.SendChain(randText(
				"（💤哔~）有点热呢……小白鼠~ 能把空调打开吗？拜托了~",
				"（💤哔~）蛇不喜欢在热的地方逗留，你明白吗~？",
			))
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
					"（❄️风速中）有点热呢……小白鼠~ 能把空调调低一点吗？拜托了~", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			} else {
				ctx.SendChain(message.Text(
					"（💤）蛇不喜欢在热的地方逗留，你明白吗~？", "\n",
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
					"（❄️风速中）有点热呢……小白鼠~ 能把空调调低一点吗？拜托了~", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			} else {
				ctx.SendChain(message.Text(
					"（💤）蛇不喜欢在热的地方逗留，你明白吗~？", "\n",
					"群温度 ", AirConditTemp[ctx.Event.GroupID], "℃",
				))
			}
		})
}
