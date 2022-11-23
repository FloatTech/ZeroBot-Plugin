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
	// 画图
	"github.com/Coloured-glaze/gg"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/zbputils/img/text"

	// 货币系统
	"github.com/FloatTech/AnimeAPI/wallet"
)

// nolint: asciicheck
// nolint: asciicheck
var (
	民政局 = &婚姻登记{
		db: &sql.Sqlite{},
	}
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
			"打是情,骂是爱,不打不亲不相爱。答应我不要分手。",
			"床头打架床尾和，夫妻没有隔夜仇。安啦安啦，不要闹变扭。",
		},
		{ // 离婚成功
			"离婚成功力\n话说你不考虑当个1？",
			"离婚成功力\n天涯何处无芳草，何必单恋一枝花？不如再摘一支（bushi",
		},
	}
)

func init() {
	engine := control.Register("qqwife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "一群一天一夫一妻制群老婆",
		Help: "- 娶群友\n- 群老婆列表\n- [允许|禁止]自由恋爱\n- [允许|禁止]牛头人\n- 设置CD为xx小时    →(默认12小时)\n- 重置花名册\n- 重置所有花名册(用于清除所有群数据及其设置)\n" +
			"--------------------------------\n以下指令存在CD,不跨天刷新,前两个受指令开关\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n自由选择对象, 自由恋爱(好感度越高成功率越高,保底30%概率)\n" +
			"- 当[对方Q号|@对方QQ]的小三\n我和你才是真爱, 为了你我愿意付出一切(好感度越高成功率越高,保底10%概率)\n" +
			"- 闹离婚\n你谁啊, 给我滚(好感度越高成功率越低)\n" +
			"- 买礼物给[对方Q号|@对方QQ]\n使用小熊饼干获取好感度\n" +
			"- 做媒 @攻方QQ @受方QQ\n身为管理, 群友的xing福是要搭把手的(攻受双方好感度越高成功率越高,保底30%概率)\n" +
			"--------------------------------\n好感度规则\n--------------------------------\n" +
			"\"娶群友\"指令好感度随机增加1~5。\n\"A牛B的C\"会导致C恨A, 好感度-5;\nB为了报复A, 好感度+5(什么柜子play)\nA为BC做媒,成功B、C对A好感度+1反之-1\n做媒成功BC好感度+1" +
			"Tips: 群老婆列表过0点刷新",
		PrivateDataFolder: "qqwife",
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
			ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
			return false
		}
		return true
	})
	// 技能CD设置
	engine.OnRegex(`^设置CD为(\d+)小时`, zero.OnlyGroup, zero.AdminPermission, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			cdTime, err := strconv.ParseFloat(ctx.State["regex_matched"].([]string)[1], 64)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]请设置纯数字\n", err))
				return
			}
			gid := ctx.Event.GroupID
			err = 民政局.setCDtime(gid, cdTime)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]设置CD时长失败\n", err))
				return
			}
			ctx.SendChain(message.Text("设置成功"))
		})
	engine.OnRegex(`^(允许|禁止)(自由恋爱|牛头人)$`, zero.OnlyGroup, zero.AdminPermission, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			status := ctx.State["regex_matched"].([]string)[1]
			mode := ctx.State["regex_matched"].([]string)[2]
			gid := ctx.Event.GroupID
			statusBool := 1
			if status == "禁止" {
				statusBool = 0
			}
			err := 民政局.修改模式(gid, mode, statusBool)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]群状态查询失败\n", err))
				return
			}
			ctx.SendChain(message.Text("设置成功"))
		})
	// 好感度系统
	engine.OnRegex(`^查好感度\s?\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]你对象好像不存在？\n", err))
				return
			}
			uid := ctx.Event.UserID
			favor, err := 民政局.getFavorability(uid, fiancee)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				return
			}
			// 输出结果
			ctx.SendChain(
				message.At(uid),
				message.Text("\n当前你们好感度为", favor),
			)
		})
	engine.OnFullMatch("娶群友", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			_, err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			}
			uid := ctx.Event.UserID
			targetinfo, status, err := 民政局.查户口(gid, uid)
			switch {
			case status == "错":
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			case (status == "攻" && targetinfo.Target == 0) ||
				(status == "受" && targetinfo.User == 0): // 如果是单身贵族
				ctx.SendChain(message.Text("今天你是单身贵族噢"))
				return
			case status == "攻": // 娶过别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你在", targetinfo.Updatetime, "娶了群友"),
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
					message.Text("\n今天你在", targetinfo.Updatetime, "被群友"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(targetinfo.User, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", targetinfo.Username, "]",
						"(", targetinfo.User, ")娶了",
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
				_, status, err := 民政局.查户口(gid, usr)
				if status == "错" {
					ctx.SendChain(message.Text("[qqwife]花名册数据读取有误，请重试\n", err))
					return
				}
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
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			}
			favor, err := 民政局.setFavorability(uid, fiancee, 1+rand.Intn(5))
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
			}
			// 请大家吃席
			ctx.SendChain(
				message.At(uid),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(fiancee), "]",
					"(", fiancee, ")哒\n当前你们好感度为", favor,
				),
			)
		})
	// 单身技能
	engine.OnRegex(`^(娶|嫁)\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, getdb, checkdog).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			choice := ctx.State["regex_matched"].([]string)[1]
			fiancee, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			// 写入CD
			err := 民政局.writeCDtime(gid, uid, 1)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			if uid == fiancee { // 如果是自己
				switch rand.Intn(3) {
				case 1:
					err := 民政局.登记(gid, uid, 0, "", "")
					if err != nil {
						ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
						return
					}
					ctx.SendChain(message.Text("今日获得成就：单身贵族"))
				default:
					ctx.SendChain(message.Text("今日获得成就：自恋狂"))
				}
				return
			}
			favor, err := 民政局.getFavorability(uid, fiancee)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				return
			}
			if favor < 30 {
				favor = 30 // 保底30%概率
			}
			if rand.Intn(101) >= favor {
				ctx.SendChain(message.Text(sendtext[1][rand.Intn(len(sendtext[1]))]))
				return
			}
			// 去民政局登记
			var choicetext string
			switch choice {
			case "娶":
				err := 民政局.登记(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
				if err != nil {
					ctx.SendChain(message.Text("[qqwife]结婚登记失败力\n", err))
					return
				}
				choicetext = "\n今天你的群老婆是"
			default:
				err := 民政局.登记(gid, fiancee, uid, ctx.CardOrNickName(fiancee), ctx.CardOrNickName(uid))
				if err != nil {
					ctx.SendChain(message.Text("[qqwife]结婚登记失败力\n", err))
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
	engine.OnRegex(`^当(\[CQ:at,qq=(\d+)\]\s?|(\d+))的小三`, zero.OnlyGroup, getdb, checkcp).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			fid := ctx.State["regex_matched"].([]string)
			fiancee, _ := strconv.ParseInt(fid[2]+fid[3], 10, 64)
			// 写入CD
			err := 民政局.writeCDtime(gid, uid, 2)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			if fiancee == uid {
				ctx.SendChain(message.Text("今日获得成就：自我攻略"))
				return
			}
			favor, err := 民政局.getFavorability(uid, fiancee)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				return
			}
			if favor < 30 {
				favor = 30 // 保底10%概率
			}
			if rand.Intn(101) >= favor/3 {
				ctx.SendChain(message.Text("失败了！可惜"))
				return
			}
			// 判断target是老公还是老婆
			var choicetext string
			userAID := uid     //攻的
			var userBID int64  //被牛的
			userCID := fiancee //受的
			fianceeinfo, gender, err := 民政局.查户口(gid, userCID)
			switch gender {
			case "单":
				ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
				return
			case "错":
				ctx.SendChain(message.Text("[qqwife]对象状态查询失败\n", err))
				return
			case "攻":
				err = 民政局.离婚休妻(gid, fianceeinfo.Target)
				if err != nil {
					ctx.SendChain(message.Text("ta不想和原来的对象分手...\n[error]", err))
					return
				}
				userAID = fiancee
				userCID = uid
				userBID = fianceeinfo.Target
				choicetext = "老公"
			case "受":
				err = 民政局.离婚休夫(gid, fianceeinfo.User)
				if err != nil {
					ctx.SendChain(message.Text("ta不想和原来的对象分手...\n[error]", err))
					return
				}
				userBID = fianceeinfo.User
				choicetext = "老婆"
			default:
				ctx.SendChain(message.Text("数据库发生问题力"))
				return
			}
			err = 民政局.登记(gid, userAID, userCID, ctx.CardOrNickName(userAID), ctx.CardOrNickName(userCID))
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]复婚登记失败力\n", err))
				return
			}
			favor, err = 民政局.setFavorability(userAID, userCID, -5)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
			}
			_, err = 民政局.setFavorability(userAID, userBID, 5)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
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
					"(", fiancee, ")哒\n当前你们好感度为", favor,
				),
			)
		})
	// 做媒技能
	engine.OnRegex(`^做媒\s?\[CQ:at,qq=(\d+)\]\s?\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, zero.AdminPermission, getdb, checkCondition).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			gayOne, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			gayZero, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			// 写入CD
			err := 民政局.writeCDtime(gid, uid, 3)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			favor, err := 民政局.getFavorability(gayOne, gayZero)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				return
			}
			if favor < 30 {
				favor = 30 // 保底30%概率
			}
			if rand.Intn(101) >= favor {
				_, err = 民政局.setFavorability(uid, gayOne, -1)
				if err != nil {
					ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				}
				_, err = 民政局.setFavorability(uid, gayZero, -1)
				if err != nil {
					ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				}
				ctx.SendChain(message.Text(sendtext[1][rand.Intn(len(sendtext[1]))]))
				return
			}
			// 去民政局登记
			err = 民政局.登记(gid, gayOne, gayZero, ctx.CardOrNickName(gayOne), ctx.CardOrNickName(gayZero))
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]结婚登记失败力\n", err))
				return
			}
			_, err = 民政局.setFavorability(uid, gayOne, 1)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
			}
			_, err = 民政局.setFavorability(uid, gayZero, 1)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
			}
			_, err = 民政局.setFavorability(gayOne, gayZero, 1)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
			}
			// 请大家吃席
			ctx.SendChain(
				message.At(uid),
				message.Text("恭喜你成功撮合了一对CP\n\n"),
				message.At(gayOne),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(gayZero, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(gayZero), "]",
					"(", gayZero, ")哒",
				),
			)
		})
	// 礼物系统
	engine.OnRegex(`^买礼物给\s?(\[CQ:at,qq=(\d+)\]|(\d+))`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			fiancee := ctx.State["regex_matched"].([]string)
			gay, _ := strconv.ParseInt(fiancee[2]+fiancee[3], 10, 64)
			if gay == uid {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.At(uid), message.Text("[qqwife]你想给自己买什么礼物呢?")))
				return
			}
			// 获取CD
			cdTime, err := 民政局.getCDtime(gid)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]获取该群技能CD错误(将以CD12H计算)\n", err))
			}
			ok, err := 民政局.compareCDtime(gid, uid, 5, cdTime)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]查询用户CD状态失败,请重试\n", err))
				return
			}
			if !ok {
				ctx.SendChain(message.Text("舔狗，今天你已经送过礼物了。"))
				return
			}
			// 获取好感度
			favor, err := 民政局.getFavorability(uid, gay)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				return
			}
			// 对接小熊饼干
			walletinfo := wallet.GetWalletOf(uid)
			if walletinfo < 1 {
				ctx.SendChain(message.Text("你钱包没钱啦！"))
				return
			}
			moneyToFavor := rand.Intn(math.Min(walletinfo, 100)) + 1
			// 计算钱对应的好感值
			newFavor := 1
			if favor > 50 {
				newFavor = moneyToFavor % 10 // 礼物厌倦
			} else {
				newFavor += rand.Intn(moneyToFavor)
			}
			// 随机对方心情
			mood := rand.Intn(2)
			if mood == 0 {
				newFavor = -newFavor
			}
			// 记录结果
			err = wallet.InsertWalletOf(uid, -moneyToFavor)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]钱包坏掉力:\n", err))
				return
			}
			lastfavor, err := 民政局.setFavorability(uid, gay, newFavor)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度数据库发生问题力\n", err))
				return
			}
			// 写入CD
			err = 民政局.writeCDtime(gid, uid, 5)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			// 输出结果
			if mood == 0 {
				ctx.SendChain(message.Text("你花了", moneyToFavor, "ATRI币买了一件女装送给了ta,ta很不喜欢,你们的好感度降低至", lastfavor))
			} else {
				ctx.SendChain(message.Text("你花了", moneyToFavor, "ATRI币买了一件女装送给了ta,ta很喜欢,你们的好感度升至", lastfavor))
			}
		})
	engine.OnFullMatchGroup([]string{"闹离婚", "办离婚"}, zero.OnlyGroup, getdb, checkdivorce).Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			// 写入CD
			err := 民政局.writeCDtime(gid, uid, 4)
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[qqwife]你的技能CD记录失败\n", err))
			}
			info, uidstatus, err := 民政局.查户口(gid, uid)
			mun := 2
			var fiancee int64
			switch uidstatus {
			case "错":
				ctx.SendChain(message.Text("[qqwife]用户状态查询失败\n", err))
				return
			case "攻":
				mun = 1
				fiancee = info.Target
			case "受":
				mun = 0
				fiancee = info.User
			}
			favor, err := 民政局.getFavorability(uid, fiancee)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]好感度库发生问题力\n", err))
				return
			}
			if favor < 20 {
				favor = 10
			}
			if rand.Intn(101) > 100-favor {
				ctx.SendChain(message.Text(sendtext[3][rand.Intn(len(sendtext[3]))]))
				return
			}
			switch mun {
			case 1:
				err = 民政局.离婚休妻(gid, fiancee)
			case 0:
				err = 民政局.离婚休夫(gid, fiancee)
			default:
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			}
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			}
			ctx.SendChain(message.Text(sendtext[4][mun]))
		})
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			ok, err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			}
			if ok {
				ctx.SendChain(message.Text("今天还没有人结婚哦"))
				return
			}
			list, number, err := 民政局.花名册(gid)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
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
			_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize*2); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("群老婆列表")
			/***********绘制标题***********/
			canvas.DrawString("群老婆列表", (1500-sl)/2, 160-h) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 250-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
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
	engine.OnFullMatch("好感度列表", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			fianceeInfo, err := 民政局.getGroupFavorability(uid)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			/***********设置图片的大小和底色***********/
			number := len(fianceeInfo)
			if number > 10 {
				number = 10
			}
			fontSize := 50.0
			canvas := gg.NewContext(1150, int(170+(50+70)*float64(number)))
			canvas.SetRGB(1, 1, 1) // 白色
			canvas.Clear()
			/***********下载字体***********/
			_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize*2); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("你的好感度排行列表")
			/***********绘制标题***********/
			canvas.DrawString("你的好感度排行列表", (1100-sl)/2, 100) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 160)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			i := 0
			for _, info := range fianceeInfo {
				if i > 9 {
					break
				}
				if info.Userinfo == "" {
					continue
				}
				fianceID, err := strconv.ParseInt(info.Userinfo, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
					return
				}
				if fianceID == 0 {
					continue
				}
				userName := ctx.CardOrNickName(fianceID)
				canvas.SetRGB255(0, 0, 0)
				canvas.DrawString(userName+"("+info.Userinfo+")", 10, float64(180+(50+70)*i))
				canvas.DrawString(strconv.Itoa(info.Favor), 1020, float64(180+60+(50+70)*i))
				canvas.DrawRectangle(10, float64(180+60+(50+70)*i)-h/2, 1000, 50)
				canvas.SetRGB255(150, 150, 150)
				canvas.Fill()
				canvas.SetRGB255(0, 0, 0)
				canvas.DrawRectangle(10, float64(180+60+(50+70)*i)-h/2, float64(info.Favor)*10, 50)
				canvas.SetRGB255(231, 27, 100)
				canvas.Fill()
				i++
			}
			data, cl := writer.ToBytes(canvas.Image())
			ctx.SendChain(message.ImageBytes(data))
			cl()
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
				ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
				return
			}
			ctx.SendChain(message.Text("重置成功"))
		})
}
