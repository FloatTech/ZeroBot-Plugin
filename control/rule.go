// Package control 控制插件的启用与优先级等
package control

import (
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

var (
	db = &sql.Sqlite{DBPath: "data/control/plugins.db"}
	// managers 每个插件对应的管理
	managers = map[string]*Control{}
	mu       = sync.RWMutex{}
	hasinit  bool
)

// Control is to control the plugins.
type Control struct {
	sync.RWMutex
	service string
	options Options
}

// newctrl returns Manager with settings.
func newctrl(service string, o *Options) *Control {
	m := &Control{service: service,
		options: func() Options {
			if o == nil {
				return Options{}
			}
			return *o
		}(),
	}
	mu.Lock()
	managers[service] = m
	mu.Unlock()
	err := db.Create(service, &grpcfg{})
	if err != nil {
		panic(err)
	}
	return m
}

// Enable enables a group to pass the Manager.
// groupID == 0 (ALL) will operate on all grps.
func (m *Control) Enable(groupID int64) {
	var c grpcfg
	m.RLock()
	err := db.Find(m.service, &c, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	m.RUnlock()
	if err != nil {
		c.GroupID = groupID
	}
	c.Disable = int64(uint64(c.Disable) & 0xffffffff_fffffffe)
	m.Lock()
	err = db.Insert(m.service, &c)
	m.Unlock()
	if err != nil {
		logrus.Errorf("[control] %v", err)
	}
}

// Disable disables a group to pass the Manager.
// groupID == 0 (ALL) will operate on all grps.
func (m *Control) Disable(groupID int64) {
	var c grpcfg
	m.RLock()
	err := db.Find(m.service, &c, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	m.RUnlock()
	if err != nil {
		c.GroupID = groupID
	}
	c.Disable |= 1
	m.Lock()
	err = db.Insert(m.service, &c)
	m.Unlock()
	if err != nil {
		logrus.Errorf("[control] %v", err)
	}
}

// Reset resets the default config of a group.
// groupID == 0 (ALL) is not allowed.
func (m *Control) Reset(groupID int64) {
	if groupID != 0 {
		m.Lock()
		err := db.Del(m.service, "WHERE gid = "+strconv.FormatInt(groupID, 10))
		m.Unlock()
		if err != nil {
			logrus.Errorf("[control] %v", err)
		}
	}
}

// IsEnabledIn 开启群
func (m *Control) IsEnabledIn(gid int64) bool {
	var c grpcfg
	var err error
	logrus.Debugln("[control] IsEnabledIn recv gid =", gid)
	if gid != 0 {
		m.RLock()
		err = db.Find(m.service, &c, "WHERE gid = "+strconv.FormatInt(gid, 10))
		m.RUnlock()
		if err == nil && gid == c.GroupID {
			logrus.Debugf("[control] plugin %s of grp %d : %d", m.service, c.GroupID, c.Disable&1)
			return c.Disable&1 == 0
		}
	}
	m.RLock()
	err = db.Find(m.service, &c, "WHERE gid = 0")
	m.RUnlock()
	if err == nil && c.GroupID == 0 {
		logrus.Debugf("[control] plugin %s of all : %d", m.service, c.Disable&1)
		return c.Disable&1 == 0
	}
	return !m.options.DisableOnDefault
}

// GetData 获取某个群的 63 字节配置信息
func (m *Control) GetData(gid int64) int64 {
	var c grpcfg
	var err error
	logrus.Debugln("[control] IsEnabledIn recv gid =", gid)
	if gid != 0 {
		m.RLock()
		err = db.Find(m.service, &c, "WHERE gid = "+strconv.FormatInt(gid, 10))
		m.RUnlock()
		if err == nil && gid == c.GroupID {
			logrus.Debugf("[control] plugin %s of grp %d : %x", m.service, c.GroupID, c.Disable>>1)
			return c.Disable >> 1
		}
	}
	m.RLock()
	err = db.Find(m.service, &c, "WHERE gid = 0")
	m.RUnlock()
	if err == nil && c.GroupID == 0 {
		logrus.Debugf("[control] plugin %s of all : %x", m.service, c.Disable>>1)
		return c.Disable >> 1
	}
	return 0
}

// SetData 为某个群设置低 63 位配置数据
func (m *Control) SetData(groupID int64, data int64) error {
	var c grpcfg
	m.RLock()
	err := db.Find(m.service, &c, "WHERE gid = "+strconv.FormatInt(groupID, 10))
	m.RUnlock()
	if err != nil {
		c.GroupID = groupID
		if m.options.DisableOnDefault {
			c.Disable = 1
		}
	}
	c.Disable |= data << 1
	logrus.Debugf("[control] set plugin %s of all : %x", m.service, data)
	m.Lock()
	err = db.Insert(m.service, &c)
	m.Unlock()
	if err != nil {
		logrus.Errorf("[control] %v", err)
	}
	return err
}

// Handler 返回 预处理器
func (m *Control) Handler() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		ctx.State["manager"] = m
		grp := ctx.Event.GroupID
		if grp == 0 {
			// 个人用户
			grp = -ctx.Event.UserID
		}
		logrus.Debugln("[control] handler get gid =", grp)
		return m.IsEnabledIn(grp)
	}
}

// Lookup returns a Manager by the service name, if
// not exist, it will returns nil.
func Lookup(service string) (*Control, bool) {
	mu.RLock()
	m, ok := managers[service]
	mu.RUnlock()
	return m, ok
}

// ForEach iterates through managers.
func ForEach(iterator func(key string, manager *Control) bool) {
	mu.RLock()
	m := copyMap(managers)
	mu.RUnlock()
	for k, v := range m {
		if !iterator(k, v) {
			return
		}
	}
}

func copyMap(m map[string]*Control) map[string]*Control {
	ret := make(map[string]*Control, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func userOrGrpAdmin(ctx *zero.Ctx) bool {
	if zero.OnlyGroup(ctx) {
		return zero.AdminPermission(ctx)
	}
	return zero.OnlyToMe(ctx)
}

func init() {
	if !hasinit {
		mu.Lock()
		if !hasinit {
			err := os.MkdirAll("data/control", 0755)
			if err != nil {
				panic(err)
			} else {
				hasinit = true
				zero.OnCommandGroup([]string{
					"启用", "enable", "禁用", "disable",
					"全局启用", "enableall", "全局禁用", "disableall",
				}, userOrGrpAdmin).Handle(func(ctx *zero.Ctx) {
					model := extension.CommandModel{}
					_ = ctx.Parse(&model)
					service, ok := Lookup(model.Args)
					if !ok {
						ctx.SendChain(message.Text("没有找到指定服务!"))
						return
					}
					grp := ctx.Event.GroupID
					if grp == 0 {
						// 个人用户
						grp = -ctx.Event.UserID
					}
					if strings.Contains(model.Command, "全局") || strings.Contains(model.Command, "all") {
						grp = 0
					}
					if strings.Contains(model.Command, "启用") || strings.Contains(model.Command, "enable") {
						service.Enable(grp)
						ctx.SendChain(message.Text("已启用服务: " + model.Args))
					} else {
						service.Disable(grp)
						ctx.SendChain(message.Text("已禁用服务: " + model.Args))
					}
				})

				zero.OnCommandGroup([]string{"还原", "reset"}, userOrGrpAdmin).Handle(func(ctx *zero.Ctx) {
					model := extension.CommandModel{}
					_ = ctx.Parse(&model)
					service, ok := Lookup(model.Args)
					if !ok {
						ctx.SendChain(message.Text("没有找到指定服务!"))
						return
					}
					grp := ctx.Event.GroupID
					if grp == 0 {
						// 个人用户
						grp = -ctx.Event.UserID
					}
					service.Reset(grp)
					ctx.SendChain(message.Text("已还原服务的默认启用状态: " + model.Args))
				})

				zero.OnCommandGroup([]string{"用法", "usage"}, userOrGrpAdmin).
					Handle(func(ctx *zero.Ctx) {
						model := extension.CommandModel{}
						_ = ctx.Parse(&model)
						service, ok := Lookup(model.Args)
						if !ok {
							ctx.SendChain(message.Text("没有找到指定服务!"))
							return
						}
						if service.options.Help != "" {
							ctx.SendChain(message.Text(service.options.Help))
						} else {
							ctx.SendChain(message.Text("该服务无帮助!"))
						}
					})

				zero.OnCommandGroup([]string{"服务列表", "service_list"}, userOrGrpAdmin).
					Handle(func(ctx *zero.Ctx) {
						msg := `---服务列表---`
						i := 0
						gid := ctx.Event.GroupID
						ForEach(func(key string, manager *Control) bool {
							i++
							msg += "\n" + strconv.Itoa(i) + `: `
							if manager.IsEnabledIn(gid) {
								msg += "●" + key
							} else {
								msg += "○" + key
							}
							return true
						})
						ctx.SendChain(message.Text(msg))
					})

				zero.OnCommandGroup([]string{"服务详情", "service_detail"}, userOrGrpAdmin).
					Handle(func(ctx *zero.Ctx) {
						var m message.Message
						m = append(m,
							message.CustomNode(
								ctx.Event.Sender.NickName,
								ctx.Event.UserID,
								"---服务详情---",
							))
						i := 0
						ForEach(func(key string, manager *Control) bool {
							service, _ := Lookup(key)
							help := service.options.Help
							i++
							msg := strconv.Itoa(i) + `: `
							if manager.IsEnabledIn(ctx.Event.GroupID) {
								msg += "●" + key
							} else {
								msg += "○" + key
							}
							msg += "\n" + help
							m = append(m,
								message.CustomNode(
									ctx.Event.Sender.NickName,
									ctx.Event.UserID,
									msg,
								))
							return true
						})

						ctx.SendGroupForwardMessage(
							ctx.Event.GroupID,
							m,
						)
					})
			}
		}
		mu.Unlock()
	}
}
