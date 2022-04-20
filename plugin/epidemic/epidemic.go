// Package epidemic 城市疫情查询
package epidemic

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/web"
)

const (
	servicename = "epidemic"
	txurl       = "https://c.m.163.com/ug/api/wuhan/app/data/list-total"
)

var (
	limit = rate.NewManager[int64](time.Second*60, 1)
)

// result 疫情查询结果
type result struct {
	Data []byte `json:"data"`
}

// epidemic 疫情数据
type epidemic struct {
	AreaTree []*area `json:"areaTree"`
}

// area 城市疫情数据
type area struct {
	Today struct {
		Confirm      string `json:"confirm"`      // 新增确诊
		Heal         string `json:"heal"`         // 新增治愈
		Dead         string `json:"dead"`         // 新增死亡
		StoreConfirm string `json:"storeConfirm"` // 新增确诊
		Input        string `json:"input"`        // 新增境外输入
	} `json:"today"`
	Total struct {
		Confirm string `json:"confirm"` // 累计确诊
		Dead    string `json:"dead"`    // 累计死亡
		Heal    string `json:"heal"`    // 累计治愈
		Input   string `json:"input"`   // 累计境外输入
	} `json:"total"`
	ExtData struct {
		NoSymptom     string `json:"noSymptom"`     // 无症状感染者
		IncrNoSymptom string `json:"incrNoSymptom"` // 新增无症状感染者
	} `json:"extData"`
	Name           string  `json:"name"`           // 城市名字
	LastUpdateTime string  `json:"lastUpdateTime"` // 更新时间
	Children       []*area `json:"children"`
}

func init() {
	engine := control.Register(servicename, &control.Options{
		DisableOnDefault: false,
		Help: "城市疫情查询\n" +
			"- xxx疫情\n",
	})
	engine.OnSuffix("疫情").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			city := ctx.State["args"].(string)
			if city == "" {
				ctx.SendChain(message.Text("你还没有输入城市名字呢！"))
				return
			}
			data, time, err := queryEpidemic(city)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if data == nil {
				ctx.SendChain(message.Text("没有找到【", city, "】城市的疫情数据."))
				return
			}
			confirm, _ := strconv.Atoi(data.Total.Confirm)
			dead, _ := strconv.Atoi(data.Today.Dead)
			heal, _ := strconv.Atoi(data.Today.Heal)
			if limit.Load(ctx.Event.UserID).Acquire() {
				temp := fmt.Sprint("【", data.Name, "】疫情数据\n",
					"新增确诊：", data.Today.Confirm, "\n",
					"新增死亡：", data.Today.Dead, "\n",
					"现有确诊：", confirm-dead-heal, "\n",
					"累计确诊：", data.Total.Confirm, "\n",
					"累计治愈：", data.Total.Heal, "\n",
					"累计死亡：", data.Total.Dead, "\n",
					"新增无症状：", data.ExtData.IncrNoSymptom, "\n",
					"无症状人数：", data.ExtData.NoSymptom, "\n",
					"更新时间：\n『", time, "』")
				txt, err := text.RenderToBase64(temp, text.FontFile, 400, 20)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(txt))); id.ID() == 0 {
					ctx.SendChain(message.Text("ERROR:可能被风控了"))
				}
			} else {
				ctx.SendChain(message.Text("您的操作太频繁了！（冷却时间为1分钟）"))
			}
		})
}

// rcity 查找城市
func rcity(a *area, cityName string) *area {
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
func queryEpidemic(findCityName string) (citydata *area, times string, err error) {
	data, err := web.GetData(txurl)
	if err != nil {
		return
	}
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return
	}
	var e epidemic
	err = json.Unmarshal(r.Data, &e)
	if err != nil {
		return
	}
	var t *area
	citydata = rcity(e.AreaTree[0], findCityName)
	return citydata, t.LastUpdateTime, nil
}
