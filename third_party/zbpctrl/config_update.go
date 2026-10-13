package control

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"

	sql "github.com/FloatTech/sqlite"
)

var ErrConfigConflict = errors.New("configuration changed")

func (m *Control[CTX]) extraLocked() (json.RawMessage, string, error) {
	if m.Options.Extra == 0 {
		return nil, "", ErrUnregisteredExtra
	}
	var row ResponseGroup
	err := m.Manager.D.Find("__resp", &row, "WHERE gid = ?", int64(m.Options.Extra))
	if err != nil && !errors.Is(err, sql.ErrNullResult) {
		return nil, "", err
	}
	var raw json.RawMessage
	if err == nil && row.Extra != "" && row.Extra != "-" {
		raw = json.RawMessage(row.Extra)
	}
	h := hmac.New(sha256.New, m.Manager.revisionKey[:])
	h.Write([]byte(strconv.FormatInt(int64(m.Options.Extra), 10)))
	h.Write([]byte{0})
	h.Write(raw)
	return raw, hex.EncodeToString(h.Sum(nil)), nil
}

// ExtraJSON reads persisted extension data independently of response gates.
func (m *Control[CTX]) ExtraJSON() (json.RawMessage, string, error) {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	return m.extraLocked()
}

// UpdateExtraJSON serializes read/merge/backup/write. No cache or runtime value
// changes before SQLite commits. expected may be empty for chat field updates.
func (m *Control[CTX]) UpdateExtraJSON(expected string, change func(json.RawMessage) (json.RawMessage, error)) (string, error) {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	raw, revision, err := m.extraLocked()
	if err != nil {
		return "", err
	}
	if expected != "" && expected != revision {
		return "", ErrConfigConflict
	}
	next, err := change(raw)
	if err != nil {
		return "", err
	}
	if bytes.Equal(raw, next) {
		return revision, nil
	}
	if !json.Valid(next) {
		return "", errors.New("invalid extension JSON")
	}
	if m.Manager.BeforeExtra != nil {
		if err = m.Manager.BeforeExtra(); err != nil {
			return "", err
		}
	}
	err = m.Manager.D.Insert("__resp", &ResponseGroup{GroupID: int64(m.Options.Extra), Extra: string(next)})
	if err != nil {
		return "", err
	}
	m.Manager.respCache[int64(m.Options.Extra)] = string(next)
	_, revision, err = m.extraLocked()
	return revision, err
}

func (m *Control[CTX]) dataLocked(gid int64) (GroupConfig, string, string, error) {
	row, mode, err := m.loadSwitchLocked(gid)
	if err != nil {
		return row, mode, "", err
	}
	raw, _ := json.Marshal([]any{m.Service, gid, row.Disable, mode})
	h := sha256.Sum256(raw)
	return row, mode, hex.EncodeToString(h[:]), nil
}

func (m *Control[CTX]) DataVersion(gid int64) (int64, string, error) {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	row, _, revision, err := m.dataLocked(gid)
	return (row.Disable >> 1) & 0x3fffffff_ffffffff, revision, err
}

// UpdateData preserves the switch/default bits and inheritance while merging
// settings under the Manager lock, avoiding lost chat/web field updates.
func (m *Control[CTX]) UpdateData(gid int64, expected string, change func(int64) (int64, error)) (string, error) {
	m.Manager.Lock()
	defer m.Manager.Unlock()
	row, mode, revision, err := m.dataLocked(gid)
	if err != nil {
		return "", err
	}
	if expected != "" && expected != revision {
		return "", ErrConfigConflict
	}
	old := (row.Disable >> 1) & 0x3fffffff_ffffffff
	next, err := change(old)
	if err != nil {
		return "", err
	}
	if next < 0 || next > 0x3fffffff_ffffffff {
		return "", errors.New("invalid configuration bits")
	}
	if next == old {
		return revision, nil
	}
	if m.BeforeData != nil {
		if err = m.BeforeData(); err != nil {
			return "", err
		}
	}
	row.Disable = (row.Disable & int64(-9223372036854775807)) | (next << 1)
	if err = m.writeSwitchLocked(row, mode); err != nil {
		return "", err
	}
	m.Cache = make(map[int64]uint8, 16)
	_, _, revision, err = m.dataLocked(gid)
	return revision, err
}
