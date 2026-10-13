package localadmin

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type AIStore struct {
	repository *chat.ConfigRepository
	groups     map[string]*ctrl.Control[*zero.Ctx]
}

var errAIValidation = errors.New("AI 设置无效")

func invalidAI(text string) error { return fmt.Errorf("%w：%s", errAIValidation, text) }
func NewAIStore(controls []*ctrl.Control[*zero.Ctx]) *AIStore {
	s := &AIStore{groups: map[string]*ctrl.Control[*zero.Ctx]{}}
	for _, c := range controls {
		if c.Service == "aichatcfg" {
			s.repository = chat.NewConfigRepository(c)
		}
		if c.Service == "aichatcfg" || c.Service == "llm" {
			s.groups[c.Service] = c
			c.Manager.Lock()
			c.BeforeData = c.Manager.BeforeSwitch
			c.Manager.Unlock()
		}
	}
	return s
}

type AISettings struct {
	Fields          map[string]any           `json:"fields"`
	Keys            map[string]bool          `json:"keys"`
	AddressRedacted map[string]bool          `json:"addressRedacted"`
	Revision        string                   `json:"revision"`
	Ready           bool                     `json:"ready"`
	Checks          map[string]AIConfigCheck `json:"checks"`
}
type AIGroupSettings struct {
	Service              string  `json:"service"`
	GroupID              string  `json:"groupId"`
	PrivateID            string  `json:"privateId,omitempty"`
	Revision             string  `json:"revision"`
	Rate                 int     `json:"rate"`
	Temperature          int     `json:"temperature"`
	EffectiveTemperature float32 `json:"effectiveTemperature"`
	UseAgent             bool    `json:"useAgent"`
	VoiceOutput          bool    `json:"voiceOutput"`
	RespondAt            bool    `json:"respondAt"`
}
type AIView struct {
	Available bool              `json:"available"`
	Global    *AISettings       `json:"global,omitempty"`
	Groups    []AIGroupSettings `json:"groups"`
}

func (s *AIStore) Get(gid int64) (AIView, error) {
	v := AIView{Groups: []AIGroupSettings{}}
	if s == nil || s.repository == nil {
		return v, nil
	}
	c, revision, err := s.repository.Read()
	if err != nil {
		return v, err
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return v, err
	}
	fields := map[string]any{}
	_ = json.Unmarshal(raw, &fields)
	keys := map[string]bool{"Key": c.Key != "", "ImageKey": c.ImageKey != "", "AgentKey": c.AgentKey != ""}
	for name := range keys {
		delete(fields, name)
	}
	redacted := map[string]bool{}
	for _, name := range []string{"API", "ImageAPI", "AgentAPI"} {
		fields[name], redacted[name] = chat.DisplayAPI(fields[name].(string))
	}
	for name, value := range fields {
		if text, ok := value.(string); ok {
			fields[name] = c.Redact(text)
		}
	}
	v.Available = true
	v.Global = &AISettings{Fields: fields, Keys: keys, AddressRedacted: redacted, Revision: revision, Ready: c.Ready()}
	v.Global.Checks = aiConfigChecks(c)
	if gid != 0 {
		for _, service := range []string{"aichatcfg", "llm"} {
			control := s.groups[service]
			if control == nil {
				continue
			}
			data, revision, err := control.DataVersion(gid)
			if err != nil {
				return AIView{}, err
			}
			stor := chat.Storage(data)
			setting := AIGroupSettings{Service: service, Revision: revision, Rate: int(stor.Rate()), Temperature: int(ctxext.Storage(data).Get(chat.BitmapTemp)), EffectiveTemperature: stor.Temp(), UseAgent: !stor.NoAgent(), VoiceOutput: !stor.NoRecord(), RespondAt: !stor.NoReplyAt()}
			if gid < 0 {
				setting.PrivateID = strconv.FormatInt(-gid, 10)
			} else {
				setting.GroupID = strconv.FormatInt(gid, 10)
			}
			v.Groups = append(v.Groups, setting)
		}
	}
	return v, nil
}

type AIKeyChange struct {
	Action  string `json:"action"`
	Value   string `json:"value"`
	Confirm bool   `json:"confirm"`
}
type AIUpdate struct {
	Revision string                     `json:"revision"`
	Fields   map[string]json.RawMessage `json:"fields"`
	Keys     map[string]AIKeyChange     `json:"keys"`
}

func validRevision(text string) bool {
	data, err := hex.DecodeString(text)
	return err == nil && len(data) == 32
}

func (s *AIStore) Save(update AIUpdate) error {
	if s == nil || s.repository == nil {
		return invalidAI("AI 配置插件未编译")
	}
	if !validRevision(update.Revision) {
		return invalidAI("AI 配置版本无效")
	}
	fields := map[string]json.RawMessage{}
	for name, value := range update.Fields {
		if name == "Key" || name == "ImageKey" || name == "AgentKey" {
			return invalidAI("密钥须通过保留、替换或清除操作设置")
		}
		fields[name] = value
	}
	for name, key := range update.Keys {
		if name != "Key" && name != "ImageKey" && name != "AgentKey" {
			return invalidAI("未知密钥设置")
		}
		switch key.Action {
		case "keep":
			if key.Value != "" || key.Confirm {
				return invalidAI("保留密钥时无需填写新值或确认")
			}
		case "replace":
			if key.Value == "" || key.Confirm {
				return invalidAI("替换密钥时须填写新值")
			}
			fields[name], _ = json.Marshal(key.Value)
		case "clear":
			if !key.Confirm || key.Value != "" {
				return invalidAI("请确认清除密钥")
			}
			fields[name] = json.RawMessage(`""`)
		default:
			return invalidAI("密钥操作无效")
		}
	}
	if err := chat.ValidateConfigFields(fields); err != nil {
		return invalidAI(err.Error())
	}
	_, err := s.repository.Update(update.Revision, fields)
	return err
}

type AIGroupUpdate struct {
	Service     string `json:"service"`
	GroupID     string `json:"groupId"`
	PrivateID   string `json:"privateId,omitempty"`
	Revision    string `json:"revision"`
	Rate        *int   `json:"rate,omitempty"`
	Temperature *int   `json:"temperature,omitempty"`
	UseAgent    *bool  `json:"useAgent,omitempty"`
	VoiceOutput *bool  `json:"voiceOutput,omitempty"`
	RespondAt   *bool  `json:"respondAt,omitempty"`
}

func aiConversationID(group, private string) (int64, error) {
	if private == "" {
		return pluginGroup(group)
	}
	if group != "" {
		return 0, invalidAI("群号与私聊 QQ 不能同时填写")
	}
	id, err := pluginGroup(private)
	if err != nil || id == 0 {
		return 0, invalidAI("私聊 QQ 须为正整数")
	}
	return -id, nil
}

func (s *AIStore) SaveGroup(update AIGroupUpdate) error {
	gid, err := aiConversationID(update.GroupID, update.PrivateID)
	if err != nil || gid == 0 || !validRevision(update.Revision) {
		return invalidAI("请选择有效群号或私聊 QQ 及配置版本")
	}
	c := s.groups[update.Service]
	if c == nil {
		return invalidAI("当前程序未编译此 AI 插件")
	}
	if update.Service == "llm" && (update.Rate != nil || update.UseAgent != nil || update.VoiceOutput != nil || update.RespondAt != nil) {
		return invalidAI("GPT/群总结仅支持温度设置")
	}
	for _, n := range []*int{update.Rate, update.Temperature} {
		if n != nil && (*n < 0 || *n > 100) {
			return invalidAI("概率和温度须为 0～100；温度 0 使用默认 70")
		}
	}
	_, err = c.UpdateData(gid, update.Revision, func(old int64) (int64, error) {
		stor := ctxext.Storage(old)
		if update.Rate != nil {
			stor = stor.Set(int64(*update.Rate), chat.BitmapRate)
		}
		if update.Temperature != nil {
			stor = stor.Set(int64(*update.Temperature), chat.BitmapTemp)
		}
		for _, pair := range []struct {
			value *bool
			mask  int64
		}{{update.UseAgent, chat.BitmapNagt}, {update.VoiceOutput, chat.BitmapNrec}, {update.RespondAt, chat.BitmapNrat}} {
			if pair.value != nil {
				x := int64(1)
				if *pair.value {
					x = 0
				}
				stor = stor.Set(x, pair.mask)
			}
		}
		return int64(stor), nil
	})
	return err
}

func (s *Server) getAI(w http.ResponseWriter, r *http.Request) {
	gid, err := aiConversationID(r.URL.Query().Get("group"), r.URL.Query().Get("private"))
	if err != nil {
		apiError(w, 400, err.Error())
		return
	}
	v, err := s.cfg.AI.Get(gid)
	if err != nil {
		apiError(w, 500, "AI 配置读取失败，请检查控制数据库")
		return
	}
	jsonReply(w, 200, v)
}
func (s *Server) saveAI(w http.ResponseWriter, r *http.Request) {
	if s.cfg.AI == nil || s.cfg.AI.repository == nil {
		apiError(w, 409, "当前程序未接入 AI 设置")
		return
	}
	var update AIUpdate
	if !readJSONLimit(w, r, &update, 131072) {
		return
	}
	if s.cfg.Logs != nil {
		for _, key := range update.Keys {
			s.cfg.Logs.AddSecrets(key.Value)
		}
	}
	err := s.cfg.AI.Save(update)
	if err != nil {
		s.aiSaveError(w, err)
		return
	}
	v, err := s.cfg.AI.Get(0)
	if err != nil {
		apiError(w, 500, "配置已保存，但读取新状态失败，请重新加载")
		return
	}
	jsonReply(w, 200, v)
}
func (s *Server) saveAIGroup(w http.ResponseWriter, r *http.Request) {
	if s.cfg.AI == nil || s.cfg.AI.repository == nil {
		apiError(w, 409, "当前程序未接入 AI 设置")
		return
	}
	var update AIGroupUpdate
	if !readJSON(w, r, &update) {
		return
	}
	err := s.cfg.AI.SaveGroup(update)
	if err != nil {
		s.aiSaveError(w, err)
		return
	}
	gid, _ := aiConversationID(update.GroupID, update.PrivateID)
	v, err := s.cfg.AI.Get(gid)
	if err != nil {
		apiError(w, 500, "配置已保存，但读取新状态失败，请重新加载")
		return
	}
	jsonReply(w, 200, v)
}
func (s *Server) aiSaveError(w http.ResponseWriter, err error) {
	if errors.Is(err, ctrl.ErrConfigConflict) {
		apiError(w, 409, "AI 配置已变化，请重新加载后核对再保存")
		return
	}
	if errors.Is(err, errAIValidation) {
		apiError(w, 400, err.Error())
		return
	}
	apiError(w, 500, "AI 配置保存失败，原设置保留；请检查备份目录及控制数据库")
}
