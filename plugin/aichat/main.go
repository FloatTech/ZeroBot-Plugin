// Package aichat 大模型聊天和Agent
package aichat

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/RomiChan/syncx"
	"github.com/fumiama/deepinfra"
	goba "github.com/fumiama/go-onebot-agent"
	"github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/airecord"
	"github.com/FloatTech/ZeroBot-Plugin/internal/modelstats"
	"github.com/FloatTech/ZeroBot-Plugin/internal/qqvoice"
	"github.com/FloatTech/ZeroBot-Plugin/internal/replystats"
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
	fastfailnorecord atomic.Bool
)

func init() {
	en.OnMessage(chat.EnsureConfig, func(ctx *zero.Ctx) bool {
		if !chat.ConfigOf(ctx).Ready() {
			return false
		}
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
		cfg := chat.ConfigOf(ctx)
		kind := replystats.Chat
		if !ctx.State[zero.StateKeyPrefixKeep+"aichatcfg_stor__"].(chat.Storage).NoAgent() && cfg.AgentAPI != "" && cfg.AgentModelName != "" && (cfg.AgentKey != "" || cfg.AgentType == 1) {
			kind = replystats.Agent
		}
		task := replystats.Default.Begin(ctx, kind)
		defer task.End()
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor := ctx.State[zero.StateKeyPrefixKeep+"aichatcfg_stor__"].(chat.Storage)
		temperature := stor.Temp()
		topp, maxn := cfg.MParams()
		mp := ctx.State[control.StateKeySyncxState].(*syncx.Map[string, any])

		logrus.Debugln("[aichat] agent mode test: noagent", stor.NoAgent(), "hasapi", cfg.AgentAPI != "", "hasmodel", cfg.AgentModelName != "")
		if !stor.NoAgent() && cfg.AgentAPI != "" && cfg.AgentModelName != "" && (cfg.AgentKey != "" || cfg.AgentType == 1) {
			logrus.Debugln("[aichat] enter agent mode")
			x := deepinfra.NewAPI(cfg.AgentAPI, string(cfg.AgentKey))
			x.SetHTTPClient(modelstats.Default.Client(&http.Client{Timeout: 30 * time.Second}, modelstats.Agent))
			mod, err := cfg.AgentType.ProtocolForAPI(cfg.AgentAPI, cfg.AgentModelName, temperature, topp, maxn, cfg.ReasoningEffort, cfg.Separator)
			if err != nil {
				task.Fail()
				logrus.Warnln("ERROR: ", err)
				return
			}
			role := agentRole(ctx)
			c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			if !ok {
				logrus.Warnln("ERROR: cannot get ctrl mamager")
			}
			ag := chat.AgentOf(ctx.Event.SelfID, c.Service)
			logrus.Debugln("[aichat] got agent")
			if cfg.ImageAPI != "" && !ag.CanViewImage() {
				mod, err := cfg.ImageType.ProtocolForAPI(cfg.ImageAPI, cfg.ImageModelName, temperature, topp, maxn, cfg.ReasoningEffort, cfg.Separator)
				if err != nil {
					task.Fail()
					logrus.Warnln("ERROR: ", err)
					return
				}
				imageAPI := deepinfra.NewAPI(cfg.ImageAPI, string(cfg.ImageKey))
				imageAPI.SetHTTPClient(modelstats.Default.Client(&http.Client{Timeout: 30 * time.Second}, modelstats.Vision))
				ag.SetViewImageAPI(imageAPI, mod)
				logrus.Debugln("[aichat] agent set img")
			}
			ctx.NoTimeout()
			logrus.Debugln("[aichat] agent set no timeout")
			hasresp := false
			generation := chat.AgentGeneration()
			err = chat.RunAgent(func(iter int) ([]zero.APIRequest, bool, error) {
				return chat.NextAgentActions(ag, zero.SuperUserPermission(ctx), iter, x, mod.Clone(), gid, role)
			}, func(req zero.APIRequest) error {
				mp.Store(chat.StateKeyAgentTriggered, struct{}{})
				if req.Action == goba.SVM { // is a fake action
					return nil
				}
				logrus.Debugln("[chat] agent triggered", gid, "add requ:", &req)
				if err := ag.AddRequest(gid, &req); err != nil {
					return err
				}
				hasresp = true
				rsp := task.Call(req.Action, req.Params, ctx.CallAction)
				if rsp.Status != "ok" || rsp.RetCode != 0 {
					task.Fail()
				}
				logrus.Debugln("[chat] agent triggered", gid, "add resp:", &rsp)
				return ag.AddResponse(gid, &goba.APIResponse{
					Status:  rsp.Status,
					Data:    json.RawMessage(rsp.Data.Raw),
					Message: rsp.Message,
					Wording: rsp.Wording,
					RetCode: rsp.RetCode,
				})
			}, func() bool { return generation == chat.AgentGeneration() })
			if err != nil {
				task.Fail()
				logrus.Warnln("[aichat] Agent 本次任务已停止，检查结束动作、记录存储或重复动作。")
				if errors.Is(err, chat.ErrAgentReloaded) {
					return
				}
			}
			if hasresp {
				return
			}
			// no response, fall back to normal chat
			logrus.Debugln("[aichat] agent fell back to normal chat")
		}

		x := deepinfra.NewAPI(cfg.API, string(cfg.Key))
		x.SetHTTPClient(modelstats.Default.Client(&http.Client{Timeout: 30 * time.Second}, modelstats.Chat))
		mod, err := cfg.Type.ProtocolForAPI(cfg.API, cfg.ModelName, temperature, topp, maxn, cfg.ReasoningEffort, cfg.Separator)
		if err != nil {
			task.Fail()
			logrus.Warnln("ERROR: ", err)
			return
		}
		data, err := x.Request(chat.GetChatContext(mod, gid, cfg.SystemP, bool(cfg.NoSystemP)))
		if err != nil {
			task.Fail()
			logrus.Warnln("[aichat] 模型请求失败，请检查接口、密钥及网络。")
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
				if ctx.Event.GroupID > 0 && !fastfailnorecord.Load() && !stor.NoRecord() {
					if qqvoice.Deliver(t, recCfg.ModelID, ctx.Event.GroupID, func(action string, params zero.Params) zero.APIResponse {
						return task.Call(action, params, ctx.CallAction)
					}) == "" {
						continue
					}
					fastfailnorecord.Store(true)
				}
				if id != nil {
					id = task.SendChain(ctx, message.Reply(id), message.Text(t))
				} else {
					id = task.SendChain(ctx, message.Text(t))
				}
				process.SleepAbout1sTo2s()
			}
		}
	})
}
