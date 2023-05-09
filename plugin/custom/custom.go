// Package custom 自定义插件
package custom

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/FloatTech/floatbox/binary"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("custom", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "自订工具",
		Help: " - /kill\n" +
			" - /发送公告\n" +
			" - @bot给主人留言<内容>\n" +
			" - 模拟xx条消息",
	})
	engine.OnFullMatchGroup([]string{"pause", "restart", "/kill"}, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			os.Exit(0)
		})
	engine.OnRegex(`^模拟([0-9]+)条消息`, isfirstsuperusers()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var err error
			e := ctx.Event
			e.RawMessage = "虚拟事件"
			e.Message = message.Message{message.Text("虚拟事件")}
			e.NativeMessage = json.RawMessage("\"虚拟事件\"")
			e.Time = time.Now().Unix()
			vev, cl := binary.OpenWriterF(func(w *binary.Writer) {
				err = json.NewEncoder(w).Encode(e)
			})
			defer cl()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			i, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			go func() {
				for j := int64(0); j < i; j++ {
					go ctx.Echo(vev)
				}
			}()
		})
	engine.OnCommand("发送公告", isfirstsuperusers()).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, zero.OnlyGroup, ctx.CheckSession())
			recv, stop := next.Repeat()
			defer stop()
			ctx.SendChain(message.Text("请输入公告内容"))
			var step int
			var origin string
			for {
				select {
				case <-time.After(time.Second * 60):
					ctx.SendChain(message.Text("时间太久啦！不发了！"))
					return
				case c := <-recv:
					switch step {
					case 0:
						origin = c.Event.RawMessage
						ctx.SendChain(message.Text("请输入\"确定\"或者\"取消\"来决定是否发送此公告"))
						step++
					case 1:
						msg := c.Event.Message.ExtractPlainText()
						if msg != "确定" && msg != "取消" {
							ctx.SendChain(message.Text("请输入\"确定\"或者\"取消\"哟"))
							continue
						}
						if msg == "确定" {
							ctx.SendChain(message.Text("正在发送..."))
							zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
								for _, g := range ctx.GetGroupList().Array() {
									gid := g.Get("group_id").Int()
									ctx.SendGroupMessage(gid, origin)
								}
								return true
							})
							return
						}
						ctx.SendChain(message.Text("已经取消发送了哟~"))
						return
					}
				}
			}
		})
	engine.OnRegex(`给主人留言.*?(.*)`, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			su := zero.BotConfig.SuperUsers[0]
			now := time.Unix(ctx.Event.Time, 0).Format("2006-01-02 15:04:05")
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			username := ctx.CardOrNickName(uid)
			botid := ctx.Event.SelfID
			botname := zero.BotConfig.NickName[0]
			rawmsg := ctx.State["regex_matched"].([]string)[1]
			rawmsg = message.UnescapeCQCodeText(rawmsg)
			msg := make(message.Message, 10)
			msg = append(msg, message.CustomNode(botname, botid, "有人给你留言啦！\n在"+now))
			if gid != 0 {
				groupname := ctx.GetGroupInfo(gid, true).Name
				msg = append(msg, message.CustomNode(botname, botid, "来自群聊:["+groupname+"]("+strconv.FormatInt(gid, 10)+")\n来自群成员:["+username+"]("+strconv.FormatInt(uid, 10)+")\n以下是留言内容"))
			} else {
				msg = append(msg, message.CustomNode(botname, botid, "来自私聊:["+username+"]("+strconv.FormatInt(uid, 10)+")\n以下是留言内容:"))
			}
			msg = append(msg, message.CustomNode(username, uid, rawmsg))
			ctx.SendPrivateForwardMessage(su, msg)
		})
}

func isfirstsuperusers() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		return ctx.Event.UserID == zero.BotConfig.SuperUsers[0]
	}
}
