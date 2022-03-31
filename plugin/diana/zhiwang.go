// Package diana 嘉然相关
package diana

import (
	"bytes"
	"io"
	"math"
	"time"

	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"

	"net/http"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 小作文查重: 回复要查的消息 查重
func init() {
	engine.OnMessage(fullmatch("查重")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			msg := ctx.Event.Message
			if msg[0].Type == "reply" {
				msg := ctx.GetMessage(message.NewMessageID(msg[0].Data["id"])).Elements[0].Data["text"]
				zhiwangjson := zhiwangapi(msg)
				if zhiwangjson == nil || zhiwangjson.Get("code").Int() != 0 {
					ctx.SendChain(message.Text("api返回错误"))
					return
				}

				if zhiwangjson.Get("data.related.#").Int() == 0 {
					ctx.SendChain(message.Text("枝网没搜到，查重率为0%，一眼丁真，鉴定为原创"))
					return
				}
				related := zhiwangjson.Get("data.related.0.reply").Map()
				rate := zhiwangjson.Get("data.related.0.rate").Float()
				ctx.SendChain(message.Text(
					"枝网文本复制检测报告(简洁)", "\n",
					"查重时间: ", time.Now().Format("2006-01-02 15:04:05"), "\n",
					"总文字复制比: ", math.Floor(rate*100), "%", "\n",
					"相似小作文：", "\n",
					related["content"].String()[:102]+".....", "\n",
					"获赞数：", related["like_num"].String(), "\n",
					zhiwangjson.Get("data.related.0.reply_url").String(), "\n",
					"作者: ", related["m_name"].String(), "\n",
					"发表时间: ", time.Unix(int64(related["ctime"].Float()), 0).Format("2006-01-02 15:04:05"), "\n",
					"查重结果仅作参考，请注意辨别是否为原创", "\n",
					"数据来源: https://asoulcnki.asia/",
				))
			}
		})
}

func zhiwangapi(text string) *gjson.Result {
	url := "https://asoulcnki.asia/v1/api/check"
	post := "{\n\"text\":\"" + text + "\"\n}"
	var jsonStr = []byte(post)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	resp.Body.Close()
	result := gjson.ParseBytes(bodyBytes)
	return &result
}

func fullmatch(src ...string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		for _, elem := range msg {
			if elem.Type == "text" {
				text := elem.Data["text"]
				text = strings.ReplaceAll(text, " ", "")
				text = strings.ReplaceAll(text, "\r", "")
				text = strings.ReplaceAll(text, "\n", "")
				for _, s := range src {
					if text == s {
						return true
					}
				}
			}
		}
		return false
	}
}
