package control

import (
	"fmt"
	"os"
	"strconv"
	"unicode"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/floatbox/file"
)

// Engine is the pre_handler, post_handler manager
type Engine struct {
	en         *zero.Engine
	prio       int
	service    string
	datafolder string
}

var priomap = make(map[int]string)      // priomap is map[prio]service
var briefmap = make(map[string]string)  // briefmap is map[brief]service
var foldermap = make(map[string]string) // foldermap is map[folder]service
var extramap = make(map[int16]string)   // extramap is map[gid]service

func newengine(service string, prio int, o *ctrl.Options[*zero.Ctx]) (e *Engine) {
	e = new(Engine)
	s, ok := priomap[prio]
	if ok && s != service {
		panic(fmt.Sprint("prio ", prio, " is used by ", s, " but ", service, " want to register"))
	}
	priomap[prio] = service
	e.en = zero.New()
	e.en.UsePreHandler(
		func(ctx *zero.Ctx) bool {
			// 防止自触发
			return ctx.Event.UserID != ctx.Event.SelfID || ctx.Event.PostType != "message"
		},
		newctrl(service, o),
		addsyncxstate,
	)
	e.prio = prio
	e.service = service
	if o.Brief != "" {
		s, ok := briefmap[o.Brief]
		if ok {
			panic("Brief \"" + o.Brief + "\" of service " + service + " has been required by service " + s)
		}
		briefmap[o.Brief] = service
	}
	if o.Extra != 0 {
		s, ok := extramap[o.Extra]
		if ok {
			logrus.Warnln("[control] extra " + strconv.Itoa(int(o.Extra)) + " of service " + service + " has been required by service " + s)
		}
		extramap[o.Extra] = service
	}
	switch {
	case o.PublicDataFolder != "":
		if unicode.IsLower([]rune(o.PublicDataFolder)[0]) {
			panic("public data folder " + o.PublicDataFolder + " must start with an upper case letter")
		}
		e.datafolder = "data/" + o.PublicDataFolder + "/"
	case o.PrivateDataFolder != "":
		if unicode.IsUpper([]rune(o.PrivateDataFolder)[0]) {
			panic("private data folder " + o.PrivateDataFolder + " must start with an lower case letter")
		}
		e.datafolder = "data/" + o.PrivateDataFolder + "/"
	default:
		e.datafolder = "data/zbp/"
	}
	if e.datafolder != "data/zbp/" {
		s, ok := foldermap[e.datafolder]
		if ok {
			panic("folder " + e.datafolder + " has been required by service " + s)
		}
		foldermap[e.datafolder] = service
	}
	if file.IsNotExist(e.datafolder) {
		err := os.MkdirAll(e.datafolder, 0755)
		if err != nil {
			panic(err)
		}
	}
	logrus.Debugln("[control] 插件", service, "已设置数据目录", e.datafolder)
	return
}

// DataFolder 本插件数据目录, 默认 data/zbp/
func (e *Engine) DataFolder() string {
	return e.datafolder
}

// IsEnabledIn 自己是否在 id (正群负个人零全局) 启用
func (e *Engine) IsEnabledIn(id int64) bool {
	c, ok := managers.Lookup(e.service)
	if !ok {
		return false
	}
	return c.IsEnabledIn(id)
}

// Delete 移除该 Engine 注册的所有 Matchers
func (e *Engine) Delete() {
	e.en.Delete()
}

// UsePreHandler 向该 Engine 添加新 PreHandler(Rule), 会在 Rule 判断前触发，如果 preHandler 没有通过，则 Rule, Matcher 不会触发
// 可用于分群组管理插件等
func (e *Engine) UsePreHandler(rules ...zero.Rule) {
	e.en.UsePreHandler(rules...)
}

// UseMidHandler 向该 Engine 添加新 MidHandler(Rule), 会在 Rule 判断后， Matcher 触发前触发，如果 midHandler 没有通过，则 Matcher 不会触发
// 可用于速率限制等
func (e *Engine) UseMidHandler(rules ...zero.Rule) {
	e.en.UseMidHandler(rules...)
}

// UsePostHandler 向该 Engine 添加新 PostHandler(Rule), 会在 Matcher 触发后触发，如果 PostHandler 返回 false, 则后续的 post handler 不会触发
// 可用于反并发等
func (e *Engine) UsePostHandler(handler ...zero.Handler) {
	e.en.UsePostHandler(handler...)
}

// On 添加新的指定消息类型的匹配器
func (e *Engine) On(typ string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.On(typ, rules...).SetPriority(e.prio))
}

// OnMessage 消息触发器
func (e *Engine) OnMessage(rules ...zero.Rule) *Matcher { return e.On("message", rules...) }

// OnNotice 系统提示触发器
func (e *Engine) OnNotice(rules ...zero.Rule) *Matcher { return e.On("notice", rules...) }

// OnRequest 请求消息触发器
func (e *Engine) OnRequest(rules ...zero.Rule) *Matcher { return e.On("request", rules...) }

// OnMetaEvent 元事件触发器
func (e *Engine) OnMetaEvent(rules ...zero.Rule) *Matcher { return e.On("meta_event", rules...) }

// OnPrefix 前缀触发器
func (e *Engine) OnPrefix(prefix string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnPrefix(prefix, rules...).SetPriority(e.prio))
}

// OnSuffix 后缀触发器
func (e *Engine) OnSuffix(suffix string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnSuffix(suffix, rules...).SetPriority(e.prio))
}

// OnCommand 命令触发器
func (e *Engine) OnCommand(commands string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnCommand(commands, rules...).SetPriority(e.prio))
}

// OnRegex 正则触发器
func (e *Engine) OnRegex(regexPattern string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnRegex(regexPattern, rules...).SetPriority(e.prio))
}

// OnKeyword 关键词触发器
func (e *Engine) OnKeyword(keyword string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnKeyword(keyword, rules...).SetPriority(e.prio))
}

// OnFullMatch 完全匹配触发器
func (e *Engine) OnFullMatch(src string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnFullMatch(src, rules...).SetPriority(e.prio))
}

// OnFullMatchGroup 完全匹配触发器组
func (e *Engine) OnFullMatchGroup(src []string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnFullMatchGroup(src, rules...).SetPriority(e.prio))
}

// OnKeywordGroup 关键词触发器组
func (e *Engine) OnKeywordGroup(keywords []string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnKeywordGroup(keywords, rules...).SetPriority(e.prio))
}

// OnCommandGroup 命令触发器组
func (e *Engine) OnCommandGroup(commands []string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnCommandGroup(commands, rules...).SetPriority(e.prio))
}

// OnPrefixGroup 前缀触发器组
func (e *Engine) OnPrefixGroup(prefix []string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnPrefixGroup(prefix, rules...).SetPriority(e.prio))
}

// OnSuffixGroup 后缀触发器组
func (e *Engine) OnSuffixGroup(suffix []string, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnSuffixGroup(suffix, rules...).SetPriority(e.prio))
}

// OnShell shell命令触发器
func (e *Engine) OnShell(command string, model any, rules ...zero.Rule) *Matcher {
	return (*Matcher)(e.en.OnShell(command, model, rules...).SetPriority(e.prio))
}
