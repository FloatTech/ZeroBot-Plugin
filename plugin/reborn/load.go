package reborn

import (
	"encoding/json"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/zbputils/control"
)

// load 加载rate数据
func load(area *rate, jsonfile string) error {
	data, err := file.GetLazyData(jsonfile, control.Md5File, true)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, area)
}
