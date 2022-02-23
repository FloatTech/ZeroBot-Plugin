// Package manager 群管
package manager

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	sql "github.com/FloatTech/sqlite"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/zbputils/control/order"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_manager/timer"
)

const (
	hint = "====群管====\n" +
		"- 禁言@QQ 1分钟\n" +
		"- 解除禁言 @QQ\n" +
		"- 我要自闭 1分钟\n" +
		"- 开启全员禁言\n" +
		"- 解除全员禁言\n" +
		"- 升为管理@QQ\n" +
		"- 取消管理@QQ\n" +
		"- 修改名片@QQ XXX\n" +
		"- 修改头衔@QQ XXX\n" +
		"- 申请头衔 XXX\n" +
		"- 踢出群聊@QQ\n" +
		"- 退出群聊 1234@bot\n" +
		"- 群聊转发 1234 XXX\n" +
		"- 私聊转发 0000 XXX\n" +
		"- 在MM月dd日的hh点mm分时(用http://url)提醒大家XXX\n" +
		"- 在MM月[每周 | 周几]的hh点mm分时(用http://url)提醒大家XXX\n" +
		"- 取消在MM月dd日的hh点mm分的提醒\n" +
		"- 取消在MM月[每周 | 周几]的hh点mm分的提醒\n" +
		"- 在\"cron\"时(用[url])提醒大家[xxx]\n" +
		"- 取消在\"cron\"的提醒\n" +
		"- 列出所有提醒\n" +
		"- 翻牌\n" +
		"- 设置欢迎语XXX（可加{at}在欢迎时@对方）\n" +
		"- 测试欢迎语\n" +
		"- [开启 | 关闭]入群验证"
)

var (
	db    = &sql.Sqlite{}
	clock timer.Clock
)

func init() { // 插件主体
	engine := control.Register("manager", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              hint,
		PrivateDataFolder: "manager",
	})

	go func() {
		defer order.DoneOnExit()()
		db.DBPath = engine.DataFolder() + "config.db"
		clock = timer.NewClock(db)
		err := db.Create("welcome", &welcome{})
		if err != nil {
			panic(err)
		}
		err = db.Create("member", &member{})
		if err != nil {
			panic(err)
		}
	}()

	// 升为管理
	engine.OnRegex(`^升为管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				true,
			)
			nickname := ctx.GetGroupMemberInfo( // 被升为管理的人的昵称
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被升为管理的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text(nickname + " 升为了管理~"))
		})
	// 取消管理
	engine.OnRegex(`^取消管理.*?(\d+)`, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupAdmin(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被取消管理的人的昵称
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被取消管理的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text("残念~ " + nickname + " 暂时失去了管理员的资格"))
		})
	// 踢出群聊
	engine.OnRegex(`^踢出群聊.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupKick(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			)
			nickname := ctx.GetGroupMemberInfo( // 被踢出群聊的人的昵称
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被踢出群聊的人的qq
				false,
			).Get("nickname").Str
			ctx.SendChain(message.Text("残念~ " + nickname + " 被放逐"))
		})
	// 退出群聊
	engine.OnRegex(`^退出群聊.*?(\d+)`, zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupLeave(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 要退出的群的群号
				true,
			)
		})
	// 开启全体禁言
	engine.OnRegex(`^开启全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始~"))
		})
	// 解除全员禁言
	engine.OnRegex(`^解除全员禁言$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				ctx.Event.GroupID,
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束~"))
		})
	// 禁言
	engine.OnRegex(`^禁言.*?(\d+).*?\s(\d+)(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			duration := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			switch ctx.State["regex_matched"].([]string)[3] {
			case "分钟":
				//
			case "小时":
				duration *= 60
			case "天":
				duration *= 60 * 24
			default:
				//
			}
			if duration >= 43200 {
				duration = 43199 // qq禁言最大时长为一个月
			}
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 要禁言的人的qq
				duration*60, // 要禁言的时间（分钟）
			)
			ctx.SendChain(message.Text("小黑屋收留成功~"))
		})
	// 解除禁言
	engine.OnRegex(`^解除禁言.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupBan(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 要解除禁言的人的qq
				0,
			)
			ctx.SendChain(message.Text("小黑屋释放成功~"))
		})
	// 自闭禁言
	engine.OnRegex(`^(我要自闭|禅定).*?(\d+)(.*)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			duration := math.Str2Int64(ctx.State["regex_matched"].([]string)[2])
			switch ctx.State["regex_matched"].([]string)[3] {
			case "分钟", "min", "mins", "m":
				break
			case "小时", "hour", "hours", "h":
				duration *= 60
			case "天", "day", "days", "d":
				duration *= 60 * 24
			default:
				break
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
		})
	// 修改名片
	engine.OnRegex(`^修改名片.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if len(ctx.State["regex_matched"].([]string)[2]) > 60 {
				ctx.SendChain(message.Text("名字太长啦！"))
				return
			}
			ctx.SetGroupCard(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被修改群名片的人
				ctx.State["regex_matched"].([]string)[2],                 // 修改成的群名片
			)
			ctx.SendChain(message.Text("嗯！已经修改了"))
		})
	// 修改头衔
	engine.OnRegex(`^修改头衔.*?(\d+).*?\s(.*)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if len(ctx.State["regex_matched"].([]string)[1]) > 18 {
				ctx.SendChain(message.Text("头衔太长啦！"))
				return
			}
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[2],                 // 修改成的群头衔
			)
			ctx.SendChain(message.Text("嗯！已经修改了"))
		})
	// 申请头衔
	engine.OnRegex(`^申请头衔(.*)`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if len(ctx.State["regex_matched"].([]string)[1]) > 18 {
				ctx.SendChain(message.Text("头衔太长啦！"))
				return
			}
			ctx.SetGroupSpecialTitle(
				ctx.Event.GroupID,
				ctx.Event.UserID,                         // 被修改群头衔的人
				ctx.State["regex_matched"].([]string)[1], // 修改成的群头衔
			)
			ctx.SendChain(message.Text("嗯！不错的头衔呢~"))
		})
	// 群聊转发
	engine.OnRegex(`^群聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 对CQ码进行反转义
			content := ctx.State["regex_matched"].([]string)[2]
			content = strings.ReplaceAll(content, "&#91;", "[")
			content = strings.ReplaceAll(content, "&#93;", "]")
			ctx.SendGroupMessage(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 需要发送的群
				content, // 需要发送的信息
			)
			ctx.SendChain(message.Text("📧 --> " + ctx.State["regex_matched"].([]string)[1]))
		})
	// 私聊转发
	engine.OnRegex(`^私聊转发.*?(\d+)\s(.*)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 对CQ码进行反转义
			content := ctx.State["regex_matched"].([]string)[2]
			content = strings.ReplaceAll(content, "&#91;", "[")
			content = strings.ReplaceAll(content, "&#93;", "]")
			ctx.SendPrivateMessage(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]), // 需要发送的人的qq
				content, // 需要发送的信息
			)
			ctx.SendChain(message.Text("📧 --> " + ctx.State["regex_matched"].([]string)[1]))
		})
	// 定时提醒
	engine.OnRegex(`^在(.{1,2})月(.{1,3}日|每?周.?)的(.{1,3})点(.{1,3})分时(用.+)?提醒大家(.*)`, zero.AdminPermission, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			dateStrs := ctx.State["regex_matched"].([]string)
			ts := timer.GetFilledTimer(dateStrs, ctx.Event.SelfID, ctx.Event.GroupID, false)
			if ts.En() {
				go clock.RegisterTimer(ts, true)
				ctx.SendChain(message.Text("记住了~"))
			} else {
				ctx.SendChain(message.Text("参数非法:" + ts.Alert))
			}
		})
	// 定时 cron 提醒
	engine.OnRegex(`^在"(.*)"时(用.+)?提醒大家(.*)`, zero.AdminPermission, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			dateStrs := ctx.State["regex_matched"].([]string)
			var url, alert string
			switch len(dateStrs) {
			case 4:
				url = dateStrs[2]
				alert = dateStrs[3]
			case 3:
				alert = dateStrs[2]
			default:
				ctx.SendChain(message.Text("参数非法!"))
				return
			}
			logrus.Debugln("[manager] cron:", dateStrs[1])
			ts := timer.GetFilledCronTimer(dateStrs[1], alert, url, ctx.Event.SelfID, ctx.Event.GroupID)
			if clock.RegisterTimer(ts, true) {
				ctx.SendChain(message.Text("记住了~"))
			} else {
				ctx.SendChain(message.Text("参数非法:" + ts.Alert))
			}
		})
	// 取消定时
	engine.OnRegex(`^取消在(.{1,2})月(.{1,3}日|每?周.?)的(.{1,3})点(.{1,3})分的提醒`, zero.AdminPermission, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			dateStrs := ctx.State["regex_matched"].([]string)
			ts := timer.GetFilledTimer(dateStrs, ctx.Event.SelfID, ctx.Event.GroupID, true)
			ti := ts.GetTimerID()
			ok := clock.CancelTimer(ti)
			if ok {
				ctx.SendChain(message.Text("取消成功~"))
			} else {
				ctx.SendChain(message.Text("没有这个定时器哦~"))
			}
		})
	// 取消 cron 定时
	engine.OnRegex(`^取消在"(.*)"的提醒`, zero.AdminPermission, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			dateStrs := ctx.State["regex_matched"].([]string)
			ts := timer.Timer{Cron: dateStrs[1], GrpID: ctx.Event.GroupID}
			ti := ts.GetTimerID()
			ok := clock.CancelTimer(ti)
			if ok {
				ctx.SendChain(message.Text("取消成功~"))
			} else {
				ctx.SendChain(message.Text("没有这个定时器哦~"))
			}
		})
	// 列出本群所有定时
	engine.OnFullMatch("列出所有提醒", zero.AdminPermission, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(clock.ListTimers(ctx.Event.GroupID)))
		})
	// 随机点名
	engine.OnFullMatchGroup([]string{"翻牌"}, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			// 无缓存获取群员列表
			list := ctx.CallAction("get_group_member_list", zero.Params{
				"group_id": ctx.Event.GroupID,
				"no_cache": true,
			}).Data
			temp := list.Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-10):]
			who := temp[rand.Intn(len(temp))]
			if who.Get("user_id").Int() == ctx.Event.SelfID {
				ctx.SendChain(message.Text("幸运儿居然是我自己"))
				return
			}
			if who.Get("user_id").Int() == ctx.Event.UserID {
				ctx.SendChain(message.Text("哎呀，就是你自己了"))
				return
			}
			nick := who.Get("card").Str
			if nick == "" {
				nick = who.Get("nickname").Str
			}
			ctx.SendChain(
				message.Text(
					nick,
					" 就是你啦！",
				),
			)
		})
	// 入群欢迎
	engine.OnNotice().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_increase" && ctx.Event.SelfID != ctx.Event.UserID {
				var w welcome
				err := db.Find("welcome", &w, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
				if err == nil {
					ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(strings.ReplaceAll(w.Msg, "{at}", "[CQ:at,qq="+strconv.FormatInt(ctx.Event.UserID, 10)+"]")))
				} else {
					ctx.SendChain(message.Text("欢迎~"))
				}
				c, ok := control.Lookup("manager")
				if ok {
					enable := c.GetData(ctx.Event.GroupID)&1 == 1
					if enable {
						uid := ctx.Event.UserID
						a := rand.Intn(100)
						b := rand.Intn(100)
						r := a + b
						ctx.SendChain(message.At(uid), message.Text(fmt.Sprintf("考你一道题：%d+%d=?\n如果60秒之内答不上来，%s就要把你踢出去了哦~", a, b, zero.BotConfig.NickName[0])))
						// 匹配发送者进行验证
						rule := func(ctx *zero.Ctx) bool {
							for _, elem := range ctx.Event.Message {
								if elem.Type == "text" {
									text := strings.ReplaceAll(elem.Data["text"], " ", "")
									ans, err := strconv.Atoi(text)
									if err == nil {
										if ans != r {
											ctx.SendChain(message.Text("答案不对哦，再想想吧~"))
											return false
										}
										return true
									}
								}
							}
							return false
						}
						next := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID), rule)
						recv, cancel := next.Repeat()
						select {
						case <-time.After(time.Minute):
							ctx.SendChain(message.Text("拜拜啦~"))
							ctx.SetGroupKick(ctx.Event.GroupID, uid, false)
							cancel()
						case <-recv:
							cancel()
							ctx.SendChain(message.Text("答对啦~"))
						}
					}
				}
			}
		})
	// 退群提醒
	engine.OnNotice().SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.NoticeType == "group_decrease" {
				userid := ctx.Event.UserID
				ctx.SendChain(message.Text(ctxext.CardOrNickName(ctx, userid), "(", userid, ")", "离开了我们..."))
			}
		})
	// 设置欢迎语
	engine.OnRegex(`^设置欢迎语([\s\S]*)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			w := &welcome{
				GrpID: ctx.Event.GroupID,
				Msg:   ctx.State["regex_matched"].([]string)[1],
			}
			err := db.Insert("welcome", w)
			if err == nil {
				ctx.SendChain(message.Text("记住啦!"))
			} else {
				ctx.SendChain(message.Text("出错啦: ", err))
			}
		})
	// 测试欢迎语
	engine.OnFullMatch("测试欢迎语", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var w welcome
			err := db.Find("welcome", &w, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
			if err == nil {
				ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(strings.ReplaceAll(w.Msg, "{at}", "[CQ:at,qq="+strconv.FormatInt(ctx.Event.UserID, 10)+"]")))
			} else {
				ctx.SendChain(message.Text("欢迎~"))
			}
		})
	// 入群后验证开关
	engine.OnRegex(`^(.*)入群验证$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			c, ok := control.Lookup("manager")
			if ok {
				data := c.GetData(ctx.Event.GroupID)
				switch option {
				case "开启", "打开", "启用":
					data |= 1
				case "关闭", "关掉", "禁用":
					data &= 0x7fffffff_fffffffe
				default:
					return
				}
				err := c.SetData(ctx.Event.GroupID, data)
				if err == nil {
					ctx.SendChain(message.Text("已", option))
					return
				}
				ctx.SendChain(message.Text("出错啦: ", err))
				return
			}
			ctx.SendChain(message.Text("找不到服务!"))
		})
	// 加群 gist 验证开关
	engine.OnRegex(`^(.*)gist加群自动审批$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			option := ctx.State["regex_matched"].([]string)[1]
			c, ok := control.Lookup("manager")
			if ok {
				data := c.GetData(ctx.Event.GroupID)
				switch option {
				case "开启", "打开", "启用":
					data |= 0x10
				case "关闭", "关掉", "禁用":
					data &= 0x7fffffff_fffffffd
				default:
					return
				}
				err := c.SetData(ctx.Event.GroupID, data)
				if err == nil {
					ctx.SendChain(message.Text("已", option))
					return
				}
				ctx.SendChain(message.Text("出错啦: ", err))
				return
			}
			ctx.SendChain(message.Text("找不到服务!"))
		})
	// 运行 CQ 码
	engine.OnRegex(`^run(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var cmd = ctx.State["regex_matched"].([]string)[1]
			cmd = strings.ReplaceAll(cmd, "&#91;", "[")
			cmd = strings.ReplaceAll(cmd, "&#93;", "]")
			// 可注入，权限为主人
			ctx.Send(cmd)
		})
	// 根据 gist 自动同意加群
	// 加群请在github新建一个gist，其文件名为本群群号的字符串的md5(小写)，内容为一行，是当前unix时间戳(10分钟内有效)。
	// 然后请将您的用户名和gist哈希(小写)按照username/gisthash的格式填写到回答即可。
	engine.OnRequest().SetBlock(false).Handle(func(ctx *zero.Ctx) {
		/*if ctx.Event.RequestType == "friend" {
			ctx.SetFriendAddRequest(ctx.Event.Flag, true, "")
		}*/
		c, ok := control.Lookup("manager")
		if ok && c.GetData(ctx.Event.GroupID)&0x10 == 0x10 && ctx.Event.RequestType == "group" && ctx.Event.SubType == "add" {
			// gist 文件名是群号的 ascii 编码的 md5
			// gist 内容是当前 uinx 时间戳，在 10 分钟内视为有效
			ans := ctx.Event.Comment[strings.Index(ctx.Event.Comment, "答案：")+len("答案："):]
			divi := strings.Index(ans, "/")
			if divi <= 0 {
				ctx.SetGroupAddRequest(ctx.Event.Flag, "add", false, "格式错误!")
				return
			}
			ghun := ans[:divi]
			hash := ans[divi+1:]
			logrus.Infoln("[manager]收到加群申请, 用户:", ghun, ", hash:", hash)
			ok, reason := checkNewUser(ctx.Event.UserID, ctx.Event.GroupID, ghun, hash)
			if ok {
				ctx.SetGroupAddRequest(ctx.Event.Flag, "add", true, "")
				process.SleepAbout1sTo2s()
				ctx.SetGroupCard(ctx.Event.GroupID, ctx.Event.UserID, ghun)
			} else {
				ctx.SetGroupAddRequest(ctx.Event.Flag, "add", false, reason)
			}
		}
	})
}
