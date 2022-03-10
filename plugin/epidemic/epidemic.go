// Package epidemic 城市疫情查询
package epidemic

import (
	"fmt"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	servicename = "queryEpidemic"
	txurl       = "https://view.inews.qq.com/g2/getOnsInfo?name=disease_h5"
)

var (
	isconfirm = false
)

func init() {
	engine := control.Register(servicename, order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "本插件用于查询城市疫情状况\n" +
			"使用方法列如： 北京疫情  \n",
	})
	engine.OnSuffix("疫情").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			region := ctx.State["regex_matched"].([]string)[1]
			process.SleepAbout1sTo2s()
			data, returnbool := queryEpidemic(region)
			if returnbool {
				ctx.SendChain(message.Text("没有找到【" + region + "】城市的疫情数据."))
				return
			}
			msgtext := "新增：" + data["confirm"] +
				" ,现有确诊：" + data["nowConfirm"] +
				" ,治愈：" + data["heal"] +
				" ,死亡：" + data["dead"] + " " + data["grade"]
			ctx.SendChain(
				message.Text("【"+data["name"]+"】疫情信息如下：\n"),
				message.Text(msgtext),
				message.Text("\n"),
				message.Text("更新时间："+data["lastUpdateTime"]),
				message.Text("\n"),
				message.Text("温馨提示：请大家做好防疫工作，出门带好口罩！"),
			)
		})
}

func queryEpidemic(findName string) (map[string]string, bool) {
	returnbool := false
	response, err := web.GetData(txurl)
	if err != nil {
		log.Errorln("[txurl-err]:", err)
	}
	info := gjson.ParseBytes(response)
	data := gjson.Parse(info.Get("data").String())
	dqnmae := data.Get("areaTree.0.children").Array()
	dataInfo := make(map[string]string)
outfor:
	for _, v := range dqnmae {
		if findName == fmt.Sprint(v.Get("name")) {
			dataInfo["name"] = v.Get("name").String()
			dataInfo["confirm"] = fmt.Sprint(v.Get("today.confirm").Int())
			dataInfo["nowConfirm"] = fmt.Sprint(v.Get("total.nowConfirm").Int())
			dataInfo["heal"] = fmt.Sprint(v.Get("total.heal").Int())
			dataInfo["dead"] = fmt.Sprint(v.Get("total.dead").Int())
			dataInfo["grade"] = ""
			dataInfo["lastUpdateTime"] = data.Get("lastUpdateTime").String()
			isconfirm = true
			break outfor
		} else {
			for _, cv := range v.Get("children").Array() {
				if findName == fmt.Sprint(cv.Get("name")) {
					dataInfo["name"] = v.Get("name").String() + "-" + cv.Get("name").String()
					dataInfo["confirm"] = fmt.Sprint(cv.Get("today.confirm").Int())
					dataInfo["nowConfirm"] = fmt.Sprint(cv.Get("total.nowConfirm").Int())
					dataInfo["heal"] = fmt.Sprint(cv.Get("total.heal").Int())
					dataInfo["dead"] = fmt.Sprint(cv.Get("total.dead").Int())
					dataInfo["grade"] = cv.Get("total.grade").String()
					dataInfo["lastUpdateTime"] = data.Get("lastUpdateTime").String()
					isconfirm = true
					break outfor
				}
			}
		}
	}
	if !isconfirm {
		returnbool = true
	}
	isconfirm = false
	return dataInfo, returnbool
}
