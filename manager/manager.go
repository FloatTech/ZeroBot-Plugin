package manager

import (
	"strconv"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	// 菜单
	zero.OnFullMatch("群管系统", zero.AdminPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"====群管====", "\n",
				"- 禁言@QQ 1分钟", "\n",
				"- 解除禁言 @QQ", "\n",
				"- 我要自闭 1分钟", "\n",
				"- 开启全员禁言", "\n",
				"- 解除全员禁言", "\n",
				"- 升为管理@QQ", "\n",
				"- 取消管理@QQ", "\n",
				"- 修改名片@QQ XXX", "\n",
				"- 修改头衔@QQ XXX", "\n",
				"- 申请头衔 XXX", "\n",
				"- 踢出群聊@QQ", "\n",
				"- 退出群聊 1234", "\n",
				"- 群聊转发 1234 XXX", "\n",
				"- 私聊转发 0000 XXX",
			))
			return
		})
	// 升为管理
	zero.OnRegex(`^升为管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				true,
			)
			nickname := ctx.GetGroupMemberInfo( // 被升为管理的人的昵称
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text(nickname + " 升为了管理~"))
			return
		})
	// 取消管理
	zero.OnRegex(`^取消管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被取消管理的人的昵称
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text("残念~ " + nickname + " 暂时失去了管理员的资格"))
			return
		})
	// 踢出群聊
	zero.OnRegex(`^踢出群聊.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupKick(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被踢出群聊的人的昵称
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text("残念~ " + nickname + " 被放逐"))
			return
		})
	// 退出群聊
	zero.OnRegex(`^退出群聊.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupLeave(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要退出的群的群号
				true,
			)
			return
		})
	// 开启全体禁言
	zero.OnRegex(`^开启全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始~"))
			return
		})
	// 解除全员禁言
	zero.OnRegex(`^解除全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束~"))
			return
		})
	// 禁言
	zero.OnRegex(`^禁言.*?(\d+).*?\s(\d+)(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			duration := strToInt(ctx.State["regex_matched"].([]string)[2])
			switch ctx.State["regex_matched"].([]string)[3] {
			case "分钟":
				//
			case "小时":
				duration = duration * 60
			case "天":
				duration = duration * 60 * 24
			default:
				//
			}
			if duration >= 43200 {
				duration = 43199 // qq禁言最大时长为一个月
			}
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要禁言的人的qq
				duration*60, // 要禁言的时间（分钟）
			)
			ctx.SendChain(message.Text("小黑屋收留成功~"))
			return
		})
	// 解除禁言
	zero.OnRegex(`^解除禁言.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 要解除禁言的人的qq
				0,
			)
			ctx.SendChain(message.Text("小黑屋释放成功~"))
			return
		})
	// 自闭禁言
	zero.OnRegex(`^我要自闭.*?(\d+)(.*)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			duration := strToInt(ctx.State["regex_matched"].([]string)[1])
			switch ctx.State["regex_matched"].([]string)[2] {
			case "分钟":
				//
			case "小时":
				duration = duration * 60
			case "天":
				duration = duration * 60 * 24
			default:
				//
			}
			if duration >= 43200 {
				duration = 43199 // qq禁言最大时长为一个月
			}
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				ctx.Event.UserID,
				duration*60, // 要自闭的时间（分钟）
			)
			ctx.SendChain(message.Text("那我就不手下留情了~"))
			return
		})
	// 修改名片
	zero.OnRegex(`^修改名片.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupCard(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被修改群名片的人
				ctx.State["regex_matched"].([]string)[2],           // 修改成的群名片
			)
			ctx.SendChain(message.Text("嗯！已经修改了"))
			return
		})
	// 修改头衔
	zero.OnRegex(`^修改头衔.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[2],           // 修改成的群头衔
			)
			ctx.SendChain(message.Text("嗯！已经修改了"))
			return
		})
	// 申请头衔
	zero.OnRegex(`^申请头衔(.*)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				ctx.Event.UserID,                         // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[1], // 修改成的群头衔
			)
			ctx.SendChain(message.Text("嗯！不错的头衔呢~"))
			return
		})
	// 群聊转发
	zero.OnRegex(`^群聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			// 对CQ码进行反转义
			content := ctx.State["regex_matched"].([]string)[2]
			content = strings.ReplaceAll(content, "&#91;", "[")
			content = strings.ReplaceAll(content, "&#93;", "]")
			ctx.SendGroupMessage(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 需要发送的群
				content, // 需要发送的信息
			)
			ctx.SendChain(message.Text("📧 --> " + ctx.State["regex_matched"].([]string)[1]))
			return
		})
	// 私聊转发
	zero.OnRegex(`^私聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			// 对CQ码进行反转义
			content := ctx.State["regex_matched"].([]string)[2]
			content = strings.ReplaceAll(content, "&#91;", "[")
			content = strings.ReplaceAll(content, "&#93;", "]")
			ctx.SendPrivateMessage(
				strToInt(ctx.State["regex_matched"].([]string)[1]), // 需要发送的人的qq
				content, // 需要发送的信息
			)
			ctx.SendChain(message.Text("📧 --> " + ctx.State["regex_matched"].([]string)[1]))
			return
		})
	// 定时提醒
	zero.OnRegex(`^在(.{1,2})月(.{1,3}日|每?周.?)的(.{1,3})点(.{1,3})分时(用.+)?提醒大家(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			dateStrs := ctx.State["regex_matched"].([]string)
			ts := getFilledTimeStamp(dateStrs, false)
			if ts.enable {
				go timer(ts, func() {
					if ts.url == "" {
						ctx.SendChain(AtAll(), message.Text(ts.alert))
					} else {
						ctx.SendChain(AtAll(), message.Text(ts.alert), ImageNoCache(ts.url))
					}
				})
				ctx.Send("记住了~")
			} else {
				ctx.Send("参数非法!")
			}
			return
		})
	// 取消定时
	zero.OnRegex(`^取消在(.{1,2})月(.{1,3}日|每?周.?)的(.{1,3})点(.{1,3})分的提醒`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			dateStrs := ctx.State["regex_matched"].([]string)
			ts := getFilledTimeStamp(dateStrs, true)
			ti := getTimerInfo(&ts)
			t, ok := timers[ti]
			if ok {
				t.enable = false
				delete(timers, ti) //避免重复取消
				ctx.Send("取消成功~")
			} else {
				ctx.Send("没有这个定时器哦~")
			}
			return
		})
	// 入群欢迎
	zero.OnNotice().SetBlock(false).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_increase" {
				ctx.SendChain(message.Text("欢迎~"))
			}
			return
		})
	// 退群提醒
	zero.OnNotice().SetBlock(false).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_decrease" {
				ctx.SendChain(message.Text("有人跑路了~"))
			}
			return
		})
	// 运行 CQ 码
	zero.OnRegex(`^run(.*)$`, zero.SuperUserPermission).SetBlock(true).SetPriority(0).
		Handle(func(ctx *zero.Ctx) {
			var cmd = ctx.State["regex_matched"].([]string)[1]
			cmd = strings.ReplaceAll(cmd, "&#91;", "[")
			cmd = strings.ReplaceAll(cmd, "&#93;", "]")
			ctx.Send(cmd)
		})
}

func strToInt(str string) int64 {
	val, _ := strconv.ParseInt(str, 10, 64)
	return val
}
