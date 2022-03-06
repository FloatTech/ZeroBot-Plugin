// Package job 定时指令触发器
package job

import (
	"hash/crc64"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/vevent"
	"github.com/fumiama/cron"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	lo      map[int64]vevent.Loop
	entries map[int64]cron.EntryID // id entryid
	mu      sync.Mutex
)

func init() {
	en := control.Register("job", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              "定时指令触发器\n- 记录在\"cron\"触发的指令\n- 取消在\"cron\"触发的指令",
		PrivateDataFolder: "job",
	})
	db.DBPath = en.DataFolder() + "job.db"
	err := db.Open()
	if err != nil {
		panic(err)
	}
	go func() {
		process.GlobalInitMutex.Lock()
		process.SleepAbout1sTo2s()
		lo = make(map[int64]vevent.Loop, len(zero.BotConfig.Driver))
		entries = map[int64]cron.EntryID{}
		for _, drv := range zero.BotConfig.Driver {
			id := drv.SelfID()
			ids := strconv.FormatInt(id, 36)
			c := &cmd{}
			lo[id] = vevent.NewLoop(id)
			err := db.Create(ids, c)
			logrus.Infoln("[job]创建表", ids)
			if err != nil {
				panic(err)
			}
			db.FindFor(ids, c, "", func() error {
				mu.Lock()
				defer mu.Unlock()
				eid, err := process.CronTab.AddFunc(c.Cron, inject(id, []byte(c.Cmd)))
				if err != nil {
					return err
				}
				entries[c.ID] = eid
				return nil
			})
		}
		logrus.Infoln("[job]本地环回初始化完成")
		process.GlobalInitMutex.Unlock()
	}()
	en.OnRegex(`^记录在"(.*)"触发的指令$`, ctxext.UserOrGrpAdmin, islonotnil, func(ctx *zero.Ctx) bool {
		ctx.SendChain(message.Text("您的下一条指令将被记录，在", ctx.State["regex_matched"].([]string)[1], "时触发"))
		select {
		case <-time.After(time.Second * 120):
			ctx.SendChain(message.Text("指令记录超时"))
			return false
		case e := <-zero.NewFutureEvent("message", 0, false, zero.CheckUser(ctx.Event.UserID)).Next():
			ctx.State["job_raw_event"] = e.RawEvent.Raw
			return true
		}
	}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := ctx.State["regex_matched"].([]string)[1]
		command := ctx.State["job_raw_event"].(string)
		c := &cmd{
			ID:   idof(cron, command),
			Cron: cron,
			Cmd:  command,
		}
		err := addcmd(ctx.Event.SelfID, c)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^取消在"(.*)"触发的指令$`, ctxext.UserOrGrpAdmin, islonotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := ctx.State["regex_matched"].([]string)[1]
		err := rmcmd(ctx.Event.SelfID, cron)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
}

func islonotnil(ctx *zero.Ctx) bool {
	return len(lo) > 0
}

func inject(bot int64, response []byte) func() {
	return func() {
		lo[bot].Echo(response)
	}
}

func idof(cron, cmd string) int64 {
	return int64(crc64.Checksum(binary.StringToBytes(cron+cmd), crc64.MakeTable(crc64.ISO)))
}

func addcmd(bot int64, c *cmd) error {
	mu.Lock()
	defer mu.Unlock()
	eid, err := process.CronTab.AddFunc(c.Cron, inject(bot, []byte(c.Cmd)))
	if err != nil {
		return err
	}
	entries[c.ID] = eid
	return db.Insert(strconv.FormatInt(bot, 36), c)
}

func rmcmd(bot int64, cron string) error {
	c := &cmd{}
	mu.Lock()
	defer mu.Unlock()
	bots := strconv.FormatInt(bot, 36)
	err := db.FindFor(bots, c, "WHERE cron='"+cron+"'", func() error {
		eid, ok := entries[c.ID]
		if ok {
			process.CronTab.Remove(eid)
			delete(entries, c.ID)
		}
		return nil
	})
	db.Del(bots, "WHERE cron='"+cron+"'")
	return err
}
