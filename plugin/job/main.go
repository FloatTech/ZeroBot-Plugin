// Package job 定时指令触发器
package job

import (
	"encoding/json"
	"hash/crc64"
	"strconv"
	"strings"
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
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	lo      map[int64]vevent.Loop
	entries map[int64]cron.EntryID // id entryid
	mu      sync.Mutex
	limit   = rate.NewLimiter(time.Second*2, 1)
)

func init() {
	en := control.Register("job", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              "定时指令触发器\n- 记录在\"cron\"触发的指令\n- 取消在\"cron\"触发的指令\n- 查看所有触发指令\n- 查看在\"cron\"触发的指令\n- 注入指令结果：任意指令",
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
			_ = db.FindFor(ids, c, "", func() error {
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
	en.OnFullMatch("查看所有触发指令", zero.SuperUserPermission, islonotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := &cmd{}
		ids := strconv.FormatInt(ctx.Event.SelfID, 36)
		mu.Lock()
		defer mu.Unlock()
		n, err := db.Count(ids)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		lst := make([]string, 0, n)
		err = db.FindFor(ids, c, "GROUP BY cron", func() error {
			lst = append(lst, c.Cron+"\n")
			return nil
		})
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(lst))
	})
	en.OnRegex(`^查看在"(.*)"触发的指令$`, zero.SuperUserPermission, islonotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := &cmd{}
		ids := strconv.FormatInt(ctx.Event.SelfID, 36)
		cron := ctx.State["regex_matched"].([]string)[1]
		mu.Lock()
		defer mu.Unlock()
		n, err := db.Count(ids)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		lst := make([]string, 0, n)
		err = db.FindFor(ids, c, "WHERE cron='"+cron+"'", func() error {
			lst = append(lst, c.Cmd+"\n")
			return nil
		})
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text(lst))
	})
	en.OnPrefix("注入指令结果：", ctxext.UserOrGrpAdmin, islonotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		command := ctx.State["args"].(string)
		if command != "" {
			vevent.NewLoopOf(vevent.NewAPICallerHook(ctx, func(rsp zero.APIResponse, err error) {
				if err == nil {
					logrus.Debugln("[job] CallerHook returned")
					id := message.NewMessageID(rsp.Data.Get("message_id").String())
					msg := ctx.GetMessage(id)
					ctx.Event.NativeMessage = json.RawMessage("\"" + msg.Elements.String() + "\"")
					ctx.Event.RawMessageID = json.RawMessage(msg.MessageId.String())
					ctx.Event.RawMessage = msg.Elements.String()
					time.Sleep(time.Second * 5) // 防止风控
					ctx.Event.Time = time.Now().Unix()
					vev, cl := binary.OpenWriterF(func(w *binary.Writer) {
						err = json.NewEncoder(w).Encode(ctx.Event)
					})
					if err != nil {
						cl()
						ctx.SendChain(message.Text("ERROR:", err))
						return
					}
					logrus.Debugln("[job] inject:", binary.BytesToString(vev))
					inject(ctx.Event.SelfID, vev)()
					cl()
				}
			})).Echo([]byte(strings.ReplaceAll(ctx.Event.RawEvent.Raw, "\"注入指令结果：", "\"")))
		}
	})
}

func islonotnil(ctx *zero.Ctx) bool {
	return len(lo) > 0
}

func inject(bot int64, response []byte) func() {
	return func() {
		if limit.Acquire() {
			lo[bot].Echo(response)
		}
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
	if err != nil {
		return err
	}
	return db.Del(bots, "WHERE cron='"+cron+"'")
}
