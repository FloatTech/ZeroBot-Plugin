package plugin_sleep_manage

import (
	"github.com/FloatTech/ZeroBot-Plugin/plugin_sleep_manage/model"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	go func() {
		defer func() {
			//recover() //可以打印panic的错误信息
			if err := recover(); err != nil { //产生了panic异常
				log.Println(err)
			}

		}() //别忘了(), 调用此匿名函数
		process.SleepAbout1sTo2s()
		_ = os.MkdirAll(dbpath, 0755)
		model.Initialize(dbfile)
	}()

}
