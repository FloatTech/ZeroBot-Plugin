// Package aichat 大模型聊天和Agent
package aichat

import (
	"encoding/json"
	"math/rand"
	"strings"

	"github.com/RomiChan/syncx"
	"github.com/fumiama/deepinfra"
	goba "github.com/fumiama/go-onebot-agent"
	"github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/airecord"
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
		Brief:            "大模型聊天和Agent",
		Help:             "- (随意聊天, 概率匹配)",

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
	fastfailnorecord = false
)

func init() {
	en.OnMessage(chat.EnsureConfig, func(ctx *zero.Ctx) bool {
		stor, ok := ctx.State[zero.StateKeyPrefixKeep+"aichatcfg_stor__"].(chat.Storage)
		if !ok {
			logrus.Warnln("ERROR: cannot get stor")
			return false
		}
		mp := ctx.State[control.StateKeySyncxState].(*syncx.Map[string, any])
		if _, ok := mp.Load(chat.StateKeyAgentHooked); !ok && !stor.NoAgent() {
			logrus.Infoln("[aichat] skip agent for ctx has not been hooked by agent")
			return false
		}
		if !(ctx.ExtractPlainText() != "" &&
			(!stor.NoReplyAt() || (stor.NoReplyAt() && !ctx.Event.IsToMe))) {
			return false
		}
		rate := stor.Rate()
		if !ctx.Event.IsToMe && rand.Intn(100) >= int(rate) {
			return false
		}
		if ctx.Event.IsToMe {
			ctx.Block()
		}
		return true
	}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor := ctx.State[zero.StateKeyPrefixKeep+"aichatcfg_stor__"].(chat.Storage)
		temperature := stor.Temp()
		topp, maxn := chat.AC.MParams()
		mp := ctx.State[control.StateKeySyncxState].(*syncx.Map[string, any])

		logrus.Debugln("[aichat] agent mode test: noagent", stor.NoAgent(), "hasapi", chat.AC.AgentAPI != "", "hasmodel", chat.AC.AgentModelName != "")
		if !stor.NoAgent() && chat.AC.AgentAPI != "" && chat.AC.AgentModelName != "" && chat.AC.Key != "" {
			logrus.Debugln("[aichat] enter agent mode")
			x := deepinfra.NewAPI(chat.AC.AgentAPI, string(chat.AC.AgentKey))
			mod, err := chat.AC.Type.Protocol(chat.AC.AgentModelName, temperature, topp, maxn)
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
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				logrus.Warnln("ERROR: cannot get ctrl mamager")
			}
			ag := chat.AgentOf(ctx.Event.SelfID, c.Service)
			logrus.Debugln("[aichat] got agent")
			if chat.AC.ImageAPI != "" && !ag.CanViewImage() {
				mod, err := chat.AC.ImageType.Protocol(chat.AC.ImageModelName, temperature, topp, maxn)
				if err != nil {
					logrus.Warnln("ERROR: ", err)
					return
				}
				ag.SetViewImageAPI(deepinfra.NewAPI(chat.AC.ImageAPI, string(chat.AC.ImageKey)), mod)
				logrus.Debugln("[aichat] agent set img")
			}
			ctx.NoTimeout()
			logrus.Debugln("[aichat] agent set no timeout")
			hasresp := false
			// ispuremsg := false
			// hassavemem := false
			for i := 0; i < 8; i++ { // 最大运行 8 轮因为问答上下文只有 16
				reqs := chat.CallAgent(ag, zero.SuperUserPermission(ctx), i+1, x, mod, gid, role)
				if len(reqs) == 0 {
					logrus.Debugln("[aichat] agent call got empty response")
					break
				}
				hasresp = true
				mp.Store(chat.StateKeyAgentTriggered, struct{}{})
				for _, req := range reqs {
					if req.Action == goba.SVM { // is a fake action
						/*if hassavemem {
							ag.AddTerminus(gid)
							logrus.Warnln("[aichat] agent call save mem multi times, force inserting EOA")
							return
						}
						hassavemem = true*/
						continue
					}
					/*if req.Action == "send_private_msg" || req.Action == "send_group_msg" {
						if ispuremsg {
							ag.AddTerminus(gid)
							logrus.Warnln("[aichat] agent call send msg multi times, force inserting EOA")
							return
						}
						ispuremsg = true
					}*/
					logrus.Debugln("[chat] agent triggered", gid, "add requ:", &req)
					ag.AddRequest(gid, &req)
					rsp := ctx.CallAction(req.Action, req.Params)
					logrus.Debugln("[chat] agent triggered", gid, "add resp:", &rsp)
					ag.AddResponse(gid, &goba.APIResponse{
						Status:  rsp.Status,
						Data:    json.RawMessage(rsp.Data.Raw),
						Message: rsp.Message,
						Wording: rsp.Wording,
						RetCode: rsp.RetCode,
					})
				}
			}
			if hasresp {
				return
			}
			// no response, fall back to normal chat
			logrus.Debugln("[aichat] agent fell back to normal chat")
		}

		x := deepinfra.NewAPI(chat.AC.API, string(chat.AC.Key))
		mod, err := chat.AC.Type.Protocol(chat.AC.ModelName, temperature, topp, maxn)
		if err != nil {
			logrus.Warnln("ERROR: ", err)
			return
		}
		data, err := x.Request(chat.GetChatContext(mod, gid, chat.AC.SystemP, bool(chat.AC.NoSystemP)))
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
				logrus.Debugln("[aichat] 回复内容:", t)
				recCfg := airecord.GetConfig()
				record := ""
				if !fastfailnorecord && !stor.NoRecord() {
					record = ctx.GetAIRecord(recCfg.ModelID, recCfg.Customgid, t)
					if record != "" {
						ctx.SendChain(message.Record(record))
						continue
					}
					fastfailnorecord = true
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
}
