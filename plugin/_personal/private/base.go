// Package base 基础指令
package base

import (
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const serviceName = "base"

var engine = control.Register(serviceName, &ctrl.Options[*zero.Ctx]{
	DisableOnDefault:  false,
	Brief:             "基础指令",
	Help:              "- /反馈[内容]\n- @bot备份代码\n- @bot上传代码\n- @bot检查更新\n- @bot重启\n重启需要将bat文件改成while或者goto循环\ntips:检查更新后如果没有问题后需要重启才OK",
	PrivateDataFolder: "base",
	OnDisable: func(ctx *zero.Ctx) {
		process.SleepAbout1sTo2s()
		ctx.SendChain(message.Text("宝↗生↘永↗梦↘！！！！"))
	},
})

func init() {
	// 重启
	go func() {
		process.GlobalInitMutex.Lock()
		defer process.GlobalInitMutex.Unlock()
		process.SleepAbout1sTo2s()
		m, ok := control.Lookup(serviceName)
		if ok {
			botQQ := m.GetData(0)
			if botQQ <= 0 {
				return
			}
			ctx := zero.GetBot(botQQ)
			gid := m.GetData(-1)
			switch {
			case gid > 0:
				ctx.SendGroupMessage(gid, message.Text("我回来了😊"))
			case gid < 0:
				ctx.SendPrivateMessage(-gid, message.Text("我回来了😊"))
			default:
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("我回来了😊"))
			}
			err := m.SetData(0, 0) // 清除缓存
			if err != nil {
				err = m.SetData(-1, 0) // 清除缓存
				if err != nil {
					ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(err))
				}
			}
		}
	}()
	zero.OnFullMatchGroup([]string{"重启", "洗手手"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m, ok := control.Lookup(serviceName)
			if ok {
				err := m.SetData(0, ctx.Event.RawEvent.Get("self_id").Int())
				if err != nil {
					ctx.SendChain(message.Text("保存botQQ号失败,", err))
				}
				gid := ctx.Event.GroupID
				if gid == 0 {
					gid = -ctx.Event.UserID
				}
				err = m.SetData(-1, gid)
				if err != nil {
					ctx.SendChain(message.Text("保存响应对象失败,", err))
				}
			}
			ctx.SendChain(message.Text("好的"))
			os.Exit(0)
		})
	// 运行 CQ 码
	zero.OnPrefix("run", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 可注入，权限为主人
			ctx.Send(message.UnescapeCQCodeText(ctx.State["args"].(string)))
		})
	// 撤回最后的发言
	zero.OnRegex(`^\[CQ:reply,id=(.*)].*`, zero.KeywordRule("多嘴", "撤回")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 获取消息id
			mid := ctx.State["regex_matched"].([]string)[1]
			// 撤回消息
			if ctx.Event.Message[1].Data["qq"] != "" {
				var nickname = zero.BotConfig.NickName[0]
				ctx.SendChain(message.Text("9494,要像", nickname, "一样乖乖的才行哟~"))
			} else {
				ctx.SendChain(message.Text("呜呜呜呜"))
			}
			ctx.DeleteMessage(message.NewMessageIDFromString(mid))
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
		})
	zero.OnNotice(func(ctx *zero.Ctx) bool {
		return ctx.Event.NoticeType == "group_recall" || ctx.Event.NoticeType == "friend_recall"
	}).SetBlock(false).Handle(func(ctx *zero.Ctx) {
		id, ok := ctx.Event.MessageID.(int64)
		if !ok {
			return
		}
		for _, msg := range zero.GetTriggeredMessages(message.NewMessageIDFromInteger(id)) {
			process.SleepAbout1sTo2s()
			ctx.DeleteMessage(msg)
		}
	})
	// 反馈信息
	zero.OnCommand("反馈").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			content := ctx.Event.Message.CQString()
			if content == "" {
				ctx.Send(
					message.ReplyWithMessage(ctx.Event.MessageID,
						message.Text("你是想反馈「空手假象」combo吗?"),
					),
				)
				return
			}
			username := ctx.CardOrNickName(uid)
			content = strings.ReplaceAll(content, zero.BotConfig.CommandPrefix+"反馈", "")
			text := "来自用户" + username + "(" + strconv.FormatInt(uid, 10) + ")的反馈"
			if gid != 0 {
				text = "来自群" + ctx.GetGroupInfo(gid, true).Name + "(" + strconv.FormatInt(gid, 10) + ")的用户\n" + username + "(" + strconv.FormatInt(uid, 10) + ")的反馈"
			}
			ctx.SendPrivateForwardMessage(zero.BotConfig.SuperUsers[0], message.Message{
				message.CustomNode(username, uid, text),
				message.CustomNode(username, uid, message.UnescapeCQCodeText(content)),
			})
			ctx.SendChain(message.Text("反馈成功"))
		})
}
