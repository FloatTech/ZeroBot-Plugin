package chat

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/RomiChan/syncx"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	goba "github.com/fumiama/go-onebot-agent"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/vevent"
)

const (
	// StateKeyAgentHooked indicates that whether the agent log is in recording
	StateKeyAgentHooked = zero.StateKeyPrefixKeep + "_chat_ag_hooked__"
	// StateKeyAgentTriggered shows that this message has already triggered an agent run
	StateKeyAgentTriggered = zero.StateKeyPrefixKeep + "_chat_ag_triggered__"
)

// agentchar 将 char.yaml 内容嵌入为默认 agent 性格
//
//go:embed char.yaml
var agentchar []byte

var agentcharcfg charcfg

// AgentCharConfig 可随时更改的 agent 性格配置
var AgentCharConfig *goba.Config

var ags = syncx.Map[int64, *goba.Agent]{}
var agentsMu sync.Mutex
var histories = map[int64]*agentHistory{}
var agentGeneration atomic.Uint64

func AgentGeneration() uint64 { return agentGeneration.Load() }

type charcfg struct {
	Sex  string `yaml:"sex"`
	Char string `yaml:"char"`
}

func init() {
	err := yaml.NewDecoder(bytes.NewReader(agentchar)).Decode(&agentcharcfg)
	if err != nil {
		panic(err)
	}
	AgentCharConfig = &goba.Config{
		Sex:   agentcharcfg.Sex,
		Chars: agentcharcfg.Char,
	}
	go func() {
		process.GlobalInitMutex.Lock()
		AgentCharConfig.Nickname = strings.Join(zero.BotConfig.NickName, "/")
		logrus.Infoln("[chat] 设置 agent 昵称:", AgentCharConfig.Nickname)
		process.GlobalInitMutex.Unlock()
	}()
}

// AgentOf id is self_id
func AgentOf(id int64, service string) *goba.Agent {
	agentsMu.Lock()
	defer agentsMu.Unlock()
	if ag, ok := ags.Load(id); ok {
		return ag
	}
	mem, err := atomicgetmemstorage(service)
	if err != nil {
		logrus.Warnln("[chat] Agent 记忆存储不可用。")
		mem = &memstorage{}
	}
	process.GlobalInitMutex.Lock()
	character := *AgentCharConfig
	process.GlobalInitMutex.Unlock()
	if cfg, err := CurrentConfig(); err == nil {
		character.Chars, character.Sex = cfg.AgentChar, cfg.AgentSex
	}
	ag := goba.NewAgent(&character, 16, 8, time.Hour*24, "", mem, true, false)
	history := histories[id]
	budget := agentContextBytes
	if cfg, err := CurrentConfig(); err == nil {
		if api, _ := DisplayAPI(cfg.AgentAPI); !strings.HasPrefix(api, "https://api.deepseek.com/") {
			budget = 65536
		}
	}
	if history == nil {
		history, err = openAgentHistory(filepath.Join(file.BOTPATH, "data", service, "history"), id, budget, CurrentConfig)
		if err == nil {
			histories[id] = history
		}
	}
	if history != nil {
		history.mu.Lock()
		history.budget = budget
		history.mu.Unlock()
		ag.SetHistoryStore(history)
	} else {
		logrus.Warnln("[chat] Agent 历史记录不可用，暂停 Agent 操作。")
		ag.SetHistoryStore(unavailableHistory{})
	}
	ags.Store(id, &ag)
	return &ag
}

// ReloadAgents applies new persona/provider settings without erasing history.
func ReloadAgents() {
	agentsMu.Lock()
	defer agentsMu.Unlock()
	agentGeneration.Add(1)
	ks := make([]int64, 0, 8)
	ags.Range(func(key int64, _ *goba.Agent) bool {
		ks = append(ks, key)
		return true
	})
	for _, k := range ks {
		ags.Delete(k)
	}
}

// ResetAgents starts fresh contexts while retaining all archived records.
func ResetAgents() error {
	agentsMu.Lock()
	for _, history := range histories {
		if err := history.reset(); err != nil {
			agentsMu.Unlock()
			return err
		}
	}
	agentsMu.Unlock()
	ReloadAgents()
	return nil
}

var checkgids = map[string]struct{}{
	"send_group_msg":          {},
	"set_group_kick":          {},
	"set_group_ban":           {},
	"set_group_whole_ban":     {},
	"set_group_card":          {},
	"set_group_name":          {},
	"set_group_special_title": {},
	"get_group_member_info":   {},
	"get_group_member_list":   {},
	"get_group_info":          {},
}

// CallAgent and check group API permission
func CallAgent(ag *goba.Agent, issudo bool, iter int, api deepinfra.API, p model.Protocol, grp int64, role goba.PermRole) []zero.APIRequest {
	reqs, _, _ := NextAgentActions(ag, issudo, iter, api, p, grp, role)
	return reqs
}

func NextAgentActions(ag *goba.Agent, issudo bool, iter int, api deepinfra.API, p model.Protocol, grp int64, role goba.PermRole) ([]zero.APIRequest, bool, error) {
	reqs, err := ag.GetAction(api, p, grp, role, iter, false)
	ended := errors.Is(err, io.EOF)
	if err != nil && !ended {
		logrus.Warnln("[chat] Agent 请求或记录保存失败，请检查模型配置、网络与本机存储。")
		return nil, false, err
	}
	logrus.Infoln("[chat] Agent 动作数:", len(reqs), "结束:", ended)
	checkedreqs := make([]zero.APIRequest, 0, len(reqs))
	for _, req := range reqs {
		if _, ok := checkgids[req.Action]; ok {
			v, ok := req.Params["group_id"].(json.Number)
			if !ok {
				logrus.Warnln("[chat] invalid group_id type", reflect.TypeOf(req.Params["group_id"]))
				continue
			}
			gid, err := v.Int64()
			if err != nil || gid <= 0 {
				logrus.Warnln("[chat] agent conv req gid err:", err)
				continue
			}
			if grp != gid && !issudo {
				logrus.Warnln("[chat] refuse to send out of grp from", grp, "to", gid)
				continue
			}
			if role == goba.PermRoleUser { // check @all
				msgb, err := json.Marshal(req.Params["message"])
				if err != nil {
					logrus.Warnln("[chat] re-marshal msg err:", err)
					continue
				}
				msgs := message.ParseMessage(msgb)
				isvalid := true
				for _, m := range msgs {
					if m.Type == "at" {
						qqs, ok := m.Data["qq"]
						if !ok {
							logrus.Warnln("[chat] invalid at message without qq")
							isvalid = false
							break
						}
						qq, err := strconv.ParseInt(qqs, 10, 64)
						if err != nil {
							logrus.Warnln("[chat] invalid at qq", qqs)
							isvalid = false
							break
						}
						if qq <= 0 {
							logrus.Warnln("[chat] invalid at qq num", qq)
							isvalid = false
							break
						}
					}
				}
				if !isvalid {
					continue
				}
			}
		}
		checkedreqs = append(checkedreqs, req)
	}
	return checkedreqs, ended, nil
}

func togobaev(ev *zero.Event) *goba.Event {
	msgid := int64(0)
	switch id := ev.MessageID.(type) {
	case int64:
		msgid = id
	case int:
		msgid = int64(id)
	case string:
		var err error
		msgid, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil
		}
	case json.Number:
		var err error
		msgid, err = id.Int64()
		if err != nil {
			return nil
		}
	default:
		if ev.PostType == "message" {
			return nil
		}
	}
	msgd := ev.NativeMessage
	return &goba.Event{
		Time:        ev.Time,
		PostType:    ev.PostType,
		MessageType: ev.MessageType,
		SubType:     ev.SubType,
		MessageID:   msgid,
		GroupID:     ev.GroupID,
		UserID:      ev.UserID,
		TargetID:    ev.TargetID,
		SelfID:      ev.SelfID,
		NoticeType:  ev.NoticeType,
		OperatorID:  ev.OperatorID,
		File:        ev.File,
		RequestType: ev.RequestType,
		Flag:        ev.Flag,
		Comment:     ev.Comment,
		Sender:      ev.Sender,
		Message:     msgd,
	}
}

func truncatecopy(params map[string]any) map[string]any {
	cpp := make(map[string]any, len(params))
	for k, v := range params {
		switch val := v.(type) {
		case string:
			if len(val) > 512 {
				cpp[k] = val[:200] + " ... " + val[len(val)-200:]
			} else {
				cpp[k] = val
			}
		case int64, int:
			cpp[k] = val
		case message.Message:
			valCopy := make(message.Message, len(val))
			copy(valCopy, val)
			for i, m := range val {
				valCopy[i] = message.Segment{
					Type: m.Type,
					Data: make(map[string]string, len(m.Data)),
				}
				for k, v := range m.Data {
					if len(v) > 512 {
						valCopy[i].Data[k] = v[:200] + " ... " + v[len(v)-200:]
					} else {
						valCopy[i].Data[k] = v
					}
				}
			}
			cpp[k] = valCopy
		case message.Segment:
			valCopy := message.Segment{
				Type: val.Type,
				Data: make(map[string]string, len(val.Data)),
			}
			for k, v := range val.Data {
				if len(v) > 512 {
					valCopy.Data[k] = v[:200] + " ... " + v[len(v)-200:]
				} else {
					valCopy.Data[k] = v
				}
			}
			cpp[k] = valCopy
		default:
			logrus.Warnln("[chat] agent unknown params typ:", reflect.TypeOf(v))
		}
	}
	return cpp
}

func logev(ctx *zero.Ctx) {
	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}
	ev := togobaev(ctx.Event)
	if ev == nil {
		return
	}
	data, _ := json.Marshal(ev)
	logrus.Debugln("[chat] agent", gid, "add ev:", binary.BytesToString(data))
	AgentOf(ctx.Event.SelfID, "aichat").AddEvent(gid, ev)

	mp := ctx.State[control.StateKeySyncxState].(*syncx.Map[string, any])

	_, hasvalue := mp.LoadOrStore(StateKeyAgentHooked, struct{}{})
	if hasvalue {
		return
	}

	vevent.HookCtxCaller(ctx, vevent.NewAPICallerReturnHook(
		ctx, func(req zero.APIRequest, rsp zero.APIResponse, _ error) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			if _, ok := mp.Load(StateKeyAgentTriggered); ok {
				logrus.Debugln("[chat] agent", gid, "skip agent triggered requ:", &req)
				return
			}
			if req.Action != "send_private_msg" &&
				req.Action != "send_group_msg" &&
				req.Action != "delete_msg" {
				logrus.Debugln("[chat] agent", gid, "skip non-msg other triggered action:", req.Action)
				return
			}
			ag := AgentOf(ctx.Event.SelfID, "aichat")
			logrus.Debugln("[chat] agent others", gid, "add requ:", &req)
			ag.AddRequest(gid, &req)
			logrus.Debugln("[chat] agent others", gid, "add resp:", &rsp)
			ag.AddResponse(gid, &goba.APIResponse{
				Status:  rsp.Status,
				Data:    json.RawMessage(rsp.Data.Raw),
				Message: rsp.Message,
				Wording: rsp.Wording,
				RetCode: rsp.RetCode,
			})
		}),
	)
}

func init() {
	zero.OnNotice().FirstPriority().SetBlock(false).Handle(logev)
	zero.OnMessage().FirstPriority().SetBlock(false).Handle(logev)
	zero.OnRequest().FirstPriority().SetBlock(false).Handle(logev)
}
