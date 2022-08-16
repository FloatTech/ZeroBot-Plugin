// Package managerplugin 自定义群管插件
package managerplugin

import (
	"strconv"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	engine := control.Register("managerplugin", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "自定义的群管插件\n" +
			" - 开启全员禁言 群号\n" +
			" - 解除全员禁言 群号\n" +
			" - 踢出并拉黑 QQ号\n" +
			" - 踢出(并拉黑)等级为[1-100]的人\n" +
			" - 反\"XX召唤术\"\n",
		PrivateDataFolder: "managerplugin",
	})
	// 指定开启某群全群禁言 Usage: 开启全员禁言123456
	engine.OnRegex(`^开启全员禁言.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始"))
		})
	// 指定解除某群全群禁言 Usage: 解除全员禁言123456
	engine.OnRegex(`^解除全员禁言.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束"))
		})
	engine.OnRegex(`^踢出并拉黑.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			gid := ctx.Event.GroupID
			ctx.SetGroupKick(gid, uid, true)
			nickname := ctx.CardOrNickName(uid)
			ctx.SendChain(message.Text("已将", nickname, "流放到边界外~"))
		})
	/*engine.OnRegex(`踢出等级为([0-9]{1,3})的人`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			setlevel := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if setlevel > 100 {
				return
			}
			gid := ctx.Event.GroupID
			banlist := ctx.GetGroupMemberListNoCache(gid).Array()
			ctx.SendChain(message.Text("正在执行中..."))
			var i int
			for _, ban := range banlist {
				banid := ban.Get("user_id").Int()
				banlevel := ban.Get("level").String()
				levelint := math.Str2Int64(banlevel)
				for _, adminid := range zero.BotConfig.SuperUsers {
					if levelint == setlevel && banid != ctx.Event.SelfID && banid != adminid {
						ctx.SetGroupKick(gid, banid, false)
						i++
					}
				}
			}
			ctx.SendChain(message.Text("本次一共踢出了", i, "个人"))
		})
	engine.OnRegex(`踢出并拉黑等级为([0-9]{1,3})的人`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			setlevel := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if setlevel > 100 {
				return
			}
			gid := ctx.Event.GroupID
			banlist := ctx.GetGroupMemberListNoCache(gid).Array()
			ctx.SendChain(message.Text("正在执行中..."))
			var i int
			for _, ban := range banlist {
				banid := ban.Get("user_id").Int()
				banlevel := ban.Get("level").String()
				levelint := math.Str2Int64(banlevel)

				for _, adminid := range zero.BotConfig.SuperUsers {
					if levelint == setlevel && banid != ctx.Event.SelfID && banid != adminid {
						ctx.SetGroupKick(gid, banid, true)
						i++
					}
				}
			}
			ctx.SendChain(message.Text("本次一共踢出了", i, "个人"))
		})
	engine.OnFullMatch("获取群成员信息", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			temp := ctx.GetGroupMemberListNoCache(gid).Array()
			l := make(message.Message, 1, 2000)
			l[0] = ctxext.FakeSenderForwardNode(ctx, message.Text("--群成员信息--"))
			for _, v := range temp {
				id := v.Get("user_id").Int()
				te := ctx.GetGroupMemberInfo(gid, id, true)
				level := te.Get("level").String()
				l = append(l, ctxext.FakeSenderForwardNode(ctx, message.Text("qq号:", id, "\n", "群等级:", level)))
			}
			ctx.SendGroupForwardMessage(gid, l)
		})*/
	engine.OnRegex(`^\[CQ:xml`, zero.OnlyGroup, zero.KeywordRule("serviceID=\"60\"")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			ctx.SetGroupKick(ctx.Event.GroupID, ctx.Event.UserID, false)
			ctx.SetGroupBan(ctx.Event.GroupID, ctx.Event.UserID, 7*24*60*60)
			ctx.SendChain(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("检测到 ["+nickname+"]("+strconv.FormatInt(ctx.Event.UserID, 10)+") 发送了干扰性消息,已处理"))...)
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
		})
}
