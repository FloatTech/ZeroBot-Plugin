package shindan

import (
	"bytes"
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

			//ctx.SendChain(message.Text("少女祈祷中......"))

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
// 如果 rule：HasTableKey 中 ctx.State["trim_key "] 是艾特
// 则 ctx.State["name"] 为 被艾特的人的 名片 昵称 群头衔
// 否则 ctx.State["name"] 为 ctx.State["trim_key "]
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

func shindanmaker(id, name string) (string, error) {
	url := "https://shindanmaker.com/" + id
	// 使每一天的结果都不同
	now := time.Now()
	seed := fmt.Sprintf("%d%d%d", now.Year(), now.Month(), now.Day())
	name = name + seed
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("_token", "2HgIQ1okTbr1SZ5EXDQWvFXQEBoerVRK8CSW0xro")
	_ = writer.WriteField("name", name)
	_ = writer.WriteField("hiddenName", "名無しのV")
	_ = writer.Close()

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Cookie", "_session=8gvUXS6oVG3hZ1vBf2QlOVGquMh0UwJEuBwHIjNf; _ga=GA1.2.531883706.1618563450; _gid=GA1.2.622904960.1618563450; trc_cookie_storage=taboola%2520global%253Auser-id%3D09ef6d7e-ddb2-4a3f-ac71-b75801d11fef-tuct7602eee; dui=eyJpdiI6ImJ0MThXNTFoZ1hVOXNWZXZFeTRDdlE9PSIsInZhbHVlIjoibUF3em14ZzhXbVJsL1k5b284Tzh5QUhHa0VSYWtYZjF3KzZNRkxtRmVIMHBUbnJ1M0JwNTEzaWhWTzZTT1NGcGVxdFhGaG11Y052a3hyUUZVZDRtNFc0ZndiRnVTSHJFL1hzak43QzNhYVd6cEFEZVJxZVNNNytpNHhQbnF0TGpmbXNNdFlUNjFwNVNubnVJbkN5NHNRPT0iLCJtYWMiOiIzOTNmNmU2NmM3ZmNlMWU5OTA1ZDFlZTJlYTRiZThlZDViMzEwN2VmN2Y0NDEyZjNkYmM5ZGJhNTllMGU1ZjE3In0%3D; XSRF-TOKEN=eyJpdiI6Ikw0eTU4eFFtSytCL3VySlpuSG1UaVE9PSIsInZhbHVlIjoiL3h5WVkvSDdEaWZVdGY1M1hmalpYUDM2Y0IzSlFIL3BFTUd6MXczekFpZWlDcHVxTk5TSVF1SzhhVm53T1dYbUd5dS8vL1FjMm9yYVlGOGJReTI4c2JoV0tnNURhMHhBODBrQ1RxYkhnQ1ZsSXoxMmlUNVNrQ241cnlFKzQzWDciLCJtYWMiOiI0NDhiMzI3NzNmOGJjNmM1ZDk1ZjYwMTFmYzk4YjQ5ZDQ3OWY0MGY1YTkwNGIyY2U5YzhiMTQ1ZGQyNDhiM2NjIn0%3D; name=eyJpdiI6IjBrTVE0RFl4Nm9kTVVMb2EvYVNPQmc9PSIsInZhbHVlIjoiV3o0b1R3azhKZDZaOE1udkFrVis1L0xsdHJoZXZsKzlkRk9DR3FTQU9XalU4c0kyc3M2ckdTMjVYSjFjT0RPeSIsIm1hYyI6IjI2ZDMzZjM0YThhNzUwOGU2NGYxNGJmZTk3YjI3ZDU4YzBmMjJkMTZkMDQyOTE2NTczODA4NzA1NDQ2OTIzMGUifQ%3D%3D; dsr=eyJpdiI6IlZPQnhVM3RWVkU5VDcxd3dmRWVBU1E9PSIsInZhbHVlIjoiREFsOEZoUHozaHBxbTJDS1V2OVR3djZtNU5TQkEreW5GTHhYaWRRbG5WeENDdXJsOG9RVDNKclgyQkJpdEZnNkFqMTQwcmJSTlBoenBxZDVSanQ5TFE9PSIsIm1hYyI6ImNiNjdjOTFmYjYzOTAwMTc3MDdhNzBhMDhhMTQ0M2Q0ODM1YzIwMDJhNmQ4MWVmZTJjZDFiMmYyM2IxYTNmNjIifQ%3D%3D; __gads=ID=b597940057bf81ba:T=1618572481:S=ALNI_MasEvf_XV_9a4OWpVPI2UNR4vOswQ; XSRF-TOKEN=eyJpdiI6InQ0MkcvTGJjZVNabnI1MUV6K1Y4b1E9PSIsInZhbHVlIjoiR2hOZ0FiTDVBM1ZPYUMzbEJGRUJiWVIyNWlHN0VRUVc1NStYMjMrWmVWRHE0R1ZQSDZXMkhWTHFYU21MczRkSDJUZnBWT1hzWnl2VEVRbXhOdzdWNEErRGM0eUYyOEdIWVBrekQ4TkdLRlcwSzVKOWJtMmJSZkVUTUVNZmprNnEiLCJtYWMiOiI5ZDc3ZDEwNjQ3NTVhMTFiYTg5YTNiM2JiNTc3NjYyYWQ1MjY2ZmE0MmMwNGQyM2I4MjRmY2I2MmEzOWRlYzdkIn0%3D; name=eyJpdiI6ImFDRTNheSsra09GYnVvWVJieDRxSWc9PSIsInZhbHVlIjoicEZpdEtqMVNOZitPRS8wdlJqVVdiZGpkdkFKek5JYlNoM3E5b2wzakxJLzdPZmJBeTBkeTRQcGZtM0pFWEtqLyIsIm1hYyI6IjBiNDA0MDI1ZjU1ZDNmNDIzODE5OWFmNjZhNDA3MTU5OWY1MzI5YTI3ZTg5YzU3YWVjZDJmNGNmZmNkZWQwZDcifQ%3D%3D; _session=8gvUXS6oVG3hZ1vBf2QlOVGquMh0UwJEuBwHIjNf")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.ContentLength <= 0 {
		return "出错啦", nil
	}
	// 解析XPATH
	doc, err := xpath.Parse(resp.Body)
	if err != nil {
		return "", err
	}
	// 取出每个返回的结果
	list := xpath.Find(doc, `//*[@id="shindanResult"]`)
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
