// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品，文案采用了Hana的zbp娶群友文案
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/imgfactory"
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
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/zbputils/img/text"
)

type 婚姻登记 struct {
	db *sql.Sqlite
	sync.RWMutex
}

// 群设置
type updateinfo struct {
	GID        int64
	Updatetime string  // 登记时间
	CanMatch   int     // 嫁婚开关
	CanNtr     int     // Ntr开关
	CDtime     float64 // CD时间
}

// 结婚证信息
type userinfo struct {
	User       int64  // 用户身份证
	Target     int64  // 对象身份证号
	Username   string // 户主名称
	Targetname string // 对象名称
	Updatetime string // 登记时间

}

var (
	民政局 = &婚姻登记{
		db: &sql.Sqlite{},
	}
	engine = control.Register("qqwife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "一群一天一夫一妻制群老婆",
		Help: "- 娶群友\n- 群老婆列表\n- [允许|禁止]自由恋爱\n- [允许|禁止]牛头人\n- 设置CD为xx小时    →(默认12小时)\n- 重置花名册\n- 重置所有花名册(用于清除所有群数据及其设置)\n- 查好感度[对方Q号|@对方QQ]\n- 好感度列表\n- 好感度数据整理 (当好感度列表出现重复名字时使用)\n" +
			"--------------------------------\n以下指令存在CD,不跨天刷新,前两个受指令开关\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n自由选择对象, 自由恋爱(好感度越高成功率越高,保底30%概率)\n" +
			"- 当[对方Q号|@对方QQ]的小三\n我和你才是真爱, 为了你我愿意付出一切(好感度越高成功率越高,保底10%概率)\n" +
			"- 闹离婚\n你谁啊, 给我滚(好感度越高成功率越低)\n" +
			"- 买礼物给[对方Q号|@对方QQ]\n使用小熊饼干获取好感度\n" +
			"- 做媒 @攻方QQ @受方QQ\n身为管理, 群友的xing福是要搭把手的(攻受双方好感度越高成功率越高,保底30%概率)\n" +
			"--------------------------------\n好感度规则\n--------------------------------\n" +
			"\"娶群友\"指令好感度随机增加1~5。\n\"A牛B的C\"会导致C恨A, 好感度-5;\nB为了报复A, 好感度+5(什么柜子play)\nA为BC做媒,成功B、C对A好感度+1反之-1\n做媒成功BC好感度+1" +
			"\nTips: 群老婆列表过0点刷新",
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
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		民政局.db.DBPath = engine.DataFolder() + "结婚登记表.db"
		err := 民政局.db.Open(time.Hour * 24)
		if err == nil {
			// 创建群配置表
			err = 民政局.db.Create("updateinfo", &updateinfo{})
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return false
			}
			// 创建CD表
			err = 民政局.db.Create("cdsheet", &cdsheet{})
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return false
			}
			// 创建好感度表
			err = 民政局.db.Create("favorability", &favorability{})
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return false
			}
			return true
		}
		ctx.SendChain(message.Text("[ERROR]:", err))
		return false
	})
)

func init() {
	engine.OnFullMatch("娶群友", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			uid := ctx.Event.UserID
			userInfo, _ := 民政局.查户口(gid, uid)
			switch {
			case userInfo != (userinfo{}) && (userInfo.Target == 0 || userInfo.User == 0): // 如果是单身贵族
				ctx.SendChain(message.Text("今天你是单身贵族噢"))
				return
			case userInfo.User == uid: // 娶过别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你在", userInfo.Updatetime, "娶了群友"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(userInfo.Target, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", userInfo.Targetname, "]",
						"(", userInfo.Target, ")哒",
					),
				)
				return
			case userInfo.Target == uid: // 嫁给别人
				ctx.SendChain(
					message.At(uid),
					message.Text("\n今天你在", userInfo.Updatetime, "被群友"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(userInfo.User, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", userInfo.Username, "]",
						"(", userInfo.User, ")娶了",
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
				usrInfo, _ := 民政局.查户口(gid, usr)
				if usrInfo != (userinfo{}) {
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
				switch rand.Intn(10) {
				case 1:
					err := 民政局.登记(gid, uid, 0, "", "")
					if err != nil {
						ctx.SendChain(message.Text("[ERROR]:", err))
						return
					}
					ctx.SendChain(message.Text("今日获得成就：单身贵族"))
				default:
					ctx.SendChain(message.Text("呜...没娶到，你可以再尝试一次"))
					return
				}
			}
			// 去民政局办证
			err = 民政局.登记(gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(fiancee))
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			favor, err := 民政局.更新好感度(uid, fiancee, 1+rand.Intn(5))
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
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
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			err := 民政局.开门时间(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			list, err := 民政局.花名册(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			number := len(list)
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
			data, err := file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(data, fontSize*2); err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("群老婆列表")
			/***********绘制标题***********/
			canvas.DrawString("群老婆列表", (1500-sl)/2, 160-h) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 250-h)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.ParseFontFace(data, fontSize); err != nil {
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
			data, err = imgfactory.ToBytes(canvas.Image())
			if err != nil {
				ctx.SendChain(message.Text("[qqwife]ERROR: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnRegex(`^重置(所有|本群|/d+)?花名册$`, zero.SuperUserPermission, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			var err error
			switch ctx.State["regex_matched"].([]string)[1] {
			case "所有":
				err = 民政局.清理花名册()
			case "本群", "":
				if ctx.Event.GroupID == 0 {
					ctx.SendChain(message.Text("该功能只能在群组使用或者指定群组"))
					return
				}
				err = 民政局.清理花名册("group" + strconv.FormatInt(ctx.Event.GroupID, 10))
			default:
				cmd := ctx.State["regex_matched"].([]string)[1]
				gid, _ := strconv.ParseInt(cmd, 10, 64) // 判断是否为群号
				if gid == 0 {
					ctx.SendChain(message.Text("请输入正确的群号"))
					return
				}
				err = 民政局.清理花名册("group" + cmd)
			}
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ctx.SendChain(message.Text("重置成功"))
		})
}

func (sql *婚姻登记) 查看设置(gid int64) (dbinfo updateinfo, err error) {
	sql.Lock()
	defer sql.Unlock()
	// 创建群表格
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		return
	}
	if !sql.db.CanFind("updateinfo", "where gid is "+strconv.FormatInt(gid, 10)) {
		// 没有记录
		return updateinfo{
			GID:      gid,
			CanMatch: 1,
			CanNtr:   1,
			CDtime:   12,
		}, nil
	}
	_ = sql.db.Find("updateinfo", &dbinfo, "where gid is "+strconv.FormatInt(gid, 10))
	return
}

func (sql *婚姻登记) 更新设置(dbinfo updateinfo) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Insert("updateinfo", &dbinfo)
}

func (sql *婚姻登记) 开门时间(gid int64) error {
	grouInfo, err := sql.查看设置(gid)
	if err != nil {
		return err
	}
	sql.Lock()
	defer sql.Unlock()
	dbinfo := updateinfo{}
	_ = sql.db.Find("updateinfo", &dbinfo, "where gid is "+strconv.FormatInt(gid, 10))
	if time.Now().Format("2006/01/02") != dbinfo.Updatetime {
		// 如果跨天了就删除
		_ = sql.db.Drop("group" + strconv.FormatInt(gid, 10))
		// 更新数据时间
		grouInfo.GID = gid
		grouInfo.Updatetime = time.Now().Format("2006/01/02")
		return sql.db.Insert("updateinfo", &grouInfo)
	}
	return nil
}

func (sql *婚姻登记) 查户口(gid, uid int64) (info userinfo, err error) {
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
	// 创建群表格
	err = sql.db.Create(gidstr, &userinfo{})
	if err != nil {
		return
	}
	uidstr := strconv.FormatInt(uid, 10)
	err = sql.db.Find(gidstr, &info, "where user = "+uidstr)
	if err != nil {
		err = sql.db.Find(gidstr, &info, "where target = "+uidstr)
	}
	return
}

// 民政局登记数据
func (sql *婚姻登记) 登记(gid, uid, target int64, username, targetname string) error {
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
	uidinfo := userinfo{
		User:       uid,
		Username:   username,
		Target:     target,
		Targetname: targetname,
		Updatetime: time.Now().Format("15:04:05"),
	}
	return sql.db.Insert(gidstr, &uidinfo)
}

func (sql *婚姻登记) 花名册(gid int64) (list [][4]string, err error) {
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
	number, _ := sql.db.Count(gidstr)
	if number <= 0 {
		return
	}
	var info userinfo
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

func (sql *婚姻登记) 清理花名册(gid ...string) error {
	sql.Lock()
	defer sql.Unlock()
	switch gid {
	case nil:
		grouplist, err := sql.db.ListTables()
		if err == nil {
			for _, listName := range grouplist {
				if listName == "favorability" {
					continue
				}
				err = sql.db.Drop(listName)
			}
		}
		return err
	default:
		err := sql.db.Drop(gid[0])
		if err == nil {
			_ = sql.db.Del("cdsheet", "where GroupID is "+strings.ReplaceAll(gid[0], "group", ""))
		}
		return err
	}
}
