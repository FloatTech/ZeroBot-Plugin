package vtbquotation

import (
	"github.com/fumiama/cron"
	log "github.com/sirupsen/logrus"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
)

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		log.Println("[vtb/cron] 开启vtb数据库日常更新")
		vtbDaily()
	}()
}

func vtbDaily() {
	c := cron.New()
	_, err := c.AddFunc("0 4 * * *", func() { vtbData() })
	if err != nil {
		log.Errorln("定时任务有错误:", err)
	} else {
		log.Println("开启vtb数据库定时任务")
		c.Start()
	}
}

func vtbData() {
	db := model.Init(dbfile)
	if db != nil {
		for _, v := range db.GetVtbList() {
			db.StoreVtb(v)
		}
		err := db.Close()
		if err != nil {
			log.Errorln("[vtb/cron]", err)
		}
	}
}
