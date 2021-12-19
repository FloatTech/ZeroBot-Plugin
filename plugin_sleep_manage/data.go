package sleepmanage

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_sleep_manage/model"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

func init() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
			}
		}()
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		model.Initialize(dbfile)
	}()
}
