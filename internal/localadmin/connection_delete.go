package localadmin

import (
	"encoding/json"
	"errors"
)

type ConnectionRemoval struct {
	Kind    string `json:"kind"`
	Index   *int   `json:"index"`
	Confirm bool   `json:"confirm"`
}

func applyConnectionRemoval(root jsonObject, remove ConnectionRemoval) ([]byte, error) {
	if !remove.Confirm || remove.Index == nil || (remove.Kind != "ws" && remove.Kind != "wss") {
		return nil, errors.New("请选择有效的连接类型、明确序号并确认删除")
	}
	key, raw, _ := field(root, remove.Kind)
	var list []json.RawMessage
	_ = json.Unmarshal(raw, &list)
	if *remove.Index < 0 || *remove.Index >= len(list) {
		return nil, errors.New("连接不存在，请重新加载")
	}
	total := 0
	for _, kind := range []string{"ws", "wss"} {
		_, items, _ := field(root, kind)
		var connections []json.RawMessage
		_ = json.Unmarshal(items, &connections)
		total += len(connections)
	}
	if total <= 1 {
		return nil, errors.New("至少保留一条连接，不能删除最后一条")
	}
	remaining := append(list[:*remove.Index], list[*remove.Index+1:]...)
	root[key], _ = json.Marshal(remaining)
	b, err := json.MarshalIndent(root, "", "  ")
	return append(b, '\n'), err
}
