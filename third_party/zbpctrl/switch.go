package control

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	db "github.com/FloatTech/sqlite"
)

var ErrSwitchConflict = errors.New("plugin settings changed")

// Explicit inheritance preserves the plugin's 62-bit data and default-flip flag.
type switchInheritance struct {
	ID string `db:"id"`
}

type SwitchState struct {
	DefaultEnabled   bool   `json:"defaultEnabled"`
	GlobalMode       string `json:"globalMode"`
	GroupMode        string `json:"groupMode"`
	EffectiveEnabled bool   `json:"effectiveEnabled"`
	Source           string `json:"source"`
	Revision         string `json:"revision"`
}

func (m *Control[CTX]) inheritanceKey(gid int64) string {
	return strconv.FormatInt(gid, 10) + ":" + m.Service
}

func (m *Control[CTX]) loadSwitchLocked(gid int64) (GroupConfig, string, error) {
	c := GroupConfig{GroupID: gid}
	err := m.Manager.D.Find(m.Service, &c, "WHERE gid = ?", gid)
	exists := err == nil
	if err != nil && !errors.Is(err, db.ErrNullResult) {
		return c, "", err
	}
	var marker switchInheritance
	err = m.Manager.D.Find("__switch_inherit", &marker, "WHERE id = ?", m.inheritanceKey(gid))
	if err != nil && !errors.Is(err, db.ErrNullResult) {
		return c, "", err
	}
	if err == nil || !exists {
		return c, "inherit", nil
	}
	if c.Disable&1 == 0 {
		return c, "enabled", nil
	}
	return c, "disabled", nil
}

func (m *Control[CTX]) stateLocked(gid int64) (SwitchState, error) {
	global, globalMode, err := m.loadSwitchLocked(0)
	if err != nil {
		return SwitchState{}, err
	}
	group := GroupConfig{GroupID: gid}
	groupMode := "inherit"
	if gid != 0 {
		group, groupMode, err = m.loadSwitchLocked(gid)
		if err != nil {
			return SwitchState{}, err
		}
	}
	s := SwitchState{DefaultEnabled: !m.Options.DisableOnDefault, GlobalMode: globalMode, GroupMode: groupMode, Source: "default", EffectiveEnabled: !m.Options.DisableOnDefault}
	if groupMode != "inherit" {
		s.Source = "group"
		s.EffectiveEnabled = groupMode == "enabled"
	}
	if globalMode != "inherit" {
		s.Source = "global"
		s.EffectiveEnabled = globalMode == "enabled"
	}
	raw, _ := json.Marshal([]any{gid, s.DefaultEnabled, globalMode, groupMode, global.Disable, group.Disable})
	digest := sha256.Sum256(raw)
	s.Revision = hex.EncodeToString(digest[:])
	return s, nil
}

func (m *Control[CTX]) SwitchState(gid int64) (SwitchState, error) {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	return m.stateLocked(gid)
}

func (m *Control[CTX]) writeSwitchLocked(c GroupConfig, mode string) error {
	table := `"` + strings.ReplaceAll(m.Service, `"`, `""`) + `"`
	return m.Manager.D.Transaction(func(tx *sql.Tx) error {
		if mode != "inherit" {
			c.Disable &^= 1
			if mode == "disabled" {
				c.Disable |= 1
			}
		}
		if _, err := tx.Exec("INSERT OR REPLACE INTO "+table+" (gid,disable) VALUES (?,?)", c.GroupID, c.Disable); err != nil {
			return err
		}
		if mode == "inherit" {
			_, err := tx.Exec("INSERT OR REPLACE INTO __switch_inherit (id) VALUES (?)", m.inheritanceKey(c.GroupID))
			return err
		}
		_, err := tx.Exec("DELETE FROM __switch_inherit WHERE id = ?", m.inheritanceKey(c.GroupID))
		return err
	})
}

// SetSwitch serializes snapshots and writes with the normal control writers.
// expected is optional for chat; the admin always supplies a read revision.
func (m *Control[CTX]) SetSwitch(gid int64, mode, expected string) (SwitchState, error) {
	if mode != "enabled" && mode != "disabled" && mode != "inherit" {
		return SwitchState{}, errors.New("invalid switch mode")
	}
	m.Manager.Lock()
	defer m.Manager.Unlock()
	state, err := m.stateLocked(gid)
	if err != nil {
		return SwitchState{}, err
	}
	if expected != "" && state.Revision != expected {
		return SwitchState{}, ErrSwitchConflict
	}
	oldMode := state.GroupMode
	if gid == 0 {
		oldMode = state.GlobalMode
	}
	if oldMode == mode {
		return state, nil
	}
	c, _, err := m.loadSwitchLocked(gid)
	if err != nil {
		return SwitchState{}, err
	}
	if m.Manager.BeforeSwitch != nil {
		if err = m.Manager.BeforeSwitch(); err != nil {
			return SwitchState{}, err
		}
	}
	if err = m.writeSwitchLocked(c, mode); err != nil {
		return SwitchState{}, err
	}
	m.Cache = make(map[int64]uint8, 16)
	return m.stateLocked(gid)
}
