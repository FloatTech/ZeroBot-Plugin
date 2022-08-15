// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品，文案采用了Hana的zbp娶群友文案
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"

	// 数据库
	sql "github.com/FloatTech/sqlite"
	// 定时器

	"github.com/wdvxdr1123/ZeroBot/extension/rate"

	// 画图
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/fogleman/gg"
)

// nolint: asciicheck
type 婚姻登记 struct {
	db   *sql.Sqlite
	dbmu sync.RWMutex
}

// 结婚证信息
type userinfo struct {
	User       int64  // 用户身份证
	Target     int64  // 对象身份证号
	Username   string // 户主名称
	Targetname string // 对象名称
	Updatetime string // 登记时间

}

// 民政局的当前时间
type updateinfo struct {
	GID        int64
	Updatetime string // 登记时间

}

func (sql *婚姻登记) checkupdate(gid int64) (updatetime string, err error) {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		return
	}
	gidstr := strconv.FormatInt(gid, 10)
	dbinfo := updateinfo{}
	err = sql.db.Find("updateinfo", &dbinfo, "where gid is "+gidstr) // 获取表格更新的时间
	if err != nil {
		updatetime = time.Now().Format("2006/01/02")
		err = sql.db.Insert("updateinfo", &updateinfo{GID: gid, Updatetime: updatetime})
		return
	}
	updatetime = dbinfo.Updatetime
	return
}

func (sql *婚姻登记) 重置(gid string) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	if gid != "ALL" {
		err := sql.db.Drop(gid)
		if err != nil {
			err = sql.db.Create(gid, &userinfo{})
			return err
		}
		gidint, _ := strconv.ParseInt(gid, 10, 64)
		updateinfo := updateinfo{
			GID:        gidint,
			Updatetime: time.Now().Format("2006/01/02"),
		}
		err = sql.db.Insert("updateinfo", &updateinfo)
		return err
	}
	grouplist, err := sql.db.ListTables()
	if err != nil {
		return err
	}
	for _, gid := range grouplist {
		err = sql.db.Drop(gid)
		if err != nil {
			continue
		}
		gidint, _ := strconv.ParseInt(gid, 10, 64)
		updateinfo := updateinfo{
			GID:        gidint,
			Updatetime: time.Now().Format("2006/01/02"),
		}
		err = sql.db.Insert("updateinfo", &updateinfo)
	}
	return err
}

func (sql *婚姻登记) 离婚休妻(gid, wife int64) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	wifestr := strconv.FormatInt(wife, 10)
	// 先判断用户是否存在
	err := sql.db.Del(gidstr, "where target = "+wifestr)
	return err
}

func (sql *婚姻登记) 离婚休夫(gid, husband int64) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	husbandstr := strconv.FormatInt(husband, 10)
	// 先判断用户是否存在
	err := sql.db.Del(gidstr, "where user = "+husbandstr)
	return err
}

func (sql *婚姻登记) 复婚(gid, uid, target int64, username, targetname string) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	uidstr := strconv.FormatInt(uid, 10)
	tagstr := strconv.FormatInt(target, 10)
	var info userinfo
	err := sql.db.Find(gidstr, &info, "where user = "+uidstr)
	if err != nil {
		err = sql.db.Find(gidstr, &info, "where user = "+tagstr)
	}
	if err != nil {
		return err
	}
	updatetime := time.Now().Format("2006/01/02")
	// 更改夫妻信息
	info.User = uid
	info.Username = username
	info.Target = target
	info.Targetname = targetname
	info.Updatetime = updatetime
	// 民政局登记数据
	err = sql.db.Insert(gidstr, &info)
	return err
}

func (sql *婚姻登记) 花名册(gid int64) (list [][4]string, number int, err error) {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	err = sql.db.Create(gidstr, &userinfo{})
	if err != nil {
		return
	}
	number, err = sql.db.Count(gidstr)
	if err != nil || number <= 0 {
		return
	}
	var info userinfo
	list = make([][4]string, 0, number)
	err = sql.db.FindFor(gidstr, &info, "GROUP BY user", func() error {
		if info.Target == 0 {
			return nil
		}
		dbinfo := [4]string{
			info.Username,
			strconv.FormatInt(info.User, 10),
			info.Targetname,
			strconv.FormatInt(info.Target, 10),
		}
		list = append(list, dbinfo)
		return nil
	})
	if len(list) == 0 {
		number = 0
	}
	return
}

func slicename(name string, canvas *gg.Context) (resultname string) {
	usermane := []rune(name) // 将每个字符单独放置
	widthlen := 0
	numberlen := 0
	for i, v := range usermane {
		width, _ := canvas.MeasureString(string(v)) // 获取单个字符的宽度
		widthlen += int(width)
		if widthlen > 350 {
			break // 总宽度不能超过350
		}
		numberlen = i
	}
	if widthlen > 350 {
		resultname = string(usermane[:numberlen-1]) + "......" // 名字切片
	} else {
		resultname = name
	}
	return
}

func (sql *婚姻登记) 查户口(gid, uid int64) (info userinfo, status int, err error) {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	uidstr := strconv.FormatInt(uid, 10)
	status = 3
	if err = sql.db.Create(gidstr, &userinfo{}); err != nil {
		status = 2
		return
	}
	err = sql.db.Find(gidstr, &info, "where user = "+uidstr)
	if err == nil {
		status = 1
		return
	}
	err = sql.db.Find(gidstr, &info, "where target = "+uidstr)
	if err == nil {
		status = 0
		return
	}
	return
}

func (sql *婚姻登记) 登记(gid, uid, target int64, username, targetname string) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	err := sql.db.Create(gidstr, &userinfo{})
	if err != nil {
		return err
	}
	updatetime := time.Now().Format("2006/01/02")
	// 填写夫妻信息
	uidinfo := userinfo{
		User:       uid,
		Username:   username,
		Target:     target,
		Targetname: targetname,
		Updatetime: updatetime,
	}
	// 民政局登记数据
	err = sql.db.Insert(gidstr, &uidinfo)
	return err
}

var (
	//nolint: asciicheck
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
			"- 娶群友\n- 群老婆列表\n" +
			"--------------------------------\n以下技能每人只能三选一\n   CD12H，不跨天刷新\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n- 当[对方Q号|@对方QQ]的小三\n- 闹离婚",
	})
	getdb := ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		民政局.db.DBPath = engine.DataFolder() + "结婚登记表.db"
		// 如果数据库不存在则下载
		// _, _ = engine.GetLazyData("结婚登记表.db", false)

		err := 民政局.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
			return false
		}
		return true
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			updatetime, err := 民政局.checkupdate(gid)
			switch {
			case err != nil:
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
				return
			case time.Now().Format("2006/01/02") != updatetime:
				if err := 民政局.重置(strconv.FormatInt(gid, 10)); err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
			}
			uid := ctx.Event.UserID
			targetinfo, status, err := 民政局.查户口(gid, uid)
			switch {
			case status == 2:
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
				return
			case status != 3 && targetinfo.Target == 0: // 如果为单身贵族
				ctx.SendChain(message.Text("今天你是单身贵族噢"))
				return
			case status == 1: // 娶过别人
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
			case status == 0: // 嫁给别人
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
				if status != 3 {
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
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
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
						ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
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
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
				choicetext = "\n今天你的群老婆是"
			default:
				err := 民政局.登记(gid, fiancee, uid, ctx.CardOrNickName(fiancee), ctx.CardOrNickName(uid))
				if err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
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
	engine.OnRegex(`^当(\[CQ:at,qq=(\d+)\]\s?|(\d+))的小三`, zero.OnlyGroup, getdb, checkcp).SetBlock(true).Limit(cdcheck, iscding).
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
			_, gender, err := 民政局.查户口(gid, fiancee)
			switch gender {
			case 3:
				ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
				return
			case 2:
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
				return
			case 1:
				// 和对象结婚登记
				err = 民政局.复婚(gid, fiancee, uid, ctx.CardOrNickName(fiancee), ctx.CardOrNickName(uid))
				if err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
				choicetext = "老公"
			case 0:
				// 和对象结婚登记
				err = 民政局.复婚(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
				if err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
				choicetext = "老婆"
			default:
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员"))
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
			updatetime, err := 民政局.checkupdate(gid)
			switch {
			case err != nil:
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
				return
			case time.Now().Format("2006/01/02") != updatetime:
				if err := 民政局.重置(strconv.FormatInt(gid, 10)); err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
				ctx.SendChain(message.Text("今天还没有人结婚哦"))
				return
			}
			list, number, err := 民政局.花名册(gid)
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
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
				ctx.SendChain(message.Text("ERROR:", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize*2); err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			sl, h := canvas.MeasureString("群老婆列表")
			/***********绘制标题***********/
			canvas.DrawString("群老婆列表", (1500-sl)/2, 160-h) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 250-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
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
	engine.OnFullMatchGroup([]string{"闹离婚", "办离婚"}, zero.OnlyGroup, getdb, checkfiancee).SetBlock(true).Limit(cdcheck, iscding2).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			info, uidstatus, err := 民政局.查户口(gid, uid)
			switch uidstatus {
			case 2:
				ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
				return
			case 1:
				if rand.Intn(10) != 1 { // 十分之一的概率成功
					ctx.SendChain(message.Text(sendtext[3][rand.Intn(len(sendtext[3]))]))
					return
				}
				err := 民政局.离婚休妻(gid, info.Target)
				if err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
				ctx.SendChain(message.Text(sendtext[4][0]))
			case 0:
				if rand.Intn(10) != 0 { // 十分之一的概率成功
					ctx.SendChain(message.Text(sendtext[3][rand.Intn(len(sendtext[3]))]))
					return
				}
				err := 民政局.离婚休夫(gid, info.User)
				if err != nil {
					ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
					return
				}
				ctx.SendChain(message.Text(sendtext[4][1]))
			}
		})
	engine.OnRegex(`^重置(所有|本群|/d+)?花名册$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			cmd := "ALL"
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
			err := 民政局.重置(cmd)
			if err != nil {
				ctx.SendChain(message.Text("数据库发生问题力\n[error]", err))
				return
			}
			ctx.SendChain(message.Text("重置成功"))
		})
}

// 以群号和昵称为限制
func cdcheck(ctx *zero.Ctx) *rate.Limiter {
	limitID := strconv.FormatInt(ctx.Event.GroupID, 10) + strconv.FormatInt(ctx.Event.UserID, 10)
	return skillCD.Load(limitID)
}
func iscding(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("你的技能现在正在CD中"))
}

// 注入判断 是否为单身
func checkdog(ctx *zero.Ctx) bool {
	// 得先判断用户是否存在才行在，再重置
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的target好像不存在？"))
		return false
	}
	// 判断是否需要重置
	gid := ctx.Event.GroupID
	updatetime, err := 民政局.checkupdate(gid)
	switch {
	case err != nil:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		return false
	case time.Now().Format("2006/01/02") != updatetime:
		if err := 民政局.重置(strconv.FormatInt(gid, 10)); err != nil {
			ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
			return false
		}
		return true // 重置后也全是单身
	}
	uid := ctx.Event.UserID
	// 获取用户信息
	uidtarget, uidstatus, err1 := 民政局.查户口(gid, uid)
	fianceeinfo, fianceestatus, err2 := 民政局.查户口(gid, fiancee)
	switch {
	case uidstatus == 2 || fianceestatus == 2:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err1, "\n", err2))
		return false
	case uidstatus == 3 && fianceestatus == 3: // 必须是两个单身
		return true
	case uidtarget.Target == fiancee: // 如果本就是一块
		ctx.SendChain(message.Text("笨蛋~你们明明已经在一起了啊w"))
		return false
	case uidstatus != 3 && uidtarget.Target == 0: // 如果是单身贵族
		ctx.SendChain(message.Text("今天的你是单身贵族噢"))
		return false
	case uidstatus == 1: // 如果如为攻
		ctx.SendChain(message.Text("笨蛋~你家里还有个吃白饭的w"))
		return false
	case uidstatus == 0: // 如果为受
		ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		return false
	case fianceestatus != 3 && fianceeinfo.Target == 0:
		ctx.SendChain(message.Text("今天的ta是单身贵族噢"))
		return false
	case fianceestatus == 1: // 如果如为攻
		ctx.SendChain(message.Text("他有别的女人了，你该放下了"))
		return false
	case fianceestatus == 0: // 如果为受
		ctx.SendChain(message.Text("这是一个纯爱的世界，拒绝NTR"))
		return false
	}
	return true
}

// 注入判断 是否满足小三要求
func checkcp(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	updatetime, err := 民政局.checkupdate(gid)
	switch {
	case err != nil:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		return false
	case time.Now().Format("2006/01/02") != updatetime:
		if err := 民政局.重置(strconv.FormatInt(gid, 10)); err != nil {
			ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		} else {
			ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
		}
		return false // 重置后全是单身
	}
	// 检查target
	fid := ctx.State["regex_matched"].([]string)
	fiancee, err := strconv.ParseInt(fid[2]+fid[3], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的对象好像不存在?"))
		return false
	}
	// 检查用户是否登记过
	uid := ctx.Event.UserID
	userinfo, uidstatus, err := 民政局.查户口(gid, uid)
	switch {
	case uidstatus == 2:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		return false
	case userinfo.Target == fiancee: // 如果本就是一块
		ctx.SendChain(message.Text("笨蛋~你们明明已经在一起了啊w"))
		return false
	case uidstatus != 3 && userinfo.Target == 0: // 如果是单身贵族
		ctx.SendChain(message.Text("今天的你是单身贵族哦"))
		return false
	case fiancee == uid: // 自我攻略
		return true
	case uidstatus == 1: // 如果如为攻
		ctx.SendChain(message.Text("打灭，不给纳小妾！"))
		return false
	case uidstatus == 0: // 如果为受
		ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		return false
	}
	fianceeinfo, fianceestatus, err := 民政局.查户口(gid, fiancee)
	switch {
	case fianceestatus == 2:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		return false
	case fianceestatus == 3:
		ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
		return false
	case fianceeinfo.Target == 0:
		ctx.SendChain(message.Text("今天的ta是单身贵族哦"))
		return false
	}
	return true
}

func checkfiancee(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	updatetime, err := 民政局.checkupdate(gid)
	switch {
	case err != nil:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		return false
	case time.Now().Format("2006/01/02") != updatetime:
		if err := 民政局.重置(strconv.FormatInt(gid, 10)); err != nil {
			ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
			return false
		}
		ctx.SendChain(message.Text("今天你还没有结婚哦"))
		return false
	}
	// 获取用户信息
	uid := ctx.Event.UserID
	_, uidstatus, err := 民政局.查户口(gid, uid)
	switch uidstatus {
	case 2:
		ctx.SendChain(message.Text("数据库发生问题力，请联系bot管理员\n[error]", err))
		return false
	case 3: // 如果是单身
		ctx.SendChain(message.Text("今天你还没有结婚哦"))
		return false
	}
	return true
}

func iscding2(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("打灭，禁止离婚  (你的技能正在CD中)"))
}
