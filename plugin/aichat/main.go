// Package aichat OpenAI聊天和群聊总结
package aichat

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	goba "github.com/fumiama/go-onebot-agent"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/airecord"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

var (
	// en data [8 temp] [8 rate] LSB
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("aichat"),
		Brief:            "OpenAI聊天",
		Help: "- 设置AI聊天触发概率10\n" +
			"- 设置AI聊天温度80\n" +
			"- 设置AI聊天(识图|Agent)接口类型[OpenAI|OLLaMA|GenAI]\n" +
			"- 设置AI聊天(不)使用Agent模式\n" +
			"- 设置AI聊天(不)支持系统提示词\n" +
			"- 设置AI聊天(识图|Agent)接口地址https://api.siliconflow.cn/v1/chat/completions\n" +
			"- 设置AI聊天(识图|Agent)密钥xxx\n" +
			"- 设置AI聊天(识图|Agent)模型名Qwen/Qwen3-8B\n" +
			"- 查看AI聊天系统提示词\n" +
			"- 重置AI聊天系统提示词\n" +
			"- 设置AI聊天系统提示词xxx\n" +
			"- 设置AI聊天分隔符</think>(留空则清除)\n" +
			"- 设置AI聊天(不)响应AT\n" +
			"- 设置AI聊天最大长度4096\n" +
			"- 设置AI聊天TopP 0.9\n" +
			"- 设置AI聊天(不)以AI语音输出\n" +
			"- 查看AI聊天配置\n" +
			"- 重置AI聊天\n" +
			"- 群聊总结 [消息数目]|群聊总结 1000\n" +
			"- /gpt [内容] （使用大模型聊天）\n",

		PrivateDataFolder: "aichat",
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
	en.OnMessage(ensureconfig, func(ctx *zero.Ctx) bool {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor, err := newstorage(ctx, gid)
		if err != nil {
			logrus.Warnln("ERROR: ", err)
			return false
		}
		ctx.State["__aichat_stor__"] = stor
		return ctx.ExtractPlainText() != "" &&
			(!stor.noreplyat() || (stor.noreplyat() && !ctx.Event.IsToMe))
	}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor := ctx.State["__aichat_stor__"].(storage)
		rate := stor.rate()
		if !ctx.Event.IsToMe && rand.Intn(100) >= int(rate) {
			return
		}
		if ctx.Event.IsToMe {
			ctx.Block()
		}
		if cfg.Key == "" {
			logrus.Warnln("ERROR: get extra err: empty key")
			return
		}
		temperature := stor.temp()
		topp, maxn := cfg.mparams()

		if !stor.noagent() && cfg.AgentAPI != "" && cfg.AgentModelName != "" {
			x := deepinfra.NewAPI(cfg.AgentAPI, string(cfg.AgentKey))
			mod, err := cfg.Type.protocol(cfg.AgentModelName, temperature, topp, maxn)
			if err != nil {
				logrus.Warnln("ERROR: ", err)
				return
			}
			role := goba.PermRoleUser
			if zero.AdminPermission(ctx) {
				role = goba.PermRoleAdmin
				if zero.SuperUserPermission(ctx) {
					role = goba.PermRoleOwner
				}
			}
			ag := chat.AgentOf(ctx.Event.SelfID)
			if cfg.ImageAPI != "" && !ag.CanViewImage() {
				mod, err := cfg.ImageType.protocol(cfg.ImageModelName, temperature, topp, maxn)
				if err != nil {
					logrus.Warnln("ERROR: ", err)
					return
				}
				ag.SetViewImageAPI(deepinfra.NewAPI(cfg.ImageAPI, string(cfg.ImageKey)), mod)
			}
			ctx.NoTimeout()
			hasresp := false
			for i := 0; i < 8; i++ { // 最大运行 8 轮因为问答上下文只有 16
				reqs := chat.CallAgent(ag, zero.SuperUserPermission(ctx), x, mod, gid, role)
				if len(reqs) == 0 {
					break
				}
				hasresp = true
				for _, req := range reqs {
					resp := ctx.CallAction(req.Action, req.Params)
					logrus.Infoln("[aichat] agent get resp:", reqs)
					ag.AddResponse(gid, &goba.APIResponse{
						Status:  resp.Status,
						Data:    json.RawMessage(resp.Data.Raw),
						Message: resp.Message,
						Wording: resp.Wording,
						RetCode: resp.RetCode,
					})
				}
			}
			if hasresp {
				ag.AddTerminus(gid)
				return
			}
			// no response, fall back to normal chat
		}

		x := deepinfra.NewAPI(cfg.API, string(cfg.Key))
		mod, err := cfg.Type.protocol(cfg.ModelName, temperature, topp, maxn)
		if err != nil {
			logrus.Warnln("ERROR: ", err)
			return
		}
		data, err := x.Request(chat.GetChatContext(mod, gid, cfg.SystemP, bool(cfg.NoSystemP)))
		if err != nil {
			logrus.Warnln("[aichat] post err:", err)
			return
		}

		txt := chat.Sanitize(strings.Trim(data, "\n 　"))
		if len(txt) > 0 {
			chat.AddChatReply(gid, txt)
			nick := zero.BotConfig.NickName[rand.Intn(len(zero.BotConfig.NickName))]
			txt = strings.ReplaceAll(txt, "{name}", ctx.CardOrNickName(ctx.Event.UserID))
			txt = strings.ReplaceAll(txt, "{me}", nick)
			id := any(nil)
			if ctx.Event.IsToMe {
				id = ctx.Event.MessageID
			}
			for _, t := range strings.Split(txt, "{segment}") {
				if t == "" {
					continue
				}
				logrus.Infoln("[aichat] 回复内容:", t)
				recCfg := airecord.GetConfig()
				record := ""
				if !stor.norecord() {
					record = ctx.GetAIRecord(recCfg.ModelID, recCfg.Customgid, t)
				}
				if record != "" {
					ctx.SendChain(message.Record(record))
				} else {
					if id != nil {
						id = ctx.SendChain(message.Reply(id), message.Text(t))
					} else {
						id = ctx.SendChain(message.Text(t))
					}
				}
				process.SleepAbout1sTo2s()
			}
		}
	})
	en.OnPrefix("设置AI聊天触发概率", zero.AdminPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBitmapHandler(bitmaprate, 0, 100))
	en.OnPrefix("设置AI聊天温度", zero.AdminPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBitmapHandler(bitmaptemp, 0, 100))
	en.OnPrefix("设置AI聊天接口类型", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetmodeltype(&cfg.Type))
	en.OnPrefix("设置AI聊天识图接口类型", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetmodeltype(&cfg.ImageType))
	en.OnPrefix("设置AI聊天Agent接口类型", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetmodeltype(&cfg.AgentType))
	en.OnPrefix("设置AI聊天接口地址", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.API))
	en.OnPrefix("设置AI聊天识图接口地址", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.ImageAPI))
	en.OnPrefix("设置AI聊天Agent接口地址", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.AgentAPI))
	en.OnPrefix("设置AI聊天密钥", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.Key))
	en.OnPrefix("设置AI聊天识图密钥", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.ImageKey))
	en.OnPrefix("设置AI聊天Agent密钥", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.AgentKey))
	en.OnPrefix("设置AI聊天模型名", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.ModelName))
	en.OnPrefix("设置AI聊天识图模型名", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.ImageModelName))
	en.OnPrefix("设置AI聊天Agent模型名", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.AgentModelName))
	en.OnPrefix("设置AI聊天系统提示词", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.SystemP))
	en.OnFullMatch("查看AI聊天系统提示词", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text(cfg.SystemP))
	})
	en.OnFullMatch("重置AI聊天系统提示词", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		cfg.SystemP = chat.SystemPrompt
		err := c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天分隔符", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.Separator))
	en.OnRegex("^设置AI聊天(不)?响应AT$", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBoolHandler(bitmapnrat))
	en.OnRegex("^设置AI聊天(不)?支持系统提示词$", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetbool(&cfg.NoSystemP))
	en.OnRegex("^设置AI聊天(不)?使用Agent模式$", ensureconfig, zero.SuperUserPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBoolHandler(bitmapnagt))
	en.OnPrefix("设置AI聊天最大长度", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetuint(&cfg.MaxN))
	en.OnPrefix("设置AI聊天TopP", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetfloat32(&cfg.TopP))
	en.OnRegex("^设置AI聊天(不)?以AI语音输出$", ensureconfig, zero.AdminPermission).SetBlock(true).
		Handle(ctxext.NewStorageSaveBoolHandler(bitmapnrec))
	en.OnFullMatch("查看AI聊天配置", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			stor, err := newstorage(ctx, gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(
				message.Text(
					"【当前AI聊天本群配置】\n",
					"• 触发概率：", int(stor.rate()), "\n",
					"• 温度：", stor.temp(), "\n",
					"• 以AI语音输出：", ModelBool(!stor.norecord()), "\n",
					"• 使用Agent：", ModelBool(!stor.noagent()), "\n",
					"• 响应@：", ModelBool(!stor.noreplyat()), "\n",
				),
				message.Text("【当前AI聊天全局配置】\n", &cfg),
			)
		})
	en.OnFullMatch("重置AI聊天", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		chat.ResetChat()
		ctx.SendChain(message.Text("成功"))
	})

	// 添加群聊总结功能
	en.OnRegex(`^群聊总结\s?(\d*)$`, ensureconfig, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).Limit(limit.LimitByGroup).Handle(func(ctx *zero.Ctx) {
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

		// 构造总结请求提示
		summaryPrompt := "请总结这个群聊内容，要求按发言顺序梳理，明确标注每个发言者的昵称，并完整呈现其核心观点、提出的问题、发表的看法或做出的回应，确保不遗漏关键信息，且能体现成员间的对话逻辑和互动关系:\n" +
			strings.Join(messages, "\n")

		stor, err := newstorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 调用大模型API进行总结
		summary, err := llmchat(summaryPrompt, stor.temp())

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
	en.OnKeyword("/gpt", ensureconfig).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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

		stor, err := newstorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// 调用大模型API进行聊天
		reply, err := llmchat(query, stor.temp())
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
	topp, maxn := cfg.mparams()

	x := deepinfra.NewAPI(cfg.API, string(cfg.Key))

	mod, err := cfg.Type.protocol(cfg.ModelName, temp, topp, maxn)
	if err != nil {
		return "", nil
	}

	data, err := x.Request(mod.User(model.NewContentText(prompt)))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(data), nil
}
