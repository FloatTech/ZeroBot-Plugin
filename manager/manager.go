package manager

import (
	"bot/manager/utils"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	// TODO 菜单
	zero.OnFullMatch("群管系统").SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(`====群管====
- 禁言@QQ 1
- 解除禁言 @QQ
- 我要自闭 1
- 开启全员禁言
- 解除全员禁言
- 升为管理@QQ
- 取消管理@QQ
- 修改名片@QQ XXX
- 修改头衔@QQ XXX
- 申请头衔 XXX
- 群聊转发 1234 XXX
- 私聊转发 0000 XXX`)
			return
		})
	// TODO 升为管理
	zero.OnRegex(`^升为管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				true,
			)
			nickname := ctx.GetGroupMemberInfo( // 被升为管理的人的昵称
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				false,
			).Get("nickname").Str
			ctx.Send(nickname + " 升为了管理~")
			return
		})
	// TODO 取消管理
	zero.OnRegex(`^取消管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被取消管理的人的昵称
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			).Get("nickname").Str
			ctx.Send("残念~ " + nickname + " 暂时失去了管理员的资格")
			return
		})
	// TODO 踢出群聊
	zero.OnRegex(`^踢出群聊.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupKick(
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被踢出群聊的人的昵称
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			).Get("nickname").Str
			ctx.Send("残念~ " + nickname + " 被放逐")
			return
		})
	// TODO 退出群聊
	zero.OnRegex(`^退出群聊.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupLeave(
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 要退出的群的群号
				true,
			)
			return
		})
	// TODO 开启全体禁言
	zero.OnRegex(`^开启全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				true,
			)
			ctx.Send("全员自闭开始~")
			return
		})
	// TODO 解除全员禁言
	zero.OnRegex(`^解除全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				false,
			)
			ctx.Send("全员自闭结束~")
			return
		})
	// TODO 禁言
	zero.OnRegex(`^禁言.*?(\d+).*?\s(\d+)(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			duration := utils.Str2Int(ctx.State["regex_matched"].([]string)[2])
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
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 要禁言的人的qq
				duration*60, // 要禁言的时间（分钟）
			)
			ctx.Send("小黑屋收留成功~")
			return
		})
	// TODO 解除禁言
	zero.OnRegex(`^解除禁言.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 要解除禁言的人的qq
				0,
			)
			ctx.Send("小黑屋释放成功~")
			return
		})
	// TODO 自闭禁言
	zero.OnRegex(`^我要自闭.*?(\d+)(.*)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			duration := utils.Str2Int(ctx.State["regex_matched"].([]string)[1])
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
			ctx.Send("那我就不手下留情了~")
			return
		})
	// TODO 修改名片
	zero.OnRegex(`^修改名片.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupCard(
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被修改群名片的人
				ctx.State["regex_matched"].([]string)[2],                // 修改成的群名片
			)
			ctx.Send("嗯！已经修改了")
			return
		})
	// TODO 修改头衔
	zero.OnRegex(`^修改头衔.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[2],                // 修改成的群头衔
			)
			ctx.Send("嗯！已经修改了")
			return
		})
	// TODO 申请头衔
	zero.OnRegex(`^申请头衔(.*)`, zero.OnlyGroup).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				ctx.Event.UserID,                         // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[1], // 修改成的群头衔
			)
			ctx.Send("嗯！不错的头衔呢~")
			return
		})
	// TODO 群聊转发
	zero.OnRegex(`^群聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendGroupMessage(
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 需要发送的群
				ctx.State["regex_matched"].([]string)[1],                // 需要发送的信息
			)
			ctx.Send("📧 --> " + ctx.State["regex_matched"].([]string)[1])
			return
		})
	// TODO 私聊转发
	zero.OnRegex(`^私聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendPrivateMessage(
				utils.Str2Int(ctx.State["regex_matched"].([]string)[1]), // 需要发送的人的qq
				ctx.State["regex_matched"].([]string)[1],                // 需要发送的信息
			)
			ctx.Send("📧 --> " + ctx.State["regex_matched"].([]string)[1])
			return
		})
	// TODO 戳一戳
	zero.OnNotice().SetBlock(false).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "notify" && ctx.Event.SubType == "poke" && ctx.Event.RawEvent.Get("target_id").Int() == utils.Str2Int(zero.BotConfig.SelfID) {
				time.Sleep(time.Second * 1)
				ctx.Send("请不要戳我 >_<")
			}
			return
		})
	// TODO 入群欢迎
	zero.OnNotice().SetBlock(false).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_increase" {
				ctx.Send("欢迎~")
			}
			return
		})
	// TODO 退群提醒
	zero.OnNotice().SetBlock(false).SetPriority(40).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_decrease" {
				ctx.Send("有人跑路了~")
			}
			return
		})
}
