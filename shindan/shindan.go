/*
基于 https://shindanmaker.com 的测定小功能
*/
package shindan

import (
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	xpath "github.com/antchfx/htmlquery"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var limit = rate.NewManager(time.Minute*5, 5)

func init() {
	table := map[string]string{
		// 可自行添加，前面为指令名，后面为网页ID
		"异世界转生":   "587874",
		"卖萌":      "360578",
		"今天是什么少女": "162207",
	}
	zero.OnMessage(HasTableKey(table), GetName()).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.Send("请稍后重试0x0...")
				return
			}

			type_ := ctx.State["key"].(string)
			name := ctx.State["name"].(string)

			text, err := shindanmaker(table[type_], name)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			ctx.SendChain(message.Text(text))
		})
}

// HasTableKey 如果 ctx.Event.RawMessage 是以 table 的 key 开头，返回真
func HasTableKey(table map[string]string) func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		msg := ctx.Event.RawMessage
		for key := range table {
			if strings.HasPrefix(msg, key) {
				temp := strings.TrimPrefix(msg, key)
				ctx.State["key"] = key
				ctx.State["trim_key"] = temp
				return true
			}
		}
		return false
	}
}

// GetName 获取名字
// 如果 ctx.State["trim_key"] 为空
// 则 ctx.State["name"] 为发送者的 名片 昵称 群头衔
// 如果 rule：HasTableKey 中 ctx.State["trim_key"] 是艾特
// 则 ctx.State["name"] 为 被艾特的人的 名片 昵称 群头衔
// 否则 ctx.State["name"] 为 ctx.State["trim_key"]
func GetName() func(ctx *zero.Ctx) bool {
	return func(ctx *zero.Ctx) bool {
		name := ctx.State["trim_key"].(string)
		arr := message.ParseMessageFromString(name)
		var qq int64 = 0
		switch {
		case name == "":
			qq = ctx.Event.UserID
		case arr[0].Type == "at":
			qq, _ = strconv.ParseInt(arr[0].Data["qq"], 10, 64)
		}
		// 获取名字
		info := ctx.GetGroupMemberInfo(ctx.Event.GroupID, qq, false)
		switch {
		case info.Get("nickname").Str != "":
			name = info.Get("nickname").Str
		case info.Get("card").Str != "":
			name = info.Get("card").Str
		case info.Get("title").Str != "":
			name = info.Get("title").Str
		}
		temp := []rune(name)
		if len(temp) > 20 {
			// 防止超长名字
			temp = temp[:20]
		}
		name = string(temp)
		ctx.State["name"] = name
		return true
	}
}

// shindanmaker 返回网页 shindanmaker.com 测定结果
func shindanmaker(id, name string) (string, error) {
	url := "https://shindanmaker.com/" + id
	// seed 使每一天的结果都不同
	now := time.Now()
	seed := fmt.Sprintf("%d%d%d", now.Year(), now.Month(), now.Day())
	name = name + seed

	// 组装参数
	client := &http.Client{}
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("_token", "D6Rn2BTOTMht2FBKNhpwOWqymZuoifLNt31RqlLj")
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("hiddenName", "名無しのV")
	_ = writer.Close()
	// 发送请求
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Cookie", "_ga=GA1.2.531883706.1618563450; trc_cookie_storage=taboola%20global%3Auser-id=09ef6d7e-ddb2-4a3f-ac71-b75801d11fef-tuct7602eee; dui=eyJpdiI6ImJ0MThXNTFoZ1hVOXNWZXZFeTRDdlE9PSIsInZhbHVlIjoibUF3em14ZzhXbVJsL1k5b284Tzh5QUhHa0VSYWtYZjF3KzZNRkxtRmVIMHBUbnJ1M0JwNTEzaWhWTzZTT1NGcGVxdFhGaG11Y052a3hyUUZVZDRtNFc0ZndiRnVTSHJFL1hzak43QzNhYVd6cEFEZVJxZVNNNytpNHhQbnF0TGpmbXNNdFlUNjFwNVNubnVJbkN5NHNRPT0iLCJtYWMiOiIzOTNmNmU2NmM3ZmNlMWU5OTA1ZDFlZTJlYTRiZThlZDViMzEwN2VmN2Y0NDEyZjNkYmM5ZGJhNTllMGU1ZjE3In0=; _session=XAj6V877yp1DE8Cb405837ySa0t6fYHM21R2HZ8Z; _gid=GA1.2.2089711003.1622961890; _gat_UA-19089743-2=1; _gat_UA-19089743-3=1; __gads=ID=b597940057bf81ba:T=1622961888:S=ALNI_MY9F-63AstFh3E3tS-DTVh08KgjJg; dsr=eyJpdiI6IlhBWTRYdk14SysyNms5VVpoMFUzMFE9PSIsInZhbHVlIjoiZUl2S2ZSL3M5a2RwT253MlpoQnJPb1NJQjZ1RUJLOWtTWnFXeWpvOG9XUnAwSGw4MGMyVDVIZjJiN0VSSUd6Vkt0V0wreEpEb3d6M2ZDZE51UzJDTGc9PSIsIm1hYyI6IjRkZjU5MjJhMTVhZjQwOGY4MjRhYzhiMjJkMzg0YTFhNzQ1YWVkODZmYjEyMjA5ODliNDdkZGQzMzVkOTdjNGIifQ==; name=eyJpdiI6Ilp0TWxIeG1scW80VWlyTFJQUG4yZWc9PSIsInZhbHVlIjoiL1NqMnJyKzhtdW1hUnFBYjhvUVhVeW9EVWdMSjdjSklvdUsrRk5Id0lrOHdtcjVDU010QkovQjAxYkZZT2Q3TSIsIm1hYyI6IjU4ZDc4NGQzNTMyMzJlZjk0YjZmNjBiMjkzNTAyYTQ0ZDg4NGNkZjhmMDY2ODk2YjdkOTdkZTY3MDlmYzdhYjkifQ==; XSRF-TOKEN=eyJpdiI6IjJrRzRiZTVYcldiL09XSURJTDJQYVE9PSIsInZhbHVlIjoicC8yVzV0cnFQQ2RXOG0xeDdUNTc1V2RmYlJIWnNiTjJ5UVFHMGlTOUhPT3VOOWtZbFFLS0d4QUxDTjdYTEdrYWdvbUJ5Y24wOVpmZjNVdTdtUUNkNmMrVDk0U3RoT0NsZmxZVWIveTh3QU9PT25aLzd1UndpUWVvTlZ2Tjd2c3IiLCJtYWMiOiIyYjdlNjFhYTAzYTIyZTdjMDIwNTZkNjIwMDJlMTI3MTZkNzhjYzdkMzIyNjdmNzFmYzI1ZmE5NzczZTVmZWVjIn0=")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 解析XPATH
	doc, err := xpath.Parse(resp.Body)
	if err != nil {
		return "", err
	}
	// 取出每个返回的结果
	list := xpath.Find(doc, `//*[@id="shindanResult"]`)
	if len(list) == 0 {
		return "", errors.New("无法查找到结果，可能 token 失效，请提交 issue")
	}
	var output = []string{}
	for child := list[0].FirstChild; child != nil; child = child.NextSibling {
		if text := xpath.InnerText(child); text != "" {
			output = append(output, text)
		} else {
			output = append(output, "\n")
		}
	}
	return strings.ReplaceAll(strings.Join(output, ""), seed, ""), nil
}
