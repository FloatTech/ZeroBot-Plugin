package reborn

import (
	"encoding/json"
	"os"

	wr "github.com/mroth/weightedrand"
	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/order"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
)

const (
	datapath = "data/Reborn"
	jsonfile = datapath + "/rate.json"
)

type rate []struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

var (
	areac *wr.Chooser
)

func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		area := make(rate, 226)
		err = load(&area)
		if err != nil {
			panic(err)
		}
		choices := make([]wr.Choice, len(area))
		for i, a := range area {
			choices[i].Item = a.Name
			choices[i].Weight = uint(a.Weight * 1e9)
		}
		areac, err = wr.NewChooser(choices...)
		if err != nil {
			panic(err)
		}
		log.Printf("[Reborn]读取%d个国家/地区", len(area))
	}()
}

// load 加载rate数据
func load(area *rate) error {
	data, err := file.GetLazyData(jsonfile, true, true)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, area)
}
