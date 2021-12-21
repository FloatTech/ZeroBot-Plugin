// Package control 控制插件的启用与优先级等
package control

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

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
	err = db.Create(service+"ban", &ban{})
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

// Ban 禁止某人在某群使用本插件
func (m *Control) Ban(uid, gid int64) {
	var err error
	var digest [16]byte
	logrus.Debugln("[control] Ban recv gid =", gid, "uid =", uid)
	if gid != 0 { // 特定群
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("%d_%d", uid, gid)))
		m.RLock()
		err = db.Insert(m.service+"ban", &ban{ID: int64(binary.LittleEndian.Uint64(digest[:8])), UserID: uid, GroupID: gid})
		m.RUnlock()
		if err == nil {
			logrus.Debugf("[control] plugin %s is banned in grp %d for usr %d.", m.service, gid, uid)
			return
		}
	}
	// 所有群
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("%d_all", uid)))
	m.RLock()
	err = db.Insert(m.service+"ban", &ban{ID: int64(binary.LittleEndian.Uint64(digest[:8])), UserID: uid, GroupID: 0})
	m.RUnlock()
	if err == nil {
		logrus.Debugf("[control] plugin %s is banned in all grp for usr %d.", m.service, uid)
	}
}

// Permit 允许某人在某群使用本插件
func (m *Control) Permit(uid, gid int64) {
	var digest [16]byte
	logrus.Debugln("[control] Permit recv gid =", gid, "uid =", uid)
	if gid != 0 { // 特定群
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("%d_%d", uid, gid)))
		m.RLock()
		_ = db.Del(m.service+"ban", "WHERE id = "+strconv.FormatInt(int64(binary.LittleEndian.Uint64(digest[:8])), 10))
		m.RUnlock()
		logrus.Debugf("[control] plugin %s is permitted in grp %d for usr %d.", m.service, gid, uid)
		return
	}
	// 所有群
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("%d_all", uid)))
	m.RLock()
	_ = db.Del(m.service+"ban", "WHERE id = "+strconv.FormatInt(int64(binary.LittleEndian.Uint64(digest[:8])), 10))
	m.RUnlock()
	logrus.Debugf("[control] plugin %s is permitted in all grp for usr %d.", m.service, uid)
}

// IsBannedIn 某人是否在某群被 ban
func (m *Control) IsBannedIn(uid, gid int64) bool {
	var b ban
	var err error
	var digest [16]byte
	logrus.Debugln("[control] IsBannedIn recv gid =", gid, "uid =", uid)
	if gid != 0 {
		digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("%d_%d", uid, gid)))
		m.RLock()
		err = db.Find(m.service+"ban", &b, "WHERE id = "+strconv.FormatInt(int64(binary.LittleEndian.Uint64(digest[:8])), 10))
		m.RUnlock()
		if err == nil && gid == b.GroupID && uid == b.UserID {
			logrus.Debugf("[control] plugin %s is banned in grp %d for usr %d.", m.service, b.GroupID, b.UserID)
			return true
		}
	}
	digest = md5.Sum(helper.StringToBytes(fmt.Sprintf("%d_all", uid)))
	m.RLock()
	err = db.Find(m.service+"ban", &b, "WHERE id = "+strconv.FormatInt(int64(binary.LittleEndian.Uint64(digest[:8])), 10))
	m.RUnlock()
	if err == nil && b.GroupID == 0 && uid == b.UserID {
		logrus.Debugf("[control] plugin %s is banned in all grp for usr %d.", m.service, b.UserID)
		return true
	}
	return false
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
func (m *Control) Handler(ctx *zero.Ctx) bool {
	ctx.State["manager"] = m
	grp := ctx.Event.GroupID
	if grp == 0 {
		// 个人用户
		return m.IsEnabledIn(-ctx.Event.UserID)
	}
	logrus.Debugln("[control] handler get gid =", grp)
	return m.IsEnabledIn(grp) && !m.IsBannedIn(ctx.Event.UserID, grp)
}

// Lookup returns a Manager by the service name, if
// not exist, it will return nil.
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

				zero.OnCommandGroup([]string{
					"禁止", "ban", "允许", "permit",
					"全局禁止", "banall", "全局允许", "permitall",
				}, zero.OnlyGroup, zero.AdminPermission).Handle(func(ctx *zero.Ctx) {
					model := extension.CommandModel{}
					_ = ctx.Parse(&model)
					args := strings.Split(model.Args, " ")
					if len(args) >= 2 {
						service, ok := Lookup(args[0])
						if !ok {
							ctx.SendChain(message.Text("没有找到指定服务!"))
							return
						}
						grp := ctx.Event.GroupID
						if strings.Contains(model.Command, "全局") || strings.Contains(model.Command, "all") {
							grp = 0
						}
						msg := "**" + args[0] + "报告**"
						if strings.Contains(model.Command, "允许") || strings.Contains(model.Command, "permit") {
							for _, usr := range args[1:] {
								uid, err := strconv.ParseInt(usr, 10, 64)
								if err == nil {
									service.Permit(uid, grp)
									msg += "\n+ 已允许" + usr
								}
							}
						} else {
							for _, usr := range args[1:] {
								uid, err := strconv.ParseInt(usr, 10, 64)
								if err == nil {
									service.Ban(uid, grp)
									msg += "\n- 已禁止" + usr
								}
							}
						}
						ctx.SendChain(message.Text(msg))
						return
					}
					ctx.SendChain(message.Text("参数错误!"))
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
						msg := "--------服务列表--------\n发送\"/用法 name\"查看详情"
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

				zero.OnCommandGroup([]string{"服务详情", "service_detail"}, userOrGrpAdmin, zero.OnlyGroup).
					Handle(func(ctx *zero.Ctx) {
						var m message.Message
						m = append(m,
							message.CustomNode(
								zero.BotConfig.NickName[0],
								ctx.Event.SelfID,
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
									zero.BotConfig.NickName[0],
									ctx.Event.SelfID,
									msg,
								))
							return true
						})

						if id := ctx.SendGroupForwardMessage(
							ctx.Event.GroupID,
							m,
						).Get("message_id").Int(); id == 0 {
							ctx.SendChain(message.Text("ERROR: 可能被风控了"))
						}
					})
			}
		}
		mu.Unlock()
	}
}
