// Package aichat OpenAI聊天
package aichat

import (
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

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
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
			"- 设置AI聊天接口地址https://xxx\n" +
			"- 设置AI聊天密钥xxx\n" +
			"- 设置AI聊天模型名xxx\n" +
			"- 查看AI聊天系统提示词\n" +
			"- 重置AI聊天系统提示词\n" +
			"- 设置AI聊天系统提示词xxx\n" +
			"- 设置AI聊天分隔符</think>(留空则清除)\n" +
			"- 设置AI聊天(不)响应AT\n" +
			"- 设置AI聊天最大长度4096\n" +
			"- 设置AI聊天TopP 0.9\n" +
			"- 查看AI聊天配置\n" +
			"- [启用|禁用]AI语音\n" +
			"- 设置AI语音群号1048452984	(tips：群里必须有AI声聊应用)\n" +
			"- 设置AI语音模型\n" +
			"- 发送AI语音xxx",
		PrivateDataFolder: "aichat",
	})
)

var (
	apitypes = map[string]uint8{
		"OpenAI": 0,
		"OLLaMA": 1,
		"GenAI":  2,
	}
	customgid = int64(1048452984)
	modelName = "lucy-voice-xueling"
)

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

		if temp <= 0 {
			temp = 70 // default setting
		}
		if temp > 100 {
			temp = 100
		}

		x := deepinfra.NewAPI(cfg.API, cfg.Key)
		var mod model.Protocol
		maxn := cfg.MaxN
		if maxn == 0 {
			maxn = 4096
		}
		topp := cfg.TopP
		if topp == 0 {
			topp = 0.9
		}

		switch cfg.Type {
		case 0:
			mod = model.NewOpenAI(
				cfg.ModelName, cfg.Separator,
				float32(temp)/100, topp, maxn,
			)
		case 1:
			mod = model.NewOLLaMA(
				cfg.ModelName, cfg.Separator,
				float32(temp)/100, topp, maxn,
			)
		case 2:
			mod = model.NewGenAI(
				cfg.ModelName,
				float32(temp)/100, topp, maxn,
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
				if id != nil {
					id = ctx.SendChain(message.Reply(id), message.Text(t))
				} else {
					id = ctx.SendChain(message.Text(t))
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
	en.OnFullMatch("查看AI聊天配置", ensureconfig, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(printConfig(cfg)))
		})
	en.OnPrefix("设置AI语音群号", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			u := strings.TrimSpace(ctx.State["args"].(string))
			num, err := strconv.ParseInt(u, 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: parse gid err: ", err))
				return
			}
			ctx.SendChain(message.Text("设置AI语音群号为", num))
			customgid = num
		})
	en.OnFullMatch("设置AI语音模型", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			jsonData := ctx.GetAICharacters(customgid, 1)

			// 转换为字符串数组
			var names []string
			// 初始化两个映射表
			nameToID := make(map[string]string)
			nameToURL := make(map[string]string)
			characters := jsonData.Get("#.characters")

			// 遍历每个角色对象
			characters.ForEach(func(_, group gjson.Result) bool {
				group.ForEach(func(_, character gjson.Result) bool {
					// 提取当前角色的三个字段
					name := character.Get("character_name").String()
					names = append(names, name)
					// 存入映射表（重复名称会覆盖，保留最后出现的条目）
					nameToID[name] = character.Get("character_id").String()
					nameToURL[name] = character.Get("preview_url").String()
					return true // 继续遍历
				})
				return true // 继续遍历
			})
			var builder strings.Builder
			// 写入开头文本
			builder.WriteString("请选择语音模型序号：\n")

			// 遍历names数组，拼接序号和名称
			for i, v := range names {
				// 将数字转换为字符串（不依赖fmt）
				numStr := strconv.Itoa(i)
				// 拼接格式："序号. 名称\n"
				builder.WriteString(numStr)
				builder.WriteString(". ")
				builder.WriteString(v)
				builder.WriteString("\n")
			}
			// 获取最终字符串
			ctx.SendChain(message.Text(builder.String()))
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.SendChain(message.Text("设置AI语音模型指令过期"))
					return
				case c := <-recv:
					msg := c.Event.Message.ExtractPlainText()
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字!"))
						continue
					}
					if num < 0 || num >= len(names) {
						ctx.SendChain(message.Text("序号非法!"))
						continue
					}
					modelName = nameToID[names[num]]
					ctx.SendChain(message.Text("已选择语音模型: ", names[num]))
					ctx.SendChain(message.Record(nameToURL[names[num]]))
					return
				}
			}
		})
}
