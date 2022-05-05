package reborn

import (
	"encoding/json"

	"github.com/FloatTech/zbputils/file"
)

// load 加载rate数据
func load(area *rate, jsonfile string) error {
	data, err := file.GetLazyData(jsonfile, true)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, area)
}
