// Package coc coc插件
package coc

import (
	"math/rand"
	"strconv"

	"github.com/FloatTech/floatbox/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine.OnRegex(`^(.|。)(r|R)([1-9]\d*)?(d|D)?([1-9]\d*)?( (.*))?$`, getsetting).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		defaultDice := 100
		cocSetting := settingGoup[ctx.Event.GroupID]
		if ctx.State["regex_matched"].([]string)[5] == "" && cocSetting.DefaultDice != 0 {
			defaultDice = cocSetting.DefaultDice
		} else if ctx.State["regex_matched"].([]string)[5] != "" {
			defaultDice, _ = strconv.Atoi(ctx.State["regex_matched"].([]string)[5])
		}
		times := 1
		if ctx.State["regex_matched"].([]string)[3] != "" {
			times, _ = strconv.Atoi(ctx.State["regex_matched"].([]string)[3])
		}
		msg := make(message.Message, 0, 2+times*2)
		msg = append(msg, message.Reply(ctx.Event.MessageID))
		if ctx.State["regex_matched"].([]string)[7] != "" {
			msg = append(msg, message.Text("因为", ctx.State["regex_matched"].([]string)[7], "进行了\n🎲 => "))
		} else {
			msg = append(msg, message.Text("🎲 => "))
		}
		sum := 0
		for i := times; i > 0; i-- {
			dice := rand.Intn(defaultDice) + 1
			sum += dice
			if i != 1 {
				msg = append(msg, message.Text(dice, " + "))
			} else {
				msg = append(msg, message.Text(dice))
			}
		}
		msg = append(msg, message.Text(" = ", sum, diceRule(cocSetting.DiceRule, sum, defaultDice*times/2, defaultDice*times, 0)))
		ctx.Send(msg)
	})
	engine.OnRegex(`^(.|。)(r|R)([1-9]\d*)?(d|D)?([1-9]\d*)?a(\S+)( (.*))?$`, getsetting).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		cocSetting := settingGoup[ctx.Event.GroupID]
		infoFile := engine.DataFolder() + strconv.FormatInt(gid, 10) + "/" + strconv.FormatInt(uid, 10) + ".json"
		if file.IsNotExist(infoFile) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请先创建角色"))
			return
		}
		cocInfo, err := loadPanel(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		getInfo := false
		thresholdMax := 100
		threshold := 50
		thresholdMin := 0
		for _, info := range cocInfo.Attribute {
			if ctx.State["regex_matched"].([]string)[6] == info.Name {
				getInfo = true
				thresholdMax = info.MaxValue
				threshold = info.Value
				thresholdMin = info.MinValue
			}
		}
		if !getInfo {
			ctx.SendChain(message.Text("[ERROR]:参数错误"))
			return
		}
		defaultDice := 100
		if ctx.State["regex_matched"].([]string)[5] == "" && cocSetting.DefaultDice != 0 {
			defaultDice = cocSetting.DefaultDice
		} else if ctx.State["regex_matched"].([]string)[5] != "" {
			defaultDice, err = strconv.Atoi(ctx.State["regex_matched"].([]string)[5])
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:骰子面数参数错误"))
				return
			}
		}
		limit := defaultDice * (threshold - thresholdMin) / (thresholdMax - thresholdMin)
		times := 1
		if ctx.State["regex_matched"].([]string)[3] != "" {
			times, _ = strconv.Atoi(ctx.State["regex_matched"].([]string)[3])
		}
		msg := make(message.Message, 0, 2+times)
		msg = append(msg, message.Reply(ctx.Event.MessageID))
		if ctx.State["regex_matched"].([]string)[8] != "" {
			msg = append(msg, message.Text("因为", ctx.State["regex_matched"].([]string)[8], "用", threshold, "的", ctx.State["regex_matched"].([]string)[6], "进行了\n"))
		}
		sum := 0
		for i := times; i > 0; i-- {
			dice := rand.Intn(defaultDice) + 1
			msg = append(msg, message.Text("🎲 => ", dice, diceRule(cocSetting.DiceRule, dice, limit, defaultDice, 0), "\n"))
			sum += dice
		}
		if times > 1 {
			msg = append(msg, message.Text("合计 = ", sum, diceRule(cocSetting.DiceRule, sum, limit*times, defaultDice*times, 0)))
		}
		ctx.Send(msg)
	})
}

func diceRule(ruleType, dice, decision, maxDice, minDice int) string {
	// 50的位置
	halflimit := float64(maxDice-minDice) / 2
	// 大成功值范围
	tenStrike := float64(maxDice-minDice) * 6 / 100
	// 成功值范围
	limit := float64(decision - minDice)
	// 大失败值范围
	fiasco := float64(maxDice-minDice) * 95 / 100
	// 骰子数
	piece := float64(dice)
	switch ruleType {
	case 1:
		switch {
		case (piece == 1 && limit < halflimit) || (limit >= halflimit && piece < tenStrike):
			return "(大成功!)"
		case (piece > fiasco && limit < halflimit) || (limit >= halflimit && dice == maxDice):
			return "(大失败!)"
		case piece <= fiasco && piece > limit:
			return "(失败)"
		case piece <= limit/2 && piece > limit/5:
			return "(困难成功)"
		case piece <= limit/5 && piece > 1:
			return "(极难成功)"
		default:
			return "(成功)"
		}
	case 2:
		switch {
		case piece < tenStrike:
			return "(大成功!)"
		case piece > fiasco:
			return "(大失败!)"
		case piece <= fiasco && piece > limit:
			return "(失败)"
		case piece <= limit/2 && piece > limit/5:
			return "(困难成功)"
		case piece <= limit/5 && piece > 1:
			return "(极难成功)"
		default:
			return "(成功)"
		}
	default:
		switch {
		case piece == 1:
			return "(大成功!)"
		case (piece > fiasco && limit < halflimit) || (limit >= halflimit && dice == maxDice):
			return "(大失败!)"
		case piece <= fiasco && piece > limit:
			return "(失败)"
		case piece <= limit/2 && piece > limit/5:
			return "(困难成功)"
		case piece <= limit/5 && piece > 1:
			return "(极难成功)"
		default:
			return "(成功)"
		}
	}
}
