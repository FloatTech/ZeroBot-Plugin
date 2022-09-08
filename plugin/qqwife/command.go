// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品，文案采用了Hana的zbp娶群友文案
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 反并发
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	// 数据库
	sql "github.com/FloatTech/sqlite"
	// 定时器
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	// 画图
	"github.com/Coloured-glaze/gg"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/zbputils/img/text"
)

//nolint: asciicheck
// nolint: asciicheck
var (
	民政局 = &婚姻登记{
		db: &sql.Sqlite{},
	}
	skillCD  = rate.NewManager[string](time.Hour*12, 1)
	sendtext = [...][]string{
		{ // 表白成功
			"是个勇敢的孩子(*/ω＼*) 今天的运气都降临在你的身边~\n\n",
			"(´･ω･`)对方答应了你 并表示愿意当今天的CP\n\n",
		},
		{ // 表白失败
			"今天的运气有一点背哦~明天再试试叭",
			"_(:з」∠)_下次还有机会 咱抱抱你w",
			"今天失败了惹. 摸摸头~咱明天还有机会",
		},
		{ // ntr成功
			"因为你的个人魅力~~今天他就是你的了w\n\n",
		},
		{ // 离婚失败
			"打是情，骂是爱，,不打不亲不相爱。答应我不要分手。",
			"床头打架床尾和，夫妻没有隔夜仇。安啦安啦，不要闹变扭。",
		},
		{ // 离婚成功
			"离婚成功力\n天涯何处无芳草，何必单恋一枝花？不如再摘一支（bushi",
			"离婚成功力\n话说你不考虑当个1？",
		},
	}
)

func init() {
	engine := control.Register("qqwife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		PrivateDataFolder: "qqwife",
		Help: "一群一天一夫一妻制群老婆\n（每天凌晨刷新CP）\n" +
			"- 娶群友\n- 群老婆列表\n- 我群老婆\n" +
			"--------------------------------\n以下技能每个CD12H，不跨天刷新\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n- 当[对方Q号|@对方QQ]的小三\n- 闹离婚",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("别着急，民政局门口排长队了！"),
				),
			)
		}),
	))
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		民政局.db.DBPath = engine.DataFolder() + "结婚登记表.db"
		err := 民政局.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
			return false
		}
		return true
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			_, err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			uid := ctx.Event.UserID
			targetinfo, status, err := 民政局.查户口(gid, uid)
			switch {
			case status == "错":
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			case (status == "攻" && targetinfo.Target == 0) ||
				(status == "受" && targetinfo.User == 0): // 如果是单身贵族
				ctx.SendChain(message.Text("今天你是单身贵族噢"))
				return
			case status == "攻": // 娶过别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你已经娶过了，群老婆是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(targetinfo.Target, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", targetinfo.Targetname, "]",
						"(", targetinfo.Target, ")哒",
					),
				)
				return
			case status == "受": // 嫁给别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你被娶了，群老公是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(targetinfo.User, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", targetinfo.Username, "]",
						"(", targetinfo.User, ")哒",
					),
				)
				return
			}
			// 无缓存获取群员列表
			temp := ctx.GetThisGroupMemberListNoCache().Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			// 将已经娶过的人剔除
			qqgrouplist := make([]int64, 0, len(temp))
			for k := 0; k < len(temp); k++ {
				usr := temp[k].Get("user_id").Int()
				_, status, _ := 民政局.查户口(gid, usr)
				if status != "单" {
					continue
				}
				qqgrouplist = append(qqgrouplist, usr)
			}
			// 没有人（只剩自己）的时候
			if len(qqgrouplist) == 1 {
				ctx.SendChain(message.Text("~群里没有ta人是单身了哦 明天再试试叭"))
				return
			}
			// 随机抽娶
			fiancee := qqgrouplist[rand.Intn(len(qqgrouplist))]
			if fiancee == uid { // 如果是自己
				ctx.SendChain(message.Text("呜...没娶到，你可以再尝试一次"))
				return
			}
			// 去民政局办证
			err = 民政局.登记(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			// 请大家吃席
			ctx.SendChain(
				message.At(uid),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(fiancee), "]",
					"(", fiancee, ")哒",
				),
			)
		})
	// 单身技能
	engine.OnRegex(`^(娶|嫁)\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, getdb, checkdog).SetBlock(true).Limit(cdcheck, iscding).
		Handle(func(ctx *zero.Ctx) {
			choice := ctx.State["regex_matched"].([]string)[1]
			fiancee, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			uid := ctx.Event.UserID
			gid := ctx.Event.GroupID
			if uid == fiancee { // 如果是自己
				switch rand.Intn(3) {
				case 1:
					err := 民政局.登记(gid, uid, 0, "", "")
					if err != nil {
						ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
						return
					}
					ctx.SendChain(message.Text("今日获得成就：单身贵族"))
				default:
					ctx.SendChain(message.Text("今日获得成就：自恋狂"))
				}
				return
			}
			if rand.Intn(2) == 0 { // 二分之一的概率表白成功
				ctx.SendChain(message.Text(sendtext[1][rand.Intn(len(sendtext[1]))]))
				return
			}
			// 去民政局登记
			var choicetext string
			switch choice {
			case "娶":
				err := 民政局.登记(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
				if err != nil {
					ctx.SendChain(message.Text("结婚登记失败力\n[error]", err))
					return
				}
				choicetext = "\n今天你的群老婆是"
			default:
				err := 民政局.登记(gid, fiancee, uid, ctx.CardOrNickName(fiancee), ctx.CardOrNickName(uid))
				if err != nil {
					ctx.SendChain(message.Text("结婚登记失败力\n[error]", err))
					return
				}
				choicetext = "\n今天你的群老公是"
			}
			// 请大家吃席
			ctx.SendChain(
				message.Text(sendtext[0][rand.Intn(len(sendtext[0]))]),
				message.At(uid),
				message.Text(choicetext),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(fiancee), "]",
					"(", fiancee, ")哒",
				),
			)
		})
	// NTR技能
	engine.OnRegex(`^当(\[CQ:at,qq=(\d+)\]\s?|(\d+))的小三`, zero.OnlyGroup, getdb, checkcp).SetBlock(true).Limit(cdcheck2, iscding).
		Handle(func(ctx *zero.Ctx) {
			fid := ctx.State["regex_matched"].([]string)
			fiancee, _ := strconv.ParseInt(fid[2]+fid[3], 10, 64)
			uid := ctx.Event.UserID
			if fiancee == uid {
				ctx.SendChain(message.Text("今日获得成就：自我攻略"))
				return
			}
			if rand.Intn(10)/4 != 0 { // 十分之三的概率NTR成功
				ctx.SendChain(message.Text("失败了！可惜"))
				return
			}
			gid := ctx.Event.GroupID
			// 判断target是老公还是老婆
			var choicetext string
			userID := uid
			targetID := fiancee
			fianceeinfo, gender, err := 民政局.查户口(gid, fiancee)
			switch gender {
			case "单":
				ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
				return
			case "错":
				ctx.SendChain(message.Text("对象状态查询失败\n[error]", err))
				return
			case "攻":
				err = 民政局.离婚休妻(gid, fianceeinfo.Target)
				if err != nil {
					ctx.SendChain(message.Text("ta不想和原来的对象分手...\n[error]", err))
					return
				}
				userID = fiancee
				targetID = uid
				choicetext = "老公"
			case "受":
				err = 民政局.离婚休夫(gid, fianceeinfo.User)
				if err != nil {
					ctx.SendChain(message.Text("ta不想和原来的对象分手...\n[error]", err))
					return
				}
				choicetext = "老婆"
			default:
				ctx.SendChain(message.Text("数据库发生问题力"))
				return
			}
			err = 民政局.登记(gid, userID, targetID, ctx.CardOrNickName(userID), ctx.CardOrNickName(targetID))
			if err != nil {
				ctx.SendChain(message.Text("复婚登记失败力\n[error]", err))
				return
			}
			// 输出结果
			ctx.SendChain(
				message.Text(sendtext[2][rand.Intn(len(sendtext[2]))]),
				message.At(uid),
				message.Text("今天你的群"+choicetext+"是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(fiancee), "]",
					"(", fiancee, ")哒",
				),
			)
		})
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			ok, err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			if ok {
				ctx.SendChain(message.Text("今天还没有人结婚哦"))
				return
			}
			list, number, err := 民政局.花名册(gid)
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			if number <= 0 {
				ctx.SendChain(message.Text("今天还没有人结婚哦"))
				return
			}
			/***********设置图片的大小和底色***********/
			fontSize := 50.0
			if number < 10 {
				number = 10
			}
			canvas := gg.NewContext(1500, int(250+fontSize*float64(number)))
			canvas.SetRGB(1, 1, 1) // 白色
			canvas.Clear()
			/***********下载字体，可以注销掉***********/
			_, err = file.GetLazyData(text.BoldFontFile, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize*2); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("群老婆列表")
			/***********绘制标题***********/
			canvas.DrawString("群老婆列表", (1500-sl)/2, 160-h) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 250-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			_, h = canvas.MeasureString("焯")
			for i, info := range list {
				canvas.DrawString(slicename(info[0], canvas), 0, float64(260+50*i)-h)
				canvas.DrawString("("+info[1]+")", 350, float64(260+50*i)-h)
				canvas.DrawString("←→", 700, float64(260+50*i)-h)
				canvas.DrawString(slicename(info[2], canvas), 800, float64(260+50*i)-h)
				canvas.DrawString("("+info[3]+")", 1150, float64(260+50*i)-h)
			}
			data, cl := writer.ToBytes(canvas.Image())
			ctx.SendChain(message.ImageBytes(data))
			cl()
		})
	engine.OnFullMatchGroup([]string{"闹离婚", "办离婚"}, zero.OnlyGroup, getdb, func(ctx *zero.Ctx) bool {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		_, uidstatus, err := 民政局.查户口(gid, uid)
		switch uidstatus {
		case "错":
			ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
			return false
		case "单":
			ctx.SendChain(message.Text("今天你还没结婚哦"))
			return false
		}
		return true
	}).SetBlock(true).Limit(cdcheck3, iscding2).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			info, uidstatus, err := 民政局.查户口(gid, uid)
			mun := 2
			switch uidstatus {
			case "错":
				ctx.SendChain(message.Text("用户状态查询失败\n[error]", err))
				return
			case "攻":
				mun = 1
			case "受":
				mun = 0
			}
			if rand.Intn(10) != 1 { // 十分之一的概率成功
				ctx.SendChain(message.Text(sendtext[3][rand.Intn(len(sendtext[3]))]))
				return
			}
			switch mun {
			case 1:
				err = 民政局.离婚休妻(gid, info.Target)
			case 0:
				err = 民政局.离婚休妻(gid, info.Target)
			default:
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			ctx.SendChain(message.Text(sendtext[4][mun]))
		})
	engine.OnRegex(`^重置(所有|本群|/d+)?花名册$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			cmd := "0"
			switch ctx.State["regex_matched"].([]string)[1] {
			case "":
				if ctx.Event.GroupID == 0 {
					ctx.SendChain(message.Text("该功能只能在群组使用或者指定群组"))
					return
				}
				cmd = strconv.FormatInt(ctx.Event.GroupID, 10)
			case "所有":
				break
			case "本群":
				if ctx.Event.GroupID == 0 {
					ctx.SendChain(message.Text("该功能只能在群组使用或者指定群组"))
					return
				}
				cmd = strconv.FormatInt(ctx.Event.GroupID, 10)
			default:
				cmd = ctx.State["regex_matched"].([]string)[1]
			}
			err := 民政局.清理花名册(cmd)
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			ctx.SendChain(message.Text("重置成功"))
		})
	engine.OnFullMatch("我群老婆", zero.OnlyGroup, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			ok, err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("群状态查询失败\n[error]", err))
				return
			}
			if ok {
				ctx.SendChain(message.Text("今天你还没结婚哦"))
				return
			}
			uid := ctx.Event.UserID
			info, uidstatus, err := 民政局.查户口(gid, uid)
			switch uidstatus {
			case "错":
				ctx.SendChain(message.Text("用户状态查询失败\n[error]", err))
				return
			case "单":
				ctx.SendChain(message.Text("今天你还没结婚哦"))
				return
			case "攻": // 娶过别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你的群老婆是"),
					message.Text(
						"\n",
						"[", info.Targetname, "]",
						"(", info.Target, ")哒",
					),
				)
				return
			case "受": // 嫁给别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你被娶了，群老公是"),
					message.Text(
						"\n",
						"[", info.Username, "]",
						"(", info.User, ")哒",
					),
				)
				return
			}
		})
}
