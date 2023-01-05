// Package warframeapi 百度内容审核
package warframeapi

import (
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/zbputils/img/text"
	"io"
	"net/http"
	"os"
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
	client    http.Client      //发起http请求的client实例
	itmeapi   wfAPIItem        //WarFrame市场的数据实例
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
			"- wf数据更新\n" +
			"- [金星|地球|火卫二]平原状态\n" +
			//"- 订阅[金星|地球|火卫二]平原[白天|夜晚]\n" +
			//"- 取消订阅[金星|地球|火卫二]平原[白天|夜晚]\n" +
			//"- wf订阅检测\n" +
			"- .wm [物品名称]\n" +
			"- [金星|地球|火卫二]平原状态\n" +
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
	loadToFuzzy()
	//尝试初始化游戏时间模拟
	//wfapi, _ := getWFAPI()
	//loadTime(wfapi)
	//gameRuntime()
	eng.OnRegex(`^(.*)平原状态$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["regex_matched"].([]string)
			switch args[1] {
			case "金星", "奥布山谷":
				ctx.SendChain(
					message.Text(
						"平原状态:", gameTimes[1].getStatus(), "\n",
						"下次更新:", gameTimes[1].getTime(),
					),
				)
			case "地球", "夜灵":
				ctx.SendChain(
					message.Text(
						"平原状态:", gameTimes[0].getStatus(), "\n",
						"下次更新:", gameTimes[0].getTime(),
					),
				)
			case "魔胎之境", "火卫二", "火卫":
				ctx.SendChain(
					message.Text(
						"平原状态:", gameTimes[2].getStatus(), "\n",
						"下次更新:", gameTimes[2].getTime(),
					),
				)
			default:
				ctx.SendChain(message.Text("ERROR: 平原不存在"))
				return
			}
		})

	eng.OnFullMatch(`警报`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := getWFAPI()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			if len(wfapi.Alerts) > 0 {
				for _, v := range wfapi.Alerts {
					if v.Active {
						sendStringArray([]string{
							"节点:" + v.Mission.Node,
							"类型:" + v.Mission.Type,
							"敌人Lv:" + fmt.Sprint(v.Mission.MinEnemyLevel) + "~" + fmt.Sprint(v.Mission.MaxEnemyLevel),
							"奖励:" + v.Mission.Reward.AsString,
							"剩余时间:" + v.Eta,
						}, ctx)
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
	eng.OnFullMatch(`仲裁`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := getWFAPI()

			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			sendStringArray([]string{
				"节点:" + wfapi.Arbitration.Node,
				"类型:" + wfapi.Arbitration.Type,
				"阵营:" + wfapi.Arbitration.Enemy,
				"剩余时间:" + fmt.Sprint(int(wfapi.Arbitration.Expiry.Sub(time.Now().UTC()).Minutes())) + "m",
			}, ctx)
		})
	eng.OnFullMatch(`每日特惠`).SetBlock(true).
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
	eng.OnFullMatch(`wf数据更新`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			wfapi, err := getWFAPI()

			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err.Error()))
				return
			}
			loadTime(wfapi)
			ctx.SendChain(message.Text("已拉取服务器时间并同步到本地模拟"))
		})
	eng.OnRegex(`^.wm (.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["regex_matched"].([]string)
			sol := fuzzy.FindNormalizedFold(args[1], itmeNames)
			var msg []string
			var name string
			switch len(sol) {
			case 0:
				ctx.SendChain(message.Text("无法查询到该物品"))
				return
			case 1:
				name = sol[0]
			default:
				for i, v := range sol {
					msg = append(msg, fmt.Sprintf("[%d] %s", i, v))
				}
				msg = append(msg, "包含多个结果，请输入编号查看(15s内),输入c直接结束会话")
				sendStringArray(msg, ctx)
				msg = []string{}
			GETNUM2:
				next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession()).Next()
				select {
				case <-time.After(time.Second * 15):
					ctx.SendChain(message.Text("会话已结束!"))
					return
				case e := <-next:
					msg := e.Event.Message.ExtractPlainText()
					if msg == "c" {
						ctx.SendChain(message.Text("会话已结束!"))
						return
					}
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字!(输入c结束会话)"))
						goto GETNUM2
					}
					name = sol[num]
				}
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
			sendStringArray(msg, ctx)
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

// 数组文字转图片并发送
func sendStringArray(texts []string, ctx *zero.Ctx) {
	b, err := text.RenderToBase64(strings.Join(texts, "\n"), text.FontFile, 800, 20)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(message.Image("base64://" + binary.BytesToString(b)))
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

func loadToFuzzy() {
	wmitems = make(map[string]items)
	itmeNames = []string{}
	for _, v := range itmeapi.Payload.Items {
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
	response, err := client.Do(reqest)
	if err != nil {
		return nil, err
	}
	data, err = io.ReadAll(response.Body)
	response.Body.Close()
	return data, err
}
func jsonSave(v interface{}, path string) {
	f, _ := os.Create(path)
	defer f.Close()
	json.NewEncoder(f).Encode(v)
}
