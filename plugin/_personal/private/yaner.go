package yaner

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	poke = rate.NewManager[int64](time.Minute*5, 6) // 戳一戳
	// Axis表情的 codechina 镜像
	res = "https://gitcode.net/weixin_49234624/zbpdata/-/raw/main/faceimg-liuyu/"
)

func init() { // 插件主体
	// 电脑状态
	zero.OnFullMatchGroup([]string{"检查身体", "自检", "启动自检", "系统状态"}, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent(),
			),
			)
		})
	// 重启
	zero.OnFullMatchGroup([]string{"重启", "restart", "kill", "洗手手"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			os.Exit(0)
		})
	// 运行 CQ 码
	zero.OnPrefix("run", zero.SuperUserPermission).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			// 可注入，权限为主人
			ctx.Send(message.UnescapeCQCodeText(ctx.State["args"].(string)))
		})
	// 撤回最后的发言
	zero.OnRegex(`^\[CQ:reply,id=(.*)].*`, zero.KeywordRule("多嘴")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取消息id
			mid := ctx.State["regex_matched"].([]string)[1]
			// 撤回消息
			if ctx.Event.Message[1].Data["qq"] != "" {
				var nickname = zero.BotConfig.NickName[0]
				ctx.SendChain(message.Text("9494，要像", nickname, "一样乖乖的才行哟~"))
			} else {
				ctx.SendChain(message.Text("呜呜呜呜"))
			}
			ctx.DeleteMessage(message.NewMessageIDFromString(mid))
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
		})
	engine := control.Register("yaner", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "基础指令",
		Help:             "柳如娮的基础指令",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"检测到唤醒环境:\n",
				"* CPU占用: ", cpuPercent(), "%\n",
				"* RAM占用: ", memPercent(), "%\n",
				"* 硬盘使用: ", diskPercent(), "\n确认ok。\n",
			))
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("嘿嘿，娮儿闪亮登场！锵↘锵↗~"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			ctx.SendChain(message.Text("宝↗生↘永↗梦↘！！！！"))
		},
	})
	// 被喊名字
	engine.OnKeywordGroup([]string{"自我介绍", "你是谁", "你谁"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("你好，我叫柳如娮。\n你可以叫我娮儿、小娮，当然你叫我机器人也可以ಠಿ_ಠ"))
		})
	engine.OnFullMatch("", zero.OnlyToMe).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			var nickname = zero.BotConfig.NickName[0]
			time.Sleep(time.Second * 1)
			switch rand.Intn(3) {
			case 1:
				ctx.SendChain(randImage("WZ.jpg", "ZZZZ.gif"))
			default:
				ctx.SendChain(message.Text(
					[]string{
						nickname + "在窥屏哦",
						"我在听",
						"请问找" + nickname + "有什么事吗",
						"？怎么了",
					}[rand.Intn(4)],
				))
			}
		})
	// 戳一戳
	engine.On("notice/notify/poke", zero.OnlyToMe).SetBlock(false).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if !poke.Load(ctx.Event.GroupID).AcquireN(1) {
				return // 最多戳6次
			}
			nickname := zero.BotConfig.NickName[0]
			switch rand.Intn(3) {
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
					"？",
				))
			}
		})
	/*/ 石头剪刀布
	engine.OnFullMatchGroup([]string{"剪刀", "石头", "布"}).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			txt := ctx.Event.Message[0].Data["text"]
			var rand_str = []string{"剪刀", "石头", "布"}
			result := rand_str[rand.Intn(len(rand_str))]
			ctx.SendChain(message.Text(result))
			process.SleepAbout1sTo2s()
			nickname := zero.BotConfig.NickName[0]
			if rand.Intn(3) != 1 {
				if txt == result {
					ctx.SendChain(randText(
						"再来！",
						"。。。。。。",
						"你的实力，"+nickname+"认可了！(o゜▽゜)o☆",
					))
				} else if (txt == "剪刀" && result == "石头") || (txt == "石头" && result == "布") || (txt == "布" && result == "剪刀") {
					ctx.SendChain(randText(
						"阿这(´･ω･`)?",
						nickname+"的猜拳有这么厉害吗(＃°Д°)",
						"唉？我居然赢了！（＾∀＾●）ﾉｼ",
						"哼恩！~我赢了！",
					))
				} else {
					ctx.SendChain(randText(
						"再来！",
						"我好菜啊ಠಿ_ಠ",
						ctx.CardOrNickName(ctx.Event.UserID)+"tql",
						"呜呜呜！(。>︿<)_θ",
					))
				}
			}
		})
	engine.OnKeyword("help").SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		if rand.Intn(2) == 0 {
			ctx.SendChain(message.Image(res + "Help.jpg"))
		}
	})
	*/
	engine.OnKeywordGroup([]string{"好吗", "行不行", "能不能", "可不可以"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			process.SleepAbout1sTo2s()
			switch rand.Intn(4) {
			case 0:
				nickname := zero.BotConfig.NickName[0]
				if rand.Intn(2) == 0 {
					ctx.SendChain(message.Text(nickname + "..." + nickname + "觉得不行"))
				} else {
					ctx.SendChain(message.Text(nickname + "..." + nickname + "觉得可以！"))
				}
			case 1:
				ctx.SendChain(randImage("Ask-YES.jpg", "Ask-NO.jpg", "Ask-YES.jpg"))
			}
		})
}

func randText(text ...string) message.MessageSegment {
	return message.Text(text[rand.Intn(len(text))])
}

func randImage(file ...string) message.MessageSegment {
	return message.Image(res + file[rand.Intn(len(file))])
}

func cpuPercent() float64 {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return -1
	}
	return math.Round(percent[0])
}

func memPercent() float64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return -1
	}
	return math.Round(memInfo.UsedPercent)
}

func diskPercent() string {
	parts, err := disk.Partitions(true)
	if err != nil {
		return err.Error()
	}
	msg := ""
	for _, p := range parts {
		diskInfo, err := disk.Usage(p.Mountpoint)
		if err != nil {
			msg += "\n  - " + err.Error()
			continue
		}
		pc := uint(math.Round(diskInfo.UsedPercent))
		if pc > 0 {
			msg += fmt.Sprintf("\n  - %s(%dM) %d%%", p.Mountpoint, diskInfo.Total/1024/1024, pc)
		}
	}
	return msg
}
