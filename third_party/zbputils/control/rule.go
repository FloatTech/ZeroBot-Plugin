// Package control 控制插件的启用与优先级等
package control

import (
	"errors"
	"os"
	"strconv"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/gg/factory"
	"github.com/FloatTech/rendercard"
	ctrl "github.com/FloatTech/zbpctrl"
)

const (
	// StorageFolder 插件控制数据目录
	StorageFolder = "data/control/"
	// Md5File ...
	Md5File = StorageFolder + "stor.spb"
	dbfile  = StorageFolder + "plugins.db"
)

var (
	// managers 每个插件对应的管理
	managers = ctrl.NewManager[*zero.Ctx](dbfile)
)

func newctrl(service string, o *ctrl.Options[*zero.Ctx]) zero.Rule {
	c := managers.NewControl(service, o)
	return func(ctx *zero.Ctx) bool {
		ctx.State["manager"] = c
		return c.Handler(ctx.Event.GroupID, ctx.Event.UserID)
	}
}

// Lookup 查找服务
func Lookup(service string) (*ctrl.Control[*zero.Ctx], bool) {
	_, ok := briefmap[service]
	if ok {
		return managers.Lookup(briefmap[service])
	}
	return managers.Lookup(service)
}

// Response 响应
func Response(gid int64) error {
	return managers.Response(gid)
}

// Silence 沉默
func Silence(gid int64) error {
	return managers.Silence(gid)
}

// CanResponse 响应状态
func CanResponse(gid int64) bool {
	return managers.CanResponse(gid)
}

// Response commands act on their current conversation and require OnlyToMe
// as well as the existing permissions. Reject prefix matches and unexpected
// arguments, while accepting trailing whitespace from QQ clients.
func responseCommandWithoutArguments(ctx *zero.Ctx) bool {
	args, ok := ctx.State["args"].(string)
	return ok && strings.TrimSpace(args) == ""
}

func responseCommandError(err error) message.Segment {
	switch {
	case errors.Is(err, ctrl.ErrAlreadyResponding):
		return message.Text("当前位置已处于响应状态。")
	case errors.Is(err, ctrl.ErrAlreadySilent):
		return message.Text("当前位置已处于沉默状态。")
	default:
		return message.Text("响应设置保存失败，原状态保留，请检查本机备份目录或数据库。")
	}
}

func init() {
	err := os.MkdirAll("data/Control", 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("data/control", 0755)
	if err != nil {
		panic(err)
	}
	zero.OnCommandGroup([]string{
		"响应", "response", "沉默", "silence",
	}, zero.UserOrGrpAdmin, zero.OnlyToMe, responseCommandWithoutArguments).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		grp := ctx.Event.GroupID
		if grp == 0 {
			// 个人用户
			grp = -ctx.Event.UserID
		}
		var msg message.Segment
		switch ctx.State["command"] {
		case "响应", "response":
			err := managers.Response(grp)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始在此工作啦~")
			} else {
				msg = responseCommandError(err)
			}
		case "沉默", "silence":
			err := managers.Silence(grp)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始休息啦~")
			} else {
				msg = responseCommandError(err)
			}
		default:
			msg = message.Text("ERROR: bad command\"", ctx.State["command"], "\"")
		}
		ctx.SendChain(msg)
	})

	zero.OnCommandGroup([]string{
		"全局响应", "allresponse", "全局沉默", "allsilence",
	}, zero.SuperUserPermission, zero.OnlyToMe, responseCommandWithoutArguments).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		var msg message.Segment
		cmd := ctx.State["command"].(string)
		switch {
		case strings.Contains(cmd, "响应") || strings.Contains(cmd, "response"):
			err := managers.Response(0)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始在全部位置工作啦~")
			} else {
				msg = responseCommandError(err)
			}
		case strings.Contains(cmd, "沉默") || strings.Contains(cmd, "silence"):
			err := managers.Silence(0)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始在未显式启用的位置休息啦~")
			} else {
				msg = responseCommandError(err)
			}
		default:
			msg = message.Text("ERROR: bad command\"", cmd, "\"")
		}
		ctx.SendChain(msg)
	})

	zero.OnCommandGroup([]string{
		"启用", "enable", "禁用", "disable",
	}, zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
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
		if strings.Contains(model.Command, "启用") || strings.Contains(model.Command, "enable") {
			if err := service.Enable(grp); err != nil {
				ctx.SendChain(message.Text("插件设置保存失败，原设置保留。"))
				return
			}
			if service.Options.OnEnable != nil {
				service.Options.OnEnable(ctx)
			} else {
				ctx.SendChain(message.Text("已启用服务: " + model.Args))
			}
		} else {
			if err := service.Disable(grp); err != nil {
				ctx.SendChain(message.Text("插件设置保存失败，原设置保留。"))
				return
			}
			if service.Options.OnDisable != nil {
				service.Options.OnDisable(ctx)
			} else {
				ctx.SendChain(message.Text("已禁用服务: " + model.Args))
			}
		}
	})

	zero.OnCommandGroup([]string{
		"此处启用所有插件", "adhocenableall", "此处禁用所有插件", "adhocdisableall",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		grp := ctx.Event.GroupID
		if grp == 0 {
			grp = -ctx.Event.UserID
		}
		condition := strings.Contains(ctx.Event.RawMessage, "启用") || strings.Contains(ctx.Event.RawMessage, "enable")
		var switchError error
		if condition {
			managers.ForEach(func(_ string, manager *ctrl.Control[*zero.Ctx]) bool {
				state, err := manager.SwitchState(grp)
				if err != nil {
					switchError = err
					return false
				}
				if !state.DefaultEnabled == condition {
					return true
				}
				switchError = manager.Enable(grp)
				if switchError != nil {
					return false
				}
				return true
			})
			if switchError != nil {
				ctx.SendChain(message.Text("批量启用未全部完成，请检查插件状态。"))
				return
			}
			ctx.SendChain(message.Text("此处启用所有插件成功"))
		} else {
			managers.ForEach(func(_ string, manager *ctrl.Control[*zero.Ctx]) bool {
				switchError = manager.Disable(grp)
				if switchError != nil {
					return false
				}
				return true
			})
			if switchError != nil {
				ctx.SendChain(message.Text("批量禁用未全部完成，请检查插件状态。"))
				return
			}
			ctx.SendChain(message.Text("此处禁用所有插件成功"))
		}
	})

	zero.OnCommandGroup([]string{
		"全局启用", "allenable", "全局禁用", "alldisable",
	}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		service, ok := Lookup(model.Args)
		if !ok {
			ctx.SendChain(message.Text("没有找到指定服务!"))
			return
		}
		if strings.Contains(model.Command, "启用") || strings.Contains(model.Command, "enable") {
			if err := service.Enable(0); err != nil {
				ctx.SendChain(message.Text("插件设置保存失败，原设置保留。"))
				return
			}
			ctx.SendChain(message.Text("已全局启用服务: " + model.Args))
		} else {
			if err := service.Disable(0); err != nil {
				ctx.SendChain(message.Text("插件设置保存失败，原设置保留。"))
				return
			}
			ctx.SendChain(message.Text("已全局禁用服务: " + model.Args))
		}
	})

	zero.OnCommandGroup([]string{"还原", "reset"}, zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
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
		if err := service.Reset(grp); err != nil {
			ctx.SendChain(message.Text("插件设置保存失败，原设置保留。"))
			return
		}
		ctx.SendChain(message.Text("已还原服务的默认启用状态: " + model.Args))
	})

	zero.OnCommandGroup([]string{
		"禁止", "ban", "允许", "permit",
	}, zero.AdminPermission, func(ctx *zero.Ctx) bool {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		args := strings.Split(model.Args, " ")
		if len(args) < 2 {
			ctx.SendChain(message.Text("参数错误!"))
			ctx.Break()
			return false
		}
		argsparsed := make([]int64, 0, len(args))
		var uid int64
		var err error
		haspermission := zero.GroupHigherPermission(func(_ *zero.Ctx) int64 { return uid })
		for _, usr := range args[1:] {
			uid, err = strconv.ParseInt(usr, 10, 64)
			if err == nil && haspermission(ctx) {
				argsparsed = append(argsparsed, uid)
			}
		}
		if len(argsparsed) == 0 {
			ctx.SendChain(message.Text("无权操作!"))
			ctx.Break()
			return false
		}
		ctx.State["__command__"] = model.Command
		ctx.State["__servicename__"] = args[0]
		ctx.State["__args__"] = argsparsed
		return true
	}, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		command := ctx.State["__command__"].(string)
		servicename := ctx.State["__servicename__"].(string)
		args := ctx.State["__args__"].([]int64)
		service, ok := Lookup(servicename)
		if !ok {
			ctx.SendChain(message.Text("没有找到指定服务!"))
			return
		}
		grp := ctx.Event.GroupID
		if grp == 0 {
			grp = -ctx.Event.UserID
		}
		msg := "**" + servicename + "报告**"
		var members map[int64]struct{}
		issu := zero.SuperUserPermission(ctx)
		if !issu {
			lst := ctx.GetGroupMemberList(ctx.Event.GroupID).Array()
			members = make(map[int64]struct{}, len(lst))
			for _, m := range lst {
				members[m.Get("user_id").Int()] = struct{}{}
			}
		}
		permit := strings.Contains(command, "允许") || strings.Contains(command, "permit")
		for _, uid := range args {
			usr := strconv.FormatInt(uid, 10)
			if _, present := members[uid]; !issu && !present {
				msg += "\nx " + usr + " 不在本群"
				continue
			}
			var err error
			if permit {
				err = service.Permit(uid, grp)
			} else {
				err = service.Ban(uid, grp)
			}
			if err != nil {
				msg += "\nx " + usr + " 操作失败，原状态保留。"
			} else if permit {
				msg += "\n+ 已允许" + usr
			} else {
				msg += "\n- 已禁止" + usr
			}
		}
		ctx.SendChain(message.Text(msg))
	})

	zero.OnCommandGroup([]string{
		"全局禁止", "allban", "全局允许", "allpermit",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		args := strings.Split(model.Args, " ")
		if len(args) >= 2 {
			service, ok := Lookup(args[0])
			if !ok {
				ctx.SendChain(message.Text("没有找到指定服务!"))
				return
			}
			msg := "**" + args[0] + "全局报告**"
			if strings.Contains(model.Command, "允许") || strings.Contains(model.Command, "permit") {
				for _, usr := range args[1:] {
					uid, err := strconv.ParseInt(usr, 10, 64)
					if err == nil {
						if err := service.Permit(uid, 0); err != nil {
							msg += "\nx " + usr + " 操作失败，原状态保留。"
						} else {
							msg += "\n+ 已允许" + usr
						}
					}
				}
			} else {
				for _, usr := range args[1:] {
					uid, err := strconv.ParseInt(usr, 10, 64)
					if err == nil {
						if err := service.Ban(uid, 0); err != nil {
							msg += "\nx " + usr + " 操作失败，原状态保留。"
						} else {
							msg += "\n- 已禁止" + usr
						}
					}
				}
			}
			ctx.SendChain(message.Text(msg))
			return
		}
		ctx.SendChain(message.Text("参数错误!"))
	})

	zero.OnCommandGroup([]string{
		"封禁", "block", "解封", "unblock",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		args := strings.Split(model.Args, " ")
		if len(args) >= 1 {
			msg := "**报告**"
			if strings.Contains(model.Command, "解") || strings.Contains(model.Command, "un") {
				for _, usr := range args {
					uid, err := strconv.ParseInt(usr, 10, 64)
					if err == nil {
						if managers.DoUnblock(uid) == nil {
							msg += "\n- 已解封" + usr
						} else {
							msg += "\nx " + usr + " 操作失败，原状态保留。"
						}
					}
				}
			} else {
				for _, usr := range args {
					uid, err := strconv.ParseInt(usr, 10, 64)
					if err == nil {
						if managers.DoBlock(uid) == nil {
							msg += "\n+ 已封禁" + usr
						} else {
							msg += "\nx " + usr + " 操作失败，原状态保留。"
						}
					}
				}
			}
			ctx.SendChain(message.Text(msg))
			return
		}
		ctx.SendChain(message.Text("参数错误!"))
	})

	zero.OnCommandGroup([]string{
		"改变默认启用状态", "allflip",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		service, ok := Lookup(model.Args)
		if !ok {
			ctx.SendChain(message.Text("没有找到指定服务!"))
			return
		}
		err := service.Flip()
		if err != nil {
			ctx.SendChain(message.Text("默认状态保存失败，原设置保留。"))
			return
		}
		ctx.SendChain(message.Text("已改变全局默认启用状态: " + model.Args))
	})

	zero.OnCommandGroup([]string{"用法", "usage"}, zero.OnlyToMe).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			model := extension.CommandModel{}
			_ = ctx.Parse(&model)
			service, ok := Lookup(model.Args)
			if !ok {
				ctx.SendChain(message.Text("没有找到指定服务!"))
				return
			}
			if service.Options.Help == "" {
				ctx.SendChain(message.Text("该服务无帮助!"))
				return
			}
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			// 处理插件帮助并且计算图像高
			plugininfo := strings.Split(strings.Trim(service.String(), "\n"), "\n")
			newplugininfo, err := rendercard.Truncate(glowsd, plugininfo, 1272-50, 38)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			imgs, err := (&rendercard.Title{
				LeftTitle:     service.Service,
				LeftSubtitle:  service.Options.Brief,
				RightTitle:    "FloatTech",
				RightSubtitle: "ZeroBot-Plugin",
				ImagePath:     kanbanpath + "kanban.png",
				TitleFontData: impactd,
				TextFontData:  glowsd,
				IsEnabled:     service.IsEnabledIn(gid),
			}).DrawTitleWithText(newplugininfo)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			data, err := factory.ToBytes(imgs) // 生成图片
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	zero.OnCommandGroup([]string{"服务列表", "service_list"}, zero.OnlyToMe).SetBlock(true).SecondPriority().
		Handle(sendServiceList)
}
