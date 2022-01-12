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

const (
	dbpath = "data/Chat/"
	dbfile = dbpath + "kimoi.json"
	prio   = 10
)

var (
	poke   = rate.NewManager(time.Minute*5, 8) // 戳一戳
	engine = control.Register("chat", &control.Options{
		DisableOnDefault: false,
		Help:             "chat\n- [BOT名字]\n- [戳一戳BOT]\n- 空调开\n- 空调关\n- 群温度\n- 设置温度[正整数]\n- mua|啾咪|摸|上你|傻|裸|贴|老婆|抱|亲|一下|咬|操|123|进去|调教|搓|让|捏|挤|略|呐|原味|胖次|内裤|内衣|衣服|ghs|批|憨批|kkp|咕|骚|喜欢|suki|好き|看|不能|砸了|透|口我|草我|自慰|onani|オナニー|炸了|色图|涩图|告白|对不起|回来|吻|软|壁咚|掰开|女友|是|喵|嗷呜|叫|拜|佬|awsl|臭|香|腿|张开|脚|脸|头发|手|pr|舔|小穴|腰|诶嘿嘿|可爱|扭蛋|鼻|眼|色气|推|床|举|手冲|饿|变|敲|爬|怕|冲|射|不穿|迫害|猫粮|揪尾巴|薄荷|早|晚安|揉|榨|掐|胸|奶子|欧派|嫩|蹭|牵手|握手|拍照|w|睡不着|欧尼酱|哥|爱你|过来|自闭|打不过|么么哒|很懂|膝枕|累了|安慰|洗澡|一起睡觉|一起|多大|姐姐|糖|嗦|牛子|🐂子|🐮子|嫌弃|紧|baka|笨蛋|插|插进来|屁股|翘|翘起来|抬|抬起|爸|傲娇|rua|咕噜咕噜|咕噜|上床|做爱|吃掉|吃|揪|种草莓|种草|掀|妹|病娇|嘻",
	})
	kimomap  = make(kimo, 256)
	chatList = make([]string, 0, 256)
)

func init() { // 插件主体
	// 被喊名字
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).SetPriority(prio).
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
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			switch {
			case poke.Load(ctx.Event.GroupID).AcquireN(3):
				// 5分钟共8块命令牌 一次消耗3块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("请不要戳", nickname, " >_<"))
			case poke.Load(ctx.Event.GroupID).Acquire():
				// 5分钟共8块命令牌 一次消耗1块命令牌
				time.Sleep(time.Second * 1)
				ctx.SendChain(message.Text("喂(#`O′) 戳", nickname, "干嘛！"))
			default:
				// 频繁触发，不回复
			}
		})
	// 群空调
	var AirConditTemp = map[int64]int{}
	var AirConditSwitch = map[int64]bool{}
	engine.OnFullMatch("空调开").SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = true
			ctx.SendChain(message.Text("❄️哔~"))
		})
	engine.OnFullMatch("空调关").SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			AirConditSwitch[ctx.Event.GroupID] = false
			delete(AirConditTemp, ctx.Event.GroupID)
			ctx.SendChain(message.Text("💤哔~"))
		})
	engine.OnRegex(`设置温度(\d+)`).SetBlock(true).SetPriority(prio).
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
	engine.OnFullMatch(`群温度`).SetBlock(true).SetPriority(prio).
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
	initChatList(func() {
		engine.OnFullMatchGroup(chatList, zero.OnlyToMe).SetBlock(true).SetPriority(prio).Handle(
			func(ctx *zero.Ctx) {
				key := ctx.MessageString()
				val := *kimomap[key]
				text := val[rand.Intn(len(val))]
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
			})
	})
}
