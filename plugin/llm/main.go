// Package llm 大模型聊天和群聊总结
package llm

import (
	"strconv"
	"strings"
	"time"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

var (
	// en data [8 temp] [8 rate] LSB
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "大模型聊天和群聊总结",
		Help: "- 群聊总结 [消息数目]|群聊总结 1000\n" +
			"- /gpt [内容] （使用大模型聊天）\n",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 {
			if ctx.Event.GroupID == 0 {
				return -ctx.Event.UserID
			}
			return ctx.Event.GroupID
		}),
		// no post option, silently quit
	))
)

var (
	limit = ctxext.NewLimiterManager(time.Second*30, 1)
)

func init() {
	// 添加群聊总结功能
	en.OnRegex(`^群聊总结\s?(\d*)$`, chat.EnsureConfig, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).Limit(limit.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("少女思考中..."))
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		p, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		if p > 1000 {
			p = 1000
		}
		if p == 0 {
			p = 200
		}
		group := ctx.GetGroupInfo(gid, false)
		if group.MemberCount == 0 {
			ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "未加入", group.Name, "(", gid, "),无法获取总结"))
			return
		}

		var messages []string

		h := ctx.GetGroupMessageHistory(gid, 0, p, false)
		h.Get("messages").ForEach(func(_, msgObj gjson.Result) bool {
			nickname := msgObj.Get("sender.nickname").Str
			text := strings.TrimSpace(message.ParseMessageFromString(msgObj.Get("raw_message").Str).ExtractPlainText())
			if text != "" {
				messages = append(messages, nickname+": "+text)
			}
			return true
		})

		if len(messages) == 0 {
			ctx.SendChain(message.Text("ERROR: 历史消息为空或者无法获得历史消息"))
			return
		}

		// 构造总结请求提示 (使用通用版省流提示词)
		// 使用反引号定义多行字符串，更清晰
		promptTemplate := `请对以下群聊对话进行【极简总结】。
要求：
1. 剔除客套与废话，直击主题。
2. 使用 Markdown 列表格式。
3. 按以下结构输出：
   - 🎯 核心议题：(一句话概括)
   - 💡 关键观点/结论：(提取3-5个重点)
   - ✅ 下一步/待办：(如果有，明确谁做什么)

群聊对话内容如下：
`
		summaryPrompt := promptTemplate + strings.Join(messages, "\n")

		stor, err := chat.NewStorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 调用大模型API进行总结
		summary, err := llmchat(summaryPrompt, stor.Temp())

		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		var b strings.Builder
		b.WriteString("群 ")
		b.WriteString(group.Name)
		b.WriteByte('(')
		b.WriteString(strconv.FormatInt(gid, 10))
		b.WriteString(") 的 ")
		b.WriteString(strconv.FormatInt(p, 10))
		b.WriteString(" 条消息总结:\n\n")
		b.WriteString(summary)

		// 分割总结内容为多段（按1000字符长度切割）
		summaryText := b.String()
		msg := make(message.Message, 0)
		for len(summaryText) > 0 {
			if len(summaryText) <= 1000 {
				msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Text(summaryText)))
				break
			}

			// 查找1000字符内的最后一个换行符，尽量在换行处分割
			chunk := summaryText[:1000]
			lastNewline := strings.LastIndex(chunk, "\n")
			if lastNewline > 0 {
				chunk = summaryText[:lastNewline+1]
			}

			msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Text(chunk)))
			summaryText = summaryText[len(chunk):]
		}
		if len(msg) > 0 {
			ctx.Send(msg)
		}
	})

	// 添加 /gpt 命令处理（同时支持回复消息和直接使用）
	en.OnKeyword("/gpt", chat.EnsureConfig).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		text := ctx.MessageString()

		var query string
		var replyContent string

		// 检查是否是回复消息 (使用MessageElement检查而不是CQ码)
		for _, elem := range ctx.Event.Message {
			if elem.Type == "reply" {
				// 提取被回复的消息ID
				replyIDStr := elem.Data["id"]
				replyID, err := strconv.ParseInt(replyIDStr, 10, 64)
				if err == nil {
					// 获取被回复的消息内容
					replyMsg := ctx.GetMessage(replyID)
					if replyMsg.Elements != nil {
						replyContent = replyMsg.Elements.ExtractPlainText()
					}
				}
				break // 找到回复元素后退出循环
			}
		}

		// 提取 /gpt 后面的内容
		parts := strings.SplitN(text, "/gpt", 2)

		var gContent string
		if len(parts) > 1 {
			gContent = strings.TrimSpace(parts[1])
		}

		// 组合内容：优先使用回复内容，如果同时有/gpt内容则拼接
		switch {
		case replyContent != "" && gContent != "":
			query = replyContent + "\n" + gContent
		case replyContent != "":
			query = replyContent
		case gContent != "":
			query = gContent
		default:
			return
		}

		stor, err := chat.NewStorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 调用大模型API进行聊天
		reply, err := llmchat(query, stor.Temp())
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		// 分割总结内容为多段（按1000字符长度切割）
		msg := make(message.Message, 0)
		for len(reply) > 0 {
			if len(reply) <= 1000 {
				msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Text(reply)))
				break
			}

			// 查找1000字符内的最后一个换行符，尽量在换行处分割
			chunk := reply[:1000]
			lastNewline := strings.LastIndex(chunk, "\n")
			if lastNewline > 0 {
				chunk = reply[:lastNewline+1]
			}

			msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Text(chunk)))
			reply = reply[len(chunk):]
		}
		if len(msg) > 0 {
			ctx.Send(msg)
		}
	})
}

// llmchat 调用大模型API包装
func llmchat(prompt string, temp float32) (string, error) {
	topp, maxn := chat.AC.MParams()

	x := deepinfra.NewAPI(chat.AC.API, string(chat.AC.Key))

	mod, err := chat.AC.Type.Protocol(chat.AC.ModelName, temp, topp, maxn)
	if err != nil {
		return "", nil
	}

	data, err := x.Request(mod.User(model.NewContentText(prompt)))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(data), nil
}
