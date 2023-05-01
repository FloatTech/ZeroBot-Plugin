package warframeapi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/RomiChan/syncx"
	"github.com/sirupsen/logrus"

	"github.com/FloatTech/floatbox/web"
)

const wfapiurl = "https://api.warframestat.us/pc"        // 星际战甲API
const wfitemurl = "https://api.warframe.market/v1/items" // 星际战甲游戏品信息列表URL

// 从WFapi获取数据
func newwfapi() (w wfapi, err error) {
	var data []byte
	data, err = web.GetData(wfapiurl)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &w)
	return
}

// 获取Warframe市场的售价表，并进行排序,cn_name为物品中文名称，onlyMaxRank表示只取最高等级的物品，返回物品售价表，物品信息，物品英文
func getitemsorder(cnName string, onlyMaxRank bool) (od orders, it *itemsInSet, n string, err error) {
	var wfapiio wfAPIItemsOrders
	data, err := web.RequestDataWithHeaders(&http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}}, fmt.Sprintf("https://api.warframe.market/v1/items/%s/orders?include=item", cnName), "GET", func(request *http.Request) error {
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Platform", "pc")
		return nil
	}, nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &wfapiio)
	if len(wfapiio.Payload.Orders) == 0 {
		err = errors.New("no such name")
	}
	od = make(orders, 0, len(wfapiio.Payload.Orders))
	// 遍历市场物品列表
	for _, v := range wfapiio.Payload.Orders {
		// 取其中类型为售卖，且去掉不在线的玩家
		if v.OrderType == "sell" && v.User.Status != "offline" {
			if !onlyMaxRank {
				od = append(od, v)
				continue
			}
			if v.ModRank == wfapiio.Include.Item.ItemsInSet[0].ModMaxRank {
				od = append(od, v)
			}
		}
	}
	// 对报价表进行排序，由低到高
	sort.Sort(od)
	// 获取物品信息
	for i, v := range wfapiio.Include.Item.ItemsInSet {
		if v.URLName == cnName {
			it = &wfapiio.Include.Item.ItemsInSet[i]
			n = v.En.ItemName
			return
		}
	}
	it = &wfapiio.Include.Item.ItemsInSet[0]
	n = wfapiio.Include.Item.ItemsInSet[0].En.ItemName
	return
}

type wmdata struct {
	wmitems   map[string]items
	itemNames []string
}

var (
	wderr error
	wd    = syncx.Lazy[*wmdata]{Init: func() (d *wmdata) {
		d, wderr = newwm()
		return
	}}
)

func newwm() (*wmdata, error) {
	var itemapi wfAPIItem // WarFrame市场的数据实例
	var wd wmdata
	data, err := web.RequestDataWithHeaders(&http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}}, wfitemurl, "GET", func(request *http.Request) error {
		request.Header.Add("Accept", "application/json")
		request.Header.Add("Language", "zh-hans")
		return nil
	}, nil)
	if err != nil {
		return &wd, err
	}
	err = json.Unmarshal(data, &itemapi)
	if err != nil {
		return &wd, err
	}
	wd.wmitems = make(map[string]items, len(itemapi.Payload.Items)*4)
	wd.itemNames = make([]string, len(itemapi.Payload.Items))
	for i, v := range itemapi.Payload.Items {
		wd.wmitems[v.ItemName] = v
		wd.itemNames[i] = v.ItemName
	}
	logrus.Infoln("[wfapi] 获取", len(wd.itemNames), "项内容")
	return &wd, nil
}
