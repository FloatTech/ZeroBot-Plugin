// Package warframeapi 星际战甲
package warframeapi

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/RomiChan/syncx"
	"github.com/lithammer/fuzzysearch/fuzzy"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	eng := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "星际战甲",
		Help: "- wf时间同步\n" +
			"- [金星|地球|火卫二]平原时间\n" +
			"- .wm [物品名称]\n" +
			"- wf仲裁\n" +
			"- wf警报\n" +
			"- wf每日特惠",
		PrivateDataFolder: "warframeapi",
	})

	// 获取具体的平原时间, 在触发后, 会启动持续时间按5分钟的时间更新模拟, 以此处理短时间内请求时, 时间不会变化的问题
	eng.OnSuffix("平原时间").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if !gameWorld.hasSync() { // 没有进行同步,就拉取一次服务器状态
				wfapi, err := newwfapi()
				if err != nil {
					ctx.SendChain(message.Text("ERROR: 获取服务器时间失败"))
				}
				gameWorld.refresh(&wfapi)
			}
			var msg any
			switch ctx.State["args"].(string) {
			case "地球", "夜灵":
				msg = gameWorld.w[0]
			case "金星", "奥布山谷":
				msg = gameWorld.w[1]
			case "魔胎之境", "火卫二", "火卫":
				msg = gameWorld.w[2]
			default:
				msg = "ERROR: 平原不存在"
			}
			ctx.SendChain(message.Text(msg))
			// 是否正在进行同步,没有就开启同步,有就不开启
			if !gameWorld.hasSync() {
				if gameWorld.setsync() {
					go func() {
						// 30*10=300=5分钟
						for i := 0; i < 30; i++ {
							time.Sleep(10 * time.Second)
							gameWorld.update() // 5分钟内每隔10秒更新一下时间
						}
						// 5分钟时间同步结束
						_ = gameWorld.resetsync()
					}()
				}
			}
		})
	eng.OnFullMatch("wf警报").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := newwfapi()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 如果返回的wfapi中, 警报数量>0
			if len(wfapi.Alerts) > 0 {
				msgs := make(message.Message, len(wfapi.Alerts))
				// 遍历警报数据, 打印警报信息
				for i, v := range wfapi.Alerts {
					msgs[i] = ctxext.FakeSenderForwardNode(ctx, message.Text(
						"激活: ", v.Active,
						"\n节点: ", v.Mission.Node,
						"\n类型: ", v.Mission.Type,
						"\n敌人等级: ", v.Mission.MinEnemyLevel, "~", v.Mission.MaxEnemyLevel,
						"\n奖励: ", v.Mission.Reward.AsString,
						"\n剩余时间:", v.Eta))
				}
				ctx.SendChain(msgs...)
			}
		})
	//TODO:订阅功能-等待重做
	// eng.OnRegex(`^(订阅|取消订阅)(.*)平原(.*)$`).SetBlock(true).
	//	Handle(func(ctx *zero.Ctx) {
	//		args := ctx.State["regex_matched"].([]string)
	//		var isEnable bool
	//		if args[1] == "订阅" {
	//			isEnable = true
	//		}
	//		updateWFAPI()
	//		status := false
	//		switch args[3] {
	//		case "fass", "白天", "温暖":
	//			status = true
	//		}
	//		switch args[2] {
	//		case "金星", "奥布山谷":
	//			//sublist = append(sublist, subList{ctx.Event.GroupID, ctx.Event.UserID, 1, status, false})
	//			if isEnable {
	//				addUseSub(ctx.Event.UserID, ctx.Event.GroupID, 1, status)
	//			} else {
	//				removeUseSub(ctx.Event.UserID, ctx.Event.GroupID, 1)
	//			}
	//			ctx.SendChain(
	//				message.At(ctx.Event.UserID),
	//				message.Text("已成功", args[1]),
	//				message.Text(gameTimes[1].Name),
	//				message.Text(status),
	//			)
	//		case "地球", "夜灵":
	//			if isEnable {
	//				addUseSub(ctx.Event.UserID, ctx.Event.GroupID, 0, status)
	//			} else {
	//				removeUseSub(ctx.Event.UserID, ctx.Event.GroupID, 0)
	//			}
	//			ctx.SendChain(
	//				message.At(ctx.Event.UserID),
	//				message.Text("已成功", args[1]),
	//				message.Text(gameTimes[0].Name),
	//				message.Text(status),
	//			)
	//		case "魔胎之境", "火卫", "火卫二":
	//			if isEnable {
	//				addUseSub(ctx.Event.UserID, ctx.Event.GroupID, 2, status)
	//			} else {
	//				removeUseSub(ctx.Event.UserID, ctx.Event.GroupID, 2)
	//			}
	//			ctx.SendChain(
	//				message.At(ctx.Event.UserID),
	//				message.Text("已成功", args[1]),
	//				message.Text(gameTimes[2].Name),
	//				message.Text(status),
	//			)
	//		default:
	//			ctx.SendChain(message.Text("ERROR: 平原不存在"))
	//			return
	//		}
	//	})
	// eng.OnFullMatch(`wf订阅检测`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
	//	rwm.Lock()
	//	var msg []message.MessageSegment
	//	for i, v := range gameTimes {
	//		nt := time.Until(v.NextTime).Seconds()
	//		switch {
	//		case nt < 0:
	//			if v.Status {
	//				v.NextTime = v.NextTime.Add(time.Duration(v.NightTime) * time.Second)
	//			} else {
	//				v.NextTime = v.NextTime.Add(time.Duration(v.DayTime) * time.Second)
	//			}
	//			v.Status = !v.Status
	//
	//			msg = callUser(i, v.Status, 0)
	//		case nt < float64(5)*60:
	//			msg = callUser(i, !v.Status, 5)
	//		case nt < float64(15)*60:
	//			if i == 2 && !v.Status {
	//				return
	//			}
	//			msg = callUser(i, !v.Status, 15)
	//		}
	//	}
	//	rwm.Unlock()
	//	if msg != nil && len(msg) > 0 {
	//		ctx.SendChain(msg...)
	//	}
	// })
	eng.OnFullMatch("wf仲裁").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 通过wfapi获取仲裁信息
			wfapi, err := newwfapi()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text(
				"节点: ", wfapi.Arbitration.Node,
				"\n类型: ", wfapi.Arbitration.Type,
				"\n阵营: ", wfapi.Arbitration.Enemy,
				"\n剩余时间: ", int(wfapi.Arbitration.Expiry.Sub(time.Now().UTC()).Minutes()), "m",
			))
		})
	eng.OnFullMatch("wf每日特惠").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := newwfapi()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(wfapi.DailyDeals) > 0 {
				msgs := make(message.Message, len(wfapi.DailyDeals))
				for i, dd := range wfapi.DailyDeals {
					msgs[i] = ctxext.FakeSenderForwardNode(ctx, message.Text(
						"物品: ", dd.Item,
						"\n价格: ", dd.OriginalPrice, "→", dd.SalePrice,
						"\n数量: (", dd.Total, "/", dd.Sold, ")",
						"\n时间: ", dd.Eta,
					))
				}
				ctx.SendChain(msgs...)
			}
		})
	// eng.OnRegex(`^入侵$`).SetBlock(true).
	// 	Handle(func(ctx *zero.Ctx) {
	// 		updateWFAPI(ctx)
	// 		for _, dd := range wfapi.dailyDeals {
	// 			imagebuild.DrawTextSend([]string{
	// 				"节点:" + wfapi.arbitration.Node,
	// 				"类型:" + wfapi.arbitration.Type,
	// 				"阵营:" + wfapi.arbitration.Enemy,
	// 				"剩余时间:" + fmt.Sprint(int(wfapi.arbitration.Expiry.Sub(time.Now().UTC()).Minutes())) + "m",
	// 			}, ctx)
	// 		}
	// 	})
	eng.OnFullMatch("wf时间同步").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := newwfapi()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			gameWorld.refresh(&wfapi)
			ctx.SendChain(message.Text("已拉取服务器时间并同步到本地模拟"))
		})
	// 根据名称从Warframe市场查询物品售价
	eng.OnPrefix(".wm ", func(ctx *zero.Ctx) bool {
		if wd.Get().wmitems == nil || wd.Get().itemNames == nil {
			if wderr != nil { // 获取失败
				ctx.SendChain(message.Text("ERROR: 获取Warframe市场物品列表失败: ", wderr))
			} else {
				ctx.SendChain(message.Text("ERROR: Warframe市场物品列表为空!"))
			}
			wd = syncx.Lazy[*wmdata]{Init: func() (d *wmdata) {
				d, wderr = newwm()
				return
			}}
			return false
		}
		return true
	}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 根据输入的名称, 从游戏物品名称列表中进行模糊搜索
			sol := fuzzy.FindNormalizedFold(ctx.State["args"].(string), wd.Get().itemNames)
			// 物品名称
			var name string

			// 根据搜搜结果, 打印找到的物品
			switch len(sol) {
			case 0: // 没有搜索到任何东西
				ctx.SendChain(message.Text("无法查询到该物品"))
				return
			case 1: // 如果只搜索到了一个
				name = sol[0]
			default: // 如果搜搜到了多个
				sb := strings.Builder{}
				if len(sol) > 25 {
					sb.WriteString("数量过多, 只显示前25\n")
					sol = sol[:25]
				}
				sb.WriteString("[0] ")
				sb.WriteString(sol[0])
				for i, v := range sol[1:] {
					sb.WriteString("\n[")
					sb.WriteString(strconv.Itoa(i + 1))
					sb.WriteString("] ")
					sb.WriteString(v)
				}
				ctx.SendChain(
					ctxext.FakeSenderForwardNode(ctx, message.Text("包含多个结果, 请输入编号查看(30s内),输入c直接结束会话")),
					ctxext.FakeSenderForwardNode(ctx, message.Text(&sb)),
				)
				itemIndex := getitemnameindex(ctx)
				if itemIndex < 0 {
					return
				}
				if itemIndex >= len(sol) || itemIndex < 0 {
					ctx.SendChain(message.Text("ERROR: 编号超出范围"))
					return
				}
				name = sol[itemIndex]
			}
			onlymaxrank := false
			msgs := message.Message{}
		GETWM:
			if onlymaxrank {
				msgs = msgs[:0]
			}

			sells, iteminfo, txt, err := getitemsorder(wd.Get().wmitems[name].URLName, onlymaxrank)
			if !onlymaxrank {
				if iteminfo.ZhHans.WikiLink == "" {
					msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx,
						message.Image("https://warframe.market/static/assets/"+wd.Get().wmitems[name].Thumb),
						message.Text("\n", wd.Get().wmitems[name].ItemName)))
				} else {
					msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx,
						message.Image("https://warframe.market/static/assets/"+wd.Get().wmitems[name].Thumb),
						message.Text("\n", wd.Get().wmitems[name].ItemName, "\nwiki: ", iteminfo.ZhHans.WikiLink)))
				}
			}

			if err != nil {
				msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx, message.Text("ERROR: ", err)))
				ctx.SendChain(msgs...)
				return
			}
			if sells == nil {
				msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx, message.Text("无可购买对象")))
				ctx.SendChain(msgs...)
				return
			}

			ismod := iteminfo.ModMaxRank != 0
			max := 5
			if len(sells) < max {
				max = len(sells)
			}
			sb := strings.Builder{}
			if ismod {
				if !onlymaxrank {
					msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx, message.Text("请输入编号选择, 或输入r获取满级报价(30s内)\n输入c直接结束会话")))
				} else {
					msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx, message.Text("请输入编号选择(30s内)\n输入c直接结束会话")))
				}
				for i := 0; i < max; i++ {
					// msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx,
					//	message.Text(fmt.Sprintf("[%d] (Rank:%d/%d)  %dP - %s\n", i, sells[i].ModRank, iteminfo.ModMaxRank, sells[i].Platinum, sells[i].User.IngameName))))
					sb.WriteString(fmt.Sprintf("[%d] (Rank:%d/%d)  %dP - %s\n", i, sells[i].ModRank, iteminfo.ModMaxRank, sells[i].Platinum, sells[i].User.IngameName))
				}
			} else {
				for i := 0; i < max; i++ {
					// msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx,
					//	message.Text(fmt.Sprintf("[%d] %dP -%s\n", i, sells[i].Platinum, sells[i].User.IngameName))))
					sb.WriteString(fmt.Sprintf("[%d] %dP -%s\n", i, sells[i].Platinum, sells[i].User.IngameName))
				}
			}
			msgs = append(msgs, ctxext.FakeSenderForwardNode(ctx, message.Text(&sb)))
			ctx.SendChain(msgs...)

			for i := 0; i < 3; i++ {
				next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Next()
				select {
				case <-time.After(time.Second * 30):
					ctx.SendChain(message.Text("会话已结束!"))
					return
				case e := <-next:
					msg := e.Event.Message.ExtractPlainText()
					// 重新获取报价
					if msg == "r" {
						onlymaxrank = true
						goto GETWM
					}
					// 主动结束会话
					if msg == "c" {
						ctx.SendChain(message.Text("会话已结束!"))
						return
					}
					i, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字! (输入c结束会话)"))
						continue
					}
					if ismod {
						ctx.SendChain(message.Text("/w ", sells[i].User.IngameName, " Hi! I want to buy: ", txt, "(Rank:", sells[i].ModRank, ") for ", sells[i].Platinum, " platinum. (warframe.market)"))
					} else {
						ctx.SendChain(message.Text("/w ", sells[i].User.IngameName, " Hi! I want to buy: ", txt, " for ", sells[i].Platinum, " platinum. (warframe.market)"))
					}
					return
				}
			}
		})
}

// 获取搜索结果中的物品具体名称index的FutureEvent
//
//	传入ctx和一个递归次数上限,返回一个int
//	如果为返回内容为负, 说明
//	-1 会话超时
//	-2 主动结束
//	-3 连续3次错误
func getitemnameindex(ctx *zero.Ctx) int {
	recv, cancel := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Repeat()
	defer cancel()
	for i := 0; i < 3; i++ {
		select {
		case <-time.After(time.Second * 30):
			// 超时15秒处理
			ctx.SendChain(message.Text("会话已超时!"))
			return -1
		case e := <-recv:
			msg := e.Event.Message.ExtractPlainText()
			// 输入c主动结束的处理
			if msg == "c" {
				ctx.SendChain(message.Text("会话已结束!"))
				return -2
			}
			// 尝试对输入进行数字转换
			num, err := strconv.Atoi(msg)
			if err != nil {
				ctx.SendChain(message.Text("请输入数字! (输入c结束会话)"))
				continue
			}
			return num
		}
	}
	ctx.SendChain(message.Text("连续输入错误, 会话已结束!"))
	return -3
}
