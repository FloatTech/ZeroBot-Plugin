package control

import (
	"encoding/json"
	"errors"
)

var (
	ErrEmptyExtra        = errors.New("empty extra")
	ErrUnregisteredExtra = errors.New("unregistered extra")
)

func (m *Control[CTX]) GetData(gid int64) int64 {
	data, _, _ := m.DataVersion(gid)
	return data
}

func (m *Control[CTX]) SetData(gid, data int64) error {
	_, err := m.UpdateData(gid, "", func(int64) (int64, error) { return data, nil })
	return err
}

func (m *Control[CTX]) GetExtra(obj any) error {
	raw, _, err := m.ExtraJSON()
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return ErrEmptyExtra
	}
	return json.Unmarshal(raw, obj)
}

func (m *Control[CTX]) SetExtra(obj any) error {
	raw, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = m.UpdateExtraJSON("", func(json.RawMessage) (json.RawMessage, error) { return raw, nil })
	return err
}
