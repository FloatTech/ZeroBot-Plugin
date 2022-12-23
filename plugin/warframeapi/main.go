package warframeapi

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/zbputils/img/text"
	"io"
	"io/ioutil"
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
	wfapi         WFAPI
	client        http.Client
	itmeapi       WFAPIItem
	wmitems       map[string]Items
	itmeNames     []string
	fileItemNames map[string]string
	itemNamesPath string
	sublist       map[int64]*SubList
	sublistPath   string
)

func init() {
	eng := control.Register("warframeapi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "warframeapi\n" +
			"- wf数据更新\n" +
			"- [金星/地球/火卫二]平原状态\n" +
			"- 订阅[金星/地球/火卫二]平原[白天/夜晚]\n" +
			"- 取消订阅[金星/地球/火卫二]平原[白天/夜晚]\n" +
			"- .wm [物品名称]\n" +
			"- 取外号 [原名称] [外号]\n" +
			"- [金星/地球/火卫二]平原状态\n" +
			"- 仲裁\n" +
			"- 警报\n" +
			"- 每日特惠",
		PrivateDataFolder: "warframeAPI",
	})

	itemNamesPath = eng.DataFolder() + "ItemNames.json"

	if isExist(itemNamesPath) {
		data, err := ioutil.ReadFile(itemNamesPath)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &fileItemNames)
		if err != nil {
			panic(err)
		}
	} else {
		fileItemNames = map[string]string{}
	}
	sublistPath = eng.DataFolder() + "Sublist.json"
	if isExist(sublistPath) {
		data, err := ioutil.ReadFile(sublistPath)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(data, &sublist)
		if err != nil {
			panic(err)
		}
	} else {
		sublist = map[int64]*SubList{}
	}
	updateWM()
	loadToFuzzy()
	udateWFAPI2()

	eng.OnRegex(`^(.*)平原状态$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["regex_matched"].([]string)
			updateWFAPI(ctx)

			switch args[1] {
			case "奥布山谷":
				fallthrough
			case "金星":
				ctx.SendChain(
					message.Text(
						"平原状态:", GameTimes[1].getStatus(), "\n",
						"下次更新:", GameTimes[1].getTime(),
					),
				)
			case "夜灵":
				fallthrough
			case "地球":
				ctx.SendChain(
					message.Text(
						"平原状态:", GameTimes[0].getStatus(), "\n",
						"下次更新:", GameTimes[0].getTime(),
					),
				)
			case "火卫":
				fallthrough
			case "火卫二":
				fallthrough
			case "魔胎之境":
				ctx.SendChain(
					message.Text(
						"平原状态:", GameTimes[2].getStatus(), "\n",
						"下次更新:", GameTimes[2].getTime(),
					),
				)
			default:
				ctx.SendChain(message.Text("Error:平原不存在"))
				return
			}
		})

	eng.OnRegex(`^警报$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
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
	eng.OnRegex(`^订阅(.*)平原(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["regex_matched"].([]string)
			updateWFAPI(ctx)
			status := false
			switch args[2] {
			case "白天":
				fallthrough
			case "fass":
				fallthrough
			case "Fass":
				fallthrough
			case "温暖":
				status = true
			}
			switch args[1] {
			case "奥布山谷":
				fallthrough
			case "金星":
				//sublist = append(sublist, SubList{ctx.Event.GroupID, ctx.Event.UserID, 1, status, false})
				addUseSub(ctx.Event.UserID, ctx.Event.GroupID, 1, status)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("已成功订阅"),
					message.Text(GameTimes[1].Name),
					message.Text(status),
				)
			case "夜灵":
				fallthrough
			case "地球":
				addUseSub(ctx.Event.UserID, ctx.Event.GroupID, 0, status)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("已成功订阅"),
					message.Text(GameTimes[0].Name),
					message.Text(status),
				)
			case "火卫":
				fallthrough
			case "火卫二":
				fallthrough
			case "魔胎之境":
				addUseSub(ctx.Event.UserID, ctx.Event.GroupID, 2, status)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("已成功订阅"),
					message.Text(GameTimes[2].Name),
					message.Text(status),
				)
			default:
				ctx.SendChain(message.Text("Error:平原不存在"))
				return
			}
		})
	eng.OnRegex(`^取消订阅(.*)平原(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["regex_matched"].([]string)
			updateWFAPI(ctx)
			status := false
			switch args[2] {
			case "白天":
				fallthrough
			case "fass":
				fallthrough
			case "Fass":
				fallthrough
			case "温暖":
				status = true
			}
			switch args[1] {
			case "奥布山谷":
				fallthrough
			case "金星":
				//sublist = append(sublist, SubList{ctx.Event.GroupID, ctx.Event.UserID, 1, status, false})
				removeUseSub(ctx.Event.UserID, ctx.Event.GroupID, 1)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("已取消订阅"),
					message.Text(GameTimes[1].Name),
					message.Text(status),
				)
			case "夜灵":
				fallthrough
			case "地球":
				removeUseSub(ctx.Event.UserID, ctx.Event.GroupID, 0)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("已取消订阅"),
					message.Text(GameTimes[0].Name),
					message.Text(status),
				)
			case "火卫":
				fallthrough
			case "火卫二":
				fallthrough
			case "魔胎之境":
				removeUseSub(ctx.Event.UserID, ctx.Event.GroupID, 2)
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text("已取消订阅"),
					message.Text(GameTimes[2].Name),
					message.Text(status),
				)
			default:
				ctx.SendChain(message.Text("Error:平原不存在"))
				return
			}
		})
	eng.OnRegex(`^仲裁$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			updateWFAPI(ctx)
			sendStringArray([]string{
				"节点:" + wfapi.Arbitration.Node,
				"类型:" + wfapi.Arbitration.Type,
				"阵营:" + wfapi.Arbitration.Enemy,
				"剩余时间:" + fmt.Sprint(int(wfapi.Arbitration.Expiry.Sub(time.Now().UTC()).Minutes())) + "m",
			}, ctx)
		})
	eng.OnRegex(`^每日特惠$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			updateWFAPI(ctx)
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
	// 		for _, dd := range wfapi.DailyDeals {
	// 			imagebuild.DrawTextSend([]string{
	// 				"节点:" + wfapi.Arbitration.Node,
	// 				"类型:" + wfapi.Arbitration.Type,
	// 				"阵营:" + wfapi.Arbitration.Enemy,
	// 				"剩余时间:" + fmt.Sprint(int(wfapi.Arbitration.Expiry.Sub(time.Now().UTC()).Minutes())) + "m",
	// 			}, ctx)
	// 		}
	// 	})
	eng.OnRegex(`^wf数据更新$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			updateWFAPI(ctx)
			LoadTime()
			ctx.SendChain(message.Text("已拉取服务器时间并同步到本地模拟"))
		})
	eng.OnRegex(`^.取外号 (.*) (.*)$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := ctx.State["regex_matched"].([]string)
		sol := fuzzy.FindNormalizedFold(args[1], itmeNames)
		var msg []string
		switch len(sol) {
		case 0:
			ctx.SendChain(message.Text("无法查询到该物品"))
			return
		case 1:
			wmitems[args[2]] = wmitems[sol[0]]
			itmeNames = append(itmeNames, args[2])
			fileItemNames[args[2]] = sol[0]
			ctx.Send(message.Text("已给[", sol[0], "]新增外号:[", args[2], "]"))
			jsonSave(fileItemNames, itemNamesPath)
		default:
			for i, v := range sol {
				msg = append(msg, fmt.Sprintf("[%d] %s", i, v))
			}
			msg = append(msg, "包含多个结果，请输入编号来确定具体的物品外号(15s内)")
			sendStringArray(msg, ctx)
			//msg = []string{}
		GETNUM:
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
					goto GETNUM
				}
				//name = sol[cc]
				//ctx.Send(message.Image("https://warframe.market/static/assets/" + wmitems[sol[cc]].Thumb))
				//msg = append(msg, wmitems[sol[cc]].ItemName)

				wmitems[args[2]] = wmitems[sol[num]]
				itmeNames = append(itmeNames, args[2])
				fileItemNames[args[2]] = sol[num]
				ctx.Send(message.Text("已给[", sol[num], "]新增外号:[", args[2], "]"))
				jsonSave(fileItemNames, itemNamesPath)
			}

		}

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
func sendStringArray(texts []string, ctx *zero.Ctx) {
	b, err := text.RenderToBase64(strings.Join(texts, "\n"), text.FontFile, 800, 20)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(message.Image("base64://" + binary.BytesToString(b)))
}

func addUseSub(QQ int64, QQGroup int64, stype int, status bool) {
	if sb, ok := sublist[QQGroup]; ok {
		if st, ok := sb.SubUser[QQ]; ok {
			st.SubType[stype] = &status
		} else {
			sublist[QQGroup].SubUser[QQ] = SubType{map[int]*bool{stype: &status}, false}
		}
	} else {
		sublist[QQGroup] = &SubList{map[int64]SubType{QQ: {map[int]*bool{stype: &status}, false}}, false, false}
	}
	jsonSave(sublist, sublistPath)
}

func removeUseSub(QQ int64, QQGroup int64, stype int) {
	if sb, ok := sublist[QQGroup]; ok {
		if _, ok := sb.SubUser[QQ]; ok {
			delete(sublist[QQGroup].SubUser[QQ].SubType, stype)
			jsonSave(sublist, sublistPath)
		}
	}
}

func updateWFAPI(ctx *zero.Ctx) {
	var data []byte
	var err error
	data, err = web.GetData("https://api.warframestat.us/pc")
	if err != nil {
		ctx.SendChain(message.Text("Error:", err.Error()))
		return
	}
	err = json.Unmarshal(data, &wfapi)
	if err != nil {
		ctx.SendChain(message.Text("Error:", err.Error()))
		return
	}

}

func udateWFAPI2() {
	var data []byte
	var err error
	data, err = web.GetData("https://api.warframestat.us/pc")
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &wfapi)
	if err != nil {
		return
	}
	gameTimeInit()
	LoadTime()
}
func updateWM() {
	var data []byte
	var err error
	data, err = getData("https://api.warframe.market/v1/items", []string{"Accept", "Language"}, []string{"application/json", "zh-hans"})
	if err != nil {
		panic(err)

	}
	err = json.Unmarshal(data, &itmeapi)
	if err != nil {
		panic(err)
	}
}

func getWMItemOrders(name string, h bool) (Orders, ItemsInSet, string, error) {
	var data []byte
	var err error
	var wfapiio WFAPIItemsOrders
	data, err = getData(fmt.Sprintf("https://api.warframe.market/v1/items/%s/orders?include=item", name), []string{"Accept", "Platform"}, []string{"application/json", "pc"})
	if err != nil {
		return nil, ItemsInSet{}, "", err
	}
	err = json.Unmarshal(data, &wfapiio)

	var sellOrders Orders
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
	wmitems = make(map[string]Items)
	itmeNames = []string{}
	for _, v := range itmeapi.Payload.Items {
		wmitems[v.ItemName] = v
		itmeNames = append(itmeNames, v.ItemName)
	}
	for k, v := range fileItemNames {
		wmitems[k] = wmitems[v]
		itmeNames = append(itmeNames, k)
	}
}

func getData(url string, head []string, headvalue []string) (data []byte, err error) {
	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)
	//增加header选项
	for i, v := range head {
		reqest.Header.Add(v, headvalue[i])
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

func isExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		fmt.Println(err)
		return false
	}
	return true
}
func jsonSave(v interface{}, path string) (bool, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return false, err
	}
	dataStr := string(data)

	// 将字符串写入指定的文件
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return false, err
	}
	defer file.Close() // 结束时关闭句柄，释放资源
	writer := bufio.NewWriter(file)
	writer.WriteString(dataStr)
	writer.Flush() // 缓存数据写入磁盘（持久化）
	return true, nil
}
