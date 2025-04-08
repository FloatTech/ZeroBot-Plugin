// Package aichat OpenAI聊天
package aichat

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/sirupsen/logrus"

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
			"- 设置AI聊天(不)响应AT",
		PrivateDataFolder: "aichat",
	})
)

var apitypes = map[string]uint8{
	"OpenAI": 0,
	"OLLaMA": 1,
	"GenAI":  2,
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

		if temp <= 0 {
			temp = 70 // default setting
		}
		if temp > 100 {
			temp = 100
		}

		x := deepinfra.NewAPI(cfg.API, cfg.Key)
		var mod model.Protocol

		switch cfg.Type {
		case 0:
			mod = model.NewOpenAI(
				cfg.ModelName, cfg.Separator,
				float32(temp)/100, 0.9, 4096,
			)
		case 1:
			mod = model.NewOLLaMA(
				cfg.ModelName, cfg.Separator,
				float32(temp)/100, 0.9, 4096,
			)
		case 2:
			mod = model.NewGenAI(
				cfg.ModelName,
				float32(temp)/100, 0.9, 4096,
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
}
