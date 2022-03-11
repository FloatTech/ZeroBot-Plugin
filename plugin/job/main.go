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
	"github.com/FloatTech/zbputils/web"
	"github.com/fumiama/cron"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	lo       map[int64]vevent.Loop
	entries  = map[int64]cron.EntryID{} // id entryid
	matchers = map[int64]*zero.Matcher{}
	mu       sync.Mutex
	limit    = rate.NewLimiter(time.Second*2, 1)
	en       = control.Register("job", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              "定时指令触发器\n- 记录以\"完全匹配关键词\"触发的指令\n- 取消以\"完全匹配关键词\"触发的指令\n- 记录在\"cron\"触发的指令\n- 取消在\"cron\"触发的指令\n- 查看所有触发指令\n- 查看在\"cron\"触发的指令\n- 查看以\"完全匹配关键词\"触发的指令\n- 注入指令结果：任意指令\n- 执行指令：任意指令",
		PrivateDataFolder: "job",
	})
)

func init() {
	db.DBPath = en.DataFolder() + "job.db"
	err := db.Open()
	if err != nil {
		panic(err)
	}
	go func() {
		process.GlobalInitMutex.Lock()
		process.SleepAbout1sTo2s()
		lo = make(map[int64]vevent.Loop, len(zero.BotConfig.Driver))
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
				if strings.HasPrefix(c.Cron, "fm:") {
					m := en.OnFullMatch(c.Cron[3:] /* skip fm: */).SetBlock(true)
					m.Handle(generalhandler(c))
					matchers[c.ID] = getmatcher(m)
					return nil
				}
				if strings.HasPrefix(c.Cron, "sm:") {
					m := en.OnFullMatch(c.Cron[3:] /* skip fm: */).SetBlock(true)
					m.Handle(superuserhandler(c))
					matchers[c.ID] = getmatcher(m)
					return nil
				}
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
	en.OnRegex(`^记录在"(.*)"触发的指令$`, ctxext.UserOrGrpAdmin, islonotnil, isfirstregmatchnotnil, logevent).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
	en.OnRegex(`^记录以"(.*)"触发的指令$`, zero.SuperUserPermission, islonotnil, isfirstregmatchnotnil, logevent).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := "fm:" + ctx.State["regex_matched"].([]string)[1]
		command := ctx.State["job_new_event"].(gjson.Result).Get("message").Raw
		logrus.Debugln("[job] get cmd:", command)
		c := &cmd{
			ID:   idof(cron, command),
			Cron: cron,
			Cmd:  command,
		}
		err := registercmd(ctx.Event.SelfID, c)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^记录以"(.*)"触发的代表我执行的指令$`, zero.SuperUserPermission, islonotnil, isfirstregmatchnotnil, logevent).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := "sm:" + ctx.State["regex_matched"].([]string)[1]
		command := ctx.State["job_raw_event"].(string)
		logrus.Debugln("[job] get cmd:", command)
		c := &cmd{
			ID:   idof(cron, command),
			Cron: cron,
			Cmd:  command,
		}
		err := registercmd(ctx.Event.SelfID, c)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^取消在"(.*)"触发的指令$`, ctxext.UserOrGrpAdmin, islonotnil, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := ctx.State["regex_matched"].([]string)[1]
		err := rmcmd(ctx.Event.SelfID, ctx.Event.UserID, cron)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^取消以"(.*)"触发的(代表我执行的)?指令$`, zero.SuperUserPermission, islonotnil, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		issu := ctx.State["regex_matched"].([]string)[2] != ""
		cron := ""
		if issu {
			cron = "sm:"
		} else {
			cron = "fm:"
		}
		cron += ctx.State["regex_matched"].([]string)[1]
		err := delcmd(ctx.Event.SelfID, cron)
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
	en.OnRegex(`^查看在"(.*)"触发的指令$`, zero.SuperUserPermission, islonotnil, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
	en.OnRegex(`^查看以"(.*)"触发的(代表我执行的)?指令$`, zero.SuperUserPermission, islonotnil, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := &cmd{}
		ids := strconv.FormatInt(ctx.Event.SelfID, 36)
		issu := ctx.State["regex_matched"].([]string)[2] != ""
		cron := ""
		if issu {
			cron = "sm:"
		} else {
			cron = "fm:"
		}
		cron += ctx.State["regex_matched"].([]string)[1]
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
	en.OnPrefix("执行指令：", ctxext.UserOrGrpAdmin, islonotnil, func(ctx *zero.Ctx) bool {
		return ctx.State["args"].(string) != ""
	}, parseArgs).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ev := strings.ReplaceAll(ctx.Event.RawEvent.Raw, "执行指令：", "")
		logrus.Debugln("[job] inject:", ev)
		inject(ctx.Event.SelfID, binary.StringToBytes(ev))()
	})
	en.OnPrefix("注入指令结果：", ctxext.UserOrGrpAdmin, islonotnil, func(ctx *zero.Ctx) bool {
		return ctx.State["args"].(string) != ""
	}, parseArgs).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		vevent.NewLoopOf(vevent.NewAPICallerHook(ctx, func(rsp zero.APIResponse, err error) {
			if err == nil {
				logrus.Debugln("[job] CallerHook returned")
				id := message.NewMessageID(rsp.Data.Get("message_id").String())
				if id.ID() == 0 {
					ctx.SendChain(message.Text("ERROR:未获取到返回结果"))
					return
				}
				msg := ctx.GetMessage(id)
				ctx.Event.NativeMessage = json.RawMessage("\"" + msg.Elements.String() + "\"")
				ctx.Event.RawMessageID = json.RawMessage(msg.MessageId.String())
				ctx.Event.RawMessage = msg.Elements.String()
				time.Sleep(time.Second * 5) // 防止风控
				ctx.Event.Time = time.Now().Unix()
				ctx.DeleteMessage(id)
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
		})).Echo(binary.StringToBytes(strings.ReplaceAll(ctx.Event.RawEvent.Raw, "注入指令结果：", "")))
	})
}

func islonotnil(ctx *zero.Ctx) bool {
	return len(lo) > 0
}

func isfirstregmatchnotnil(ctx *zero.Ctx) bool {
	return ctx.State["regex_matched"].([]string)[1] != ""
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

func registercmd(bot int64, c *cmd) error {
	mu.Lock()
	defer mu.Unlock()
	m := en.OnFullMatch(c.Cron[3:] /* skip fm: or sm: */).SetBlock(true)
	if strings.HasPrefix(c.Cron, "sm:") {
		m.Handle(superuserhandler(c))
	} else {
		m.Handle(generalhandler(c))
	}
	matchers[c.ID] = getmatcher(m)
	return db.Insert(strconv.FormatInt(bot, 36), c)
}

func generalhandler(c *cmd) zero.Handler {
	return func(ctx *zero.Ctx) {
		ctx.Event.NativeMessage = json.RawMessage(c.Cmd) // c.Cmd only have message
		ctx.Event.Time = time.Now().Unix()
		var err error
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
}

func superuserhandler(c *cmd) zero.Handler {
	e := &zero.Event{Sender: new(zero.User)}
	err := json.Unmarshal(binary.StringToBytes(c.Cmd), e)
	return func(ctx *zero.Ctx) {
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.Event.UserID = e.UserID
		ctx.Event.RawMessage = e.RawMessage
		ctx.Event.Sender = e.Sender
		ctx.Event.NativeMessage = e.NativeMessage
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
}

func rmcmd(bot, caller int64, cron string) error {
	c := &cmd{}
	mu.Lock()
	defer mu.Unlock()
	bots := strconv.FormatInt(bot, 36)
	e := new(zero.Event)
	var delcmd []string
	err := db.FindFor(bots, c, "WHERE cron='"+cron+"'", func() error {
		err := json.Unmarshal(binary.StringToBytes(c.Cmd), e)
		if err != nil {
			return err
		}
		if e.UserID != caller {
			return nil
		}
		eid, ok := entries[c.ID]
		if ok {
			process.CronTab.Remove(eid)
			delete(entries, c.ID)
			delcmd = append(delcmd, "id="+strconv.FormatInt(c.ID, 10))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(delcmd) > 0 {
		return db.Del(bots, "WHERE "+strings.Join(delcmd, " or "))
	}
	return nil
}

func delcmd(bot int64, cron string) error {
	c := &cmd{}
	mu.Lock()
	defer mu.Unlock()
	bots := strconv.FormatInt(bot, 36)
	var delcmd []string
	err := db.FindFor(bots, c, "WHERE cron='"+cron+"'", func() error {
		m, ok := matchers[c.ID]
		if ok {
			m.Delete()
			delete(matchers, c.ID)
			delcmd = append(delcmd, "id="+strconv.FormatInt(c.ID, 10))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(delcmd) > 0 {
		return db.Del(bots, "WHERE "+strings.Join(delcmd, " or "))
	}
	return nil
}

func parseArgs(ctx *zero.Ctx) bool {
	cmds := ctx.State["args"].(string)
	if !strings.Contains(cmds, "?::") && !strings.Contains(cmds, "!::") {
		return true
	}
	args := make(map[int]string)
	for strings.Contains(ctx.Event.RawEvent.Raw, "?::") {
		start := strings.Index(ctx.Event.RawEvent.Raw, "?::")
		msgend := strings.Index(ctx.Event.RawEvent.Raw[start+3:], "::")
		if msgend < 0 {
			ctx.SendChain(message.Text("ERROR:找不到结束的::"))
			return false
		}
		msgend += start + 3
		numend := strings.Index(ctx.Event.RawEvent.Raw[msgend+2:], "!")
		if numend <= 0 {
			ctx.SendChain(message.Text("ERROR:找不到结束的!"))
			return false
		}
		numend += msgend + 2
		logrus.Debugln("[job]", start, msgend, numend)
		msg := ctx.Event.RawEvent.Raw[start+3 : msgend]
		arg, err := strconv.Atoi(ctx.Event.RawEvent.Raw[msgend+2 : numend])
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		arr, ok := args[arg]
		if !ok {
			var id message.MessageID
			if msg == "" {
				id = ctx.SendChain(message.At(ctx.Event.UserID), message.Text("请输入参数", arg))
			} else {
				id = ctx.SendChain(message.At(ctx.Event.UserID), message.Text("[", arg, "] ", msg))
			}
			select {
			case <-time.After(time.Second * 120):
				ctx.Send(message.ReplyWithMessage(id, message.Text("参数读取超时")))
				if msg[0] != '?' {
					return false
				}
			case e := <-zero.NewFutureEvent("message", 0, true, zero.CheckUser(ctx.Event.UserID)).Next():
				args[arg] = e.Message.String()
				arr = args[arg]
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Reply(e.MessageID), message.Text("已记录"))
				process.SleepAbout1sTo2s()
			}
		}
		ctx.Event.RawEvent.Raw = ctx.Event.RawEvent.Raw[:start] + arr + ctx.Event.RawEvent.Raw[numend+1:]
	}
	args = make(map[int]string)
	for strings.Contains(ctx.Event.RawEvent.Raw, "!::") {
		start := strings.Index(ctx.Event.RawEvent.Raw, "!::")
		msgend := strings.Index(ctx.Event.RawEvent.Raw[start+3:], "::")
		if msgend < 0 {
			ctx.SendChain(message.Text("ERROR:找不到结束的::"))
			return false
		}
		msgend += start + 3
		numend := strings.Index(ctx.Event.RawEvent.Raw[msgend+2:], "!")
		if numend <= 0 {
			ctx.SendChain(message.Text("ERROR:找不到结束的!"))
			return false
		}
		numend += msgend + 2
		logrus.Debugln("[job]", start, msgend, numend)
		u := ctx.Event.RawEvent.Raw[start+3 : msgend]
		if u == "" {
			return false
		}
		arg, err := strconv.Atoi(ctx.Event.RawEvent.Raw[msgend+2 : numend])
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		arr, ok := args[arg]
		if !ok {
			isnilable := u[0] == '?'
			if isnilable {
				u = u[1:]
				if u == "" {
					return false
				}
			}
			b, err := web.GetData(u)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				if !isnilable {
					return false
				}
			}
			if len(b) > 0 {
				type fakejson struct {
					Arg string `json:"arg"`
				}
				f := fakejson{Arg: binary.BytesToString(b)}
				w := binary.SelectWriter()
				defer binary.PutWriter(w)
				_ = json.NewEncoder(w).Encode(&f)
				arr = w.String()[8 : w.Len()-3]
				args[arg] = arr
			}
		}
		w := binary.SelectWriter()
		w.WriteString(ctx.Event.RawEvent.Raw[:start])
		w.WriteString(arr)
		w.WriteString(ctx.Event.RawEvent.Raw[numend+1:])
		ctx.Event.RawEvent.Raw = string(w.Bytes())
		binary.PutWriter(w)
	}
	return true
}

func logevent(ctx *zero.Ctx) bool {
	ctx.SendChain(message.Text("您的下一条指令将被记录，在", ctx.State["regex_matched"].([]string)[1], "时触发"))
	select {
	case <-time.After(time.Second * 120):
		ctx.SendChain(message.Text("指令记录超时"))
		return false
	case e := <-zero.NewFutureEvent("message", 0, true, zero.CheckUser(ctx.Event.UserID)).Next():
		ctx.State["job_raw_event"] = e.RawEvent.Raw
		ctx.State["job_new_event"] = e.RawEvent
		return true
	}
}
