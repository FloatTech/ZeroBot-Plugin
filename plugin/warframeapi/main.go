// Package warframeapi 百度内容审核
package warframeapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/lithammer/fuzzysearch/fuzzy"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	wmitems   map[string]items // WarFrame市场的中文名称对应的物品的字典
	itmeNames []string         // 物品名称列表
	rt        runtime
)

// 时间同步状态
type runtime struct {
	rwm    sync.RWMutex
	enable bool // 是否启动
}

const wfapiurl = "https://api.warframestat.us/pc"        // 星际战甲API
const wfitemurl = "https://api.warframe.market/v1/items" // 星际战甲游戏品信息列表URL

func init() {
	eng := control.Register("warframeapi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "warframeapi\n" +
			"- wf时间同步\n" +
			"- [金星|地球|火卫二]平原时间\n" +
			"- .wm [物品名称]\n" +
			"- 仲裁\n" +
			"- 警报\n" +
			"- 每日特惠",
		PrivateDataFolder: "warframeapi",
	})
	updateWM()

	// 获取具体的平原时间，在触发后，会启动持续时间按5分钟的时间更新模拟，以此处理短时间内请求时，时间不会变化的问题
	eng.OnSuffix("平原时间").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if !rt.enable { // 没有进行同步,就拉取一次服务器状态
				wfapi, err := wfapiGetData()
				if err != nil {
					ctx.SendChain(message.Text("Error:获取服务器时间失败"))
				}
				loadTime(wfapi)
			}
			switch ctx.State["args"].(string) {
			case "地球", "夜灵":
				ctx.SendChain(message.Text(gameTimes[0]))
			case "金星", "奥布山谷":
				ctx.SendChain(message.Text(gameTimes[1]))
			case "魔胎之境", "火卫二", "火卫":
				ctx.SendChain(message.Text(gameTimes[2]))
			default:
				ctx.SendChain(message.Text("ERROR: 平原不存在"))
			}
			// 是否正在进行同步,没有就开启同步,有就不开启
			if !rt.enable {
				// 设置标志位
				rt.rwm.Lock()
				if rt.enable { // 预检测，防止其他线程同时进来
					return
				}
				rt.enable = true
				rt.rwm.Unlock()

				go func() {
					// 30*10=300=5分钟
					for i := 0; i < 30; i++ {
						time.Sleep(10 * time.Second)
						timeDet() // 5分钟内每隔10秒更新一下时间
					}
					// 5分钟时间同步结束
					rt.rwm.Lock()
					rt.enable = false
					rt.rwm.Unlock()
				}()
			}
		})
	eng.OnFullMatch("警报").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := wfapiGetData()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			// 如果返回的wfapi中，警报数量>0
			if len(wfapi.Alerts) > 0 {
				// 遍历警报数据，打印警报信息
				for _, v := range wfapi.Alerts {
					// 如果警报处于激活状态
					if v.Active {
						ctx.SendChain(stringArrayToImage([]string{
							"节点:" + v.Mission.Node,
							"类型:" + v.Mission.Type,
							"敌人Lv:" + fmt.Sprint(v.Mission.MinEnemyLevel) + "~" + fmt.Sprint(v.Mission.MaxEnemyLevel),
							"奖励:" + v.Mission.Reward.AsString,
							"剩余时间:" + v.Eta,
						}))
					}
				}
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
	eng.OnFullMatch("仲裁").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 通过wfapi获取仲裁信息
			wfapi, err := wfapiGetData()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			ctx.SendChain(stringArrayToImage([]string{
				"节点:" + wfapi.Arbitration.Node,
				"类型:" + wfapi.Arbitration.Type,
				"阵营:" + wfapi.Arbitration.Enemy,
				"剩余时间:" + fmt.Sprint(int(wfapi.Arbitration.Expiry.Sub(time.Now().UTC()).Minutes())) + "m",
			}))
		})
	eng.OnFullMatch("每日特惠").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := wfapiGetData()

			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			for _, dd := range wfapi.DailyDeals {
				ctx.SendChain(
					message.Text(
						"物品:", dd.Item, "\n",
						"价格:", dd.OriginalPrice, "→", dd.SalePrice, "\n",
						"数量:(", dd.Total, "/", dd.Sold, ")\n",
						"时间:", dd.Eta,
					),
				)
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
			wfapi, err := wfapiGetData()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			loadTime(wfapi)
			ctx.SendChain(message.Text("已拉取服务器时间并同步到本地模拟"))
		})
	// 根据名称从Warframe市场查询物品售价
	eng.OnPrefix(".wm ").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 根据输入的名称，从游戏物品名称列表中进行模糊搜索
			sol := fuzzy.FindNormalizedFold(ctx.State["args"].(string), itmeNames)
			var msg []string
			// 物品名称
			var name string

			// 根据搜搜结果，打印找到的物品
			switch len(sol) {
			case 0: // 没有搜索到任何东西
				ctx.SendChain(message.Text("无法查询到该物品"))
				return
			case 1: // 如果只搜索到了一个
				name = sol[0]
			default: // 如果搜搜到了多个
				// 遍历搜索结果，并打印为图片展出
				for i, v := range sol {
					msg = append(msg, fmt.Sprintf("[%d] %s", i, v))
				}
				msg = append(msg, "包含多个结果，请输入编号查看(15s内),输入c直接结束会话")
				ctx.SendChain(stringArrayToImage(msg))
				msg = []string{}

				itemIndex := itemNameFutureEvent(ctx, 2)
				if itemIndex == -1 {
					return
				}
				name = sol[itemIndex]
			}
			Mf := false
		GETWM:
			if Mf {
				msg = []string{}
			}
			sells, itmeinfo, txt, err := wmItemOrders(wmitems[name].URLName, Mf)
			if !Mf {
				if itmeinfo.ZhHans.WikiLink == "" {
					ctx.Send([]message.MessageSegment{
						message.Image("https://warframe.market/static/assets/" + wmitems[name].Thumb),
						message.Text(wmitems[name].ItemName, "\n"),
					})
				} else {
					ctx.Send([]message.MessageSegment{
						message.Image("https://warframe.market/static/assets/" + wmitems[name].Thumb),
						message.Text(wmitems[name].ItemName, "\n"),
						message.Text("wiki:", itmeinfo.ZhHans.WikiLink),
					})
				}
			}

			msg = append(msg, wmitems[name].ItemName)

			if err != nil {
				ctx.Send(message.Text("Error:", err.Error()))
				return
			}
			if sells == nil {
				ctx.Send(message.Text("无可购买对象"))
				return
			}

			ismod := false
			if itmeinfo.ModMaxRank != 0 {
				ismod = true
			}

			max := 5
			if len(sells) <= max {
				max = len(sells)
			}
			for i := 0; i < max; i++ {
				if ismod {
					msg = append(msg, fmt.Sprintf("[%d](Rank:%d/%d)  %dP - %s\n", i, sells[i].ModRank, itmeinfo.ModMaxRank, sells[i].Platinum, sells[i].User.IngameName))
				} else {
					msg = append(msg, fmt.Sprintf("[%d] %dP -%s\n", i, sells[i].Platinum, sells[i].User.IngameName))
				}
			}

			if ismod && !Mf {
				msg = append(msg, "请输入编号选择，或输入r获取满级报价(30s内)\n输入c直接结束会话")
			} else {
				msg = append(msg, "请输入编号选择(30s内)\n输入c直接结束会话")
			}
			ctx.SendChain(stringArrayToImage(msg))

		GETNUM3:
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Next()
			select {
			case <-time.After(time.Second * 30):
				ctx.SendChain(message.Text("会话已结束!"))
				return
			case e := <-next:
				msg := e.Event.Message.ExtractPlainText()
				// 重新获取报价
				if msg == "r" {
					Mf = true
					goto GETWM
				}
				// 主动结束会话
				if msg == "c" {
					ctx.SendChain(message.Text("会话已结束!"))
					return
				}
				i, err := strconv.Atoi(msg)
				if err != nil {
					ctx.SendChain(message.Text("请输入数字!(输入c结束会话)"))
					goto GETNUM3
				}
				if err == nil {
					if ismod {
						ctx.Send(message.Text("/w ", sells[i].User.IngameName, " Hi! I want to buy: ", txt, "(Rank:", sells[i].ModRank, ") for ", sells[i].Platinum, " platinum. (warframe.market)"))
					} else {
						ctx.Send(message.Text("/w ", sells[i].User.IngameName, " Hi! I want to buy: ", txt, " for ", sells[i].Platinum, " platinum. (warframe.market)"))
					}
				}
			}
		})
}

// 获取搜索结果中的物品具体名称index的FutureEvent,传入ctx和一个递归次数上限,返回一个int，如果为返回内容为-1，说明会话超时，或主动结束，或超出递归
func itemNameFutureEvent(ctx *zero.Ctx, count int) int {
	next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Next()
	select {
	case <-time.After(time.Second * 15):
		// 超时15秒处理
		ctx.SendChain(message.Text("会话已超时!"))
		return -1
	case e := <-next:
		msg := e.Event.Message.ExtractPlainText()
		// 输入c主动结束的处理
		if msg == "c" {
			ctx.SendChain(message.Text("会话已结束!"))
			return -1
		}
		// 尝试对输入进行数字转换
		num, err := strconv.Atoi(msg)
		// 如果出错，说明输入的并非数字，则重新触发该内容
		if err != nil {
			// 查看是否超时
			if count == 0 {
				ctx.SendChain(message.Text("连续输入错误，会话已结束!"))
				return -1
			}
			ctx.SendChain(message.Text("请输入数字!(输入c结束会话)[", count, "]"))
			count--
			return itemNameFutureEvent(ctx, count)
		}
		return num
	}
}

// 数组字符串转图片
func stringArrayToImage(texts []string) message.MessageSegment {
	b, err := text.RenderToBase64(strings.Join(texts, "\n"), text.FontFile, 400, 20)
	if err != nil {
		return message.Text("ERROR: ", err)
	}
	return message.Image("base64://" + binary.BytesToString(b))
}

// 从WFapi获取数据
func wfapiGetData() (wfAPI, error) {
	var wfapi wfAPI // WarFrameAPI的数据实例
	var data []byte
	var err error
	data, err = web.GetData(wfapiurl)
	if err != nil {
		return wfapi, err
	}
	err = json.Unmarshal(data, &wfapi)
	if err != nil {
		return wfapi, err
	}
	return wfapi, nil
}

// 从WF市场获取物品数据信息
func updateWM() {
	var itmeapi wfAPIItem // WarFrame市场的数据实例

	data, err := web.RequestDataWithHeaders(&http.Client{}, wfitemurl, "GET", func(request *http.Request) error {
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Language", "zh-hans")
		return nil
	}, nil)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &itmeapi)
	if err != nil {
		panic(err)
	}
	loadToFuzzy(itmeapi)
}

// 获取Warframe市场的售价表，并进行排序,cn_name为物品中文名称，onlyMaxRank表示只取最高等级的物品，返回物品售价表，物品信息，物品英文
func wmItemOrders(cnName string, onlyMaxRank bool) (orders, itemsInSet, string, error) {
	var wfapiio wfAPIItemsOrders
	data, err := web.RequestDataWithHeaders(&http.Client{}, fmt.Sprintf("https://api.warframe.market/v1/items/%s/orders?include=item", cnName), "GET", func(request *http.Request) error {
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Platform", "pc")
		return nil
	}, nil)
	if err != nil {
		return nil, itemsInSet{}, "", err
	}
	err = json.Unmarshal(data, &wfapiio)
	var sellOrders orders
	// 遍历市场物品列表
	for _, v := range wfapiio.Payload.Orders {
		// 取其中类型为售卖，且去掉不在线的玩家
		if v.OrderType == "sell" && v.User.Status != "offline" {
			// 如果需要满级报价
			if onlyMaxRank && v.ModRank == wfapiio.Include.Item.ItemsInSet[0].ModMaxRank {
				sellOrders = append(sellOrders, v)
			} else if !onlyMaxRank {
				sellOrders = append(sellOrders, v)
			}
		}
	}
	// 对报价表进行排序，由低到高
	sort.Sort(sellOrders)
	// 获取物品信息
	for i, v := range wfapiio.Include.Item.ItemsInSet {
		if v.URLName == cnName {
			return sellOrders, wfapiio.Include.Item.ItemsInSet[i], wfapiio.Include.Item.ItemsInSet[i].En.ItemName, err
		}
	}
	return sellOrders, wfapiio.Include.Item.ItemsInSet[0], wfapiio.Include.Item.ItemsInSet[0].En.ItemName, err
}

func loadToFuzzy(wminfo wfAPIItem) {
	wmitems = make(map[string]items)
	itmeNames = []string{}
	for _, v := range wminfo.Payload.Items {
		wmitems[v.ItemName] = v
		itmeNames = append(itmeNames, v.ItemName)
	}
}
