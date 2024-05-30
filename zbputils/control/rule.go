// Package control 控制插件的启用与优先级等
package control

import (
	"encoding/base64"
	"image"
	"image/jpeg"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/rendercard"
	ctrl "github.com/FloatTech/zbpctrl"

	"github.com/FloatTech/zbputils/ctxext"
)

const (
	// StorageFolder 插件控制数据目录
	StorageFolder = "data/control/"
	// Md5File ...
	Md5File = StorageFolder + "stor.spb"
	dbfile  = StorageFolder + "plugins.db"
	lnfile  = StorageFolder + "lnperpg.txt"
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

func init() {
	err := os.MkdirAll("data/Control", 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("data/control", 0755)
	if err != nil {
		panic(err)
	}
	// 载入用户配置
	if file.IsExist(lnfile) {
		data, err := os.ReadFile(lnfile)
		if err != nil {
			logrus.Warnln("[control] 读取配置文件失败,将使用默认的显示行数:", err)
		} else {
			mun, err := strconv.Atoi(binary.BytesToString(data))
			if err != nil {
				logrus.Warnln("[control] 获取设置的服务列表显示行数错误,将使用默认的显示行数:", err)
			} else if mun > 0 {
				lnperpg = mun
				logrus.Infoln("[control] 获取到当前设置的服务列表显示行数为:", lnperpg)
			}
		}
	}
	zero.OnCommandGroup([]string{
		"响应", "response", "沉默", "silence",
	}, zero.UserOrGrpAdmin, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		grp := ctx.Event.GroupID
		if grp == 0 {
			// 个人用户
			grp = -ctx.Event.UserID
		}
		var msg message.MessageSegment
		switch ctx.State["command"] {
		case "响应", "response":
			err := managers.Response(grp)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始在此工作啦~")
			} else {
				msg = message.Text("ERROR: ", err)
			}
		case "沉默", "silence":
			err := managers.Silence(grp)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始休息啦~")
			} else {
				msg = message.Text("ERROR: ", err)
			}
		default:
			msg = message.Text("ERROR: bad command\"", ctx.State["command"], "\"")
		}
		ctx.SendChain(msg)
	})

	zero.OnCommandGroup([]string{
		"全局响应", "allresponse", "全局沉默", "allsilence",
	}, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		var msg message.MessageSegment
		cmd := ctx.State["command"].(string)
		switch {
		case strings.Contains(cmd, "响应") || strings.Contains(cmd, "response"):
			err := managers.Response(0)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始在全部位置工作啦~")
			} else {
				msg = message.Text("ERROR: ", err)
			}
		case strings.Contains(cmd, "沉默") || strings.Contains(cmd, "silence"):
			err := managers.Silence(0)
			if err == nil {
				msg = message.Text(zero.BotConfig.NickName[0], "将开始在未显式启用的位置休息啦~")
			} else {
				msg = message.Text("ERROR: ", err)
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
			service.Enable(grp)
			if service.Options.OnEnable != nil {
				service.Options.OnEnable(ctx)
			} else {
				ctx.SendChain(message.Text("已启用服务: " + model.Args))
			}
		} else {
			service.Disable(grp)
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
		if condition {
			managers.ForEach(func(key string, manager *ctrl.Control[*zero.Ctx]) bool {
				if manager.Options.DisableOnDefault == condition {
					return true
				}
				manager.Enable(grp)
				return true
			})
			ctx.SendChain(message.Text("此处启用所有插件成功"))
		} else {
			managers.ForEach(func(key string, manager *ctrl.Control[*zero.Ctx]) bool {
				manager.Disable(grp)
				return true
			})
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
			service.Enable(0)
			ctx.SendChain(message.Text("已全局启用服务: " + model.Args))
		} else {
			service.Disable(0)
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
		service.Reset(grp)
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
		haspermission := zero.GroupHigherPermission(func(ctx *zero.Ctx) int64 { return uid })
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
		if strings.Contains(command, "允许") || strings.Contains(command, "permit") {
			for _, uid := range args {
				usr := strconv.FormatInt(uid, 10)
				if issu {
					service.Permit(uid, grp)
					msg += "\n+ 已允许" + usr
				} else {
					_, ok := members[uid]
					if ok {
						service.Permit(uid, grp)
						msg += "\n+ 已允许" + usr
					} else {
						msg += "\nx " + usr + " 不在本群"
					}
				}
			}
		} else {
			for _, uid := range args {
				usr := strconv.FormatInt(uid, 10)
				if issu {
					service.Ban(uid, grp)
					msg += "\n- 已禁止" + usr
				} else {
					_, ok := members[uid]
					if ok {
						service.Ban(uid, grp)
						msg += "\n- 已禁止" + usr
					} else {
						msg += "\nx " + usr + " 不在本群"
					}
				}
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
						service.Permit(uid, 0)
						msg += "\n+ 已允许" + usr
					}
				}
			} else {
				for _, usr := range args[1:] {
					uid, err := strconv.ParseInt(usr, 10, 64)
					if err == nil {
						service.Ban(uid, 0)
						msg += "\n- 已禁止" + usr
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
						}
					}
				}
			} else {
				for _, usr := range args {
					uid, err := strconv.ParseInt(usr, 10, 64)
					if err == nil {
						if managers.DoBlock(uid) == nil {
							msg += "\n+ 已封禁" + usr
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
			ctx.SendChain(message.Text("ERROR: ", err))
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
			data, err := imgfactory.ToBytes(imgs) // 生成图片
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	zero.OnCommandGroup([]string{"服务列表", "service_list"}, zero.OnlyToMe).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			var imgs []image.Image
			imgs, err = drawservicesof(gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(imgs) > 1 {
				wg := sync.WaitGroup{}
				msg := make(message.Message, len(imgs))
				wg.Add(len(imgs))
				for i := 0; i < len(imgs); i++ {
					go func(i int) {
						defer wg.Done()
						msg[i] = ctxext.FakeSenderForwardNode(ctx, message.Image(binary.BytesToString(binary.NewWriterF(func(w *binary.Writer) {
							w.WriteString("base64://")
							encoder := base64.NewEncoder(base64.StdEncoding, w)
							var opt jpeg.Options
							opt.Quality = 70
							if err1 := jpeg.Encode(encoder, imgs[i], &opt); err1 != nil {
								err = err1
								return
							}
							_ = encoder.Close()
						}))))
					}(i)
				}
				wg.Wait()
				if id := ctx.Send(msg); id.ID() == 0 {
					ctx.SendChain(message.Text("ERROR: 可能被风控了"))
				}
			} else {
				b64, err := imgfactory.ToBase64(imgs[0])
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(b64))); id.ID() == 0 {
					ctx.SendChain(message.Text("ERROR: 可能被风控了"))
				}
			}
		})

	zero.OnCommand("设置服务列表显示行数", zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).SecondPriority().Handle(func(ctx *zero.Ctx) {
		model := extension.CommandModel{}
		_ = ctx.Parse(&model)
		mun, err := strconv.Atoi(model.Args)
		if err != nil {
			ctx.SendChain(message.Text("请输入正确的数字"))
			return
		}
		err = os.WriteFile(lnfile, binary.StringToBytes(model.Args), 0644)
		if err != nil {
			ctx.SendChain(message.Text(err))
			return
		}
		lnperpg = mun
		// 清除缓存
		titlecache = nil
		fullpageshadowcache = nil
		ctx.SendChain(message.Text("已设置列表单页显示数为 " + strconv.Itoa(lnperpg)))
	})
}
