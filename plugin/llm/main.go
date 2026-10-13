// Package llm 大模型聊天和群聊总结
package llm

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/internal/replystats"
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
	limit = newRequestQuota(30*time.Second, 4096)
)

func init() {
	// 添加群聊总结功能
	en.OnRegex(`^群聊总结\s?(\d*)$`, chat.EnsureConfig, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if !allowRequest(ctx) {
			return
		}
		task := replystats.Default.Begin(ctx, replystats.GPT)
		defer task.End()
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
			task.Fail()
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
			task.Fail()
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
			task.Fail()
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 调用大模型API进行总结
		summary, err := requestLLM(chat.ConfigOf(ctx), summaryPrompt, stor.Temp())

		if err != nil {
			task.Fail()
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

		sendLLMReply(ctx, b.String(), task)
	})

	// 添加 /gpt 命令处理（同时支持回复消息和直接使用）
	en.OnMessage(func(ctx *zero.Ctx) bool {
		arg, matched := parseGPT(ctx.Event.Message.ExtractPlainText())
		if matched {
			ctx.State["gpt_query"] = arg
		}
		return matched
	}, chat.EnsureConfig).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}

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

		gContent := ctx.State["gpt_query"].(string)

		// Distinguish source text from the user's instruction about that text.
		query = buildGPTQuery(replyContent, gContent)
		if query == "" {
			ctx.SendChain(message.Text("请在 /gpt 后填写问题，或回复一条消息后发送 /gpt。"))
			return
		}

		stor, err := chat.NewStorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if !allowRequest(ctx) {
			return
		}
		task := replystats.Default.Begin(ctx, replystats.GPT)
		defer task.End()
		// 调用大模型API进行聊天
		reply, err := requestLLM(chat.ConfigOf(ctx), query, stor.Temp())
		if err != nil {
			task.Fail()
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		sendLLMReply(ctx, reply, task)
	})
}

// Short replies stay readable in the conversation; longer replies retain
// their complete content in the existing merged-forward format.
func sendLLMReply(ctx *zero.Ctx, text string, task *replystats.Task) {
	if text == "" {
		return
	}
	if utf8.RuneCountInString(text) <= 200 {
		task.SendChain(ctx, message.Text(text))
		return
	}
	msg := make(message.Message, 0)
	for _, chunk := range splitReply(text, 1000) {
		msg = append(msg, ctxext.FakeSenderForwardNode(ctx, message.Text(chunk)))
	}
	task.Send(ctx, msg)
}

func allowRequest(ctx *zero.Ctx) bool {
	location := ctx.Event.GroupID
	if location == 0 {
		location = -ctx.Event.UserID
	}
	if limit.Allow(location, time.Now()) {
		return true
	}
	ctx.SendChain(message.Text("请求较频繁，同一群或私聊每 30 秒最多一次，请稍后重试。"))
	return false
}
