package goba

import (
	"fmt"

	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/sirupsen/logrus"
)

// HistoryStore preserves complete records separately from the finite model
// request. Implementations must isolate conversations and return read/write
// failures without including message text or credentials in errors.
type HistoryStore interface {
	Append(grp int64, kind string, item fmt.Stringer) error
	Modelize(p model.Protocol, grp int64, system string, userSystem bool, pre func(*fmt.Stringer)) (deepinfra.Model, error)
}

func (ag *Agent) SetHistoryStore(history HistoryStore) { ag.history = history }

func (ag *Agent) record(grp int64, kind string, item fmt.Stringer, bot bool) error {
	if ag.history != nil {
		if err := ag.history.Append(grp, kind, item); err != nil {
			logrus.Warnln("[goba] 对话记录保存失败，Agent 停止本次操作。")
			return err
		}
		return nil
	}
	ag.log.Add(grp, item, bot)
	return nil
}
