// Package crypter 处理函数
package crypter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/FloatTech/AnimeAPI/airecord"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var faceTagRe = regexp.MustCompile(`\{\{face:(\d+)\}\}`)

func parseID(v interface{}) int64 {
	n, _ := strconv.ParseInt(fmt.Sprintf("%v", v), 10, 64)
	return n
}

func serializeMsg(segs message.Message) string {
	var sb strings.Builder
	for _, seg := range segs {
		switch seg.Type {
		case "text":
			sb.WriteString(fmt.Sprintf("%v", seg.Data["text"]))
		case "face":
			fmt.Fprintf(&sb, "{{face:%v}}", seg.Data["id"])
		}
	}
	return sb.String()
}

func deserializeMsg(s string) message.Message {
	var msg message.Message
	last := 0
	for _, loc := range faceTagRe.FindAllStringSubmatchIndex(s, -1) {
		if loc[0] > last {
			msg = append(msg, message.Text(s[last:loc[0]]))
		}
		id, _ := strconv.Atoi(s[loc[2]:loc[3]])
		msg = append(msg, message.Face(id))
		last = loc[1]
	}
	if last < len(s) {
		msg = append(msg, message.Text(s[last:]))
	}
	return msg
}

func getInput(ctx *zero.Ctx, cmds ...string) string {
	full := serializeMsg(ctx.Event.Message)
	for _, cmd := range cmds {
		if idx := strings.Index(full, cmd); idx >= 0 {
			return strings.TrimSpace(full[idx+len(cmd):])
		}
	}
	return ""
}

func getReplyContent(ctx *zero.Ctx) string {
	for _, seg := range ctx.Event.Message {
		if seg.Type == "reply" {
			if msgID := parseID(seg.Data["id"]); msgID > 0 {
				if msg := ctx.GetMessage(msgID); msg.Elements != nil {
					return serializeMsg(msg.Elements)
				}
			}
		}
	}
	return ""
}

func getReplyFaceIDs(ctx *zero.Ctx) []int {
	for _, seg := range ctx.Event.Message {
		if seg.Type == "reply" {
			if msgID := parseID(seg.Data["id"]); msgID > 0 {
				return extractFaceIDs(ctx.GetMessage(msgID).Elements)
			}
		}
	}
	return nil
}

func extractFaceIDs(segs message.Message) []int {
	var ids []int
	for _, seg := range segs {
		if seg.Type == "face" {
			if id := int(parseID(seg.Data["id"])); id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// hou
func houEncryptHandler(ctx *zero.Ctx) {
	text := getInput(ctx, "h加密", "齁语加密")
	result := encodeHou(text)
	recCfg := airecord.GetConfig()
	if record := ctx.GetAIRecord(recCfg.ModelID, recCfg.Customgid, result); record != "" {
		ctx.SendChain(message.Record(record))
	} else {
		ctx.SendChain(message.Text(result))
	}
}

func houDecryptHandler(ctx *zero.Ctx) {
	text := getInput(ctx, "h解密", "齁语解密")
	if text == "" {
		text = getReplyContent(ctx)
	}
	if text == "" {
		ctx.SendChain(message.Text("请输入密文或回复加密消息"))
		return
	}
	ctx.SendChain(deserializeMsg(decodeHou(text))...)
}

// fumo
func fumoEncryptHandler(ctx *zero.Ctx) {
	ctx.SendChain(message.Text(encryptFumo(getInput(ctx, "fumo加密"))))
}

func fumoDecryptHandler(ctx *zero.Ctx) {
	text := getInput(ctx, "fumo解密")
	if text == "" {
		text = getReplyContent(ctx)
	}
	if text == "" {
		ctx.SendChain(message.Text("请输入密文或回复加密消息"))
		return
	}
	ctx.SendChain(deserializeMsg(decryptFumo(text))...)
}

// qq表情
func qqEmojiEncryptHandler(ctx *zero.Ctx) {
	text := getInput(ctx, "qq加密")
	if text == "" {
		ctx.SendChain(message.Text("请输入要加密的文本"))
		return
	}
	ctx.SendChain(encodeQQEmoji(text)...)
}

func qqEmojiDecryptHandler(ctx *zero.Ctx) {
	faceIDs := extractFaceIDs(ctx.Event.Message)
	if len(faceIDs) == 0 {
		faceIDs = getReplyFaceIDs(ctx)
	}
	if len(faceIDs) == 0 {
		ctx.SendChain(message.Text("请回复QQ表情加密消息进行解密"))
		return
	}
	ctx.SendChain(deserializeMsg(decodeQQEmoji(faceIDs))...)
}
