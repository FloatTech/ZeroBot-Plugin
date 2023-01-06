// Package warframeapi 百度内容审核
package warframeapi

import (
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/zbputils/img/text"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/lithammer/fuzzysearch/fuzzy"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	wmitems   map[string]items //WarFrame市场的中文名称对应的物品的字典
	itmeNames []string         //物品名称列表
	//TODO:订阅功能-等待重做
	//sublist     map[int64]*subList //订阅列表
	//sublistPath string             //订阅列表存储路径
)

func init() {
	eng := control.Register("warframeapi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "warframeapi\n" +
			"- wf时间同步\n" +
			"- [金星|地球|火卫二]平原时间\n" +
			//"- 订阅[金星|地球|火卫二]平原[白天|夜晚]\n" +
			//"- 取消订阅[金星|地球|火卫二]平原[白天|夜晚]\n" +
			//"- wf订阅检测\n" +
			"- .wm [物品名称]\n" +
			"- 仲裁\n" +
			"- 警报\n" +
			"- 每日特惠",
		PrivateDataFolder: "warframeapi",
	})
	//TODO:订阅功能-等待重做
	//订阅名单文件路径
	//sublistPath = eng.DataFolder() + "Sublist.json"
	//if file.IsExist(sublistPath) {
	//	data, err := os.ReadFile(sublistPath)
	//	if err != nil {
	//		panic(err)
	//	}
	//	err = json.Unmarshal(data, &sublist)
	//	if err != nil {
	//		panic(err)
	//	}
	//} else {
	//	sublist = map[int64]*subList{}
	//}
	updateWM()

	//初始化游戏时间模拟
	go gameRuntime()
	eng.OnSuffix("平原时间").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			//根据具体输入的平原，来显示时间
			switch ctx.State["args"].(string) {
			case "地球", "夜灵":
				ctx.SendChain(getTimeString(0))
			case "金星", "奥布山谷":
				ctx.SendChain(getTimeString(1))
			case "魔胎之境", "火卫二", "火卫":
				ctx.SendChain(getTimeString(2))
			default:
				ctx.SendChain(message.Text("ERROR: 平原不存在"))
				return
			}
		})

	eng.OnFullMatch("警报").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := getWFAPI()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			//如果返回的wfapi中，警报数量>0
			if len(wfapi.Alerts) > 0 {
				//遍历警报数据，打印警报信息
				for _, v := range wfapi.Alerts {
					//如果警报处于激活状态
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
	//eng.OnRegex(`^(订阅|取消订阅)(.*)平原(.*)$`).SetBlock(true).
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
	//eng.OnFullMatch(`wf订阅检测`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
	//})
	eng.OnFullMatch("仲裁").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			//通过wfapi获取仲裁信息
			wfapi, err := getWFAPI()
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
			wfapi, err := getWFAPI()

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
			wfapi, err := getWFAPI()
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
			//根据输入的名称，从游戏物品名称列表中进行模糊搜索
			sol := fuzzy.FindNormalizedFold(ctx.State["args"].(string), itmeNames)
			var msg []string
			//物品名称
			var name string

			//根据搜搜结果，打印找到的物品
			switch len(sol) {
			case 0: //没有搜索到任何东西
				ctx.SendChain(message.Text("无法查询到该物品"))
				return
			case 1: //如果只搜索到了一个
				name = sol[0]
			default: //如果搜搜到了多个
				//遍历搜索结果，并打印为图片展出
				for i, v := range sol {
					msg = append(msg, fmt.Sprintf("[%d] %s", i, v))
				}
				msg = append(msg, "包含多个结果，请输入编号查看(15s内),输入c直接结束会话")
				ctx.SendChain(stringArrayToImage(msg))
				msg = []string{}

				itemIndex := getItemNameFutureEvent(ctx, 2)
				if itemIndex == -1 {
					return
				}
				name = sol[itemIndex]
				//GETNUM2: //获取用户具体想查看哪个物品
				//next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Next()
				//select {
				//case <-time.After(time.Second * 15):
				//	ctx.SendChain(message.Text("会话已结束!"))
				//	return
				//case e := <-next:
				//	msg := e.Event.Message.ExtractPlainText()
				//	if msg == "c" {
				//		ctx.SendChain(message.Text("会话已结束!"))
				//		return
				//	}
				//	num, err := strconv.Atoi(msg)
				//	if err != nil {
				//		ctx.SendChain(message.Text("请输入数字!(输入c结束会话)"))
				//		goto GETNUM2
				//	}
				//	name = sol[num]
				//}
			}
			Mf := false
		GETWM:
			if Mf {
				msg = []string{}
			}
			sells, itmeinfo, txt, err := getWMItemOrders(wmitems[name].URLName, Mf)
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
			msg = append(msg, wmitems[name].ItemName)
			ismod := false
			if err != nil {
				ctx.Send(message.Text("Error:", err.Error()))
				return
			}
			if itmeinfo.ModMaxRank != 0 {
				ismod = true
			}
			max := 5
			if sells == nil {
				ctx.Send(message.Text("无可购买对象"))
				return
			}

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
				if msg == "r" {
					Mf = true
					goto GETWM
				}
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
						ctx.Send(message.Text(fmt.Sprintf("/w %s Hi! I want to buy: %s(Rank:%d) for %d platinum. (warframe.market)", sells[i].User.IngameName, txt, sells[i].ModRank, sells[i].Platinum)))

					} else {
						ctx.Send(message.Text(fmt.Sprintf("/w %s Hi! I want to buy: %s for %d platinum. (warframe.market)", sells[i].User.IngameName, txt, sells[i].Platinum)))
					}
					return
				}

				return
			}

		})

}

// 获取搜索结果中的物品具体名称index的FutureEvent,传入ctx和一个递归次数上限,返回一个int，如果为返回内容为-1，说明会话超时，或主动结束，或超出递归
func getItemNameFutureEvent(ctx *zero.Ctx, count int) int {
	next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Next()
	select {
	case <-time.After(time.Second * 15):
		//超时15秒处理
		ctx.SendChain(message.Text("会话已超时!"))
		return -1
	case e := <-next:
		msg := e.Event.Message.ExtractPlainText()
		//输入c主动结束的处理
		if msg == "c" {
			ctx.SendChain(message.Text("会话已结束!"))
			return -1
		}
		//尝试对输入进行数字转换
		num, err := strconv.Atoi(msg)
		//如果出错，说明输入的并非数字，则重新触发该内容
		if err != nil {
			//查看是否超时
			if count == 0 {
				ctx.SendChain(message.Text("连续输入错误，会话已结束!"))
				return -1
			} else {
				ctx.SendChain(message.Text("请输入数字!(输入c结束会话)[", count, "]"))
				count--
				return getItemNameFutureEvent(ctx, count)
			}
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

//TODO:订阅功能-等待重做
// 添加用户订阅
//func addUseSub(qq int64, qqGroup int64, stype int, status bool) {
//	if sb, ok := sublist[qqGroup]; ok {
//		if st, ok := sb.SubUser[qq]; ok {
//			st.SubType[stype] = &status
//		} else {
//			sublist[qqGroup].SubUser[qq] = subType{map[int]*bool{stype: &status}, false}
//		}
//	} else {
//		sublist[qqGroup] = &subList{map[int64]subType{qq: {map[int]*bool{stype: &status}, false}}, false, false}
//	}
//	jsonSave(sublist, sublistPath)
//}
//
//// 移除用户订阅
//func removeUseSub(qq int64, qqGroup int64, stype int) {
//	if sb, ok := sublist[qqGroup]; ok {
//		if _, ok := sb.SubUser[qq]; ok {
//			delete(sublist[qqGroup].SubUser[qq].SubType, stype)
//			jsonSave(sublist, sublistPath)
//		}
//	}
//}

// 从WFapi获取数据
func getWFAPI() (wfAPI, error) {
	var wfapi wfAPI //WarFrameAPI的数据实例
	var data []byte
	var err error
	data, err = web.GetData("https://api.warframestat.us/pc")
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
	var itmeapi wfAPIItem //WarFrame市场的数据实例
	var data []byte
	var err error
	data, err = getData("https://api.warframe.market/v1/items", map[string]string{"Accept": "application/json", "Language": "zh-hans"})
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &itmeapi)
	if err != nil {
		panic(err)
	}
	loadToFuzzy(itmeapi)
}

func getWMItemOrders(name string, h bool) (orders, itemsInSet, string, error) {
	var data []byte
	var err error
	var wfapiio wfAPIItemsOrders
	data, err = getData(fmt.Sprintf("https://api.warframe.market/v1/items/%s/orders?include=item", name), map[string]string{"Accept": "application/json", "Platform": "pc"})
	if err != nil {
		return nil, itemsInSet{}, "", err
	}
	err = json.Unmarshal(data, &wfapiio)

	var sellOrders orders
	for _, v := range wfapiio.Payload.Orders {
		if v.OrderType == "sell" && v.User.Status != "offline" {
			if h && v.ModRank == wfapiio.Include.Item.ItemsInSet[0].ModMaxRank {
				sellOrders = append(sellOrders, v)
			} else if !h {
				sellOrders = append(sellOrders, v)
			}

		}
	}
	sort.Sort(sellOrders)

	for i, v := range wfapiio.Include.Item.ItemsInSet {
		if v.URLName == name {
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

func getData(url string, head map[string]string) (data []byte, err error) {
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)
	//增加header选项
	for i, v := range head {
		reqest.Header.Add(i, v)
	}
	if err != nil {
		return nil, err
	}
	//处理返回结果
	//发起http请求的client实例
	client := http.Client{}
	response, err := client.Do(reqest)
	if err != nil {
		return nil, err
	}
	data, err = io.ReadAll(response.Body)
	response.Body.Close()
	return data, err
	//func jsonSave(v interface{}, path string) {
	//	f, _ := os.Create(path)
	//	defer f.Close()
	//	json.NewEncoder(f).Encode(v)
	//}

}
