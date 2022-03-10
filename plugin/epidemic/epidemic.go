// Package epidemic 城市疫情查询
package epidemic

import (
	"fmt"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	servicename = "qe"
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
	engine.OnRegex(`^(.*)疫情$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			region := ctx.State["regex_matched"].([]string)[1]
			process.SleepAbout1sTo2s()
			ctx.SendChain(queryEpidemic(region))
		})
}

func queryEpidemic(findName string) message.MessageSegment {
	var returnText string
	params := make(map[string]interface{})
	response := sendRequest(txurl, params, "GET")
	info := gjson.ParseBytes(response)
	data := gjson.Parse(info.Get("data").String())
	dqnmae := data.Get("areaTree.0.children").Array()
	tmp := `【%s】新增：%d ,现有确诊：%d ,治愈：%d ,死亡：%d `
	tmp1 := `【%s-%s】新增：%d ,现有确诊：%d ,治愈：%d ,死亡：%d %s`
outfor:
	for _, v := range dqnmae {
		if findName == fmt.Sprint(v.Get("name")) {
			returnText = fmt.Sprintf(tmp,
				v.Get("name").String(),
				v.Get("today.confirm").Int(),
				v.Get("total.nowConfirm").Int(),
				v.Get("total.heal").Int(),
				v.Get("total.dead").Int(),
			)
			isconfirm = true
			break outfor
		} else {
			for _, cv := range v.Get("children").Array() {
				if findName == fmt.Sprint(cv.Get("name")) {
					returnText = fmt.Sprintf(tmp1,
						v.Get("name").String(),
						cv.Get("name").String(),
						cv.Get("today.confirm").Int(),
						cv.Get("total.nowConfirm").Int(),
						cv.Get("total.heal").Int(),
						cv.Get("total.dead").Int(),
						cv.Get("total.grade").String(),
					)
					isconfirm = true
					break outfor
				}
			}
		}
	}
	if !isconfirm {
		returnText = " 没有找到【" + findName + "】城市的疫情数据."
	}
	isconfirm = false
	return message.Text(returnText + "\n" + fmt.Sprintf(" 更新时间：%s", data.Get("lastUpdateTime").String()) + "\n" + "温馨提示：请大家做好防疫工作，出门带好口罩！")
}
