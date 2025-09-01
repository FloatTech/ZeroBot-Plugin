// Package aichat OpenAI聊天和群聊总结
package aichat

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
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
			"- 设置AI聊天接口类型[OpenAI|OLLaMA|GenAI]\n" +
			"- 设置AI聊天(不)支持系统提示词\n" +
			"- 设置AI聊天接口地址https://api.siliconflow.cn/v1/chat/completions\n" +
			"- 设置AI聊天密钥xxx\n" +
			"- 设置AI聊天模型名Qwen/Qwen3-8B\n" +
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
	})
)

var (
	apitypes = map[string]uint8{
		"OpenAI": 0,
		"OLLaMA": 1,
		"GenAI":  2,
	}
	apilist = [3]string{"OpenAI", "OLLaMA", "GenAI"}
	limit   = ctxext.NewLimiterManager(time.Second*30, 1)
)

// getModelParams 获取模型参数：温度(float32(temp)/100)、TopP和最大长度
func getModelParams(temp int64) (temperature float32, topp float32, maxn uint) {
	// 处理温度参数
	if temp <= 0 {
		temp = 70 // default setting
	}
	if temp > 100 {
		temp = 100
	}
	temperature = float32(temp) / 100

	// 处理TopP参数
	topp = cfg.TopP
	if topp == 0 {
		topp = 0.9
	}

	// 处理最大长度参数
	maxn = cfg.MaxN
	if maxn == 0 {
		maxn = 4096
	}

	return temperature, topp, maxn
}

func init() {
	en.OnMessage(ensureconfig, func(ctx *zero.Ctx) bool {
		return ctx.ExtractPlainText() != "" &&
			(!cfg.NoReplyAT || (cfg.NoReplyAT && !ctx.Event.IsToMe))
	}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			return
		}
		rate := c.GetData(gid)
		temp := (rate >> 8) & 0xff
		rate &= 0xff
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

		temperature, topp, maxn := getModelParams(temp)

		x := deepinfra.NewAPI(cfg.API, cfg.Key)
		var mod model.Protocol
		switch cfg.Type {
		case 0:
			mod = model.NewOpenAI(
				cfg.ModelName, cfg.Separator,
				temperature, topp, maxn,
			)
		case 1:
			mod = model.NewOLLaMA(
				cfg.ModelName, cfg.Separator,
				temperature, topp, maxn,
			)
		case 2:
			mod = model.NewGenAI(
				cfg.ModelName,
				temperature, topp, maxn,
			)
		default:
			logrus.Warnln("[aichat] unsupported AI type", cfg.Type)
			return
		}

		data, err := x.Request(chat.Ask(mod, gid, cfg.SystemP, cfg.NoSystemP))
		if err != nil {
			logrus.Warnln("[aichat] post err:", err)
			return
		}

		txt := chat.Sanitize(strings.Trim(data, "\n 　"))
		if len(txt) > 0 {
			chat.Reply(gid, txt)
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
				if !cfg.NoRecord {
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
	en.OnPrefix("设置AI聊天触发概率", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		r, err := strconv.Atoi(args)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: parse rate err: ", err))
			return
		}
		if r > 100 {
			r = 100
		} else if r < 0 {
			r = 0
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		val := c.GetData(gid) & (^0xff)
		err = c.SetData(gid, val|int64(r&0xff))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set data err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天温度", zero.AdminPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		r, err := strconv.Atoi(args)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: parse rate err: ", err))
			return
		}
		if r > 100 {
			r = 100
		} else if r < 0 {
			r = 0
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		val := c.GetData(gid) & (^0xff00)
		err = c.SetData(gid, val|(int64(r&0xff)<<8))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set data err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天接口类型", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			ctx.SendChain(message.Text("ERROR: no such plugin"))
			return
		}
		typ, ok := apitypes[args]
		if !ok {
			ctx.SendChain(message.Text("ERROR: 未知类型 ", args))
			return
		}
		cfg.Type = int(typ)
		err := c.SetExtra(&cfg)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set extra err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天接口地址", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.API))
	en.OnPrefix("设置AI聊天密钥", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.Key))
	en.OnPrefix("设置AI聊天模型名", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetstr(&cfg.ModelName))
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
		Handle(newextrasetbool(&cfg.NoReplyAT))
	en.OnRegex("^设置AI聊天(不)?支持系统提示词$", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetbool(&cfg.NoSystemP))
	en.OnPrefix("设置AI聊天最大长度", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetuint(&cfg.MaxN))
	en.OnPrefix("设置AI聊天TopP", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetfloat32(&cfg.TopP))
	en.OnRegex("^设置AI聊天(不)?以AI语音输出$", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(newextrasetbool(&cfg.NoRecord))
	en.OnFullMatch("查看AI聊天配置", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				ctx.SendChain(message.Text("ERROR: no such plugin"))
				return
			}
			gid := ctx.Event.GroupID
			rate := c.GetData(gid) & 0xff
			temp := (c.GetData(gid) >> 8) & 0xff
			if temp <= 0 {
				temp = 70 // default setting
			}
			if temp > 100 {
				temp = 100
			}
			ctx.SendChain(message.Text(printConfig(rate, temp, cfg)))
		})
	en.OnFullMatch("重置AI聊天", ensureconfig, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		chat.Reset()
		ctx.SendChain(message.Text("成功"))
	})

	// 添加群聊总结功能
	en.OnRegex(`^群聊总结\s?(\d*)$`, ensureconfig, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).Limit(limit.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("少女思考中..."))
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			return
		}
		rate := c.GetData(gid)
		temp := (rate >> 8) & 0xff
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

		// 调用大模型API进行总结
		summary, err := llmchat(summaryPrompt, temp)

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
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			return
		}
		rate := c.GetData(gid)
		temp := (rate >> 8) & 0xff
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

		// 调用大模型API进行聊天
		reply, err := llmchat(query, temp)
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
func llmchat(prompt string, temp int64) (string, error) {
	temperature, topp, maxn := getModelParams(temp) // 使用默认温度70

	x := deepinfra.NewAPI(cfg.API, cfg.Key)
	var mod model.Protocol
	switch cfg.Type {
	case 0:
		mod = model.NewOpenAI(
			cfg.ModelName, cfg.Separator,
			temperature, topp, maxn,
		)
	case 1:
		mod = model.NewOLLaMA(
			cfg.ModelName, cfg.Separator,
			temperature, topp, maxn,
		)
	case 2:
		mod = model.NewGenAI(
			cfg.ModelName,
			temperature, topp, maxn,
		)
	default:
		logrus.Warnln("[aichat] unsupported AI type", cfg.Type)
		return "", errors.New("不支持的AI类型")
	}

	data, err := x.Request(mod.User(prompt))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(data), nil
}
