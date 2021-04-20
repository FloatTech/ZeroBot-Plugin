package huahua

import (
	"math/rand"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	PRIO   = -1
	ENABLE = true
	RES    = "https://raw.dihe.moe/Yiwen-Chan/ZeroBot-Plugin/master/huahua/"
)

func init() {
	zero.OnFullMatch("椛椛醒醒", zero.AdminPermission).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ENABLE = true
			ctx.SendChain(randText("啊……好困啊……"))
		})
	zero.OnFullMatch("椛椛睡吧", zero.AdminPermission).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ENABLE = false
			ctx.SendChain(randText("那晚安咯……"))
		})
	zero.OnRegex("^。{1,6}$", HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			count := strings.Count(ctx.Event.Message.CQString(), "。")
			ctx.SendChain(message.Text(
				[]string{
					"一个泡泡。。",
					"两个泡泡。",
					"三个泡泡",
					"四个泡泡。。",
					"五个泡泡",
					"六个泡泡",
				}[count-1],
			))
		})
	zero.OnFullMatchGroup([]string{"！", "!"}, HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("一个感叹号"))
		})
	zero.OnFullMatchGroup([]string{"！？", "？！", "!?", "?!"}, HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("惊讶"))
		})
	zero.OnFullMatch("…", HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("不知道该说什么", "啊咧？"))
		})
	zero.OnKeyword("不记得", HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("忘川河流淌而过"))
		})
	zero.OnKeywordGroup([]string{"可怜", "好累啊"}, HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("给你抱抱", "摸摸头(๑•́ωก̀๑)"))
		})
	zero.OnKeyword("好冷", HuaHuaSwitch(), HuaHuaChance(80)).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("多穿点", "赶紧躺被窝"))
		})
	// 喊道椛椛
	zero.OnFullMatch("啾啾", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("啾啾"))
		})
	zero.OnFullMatch("乖", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("好饿……先请我吃饭饭"))
		})
	zero.OnKeyword("听话", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("我只听主人的话"))
		})
	zero.OnFullMatch("举高高", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("好重，举不动！"))
		})
	zero.OnKeyword("何在", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("我在二次元"))
		})
	zero.OnFullMatch("傲娇", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("哼"))
		})
	zero.OnFullMatch("卖萌", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("(´ฅω•ฅ｀)ﾁﾗｯ"))
		})
	zero.OnFullMatch("变身", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("人家才不是马猴烧酒呐！"))
		})
	zero.OnFullMatch("可爱", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("٩(๑òωó๑)۶"))
		})
	zero.OnKeyword("吃饭了吗", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("等你投食"))
		})
	zero.OnKeyword("吃鱼", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("姬气人吃不得这些东西"))
		})
	zero.OnKeyword("唱歌", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("不唱"))
		})
	zero.OnKeyword("在不在", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("好啦好啦"))
		})
	zero.OnKeyword("坏掉了", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("我不是我没有"))
		})
	zero.OnKeyword("夸人", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("真是太厉害了呢"))
		})
	zero.OnKeyword("女仆模式", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("狗修金sama"))
		})
	zero.OnKeyword("抱抱", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("请不要乱摸"))
		})
	zero.OnKeyword("揉揉", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("不要不可以"))
		})
	zero.OnKeyword("摸头", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("不行，你摸不到"))
		})
	zero.OnKeyword("摸摸", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("lsp走开"))
		})
	zero.OnKeyword("放屁", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("我只是个姬气人，你想多了"))
		})
	zero.OnKeyword("数数", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("就不数"))
		})
	zero.OnKeyword("爱我", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("才不爱你呢"))
		})
	zero.OnKeyword("被玩坏了", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("你才被玩坏了"))
		})
	zero.OnKeyword("跳舞", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("跳你妹，我只是姬气人"))
		})
	zero.OnKeyword("过肩摔", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("椛椛不会"))
		})
	zero.OnKeyword("还能塞得下", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("凑不要脸"))
		})
	zero.OnKeyword("钉宫三连", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("八嘎hentai无路赛"))
		})
	zero.OnKeyword("攻略", HuaHuaSwitch(), HuaHuaChance(80), zero.OnlyToMe).SetBlock(true).SetPriority(PRIO).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(randText("lsp别想泡我"))
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

// HuaHuaSwitch 控制 HuaHua 的开关
func HuaHuaSwitch() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		return ENABLE
	}
}

// HuaHuaChance 概率响应，输入回应 1-100
func HuaHuaChance(percent int) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		if r := rand.Intn(100); r <= percent-1 {
			return true
		}
		return false
	}
}
