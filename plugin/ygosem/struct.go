// Package ygosem 基于ygosem的插件功能
package ygosem

import (
	"errors"
	"regexp"
	"strings"

	"github.com/FloatTech/floatbox/web"
)

type gameCardInfo struct {
	Name    string // 卡名
	ID      string // 卡密
	Type    string // 种类
	Race    string // 种族
	Attr    string // 属性
	Level   string // 等级
	Atk     string // 攻击力
	Def     string // 防御力
	Depict  string // 效果
	Maxcard string // 是否是分享的开关
}

// 正则筛选数据
func regexpmatch(rule, str string) [][]string {
	return regexp.MustCompile(rule).FindAllStringSubmatch(str, -1)
}

// 正则返回第n组的数据
func regexpmatchByRaw(rule, str string, n int) []string {
	return regexp.MustCompile(rule).FindAllStringSubmatch(str, -1)[n]
}

// 正则返回第0组的数据
func regexpmatchByZero(rule, str string) []string {
	return regexpmatchByRaw(rule, str, 0)
}

// 获取卡名列表
func getYGolist(body string) (cardsname map[string]string) {
	nameList := regexpmatch(`<div class="icon_size" style="white-space: nowrap;">\s*<a href="..(.*)" target="_blank">(.*)</a>\s*</div>`, body)
	if len(nameList) != 0 {
		cardsname = make(map[string]string, len(nameList)*2)
		for _, names := range nameList {
			cardsname[names[2]] = names[1]
		}
	}
	return
}

// 获取卡面信息
func getCarddata(body string) (cardata gameCardInfo) {
	// 获取卡名
	cardName := regexpmatchByZero(`<b>中文名</b> </span>&nbsp;<span class="item_box_value">\s*(.*)</span>\s*</div>`, body)
	if len(cardName) == 0 {
		return
	}
	cardata.Name = cardName[1]
	// 获取卡密
	cardID := regexpmatchByZero(`<b>卡片密码</b> </span>&nbsp;<span class="item_box_value">\s*(.*)\s*</span>`, body)
	cardata.ID = cardID[1]
	// 种类
	cardType := regexpmatchByZero(`<b>卡片种类</b> </span>&nbsp;<span class="item_box_value" id="dCnType">\s*(.*?)\s*</span>\s*<span`, body)
	cardata.Type = cardType[1]
	if strings.Contains(cardType[1], "怪兽") {
		// 种族
		cardRace := regexpmatchByZero(`<span id="dCnRace" class="item_box_value">\s*(.*)\s*</span>\s*<span id="dEnRace"`, body)
		cardata.Race = cardRace[1]
		// 属性
		cardAttr := regexpmatchByZero(`<b>属性</b> </span>&nbsp;<span class="item_box_value" id="attr">\s*(.*)\s*</span>`, body)
		cardata.Attr = cardAttr[1]
		/*星数*/
		switch {
		case strings.Contains(cardType[1], "连接"):
			cardLevel := regexpmatchByZero(`<span class="item_box_value">(LINK.*)</span>`, body)
			cardata.Level = cardLevel[1]
		default:
			cardLevel := regexpmatchByZero(`<b>星数/阶级</b> </span><span class=\"item_box_value\">\s*(.*)\s*</span>`, body)
			cardata.Level = cardLevel[1]
			// 守备力
			cardDef := regexpmatchByZero(`<b>DEF</b></span>\s*&nbsp;<span class="item_box_value">\s*(\d+|\?|？)\s*</span>\s*</div>`, body)
			cardata.Def = cardDef[1]
		}
		// 攻击力
		cardAtk := regexpmatchByZero(`<b>ATK</b> </span>&nbsp;<span class=\"item_box_value\">\s*(\d+|\?|？)\s*</span>`, body)
		cardata.Atk = cardAtk[1]
	}
	/*效果*/
	cardDepict := regexpmatchByZero(`<div class="item_box_text" id="cardDepict">\s*(?s:(.*?))\s*</div>`, body)
	cardata.Depict = cardDepict[1]
	return
}

// 获取卡图
func getPic(body string, choosepic bool) (imageBytes []byte, err error) {
	// 获取卡图连接
	cardpic := regexpmatchByZero(`picsCN(/\d+/\d+).jpg`, body)
	if len(cardpic) == 0 {
		return nil, errors.New("getPic正则匹配失败")
	}
	choose := "larg/"
	if !choosepic {
		choose = "picsCN/"
	}
	picHref := "https://www.ygo-sem.cn/yugioh/" + choose + cardpic[1] + ".jpg"
	// 读取获取的[]byte数据
	return web.RequestDataWith(web.NewDefaultClient(), picHref, reqconf[0], reqconf[1], reqconf[2], nil)
}

// 获取描述
func getDescribe(body string) string {
	cardName := regexpmatchByZero(`<b>中文名</b> </span>&nbsp;<span class="item_box_value">\s*(?s:(.*?))\s*</span>\s*</div>`, body)
	if len(cardName) == 0 {
		return "查无此卡"
	}
	describeinfo := regexpmatchByZero(`<span class="cont-list">\s*(?s:(.*?))\s*<span style="display:block;`, body)
	if len(describeinfo) == 0 {
		return "无相关描述,请期待更新"
	}
	getdescribe := strings.ReplaceAll(describeinfo[1], "\r\n", "")
	getdescribe = strings.ReplaceAll(getdescribe, " ", "")
	href1 := regexpmatch(`<span(.*?)data-content=(.*?)'>(.*?)</span>`, getdescribe)
	if len(href1) != 0 {
		for _, hrefv := range href1 {
			getdescribe = strings.ReplaceAll(getdescribe, hrefv[0], "「"+hrefv[3]+"」")
		}
	}
	href2 := regexpmatch(`<ahref='(.*?)'target='_blank'>(.*?)</a>`, getdescribe)
	if len(href2) != 0 {
		for _, hrefv := range href2 {
			getdescribe = strings.ReplaceAll(getdescribe, hrefv[0], hrefv[2])
		}
	}
	getdescribe = strings.ReplaceAll(getdescribe, "</span>", "")
	getdescribe = strings.ReplaceAll(getdescribe, "<br/>", "\r\n")
	getdescribe = strings.ReplaceAll(getdescribe, "<br />", "\n")
	return "卡名：" + cardName[1] + "\n\n描述:\n" + getdescribe
}

// 获取调整
func getAdjustment(body string) string {
	adjustment := regexpmatch(`<div class="accordion-inner" id="adjust">\s*<table class="table">\s*<tbody>\s*<tr>\s*<td>\s*(?s:(.*?))\s*</td>`, body)
	if len(adjustment) == 0 {
		return "无相关调整，可以尝试搜索相关效果的旧卡"
	}
	adjust := strings.ReplaceAll(adjustment[0][1], "<br/>", "\n")
	return strings.ReplaceAll(adjust, "<br />", "\n")
}
