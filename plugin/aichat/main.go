// Package aichat OpenAI聊天
package aichat

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/fumiama/deepinfra"
	"github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
)

var (
	api *deepinfra.API
	en  = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("aichat"),
		Brief:            "OpenAI聊天",
		Help: "- 设置AI聊天触发概率10\n" +
			"- 设置AI聊天温度80\n" +
			"- 设置AI聊天密钥xxx\n" +
			"- 设置AI聊天模型名xxx\n" +
			"- 设置AI聊天系统提示词xxx\n" +
			"- 设置AI聊天分隔符</think>(留空则清除)",
		PrivateDataFolder: "aichat",
	})
)

var (
	modelname    = "deepseek-ai/DeepSeek-R1"
	systemprompt = "你正在QQ群与用户聊天，用户发送了消息。按自己的心情简短思考后条理清晰地回复。"
	sepstr       = ""
)

func init() {
	mf := en.DataFolder() + "model.txt"
	sf := en.DataFolder() + "system.txt"
	pf := en.DataFolder() + "sep.txt"
	if file.IsExist(mf) {
		data, err := os.ReadFile(mf)
		if err != nil {
			logrus.Warnln("read model", err)
		} else {
			modelname = string(data)
		}
	}
	if file.IsExist(sf) {
		data, err := os.ReadFile(sf)
		if err != nil {
			logrus.Warnln("read system", err)
		} else {
			systemprompt = string(data)
		}
	}
	if file.IsExist(pf) {
		data, err := os.ReadFile(pf)
		if err != nil {
			logrus.Warnln("read sep", err)
		} else {
			sepstr = string(data)
		}
	}

	en.OnMessage(func(ctx *zero.Ctx) bool {
		return ctx.ExtractPlainText() != ""
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
		key := ""
		err := c.GetExtra(&key)
		if err != nil {
			logrus.Warnln("ERROR: get extra err:", err)
			return
		}
		if key == "" {
			logrus.Warnln("ERROR: get extra err: empty key")
			return
		}
		var x deepinfra.API
		y := &x
		if api == nil {
			x = deepinfra.NewAPI(deepinfra.APIDeepInfra, key)
			atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&api)), unsafe.Pointer(&x))
		} else {
			y = api
		}
		if temp <= 0 {
			temp = 70 // default setting
		}
		if temp > 100 {
			temp = 100
		}
		data, err := y.Request(chat.Ask(ctx, float32(temp)/100, modelname, systemprompt, sepstr))
		if err != nil {
			logrus.Warnln("[niniqun] post err:", err)
			return
		}
		txt := strings.Trim(data, "\n 　")
		if len(txt) > 0 {
			chat.Reply(ctx, txt)
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
	en.OnPrefix("设置AI聊天密钥", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
		err := c.SetExtra(&args)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天模型名", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		modelname = args
		err := os.WriteFile(mf, []byte(args), 0644)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天系统提示词", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		systemprompt = args
		err := os.WriteFile(sf, []byte(args), 0644)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
	en.OnPrefix("设置AI聊天分隔符", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			sepstr = ""
			_ = os.Remove(pf)
			ctx.SendChain(message.Text("清除成功"))
			return
		}
		sepstr = args
		err := os.WriteFile(pf, []byte(args), 0644)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("设置成功"))
	})
}
