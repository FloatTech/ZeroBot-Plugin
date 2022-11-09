// Package jikipedia 小鸡词典
// 修改自https://github.com/TeamPGM/PagerMaid_Plugins_Pyro ，非常感谢！！
package jikipedia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	url = "https://api.jikipedia.com/go/search_entities"
)

type value struct {
	Phrase string `json:"phrase"`
	Page   int    `json:"page"`
	Size   int    `json:"size"`
}

func init() {
	// 初始化engine
	engine := control.Register("jikipedia", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "小鸡词典",
		Help:             "- [查梗|小鸡词典][梗]",
	},
	)
	engine.OnPrefixGroup([]string{"小鸡词典", "查梗"}).Limit(ctxext.LimitByGroup).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			keyWord := strings.Trim(ctx.State["args"].(string), " ")

			definition, err := parseKeyword(keyWord)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if definition.String() == "" {
				ctx.SendChain(message.Text("好像什么都没查到，换个关键词试一试？"))
				return
			}
			imgURL := definition.Get("images.0.scaled.path").String()
			ctx.SendChain(message.Text("【标题】:", definition.Get("term.title"),
				"\n【释义】:", definition.Get("plaintext"),
				"\n【原文】:https://jikipedia.com/definition/", definition.Get("id")),
				message.Image(imgURL))
		},
	)
}

func parseKeyword(keyWord string) (definition gjson.Result, err error) {
	client := &http.Client{}

	values := value{Phrase: keyWord, Page: 1, Size: 10}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return
	}
	var request *http.Request
	request, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	request.Header = http.Header{
		"Accept":             {"application/json, text/plain, */*"},
		"Accept-Encoding":    {"gzip, deflate, br"},
		"Accept-Language":    {"zh-CN,zh-TW;q=0.9,zh;q=0.8"},
		"Client":             {"web"},
		"Client-Version":     {"2.7.2g"},
		"Connection":         {"keep-alive"},
		"Host":               {"api.jikipedia.com"},
		"Origin":             {"https://jikipedia.com"},
		"Referer":            {"https://jikipedia.com/"},
		"Sec-Fetch-Dest":     {"empty"},
		"Sec-Fetch-Mode":     {"cors"},
		"Sec-Fetch-Site":     {"same-site"},
		"Token":              {""},
		"User-Agent":         {"Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.0.0 Mobile Safari/537.36"},
		"XID":                {"uNo5bL1nyNCp/Gm7lJAHQ91220HLbMT8jqk9IJYhtHA4ofP+zgxwM6lSDIKiYoppP2k1IW/1Vxc2vOVGxOOVReebsLmWPHhTs7NCRygfDkE="},
		"sec-ch-ua":          {`" Not A;Brand";v="99", "Chromium";v="102", "Google Chrome";v="102"`},
		"sec-ch-ua-mobile":   {"?1"},
		"sec-ch-ua-platform": {`"Android"`},
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	var response *http.Response
	response, err = client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		extraInfo := ""
		if response.StatusCode == 423 {
			extraInfo = "\n调用过多被网站暂时封禁，请等待数个小时后使用该功能~"
		}
		s := fmt.Sprintf("status code: %d%s", response.StatusCode, extraInfo)
		err = errors.New(s)
		return
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	gjson.Get(binary.BytesToString(data), "data").ForEach(func(key, value gjson.Result) bool {
		definition = value.Get("definitions.0")
		return definition.String() == ""
	})
	return
}
