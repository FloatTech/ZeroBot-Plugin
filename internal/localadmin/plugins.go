package localadmin

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type PluginState = ctrl.SwitchState

type PluginStore struct {
	controls        map[string]*ctrl.Control[*zero.Ctx]
	responseManager *ctrl.Manager[*zero.Ctx]
}

func NewPluginStore(controls []*ctrl.Control[*zero.Ctx], backups string) *PluginStore {
	s := &PluginStore{controls: map[string]*ctrl.Control[*zero.Ctx]{}}
	managers := map[*ctrl.Manager[*zero.Ctx]]bool{}
	for _, c := range controls {
		s.controls[c.Service] = c
		if !managers[c.Manager] {
			manager := c.Manager
			manager.Lock()
			manager.BeforeSwitch = func() error {
				_, err := BackupControlDatabase(&manager.D, backups)
				return err
			}
			manager.BeforeExtra = manager.BeforeSwitch
			manager.BeforePolicy = manager.BeforeSwitch
			manager.Unlock()
			managers[manager] = true
		}
	}
	if len(managers) == 1 {
		for manager := range managers {
			s.responseManager = manager
		}
	}
	return s
}

func pluginGroup(text string) (int64, error) {
	if text == "" || text == "0" {
		return 0, nil
	}
	id, err := strconv.ParseInt(text, 10, 64)
	if err != nil || id <= 0 || strconv.FormatInt(id, 10) != text {
		return 0, errors.New("群号须为正整数；全局设置使用 0")
	}
	return id, nil
}

func (s *Server) getPlugins(w http.ResponseWriter, r *http.Request) {
	gid, err := pluginGroup(r.URL.Query().Get("group"))
	if err != nil {
		apiError(w, 400, err.Error())
		return
	}
	plugins := append([]Plugin{}, s.cfg.Plugins...)
	var response *ctrl.ResponseState
	if store := s.cfg.PluginControl; store != nil {
		for i := range plugins {
			c := store.controls[plugins[i].Name]
			if c == nil {
				apiError(w, 500, "插件注册状态不一致")
				return
			}
			state, err := c.SwitchState(gid)
			if err != nil {
				apiError(w, 500, "插件状态读取失败，请检查本机控制数据库")
				return
			}
			plugins[i].State = &state
		}
		if store.responseManager != nil {
			state, err := store.responseManager.ReadResponseState(gid)
			if err != nil {
				apiError(w, 500, "响应状态读取失败，请检查本机控制数据库")
				return
			}
			response = &state
		}
	}
	jsonReply(w, 200, map[string]any{"plugins": plugins, "groupId": strconv.FormatInt(gid, 10), "controlAvailable": s.cfg.PluginControl != nil, "response": response})
}

func (s *Server) setPluginSwitch(w http.ResponseWriter, r *http.Request) {
	var update struct {
		Name     string `json:"name"`
		GroupID  string `json:"groupId"`
		Mode     string `json:"mode"`
		Revision string `json:"revision"`
	}
	if !readJSON(w, r, &update) {
		return
	}
	if s.cfg.PluginControl == nil {
		apiError(w, 409, "当前程序未接入插件控制")
		return
	}
	gid, err := pluginGroup(update.GroupID)
	decoded, revisionErr := hex.DecodeString(update.Revision)
	if err != nil || update.GroupID == "" || revisionErr != nil || len(decoded) != 32 || (update.Mode != "inherit" && update.Mode != "enabled" && update.Mode != "disabled") {
		apiError(w, 400, "请选择有效的群号、开关状态及配置版本")
		return
	}
	c := s.cfg.PluginControl.controls[update.Name]
	if c == nil {
		apiError(w, 404, "当前程序没有注册这个插件")
		return
	}
	state, err := c.SetSwitch(gid, update.Mode, update.Revision)
	if errors.Is(err, ctrl.ErrSwitchConflict) {
		apiError(w, 409, "插件设置已变化，请刷新状态后再保存")
		return
	}
	if err != nil {
		apiError(w, 500, "插件设置保存失败，原设置保留；请检查备份目录及控制数据库")
		return
	}
	jsonReply(w, 200, map[string]any{"name": c.Service, "groupId": strconv.FormatInt(gid, 10), "state": state})
}
