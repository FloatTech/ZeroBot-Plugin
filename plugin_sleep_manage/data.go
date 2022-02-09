package sleepmanage

import (
	"os"

	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/zbputils/control/order"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_sleep_manage/model"
)

func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		model.Initialize(dbfile)
	}()
}
