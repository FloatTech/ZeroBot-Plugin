// Package epidemic 城市疫情查询
package epidemic

import (
	"encoding/json"
	"strconv"
	"strings"

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

type Result struct {
	Ret  int    `json:"ret"`
	Data string `json:"data"`
}

type Epidemic struct {
	LastUpdateTime string `json:"lastUpdateTime"`
	AreaTree       []Area `json:"areaTree"`
}

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
	Children []Area `json:"children"`
}

type CityInfo struct {
	Name           string
	Confirm        int
	NowConfirm     int
	Dead           int
	Heal           int
	Grade          string
	LastUpdateTime string
}

func init() {
	engine := control.Register(servicename, order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "本插件用于查询城市疫情状况\n" +
			"使用方法列如： 北京疫情  \n",
	})
	engine.OnSuffix("疫情").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := ""
			for _, elem := range ctx.Event.Message {
				if elem.Type == "text" {
					text = strings.ReplaceAll(elem.Data["text"], " ", "")
					text = text[:strings.LastIndex(text, "疫情")]
					break
				}
			}
			process.SleepAbout1sTo2s()
			data := queryEpidemic(text)
			if data == nil {
				ctx.SendChain(message.Text("没有找到【" + text + "】城市的疫情数据."))
				return
			}
			msgtext := "【" + data.Name + "】疫情数据:\n" +
				"新增：" + strconv.Itoa(data.Confirm) +
				" ,现有确诊：" + strconv.Itoa(data.NowConfirm) +
				" ,治愈：" + strconv.Itoa(data.Heal) +
				" ,死亡：" + strconv.Itoa(data.Dead) + " " + data.Grade
			ctx.SendChain(
				message.Text(msgtext),
				message.Text("\n"),
				message.Text("更新时间："+data.LastUpdateTime),
				message.Text("\n"),
				message.Text("温馨提示：请大家做好防疫工作，出门带好口罩！"),
			)
		})
}

// Rcity 查找城市
func Rcity(a *Area, cityName string) *Area {
	if a == nil {
		return nil
	}
	if a.Name == cityName {
		return a
	}
	for _, v := range a.Children {
		if v.Name == cityName {
			return &v
		}
		c := Rcity(&v, cityName)
		if c != nil {
			return c
		}
	}
	return nil
}

// queryEpidemic 查询城市疫情
func queryEpidemic(findCityName string) *CityInfo {
	response, err := web.GetData(txurl)
	if err != nil {
		log.Errorln("[txurl-err]:", err)
	}
	var r Result
	_ = json.Unmarshal(response, &r)
	var e Epidemic
	_ = json.Unmarshal([]byte(r.Data), &e)
	citydata := Rcity(&e.AreaTree[0], findCityName)
	if citydata == nil {
		return nil
	} else {
		return &CityInfo{
			Name:           citydata.Name,
			Confirm:        citydata.Today.Confirm,
			NowConfirm:     citydata.Total.NowConfirm,
			Dead:           citydata.Total.Dead,
			Heal:           citydata.Total.Heal,
			Grade:          citydata.Total.Grade,
			LastUpdateTime: e.LastUpdateTime,
		}
	}
}
