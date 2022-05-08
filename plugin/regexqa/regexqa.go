package regexqa

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/kv"
)

var global = context{
	group: make(map[int64]*regexGroup),
}

type context struct {
	mu    sync.RWMutex
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
}

var transformRegex = regexp.MustCompile(`<<.+?>>`)

func transformPattern(pattern string) string {
	pattern = transformRegex.ReplaceAllStringFunc(pattern, func(s string) string {
		s = strings.Trim(s, "<>")
		return `(?P<` + s + `>.+?)`
	})
	return "^" + pattern + "$"
}

func saveRegex(ctx *zero.Ctx) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(&global.group)
	if err != nil {
		ctx.Send("无法保存正则表达式")
		return
	}
	err = bucket.Put([]byte("global"), buf.Bytes())
	if err != nil {
		ctx.Send("无法保存正则表达式")
	}
}

var bucket = kv.New("regexqa")

func initRegex() {
	got, err := bucket.Get([]byte("global"))
	if err == nil {
		gob.NewDecoder(bytes.NewReader(got)).Decode(&global.group)
		for _, v := range global.group {
			for i := range v.All {
				v.All[i].regex = regexp.MustCompile(transformPattern(v.All[i].Pattern))
			}
			for _, insts := range v.Private {
				for j := range insts {
					insts[j].regex = regexp.MustCompile(transformPattern(insts[j].Pattern))
				}
			}
		}
	}
}

func init() {
	en := control.Register("regexqa", &control.Options{
		DisableOnDefault: true,
		Help:             "",
	})
	initRegex()
	en.OnRegex(`^(我|大家|有人)(说|问)(.*)你(答|说)`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			global.mu.Lock()
			defer global.mu.Unlock()

			matched := ctx.State["regex_matched"].([]string)
			all := true
			if matched[1] == "我" {
				all = false
			}
			if all && ctx.Event.Sender.Role == "member" {
				ctx.Send("非管理员无法设置全局问答")
				return
			}
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			pattern := matched[3]
			template := strings.TrimPrefix(ctx.MessageString(), matched[0])
			if global.group[gid] == nil {
				global.group[gid] = &regexGroup{
					Private: make(map[int64][]inst),
				}
			}
			compiled, err := regexp.Compile(transformPattern(pattern))
			if err != nil {
				ctx.Send("无法编译正则表达式")
				return
			}
			regexInst := inst{
				regex:    compiled,
				Pattern:  pattern,
				Template: template,
			}
			rg := global.group[gid]
			if all {
				rg.All = append(rg.All, regexInst)
			} else {
				rg.Private[uid] = append(rg.Private[uid], regexInst)
			}
			saveRegex(ctx)
		})

	en.OnRegex(`^(查看|看看)(我|大家|有人)(说|问)`).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			global.mu.RLock()
			defer global.mu.RUnlock()

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

			var writer bytes.Buffer
			if all {
				writer.WriteString("该群设置的“有人问”有：")
			} else {
				writer.WriteString(fmt.Sprintf("你在该群设置的含有 %s 的问题有：", arg))
			}
			show := func(insts []inst) []inst {
				for i := range insts {
					if strings.Contains(insts[i].Pattern, arg) {
						writer.WriteString(strings.Trim(insts[i].Pattern, "^$"))
						writer.WriteByte('\n')
					}
				}
				return insts
			}

			if all {
				show(rg.All)
			} else {
				show(rg.Private[uid])
			}
			ctx.Send(writer.String())
			saveRegex(ctx)
		})

	en.OnRegex(`^删除(大家|有人|我)(说|问)`).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			global.mu.Lock()
			defer global.mu.Unlock()

			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			matched := ctx.State["regex_matched"].([]string)
			pattern := strings.TrimPrefix(ctx.MessageString(), matched[0])
			rg := global.group[gid]
			if rg == nil {
				return
			}
			all := true
			if matched[1] == "我" {
				all = false
			}
			if all && ctx.Event.Sender.Role == "member" {
				ctx.Send("非管理员无法删除全局问答")
				return
			}
			deleteInst := func(insts []inst) ([]inst, bool) {
				for i := range insts {
					if insts[i].Pattern == pattern {
						insts[i] = insts[len(insts)-1]
						insts = insts[:len(insts)-1]
						return insts, true
					}
				}
				return insts, false
			}
			var ok bool
			if matched[1] == "我" {
				rg.Private[uid], ok = deleteInst(rg.Private[uid])
			} else {
				rg.All, ok = deleteInst(rg.All)
			}
			if ok {
				ctx.Send("删除成功")
				saveRegex(ctx)
			} else {
				ctx.Send("没有找到对应的问答词条")
			}
		})

	en.OnMessage().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.GroupID != 0 {
				return
			}
			global.mu.RLock()
			defer global.mu.RUnlock()

			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			rg := global.group[gid]
			if rg == nil {
				return
			}
			if runInsts(ctx, rg.All) {
				return
			}
			runInsts(ctx, rg.Private[uid])
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
			ctx.Send(template)
			return true
		}
	}
	return false
}
