// Package epidemic 城市疫情查询
package epidemic

import (
	"encoding/json"
	"strconv"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	servicename = "epidemic"
	txurl       = "https://view.inews.qq.com/g2/getOnsInfo?name=disease_h5"
)

// Result 疫情查询结果
type Result struct {
	Ret  int    `json:"ret"`
	Data string `json:"data"`
}

// Epidemic 疫情数据
type Epidemic struct {
	LastUpdateTime string `json:"lastUpdateTime"`
	AreaTree       []Area `json:"areaTree"`
}

// Area 城市疫情数据
type Area struct {
	Name  string `json:"name"`
	Today struct {
		Confirm int `json:"confirm"`
	} `json:"today"`
	Total struct {
		NowConfirm int    `json:"nowConfirm"`
		Confirm    int    `json:"confirm"`
		Dead       int    `json:"dead"`
		Heal       int    `json:"heal"`
		Grade      string `json:"grade"`
	} `json:"total"`
	Children []*Area `json:"children"`
}

func init() {
	engine := control.Register(servicename, order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "本插件用于查询城市疫情状况\n" +
			"使用方法列如： 北京疫情  \n",
	})
	engine.OnSuffix("疫情").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ctx.State["args"].(string)
			if text == "" {
				ctx.SendChain(message.Text("你还没有输入城市名字呢！"))
				return
			}
			process.SleepAbout1sTo2s()
			data, times := queryEpidemic(text)
			if data == nil {
				ctx.SendChain(message.Text("没有找到【" + text + "】城市的疫情数据."))
				return
			}
			msgtext := "【" + data.Name + "】疫情数据:\n" +
				"新增：" + strconv.Itoa(data.Today.Confirm) +
				" ,现有确诊：" + strconv.Itoa(data.Total.NowConfirm) +
				" ,治愈：" + strconv.Itoa(data.Total.Heal) +
				" ,死亡：" + strconv.Itoa(data.Total.Dead) + " " + data.Total.Grade
			ctx.SendChain(
				message.Text(msgtext),
				message.Text("\n"),
				message.Text("更新时间："+times),
				message.Text("\n"),
				message.Text("温馨提示：请大家做好防疫工作，出门带好口罩！"),
			)
		})
}

// rcity 查找城市
func rcity(a *Area, cityName string) *Area {
	if a == nil {
		return nil
	}
	if a.Name == cityName {
		return a
	}
	for _, v := range a.Children {
		if v.Name == cityName {
			return v
		}
		c := rcity(v, cityName)
		if c != nil {
			return c
		}
	}
	return nil
}

// queryEpidemic 查询城市疫情
func queryEpidemic(findCityName string) (citydata *Area, times string) {
	response, err := web.GetData(txurl)
	if err != nil {
		log.Errorln("[txurl-err]:", err)
		return nil, ""
	}
	var r Result
	err = json.Unmarshal(response, &r)
	if err != nil {
		log.Errorln("[txjson-Result-err]:", err)
		return nil, ""
	}
	var e Epidemic
	err = json.Unmarshal([]byte(r.Data), &e)
	if err != nil {
		log.Errorln("[txjson-Epidemic-err]:", err)
		return nil, ""
	}
	citydata = rcity(&e.AreaTree[0], findCityName)
	return citydata, e.LastUpdateTime
}
