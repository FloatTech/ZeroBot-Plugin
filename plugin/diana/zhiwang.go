// Package diana 嘉然相关
package diana

import (
	"bytes"
	"math"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
	"github.com/wdvxdr1123/ZeroBot/message"

	zero "github.com/wdvxdr1123/ZeroBot"
)

// 小作文查重: 回复要查的消息 查重
func init() {
	engine.OnMessage(func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		if msg[0].Type != "reply" {
			return false
		}
		for _, elem := range msg {
			if elem.Type == "text" {
				text := elem.Data["text"]
				text = strings.ReplaceAll(text, " ", "")
				text = strings.ReplaceAll(text, "\r", "")
				text = strings.ReplaceAll(text, "\n", "")
				if text == "查重" {
					return true
				}
			}
		}
		return false
	}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		msg := ctx.GetMessage(message.NewMessageIDFromString(ctx.Event.Message[0].Data["id"])).Elements[0].Data["text"]
		result, err := zhiwangapi(msg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		if result.Get("code").Int() != 0 {
			ctx.SendChain(message.Text("api返回错误:", result.Get("code").Int()))
			return
		}
		if result.Get("data.related.#").Int() == 0 {
			ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("枝网没搜到，查重率为0%，鉴定为原创")))
			return
		}
		related := result.Get("data.related.0.reply").Map()
		rate := result.Get("data.related.0.rate").Float()
		relatedcontent := related["content"].String()
		if len(relatedcontent) > 102 {
			relatedcontent = relatedcontent[:102] + "....."
		}
		ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.Text(
			"枝网文本复制检测报告(简洁)", "\n",
			"查重时间: ", time.Now().Format("2006-01-02 15:04:05"), "\n",
			"总文字复制比: ", math.Floor(rate*100), "%", "\n",
			"相似小作文：", "\n", relatedcontent, "\n",
			"获赞数：", related["like_num"].String(), "\n",
			result.Get("data.related.0.reply_url").String(), "\n",
			"作者: ", related["m_name"].String(), "\n",
			"发表时间: ", time.Unix(int64(related["ctime"].Float()), 0).Format("2006-01-02 15:04:05"), "\n",
			"查重结果仅作参考，请注意辨别是否为原创", "\n",
			"数据来源: https://asoulcnki.asia/",
		)))
	})
}

func zhiwangapi(text string) (*gjson.Result, error) {
	b, cl := binary.OpenWriterF(func(w *binary.Writer) {
		w.WriteString("{\n\"text\":\"")
		w.WriteString(text)
		w.WriteString("\"\n}")
	})

	data, err := web.PostData("https://asoulcnki.asia/v1/api/check", "application/json", bytes.NewReader(b))
	cl()
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(data)
	return &result, nil
}
