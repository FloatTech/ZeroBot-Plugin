// Package sleepmanage 睡眠管理
package sleepmanage

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"

	"github.com/FloatTech/zbputils/control/order"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_sleep_manage/model"
)

func init() {
	engine := control.Register("sleepmanage", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              "sleepmanage\n- 早安\n- 晚安",
		PrivateDataFolder: "sleep",
	})
	dbfile := engine.DataFolder() + "manage.db"
	engine.OnFullMatch("早安", isMorning, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln(err)
				return
			}
			position, getUpTime := db.GetUp(ctx.Event.GroupID, ctx.Event.UserID)
			log.Println(position, getUpTime)
			hour, minute, second := timeDuration(getUpTime)
			if (hour == 0 && minute == 0 && second == 0) || hour >= 24 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("早安成功！你是今天第%d个起床的", position)))
			} else {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("早安成功！你的睡眠时长为%d时%d分%d秒,你是今天第%d个起床的", hour, minute, second, position)))
			}
			db.Close()
		})
	engine.OnFullMatch("晚安", isEvening, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln(err)
				return
			}
			position, sleepTime := db.Sleep(ctx.Event.GroupID, ctx.Event.UserID)
			log.Println(position, sleepTime)
			hour, minute, second := timeDuration(sleepTime)
			if (hour == 0 && minute == 0 && second == 0) || hour >= 24 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("晚安成功！你是今天第%d个睡觉的", position)))
			} else {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("晚安成功！你的清醒时长为%d时%d分%d秒,你是今天第%d个睡觉的", hour, minute, second, position)))
			}
			db.Close()
		})
}

func timeDuration(time time.Duration) (hour, minute, second int64) {
	hour = int64(time) / (1000 * 1000 * 1000 * 60 * 60)
	minute = (int64(time) - hour*(1000*1000*1000*60*60)) / (1000 * 1000 * 1000 * 60)
	second = (int64(time) - hour*(1000*1000*1000*60*60) - minute*(1000*1000*1000*60)) / (1000 * 1000 * 1000)
	return hour, minute, second
}

// 只统计6点到12点的早安
func isMorning(ctx *zero.Ctx) bool {
	now := time.Now().Hour()
	return now >= 6 && now <= 12
}

// 只统计21点到凌晨3点的晚安
func isEvening(ctx *zero.Ctx) bool {
	now := time.Now().Hour()
	return now >= 21 || now <= 3
}
