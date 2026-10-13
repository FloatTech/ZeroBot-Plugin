package job

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/process"
	sql "github.com/FloatTech/sqlite"

	"github.com/FloatTech/zbputils/ctxext"
)

var global = context{
	group: make(map[int64]*regexGroup),
}

type context struct {
	group map[int64]*regexGroup
}

type regexGroup struct {
	All     []inst
	Private map[int64][]inst
}

type inst struct {
	regex    *regexp.Regexp
	Pattern  string
	Template string
	IsInject bool
}

var transformRegex = regexp.MustCompile(`<<.+?>>`)

func transformPattern(pattern string) string {
	pattern = transformRegex.ReplaceAllStringFunc(pattern, func(s string) string {
		s = strings.Trim(s, "<>")
		return `(?P<` + s + `>.+?)`
	})
	return "^" + pattern + "$"
}

// isPrivate:false & id:0 is global
func saveRegex(gid, uid int64, bots, pattern, template string) error {
	cr := "rm:"
	if uid > 0 {
		cr = "rp:" + strconv.FormatInt(uid, 36) + ":"
	}
	cr += strconv.FormatInt(gid, 36) + ":" + pattern
	return db.Insert(bots, &cmd{
		ID:   idof(cr, template),
		Cron: cr,
		Cmd:  template,
	})
}

// isPrivate:false & id:0 is global
func saveInjectRegex(gid, uid int64, bots, pattern, template string) error {
	cr := "im:"
	if uid > 0 {
		cr = "ip:" + strconv.FormatInt(uid, 36) + ":"
	}
	cr += strconv.FormatInt(gid, 36) + ":" + pattern
	return db.Insert(bots, &cmd{
		ID:   idof(cr, template),
		Cron: cr,
		Cmd:  template,
	})
}

// isPrivate:false & id:0 is global
func removeRegex(gid, uid int64, bots, pattern string) error {
	cr := "rm:"
	if uid > 0 {
		cr = "rp:" + strconv.FormatInt(uid, 36) + ":"
	}
	cr += strconv.FormatInt(gid, 36) + ":" + pattern
	c := &cmd{}
	var delids []int64
	_ = db.FindFor(bots, c, "WHERE cron = ?", func() error {
		delids = append(delids, c.ID)
		return nil
	}, cr)
	if len(delids) > 0 {
		q, s := sql.QuerySet("WHERE id", "IN", delids)
		return db.Del(bots, q, s...)
	}
	return nil
}

// isPrivate:false & id:0 is global
func removeInjectRegex(gid, uid int64, bots, pattern string) error {
	cr := "im:"
	if uid > 0 {
		cr = "ip:" + strconv.FormatInt(uid, 36) + ":"
	}
	cr += strconv.FormatInt(gid, 36) + ":" + pattern
	c := &cmd{}
	var delids []int64
	_ = db.FindFor(bots, c, "WHERE cron = ?", func() error {
		delids = append(delids, c.ID)
		return nil
	}, cr)
	if len(delids) > 0 {
		q, s := sql.QuerySet("WHERE id", "IN", delids)
		return db.Del(bots, q, s...)
	}
	return nil
}

func init() {
	en.OnRegex(`^(我|大家|有人)(说|问)(.*)你(答|说|做|执行)`, zero.OnlyGroup, zero.OnlyToMe).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()

		matched := ctx.State["regex_matched"].([]string)
		all := true
		if matched[1] == "我" {
			all = false
		}
		if all && !zero.AdminPermission(ctx) {
			ctx.SendChain(message.Text("非管理员/主人无法设置全局问答"))
			return
		}
		isInject := false
		if matched[4] == "做" || matched[4] == "执行" {
			if !zero.AdminPermission(ctx) {
				ctx.SendChain(message.Text("非管理员/主人无法设置注入"))
				return
			}
			isInject = true
		}
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		pattern := message.UnescapeCQCodeText(matched[3])
		template := strings.TrimPrefix(ctx.MessageString(), matched[0])
		if global.group[gid] == nil {
			global.group[gid] = &regexGroup{
				Private: make(map[int64][]inst),
			}
		}
		if global.group[gid].Private == nil {
			global.group[gid].Private = make(map[int64][]inst)
		}
		compiled, err := regexp.Compile(transformPattern(pattern))
		if err != nil {
			ctx.SendChain(message.Text("ERROR:无法编译正则表达式:", err))
			return
		}
		regexInst := inst{
			regex:    compiled,
			Pattern:  pattern,
			Template: template,
			IsInject: isInject,
		}
		rg := global.group[gid]
		if all {
			if isInject {
				err = saveInjectRegex(gid, 0, strconv.FormatInt(ctx.Event.SelfID, 36), pattern, template)
			} else {
				err = saveRegex(gid, 0, strconv.FormatInt(ctx.Event.SelfID, 36), pattern, template)
			}
			if err == nil {
				rg.All = append(rg.All, regexInst)
			}
		} else {
			if isInject {
				err = saveInjectRegex(gid, uid, strconv.FormatInt(ctx.Event.SelfID, 36), pattern, template)
			} else {
				err = saveRegex(gid, uid, strconv.FormatInt(ctx.Event.SelfID, 36), pattern, template)
			}
			if err == nil {
				rg.Private[uid] = append(rg.Private[uid], regexInst)
			}
		}
		if err != nil {
			ctx.SendChain(message.Text("ERROR:无法保存正则表达式:", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})

	en.OnRegex(`^(查看|看看)(我|大家|有人)(说|问)`, zero.OnlyGroup, zero.OnlyToMe).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		mu.RLock()
		defer mu.RUnlock()

		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		matched := ctx.State["regex_matched"].([]string)
		all := true
		if matched[2] == "我" {
			all = false
		}
		arg := strings.TrimPrefix(ctx.MessageString(), matched[0])
		rg := global.group[gid]
		if rg == nil {
			return
		}

		w := binary.SelectWriter()
		defer binary.PutWriter(w)
		if all {
			w.WriteString("该群设置的“有人问”有：\n")
		} else {
			_, _ = fmt.Fprintf(w, "你在该群设置的含有 %s 的问题有：\n", arg)
		}
		show := func(insts []inst) {
			for i := range insts {
				if strings.Contains(insts[i].Pattern, arg) {
					w.WriteString(strings.Trim(insts[i].Pattern, "^$"))
					if insts[i].IsInject {
						w.WriteString("(做)")
					}
					_ = w.WriteByte('\n')
				}
			}
		}

		if all {
			show(rg.All)
		} else {
			show(rg.Private[uid])
		}
		ctx.SendChain(message.Text(w.String()))
	})

	en.OnRegex(`^删除(大家|有人|我)(说|问|让你做|让你执行)`, zero.OnlyGroup, zero.OnlyToMe).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		mu.Lock()
		defer mu.Unlock()

		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		matched := ctx.State["regex_matched"].([]string)
		pattern := strings.TrimPrefix(ctx.MessageString(), matched[0])
		escapedpattern := message.UnescapeCQCodeText(pattern)
		if pattern == escapedpattern {
			escapedpattern = ""
		}
		rg := global.group[gid]
		if rg == nil {
			return
		}
		all := true
		if matched[1] == "我" {
			all = false
		}
		if all && !zero.AdminPermission(ctx) {
			ctx.SendChain(message.Text("非管理员/主人无法删除全局问答"))
			return
		}
		isInject := false
		if matched[2] == "让你做" || matched[2] == "让你执行" {
			if !zero.AdminPermission(ctx) {
				ctx.SendChain(message.Text("非管理员/主人无法删除注入"))
				return
			}
			isInject = true
		}
		var deleteInst func(insts []inst) ([]inst, error)
		if escapedpattern == "" {
			deleteInst = func(insts []inst) ([]inst, error) {
				for i := range insts {
					if insts[i].Pattern == pattern {
						insts[i] = insts[len(insts)-1]
						insts = insts[:len(insts)-1]
						return insts, nil
					}
				}
				return insts, errors.New("没有找到对应的问答词条")
			}
		} else {
			deleteInst = func(insts []inst) ([]inst, error) {
				for i := range insts {
					if insts[i].Pattern == pattern || insts[i].Pattern == escapedpattern {
						insts[i] = insts[len(insts)-1]
						insts = insts[:len(insts)-1]
						return insts, nil
					}
				}
				return insts, errors.New("没有找到对应的问答词条")
			}
		}
		removeInDB := func(f func(int64, int64, string, string) error, gid, uid int64, bots, pattern string) (err error) {
			err = f(gid, uid, bots, pattern)
			if err != nil && escapedpattern == "" {
				return
			}
			if escapedpattern != "" {
				err = f(gid, uid, bots, escapedpattern)
			}
			return
		}
		var err error
		var f func(int64, int64, string, string) error
		if isInject {
			f = removeInjectRegex
		} else {
			f = removeRegex
		}
		if matched[1] == "我" {
			err = removeInDB(f, gid, uid, strconv.FormatInt(ctx.Event.SelfID, 36), pattern)
			if err == nil {
				rg.Private[uid], err = deleteInst(rg.Private[uid])
			}
		} else {
			err = removeInDB(f, gid, 0, strconv.FormatInt(ctx.Event.SelfID, 36), pattern)
			if err == nil {
				rg.All, err = deleteInst(rg.All)
			}
		}
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("删除成功"))
	})

	en.On(`message/group`, func(ctx *zero.Ctx) bool {
		mu.RLock()
		defer mu.RUnlock()

		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		rg := global.group[gid]
		if rg == nil {
			return false
		}
		if runInsts(ctx, rg.All) {
			return true
		}
		return runInsts(ctx, rg.Private[uid])
	}).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		template := ctx.State["regqa_template"].(string)
		if ctx.State["regqa_isinject"].(bool) {
			ctx.Event.NativeMessage = json.RawMessage("\"" + template + "\"")
			ctx.Event.RawMessage = template
			process.SleepAbout1sTo2s() // 防止风控
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
		} else {
			ctx.Send(template)
		}
	})
}

func runInsts(ctx *zero.Ctx, insts []inst) bool {
	msg := ctx.MessageString()
	for _, inst := range insts {
		if matched := inst.regex.FindStringSubmatch(msg); matched != nil {
			template := inst.Template
			sub := inst.regex.SubexpNames()
			for i := 1; i < len(matched); i++ {
				if sub[i] != "" {
					template = strings.ReplaceAll(template, "<<"+sub[i]+">>", matched[i])
				}
				template = strings.ReplaceAll(template, "$"+strconv.Itoa(i), matched[i])
			}
			ctx.State["regqa_template"] = template
			ctx.State["regqa_isinject"] = inst.IsInject
			return true
		}
	}
	return false
}
