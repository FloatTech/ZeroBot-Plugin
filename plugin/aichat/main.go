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
	"github.com/FloatTech/zbputils/control"
)

var (
	api *deepinfra.API
	en  = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Extra:             control.ExtraFromString("aichat"),
		Brief:             "OpenAI聊天",
		Help:              "- 设置AI聊天触发概率10\n- 设置AI聊天密钥xxx\n- 设置AI聊天模型名xxx\n- 设置AI聊天系统提示词xxx",
		PrivateDataFolder: "aichat",
	})
	lst = newlist()
)

var (
	modelname    = "deepseek-ai/DeepSeek-R1"
	systemprompt = "你正在QQ群与用户聊天，用户发送了消息。按自己的心情简短思考后，条理清晰地回应**一句话**，禁止回应多句。"
)

func init() {
	mf := en.DataFolder() + "model.txt"
	sf := en.DataFolder() + "system.txt"
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

	en.OnMessage(func(ctx *zero.Ctx) bool {
		txt := ctx.ExtractPlainText()
		ctx.State["aichat_txt"] = txt
		return txt != ""
	}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		lst.add(ctx.Event.GroupID, ctx.State["aichat_txt"].(string))
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		if !ok {
			return
		}
		rate := c.GetData(gid)
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
		data, err := y.Request(lst.body(modelname, systemprompt, gid))
		if err != nil {
			logrus.Warnln("[niniqun] post err:", err)
			return
		}
		txt := strings.Trim(data, "\n 　")
		if len(txt) > 0 {
			lst.add(ctx.Event.GroupID, txt)
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
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		c.SetData(gid, int64(r&0xff))
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
	})
}
