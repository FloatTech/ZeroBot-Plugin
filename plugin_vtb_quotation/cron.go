package plugin_vtb_quotation

import (
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/firstVtb"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/secondVtb"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	engine.OnMessage(atriRule).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		log.Println("定时任务只创建一次")
		AtriRule = false
		log.Println("开启vtb数据库日常更新")
		vtbDaily()
	})
}

func vtbDaily() {
	log.Println("创建vtb数据库定时任务")
	c := cron.New()
	if err := c.AddFunc("0 0 4 * * ?", func() { vtbData() }); err != nil {
		log.Println("定时任务有错误:", err)
	} else {
		log.Println("开启vtb数据库定时任务")
		c.Start()
	}
}
func vtbData() {
	model.Init()
	vtbListStr := firstVtb.GetVtbListStr()
	uidList := firstVtb.DealVtbListStr(vtbListStr)
	log.Println(uidList)
	for _, v := range uidList {
		vtbStr := secondVtb.GetVtbStr(v)
		secondVtb.DealVtbStr(vtbStr, v)
	}
	model.Db.Close()
}
