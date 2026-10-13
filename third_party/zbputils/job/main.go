// Package job 定时指令触发器
package job

import (
	"encoding/json"
	"errors"
	"hash/crc64"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/fumiama/cron"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/floatbox/web"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/vevent"
)

var (
	entries  = map[int64]cron.EntryID{} // id entryid
	matchers = map[int64]*zero.Matcher{}
	mu       sync.RWMutex
	en       = control.Register("job", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "定时指令触发器",
		Help:              "- 记录以\"完全匹配关键词\"触发的指令\n- 取消以\"完全匹配关键词\"触发的指令\n- 记录在\"cron\"触发的(别名xxx的)指令\n- 取消在\"cron\"触发的指令\n- 查看所有触发指令\n- 查看在\"cron\"触发的指令\n- 查看以\"完全匹配关键词\"触发的指令\n- 注入指令结果：任意指令\n- 执行指令：任意指令\n- [我|大家|有人][说|问][正则表达式]你[答|说|做|执行][模版]\n- [查看|看看][我|大家|有人][说|问][正则表达式]\n- 删除[大家|有人|我][说|问|让你做|让你执行][正则表达式]",
		PrivateDataFolder: "job",
	})
)

func init() {
	err := db.Open(time.Hour)
	if err != nil {
		panic(err)
	}
	go func() {
		process.GlobalInitMutex.Lock()
		process.SleepAbout1sTo2s()
		zero.RangeBot(func(id int64, _ *zero.Ctx) bool {
			ids := strconv.FormatInt(id, 36)
			c := &cmd{}
			err := db.Create(ids, c)
			logrus.Debugln("[job]创建表", ids)
			if err != nil {
				panic(err)
			}
			err = db.FindFor(ids, c, "", func() error {
				mu.Lock()
				defer mu.Unlock()
				if strings.HasPrefix(c.Cron, "fm:") {
					m := en.OnFullMatch(c.Cron[3:] /* skip fm: */).SetBlock(true)
					m.Handle(generalhandler(c.Cmd))
					matchers[c.ID] = (*zero.Matcher)(m)
					return nil
				}
				if strings.HasPrefix(c.Cron, "sm:") {
					m := en.OnFullMatch(c.Cron[3:] /* skip sm: */).SetBlock(true)
					h, err := superuserhandler(binary.StringToBytes(c.Cmd))
					if err != nil {
						return nil
					}
					m.Handle(h)
					matchers[c.ID] = (*zero.Matcher)(m)
					return nil
				}
				if strings.HasPrefix(c.Cron, "rm:") || strings.HasPrefix(c.Cron, "im:") {
					patttens := strings.SplitN(c.Cron, ":", 3)
					if len(patttens) != 3 {
						return errors.New("error regex match global pattern")
					}
					grp, err := strconv.ParseInt(patttens[1], 36, 64)
					if err != nil {
						return err
					}
					if global.group[grp] == nil {
						global.group[grp] = new(regexGroup)
					}
					tmpl := make([]byte, len(c.Cmd))
					copy(tmpl, c.Cmd)
					global.group[grp].All = append(global.group[grp].All, inst{
						regex:    regexp.MustCompile(transformPattern(patttens[2])),
						Pattern:  patttens[2],
						Template: binary.BytesToString(tmpl),
						IsInject: patttens[0][0] == 'i',
					})
					return nil
				}
				if strings.HasPrefix(c.Cron, "rp:") || strings.HasPrefix(c.Cron, "ip:") {
					patttens := strings.SplitN(c.Cron, ":", 4)
					if len(patttens) != 4 {
						return errors.New("error regex match private pattern")
					}
					uid, err := strconv.ParseInt(patttens[1], 36, 64)
					if err != nil {
						return err
					}
					gid, err := strconv.ParseInt(patttens[2], 36, 64)
					if err != nil {
						return err
					}
					if global.group[gid] == nil {
						global.group[gid] = new(regexGroup)
					}
					tmpl := make([]byte, len(c.Cmd))
					copy(tmpl, c.Cmd)
					if global.group[gid].Private == nil {
						global.group[gid].Private = make(map[int64][]inst)
					}
					global.group[gid].Private[uid] = append(global.group[gid].Private[uid], inst{
						regex:    regexp.MustCompile(transformPattern(patttens[3])),
						Pattern:  patttens[3],
						Template: binary.BytesToString(tmpl),
						IsInject: patttens[0][0] == 'i',
					})
					return nil
				}
				cr, _, _ := strings.Cut(c.Cron, ":->")
				eid, err := process.CronTab.AddFunc(cr, inject(zero.GetBot(id), []byte(c.Cmd)))
				if err != nil {
					return err
				}
				entries[c.ID] = eid
				return nil
			})
			if err != nil && err != sql.ErrNullResult {
				panic(err)
			}
			return true
		})
		logrus.Infoln("[job]本地环回初始化完成")
		process.GlobalInitMutex.Unlock()
	}()
	en.OnRegex(`^记录在"(.*)"触发的(别名.*的)?指令$`, zero.UserOrGrpAdmin, isfirstregmatchnotnil, logevent).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := ctx.State["regex_matched"].([]string)[1]
		alias := ctx.State["regex_matched"].([]string)[2]
		command := ctx.State["job_raw_event"].(string)
		if alias != "" {
			cron += ":->" + alias[len("别名"):len(alias)-len("的")]
		}
		c := &cmd{
			ID:   idof(cron, command),
			Cron: cron,
			Cmd:  command,
		}
		err := addcmd(ctx, c)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^记录以"(.*)"触发的指令$`, zero.SuperUserPermission, isfirstregmatchnotnil, logevent).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^记录以"(.*)"触发的代表我执行的指令$`, zero.SuperUserPermission, isfirstregmatchnotnil, logevent).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^取消在"(.*)"触发的指令$`, zero.UserOrGrpAdmin, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cron := ctx.State["regex_matched"].([]string)[1]
		err := rmcmd(ctx.Event.SelfID, ctx.Event.UserID, cron, zero.AdminPermission(ctx))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnRegex(`^取消以"(.*)"触发的(代表我执行的)?指令$`, zero.SuperUserPermission, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("成功!"))
	})
	en.OnFullMatch("查看所有触发指令", zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := &cmd{}
		ids := strconv.FormatInt(ctx.Event.SelfID, 36)
		mu.Lock()
		defer mu.Unlock()
		n, err := db.Count(ids)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		lst := make([]string, 0, n+2)
		q := ""
		args := []any{}
		if ctx.Event.GroupID != 0 {
			grp := strconv.FormatInt(ctx.Event.GroupID, 36)
			q = "WHERE cron LIKE 'fm:%' OR cron LIKE 'sm:%' OR cron LIKE ? OR cron LIKE ? "
			args = []any{"_m:" + grp + ":%", "_p:" + grp + ":?:%"}
			lst = append(lst, "在本群的触发指令]\n")
		} else {
			lst = append(lst, "全部触发指令]\n")
		}
		q += "GROUP BY cron"
		err = db.FindFor(ids, c, q, func() error {
			lst = append(lst, c.Cron+"\n")
			return nil
		}, args...)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		lst = append(lst, "[END")
		ctx.SendChain(message.Text(lst))
	})
	en.OnRegex(`^查看在"(.*)"触发的指令$`, zero.SuperUserPermission, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := &cmd{}
		ids := strconv.FormatInt(ctx.Event.SelfID, 36)
		cron := ctx.State["regex_matched"].([]string)[1]
		mu.Lock()
		defer mu.Unlock()
		n, err := db.Count(ids)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		lst := make([]string, 0, n)
		err = db.FindFor(ids, c, "WHERE cron = ?", func() error {
			lst = append(lst, c.Cmd+"\n")
			return nil
		}, cron)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(lst))
	})
	en.OnRegex(`^查看以"(.*)"触发的(代表我执行的)?指令$`, zero.SuperUserPermission, isfirstregmatchnotnil).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		lst := make([]string, 0, n)
		err = db.FindFor(ids, c, "WHERE cron = ?", func() error {
			lst = append(lst, c.Cmd+"\n")
			return nil
		}, cron)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text(lst))
	})
	en.OnPrefix("执行指令：", zero.UserOrGrpAdmin, func(ctx *zero.Ctx) bool {
		return ctx.State["args"].(string) != ""
	}, parseArgs).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ev := strings.ReplaceAll(ctx.Event.RawEvent.Raw, "执行指令：", "")
		logrus.Debugln("[job] inject:", ev)
		inject(ctx, binary.StringToBytes(ev))()
	})
	en.OnPrefix("注入指令结果：", zero.UserOrGrpAdmin, func(ctx *zero.Ctx) bool {
		return ctx.State["args"].(string) != ""
	}, parseArgs).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		hook := vevent.NewAPICallerReturnHook(ctx, func(_ zero.APIRequest, rsp zero.APIResponse, err error) {
			if err == nil {
				logrus.Debugln("[job] CallerHook returned")
				id := message.NewMessageIDFromInteger(rsp.Data.Get("message_id").Int())
				if id.ID() == 0 {
					ctx.SendChain(message.Text("ERROR:未获取到返回结果"))
					return
				}
				msg := ctx.GetMessage(id)
				ctx.Event.NativeMessage = json.RawMessage("\"" + msg.Elements.String() + "\"")
				ctx.Event.RawMessageID = json.RawMessage(msg.MessageID.String())
				ctx.Event.RawMessage = msg.Elements.String()
				process.SleepAbout1sTo2s() // 防止风控
				ctx.Event.Time = time.Now().Unix()
				ctx.DeleteMessage(id)
				vev, cl := binary.OpenWriterF(func(w *binary.Writer) {
					err = json.NewEncoder(w).Encode(ctx.Event)
				})
				if err != nil {
					cl()
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				logrus.Debugln("[job] inject:", binary.BytesToString(vev))
				defer func() {
					_ = recover()
					cl()
				}()
				ctx.Echo(vev)
			}
		})
		hookedctx := *ctx //nolint: govet
		vevent.HookCtxCaller(&hookedctx, hook)
		hookedctx.Echo(binary.StringToBytes(strings.ReplaceAll(ctx.Event.RawEvent.Raw, "注入指令结果：", "")))
	})
}

func isfirstregmatchnotnil(ctx *zero.Ctx) bool {
	return ctx.State["regex_matched"].([]string)[1] != ""
}

func inject(ctx *zero.Ctx, response []byte) func() {
	return func() { ctx.Echo(response) }
}

func idof(cron, cmd string) int64 {
	return int64(crc64.Checksum(binary.StringToBytes(cron+cmd), crc64.MakeTable(crc64.ISO)))
}

func addcmd(ctx *zero.Ctx, c *cmd) error {
	mu.Lock()
	defer mu.Unlock()
	cr, _, _ := strings.Cut(c.Cron, ":->")
	eid, err := process.CronTab.AddFunc(cr, inject(ctx, []byte(c.Cmd)))
	if err != nil {
		return err
	}
	entries[c.ID] = eid
	return db.Insert(strconv.FormatInt(ctx.Event.SelfID, 36), c)
}

func registercmd(bot int64, c *cmd) error {
	mu.Lock()
	defer mu.Unlock()
	m := en.OnFullMatch(c.Cron[3:] /* skip fm: or sm: */).SetBlock(true)
	if strings.HasPrefix(c.Cron, "sm:") {
		h, err := superuserhandler(binary.StringToBytes(c.Cmd))
		if err != nil {
			return err
		}
		m.Handle(h)
	} else {
		m.Handle(generalhandler(c.Cmd))
	}
	matchers[c.ID] = (*zero.Matcher)(m)
	return db.Insert(strconv.FormatInt(bot, 36), c)
}

func generalhandler(command string) zero.Handler {
	cmdraw := make(json.RawMessage, len(command))
	copy(cmdraw, command)
	return func(ctx *zero.Ctx) {
		ctx.Event.NativeMessage = cmdraw
		ctx.Event.Time = time.Now().Unix()
		var err error
		vev, cl := binary.OpenWriterF(func(w *binary.Writer) {
			err = json.NewEncoder(w).Encode(ctx.Event)
		})
		if err != nil {
			cl()
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		logrus.Debugln("[job] inject:", binary.BytesToString(vev))
		defer func() {
			_ = recover()
			cl()
		}()
		ctx.Echo(vev)
	}
}

func superuserhandler(rsp []byte) (zero.Handler, error) {
	e := &zero.Event{Sender: new(zero.User)}
	err := json.Unmarshal(rsp, e)
	if err != nil {
		return nil, err
	}
	return func(ctx *zero.Ctx) {
		ctx.Event.UserID = e.UserID
		ctx.Event.RawMessage = e.RawMessage
		ctx.Event.Sender = e.Sender
		ctx.Event.NativeMessage = e.NativeMessage
		vev, cl := binary.OpenWriterF(func(w *binary.Writer) {
			err = json.NewEncoder(w).Encode(ctx.Event)
		})
		if err != nil {
			cl()
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		logrus.Debugln("[job] inject:", binary.BytesToString(vev))
		defer func() {
			_ = recover()
			cl()
		}()
		ctx.Echo(vev)
	}, nil
}

func rmcmd(bot, caller int64, cron string, force bool) error {
	c := &cmd{}
	mu.Lock()
	defer mu.Unlock()
	bots := strconv.FormatInt(bot, 36)
	e := new(zero.Event)
	var delids []int64
	err := db.FindFor(bots, c, "WHERE cron = ? OR cron LIKE ?", func() error {
		err := json.Unmarshal(binary.StringToBytes(c.Cmd), e)
		if err != nil {
			return err
		}
		if !force && e.UserID != caller {
			return nil
		}
		eid, ok := entries[c.ID]
		if ok {
			process.CronTab.Remove(eid)
			delete(entries, c.ID)
			delids = append(delids, c.ID)
		}
		return nil
	}, cron, cron+":->%")
	if err != nil {
		return err
	}
	if len(delids) > 0 {
		q, s := sql.QuerySet("WHERE id", "IN", delids)
		return db.Del(bots, q, s...)
	}
	return nil
}

func delcmd(bot int64, cron string) error {
	c := &cmd{}
	mu.Lock()
	defer mu.Unlock()
	bots := strconv.FormatInt(bot, 36)
	var delids []int64
	err := db.FindFor(bots, c, "WHERE cron = ?", func() error {
		m, ok := matchers[c.ID]
		if ok {
			m.Delete()
			delete(matchers, c.ID)
			delids = append(delids, c.ID)
		}
		return nil
	}, cron)
	if err != nil {
		return err
	}
	if len(delids) > 0 {
		q, s := sql.QuerySet("WHERE id", "IN", delids)
		return db.Del(bots, q, s...)
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		arr, ok := args[arg]
		if !ok {
			var id message.ID
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
			case c := <-zero.NewFutureEvent("message", 0, true, zero.CheckUser(ctx.Event.UserID)).Next():
				args[arg] = c.Event.Message.String()
				arr = args[arg]
				process.SleepAbout1sTo2s()
				ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("已记录"))
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
			ctx.SendChain(message.Text("ERROR: ", err))
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
				ctx.SendChain(message.Text("ERROR: ", err))
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
	ctx.SendChain(message.Text("您的下一条指令将被记录, 在", ctx.State["regex_matched"].([]string)[1], "时触发"))
	select {
	case <-time.After(time.Second * 120):
		ctx.SendChain(message.Text("指令记录超时"))
		return false
	case c := <-zero.NewFutureEvent("message", 0, true, zero.CheckUser(ctx.Event.UserID)).Next():
		ctx.State["job_raw_event"] = c.Event.RawEvent.Raw
		ctx.State["job_new_event"] = c.Event.RawEvent
		return true
	}
}
