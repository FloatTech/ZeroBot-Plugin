package qqwife

import (
	"strconv"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 画图
	"github.com/Coloured-glaze/gg"
)

// nolint: asciicheck
//nolint: asciicheck
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

func (sql *婚姻登记) 开门时间(gid int64) (ok bool, err error) {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	ok = false
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		return
	}
	gidstr := strconv.FormatInt(gid, 10)
	dbinfo := updateinfo{}
	// 获取表格更新的时间
	err = sql.db.Find("updateinfo", &dbinfo, "where gid is "+gidstr)
	if err != nil {
		// 如果没有登记过就记录
		err = sql.db.Insert("updateinfo", &updateinfo{
			GID:        gid,
			Updatetime: time.Now().Format("2006/01/02"),
		})
		if err == nil {
			ok = true
		}
		return
	}
	// 开门了就拿新的花名册
	if time.Now().Format("2006/01/02") == dbinfo.Updatetime {
		return
	}
	err = sql.db.Drop(gidstr)
	if err != nil {
		return
	}
	updateinfo := updateinfo{
		GID:        gid,
		Updatetime: time.Now().Format("2006/01/02"),
	}
	err = sql.db.Insert("updateinfo", &updateinfo)
	if err == nil {
		ok = true
	}
	return
}

func (sql *婚姻登记) 清理花名册(gid string) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	grouplist, err := sql.db.ListTables()
	if err != nil {
		return err
	}
	if gid != "0" {
		grouplist = []string{gid}
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

func (sql *婚姻登记) 查户口(gid, uid int64) (info userinfo, status string, err error) {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	uidstr := strconv.FormatInt(uid, 10)
	status = "单"
	err = sql.db.Create(gidstr, &userinfo{})
	if err != nil {
		status = "错"
		return
	}
	err = sql.db.Find(gidstr, &info, "where user = "+uidstr)
	if err == nil {
		status = "攻"
		return
	}
	err = sql.db.Find(gidstr, &info, "where target = "+uidstr)
	if err == nil {
		status = "受"
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

func (sql *婚姻登记) 离婚休妻(gid, wife int64) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	wifestr := strconv.FormatInt(wife, 10)
	return sql.db.Del(gidstr, "where target = "+wifestr)
}

func (sql *婚姻登记) 离婚休夫(gid, husband int64) error {
	sql.dbmu.Lock()
	defer sql.dbmu.Unlock()
	gidstr := strconv.FormatInt(gid, 10)
	husbandstr := strconv.FormatInt(husband, 10)
	return sql.db.Del(gidstr, "where user = "+husbandstr)
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

// 以群号和昵称为限制
func cdcheck(ctx *zero.Ctx) *rate.Limiter {
	limitID := strconv.FormatInt(ctx.Event.GroupID, 10) + strconv.FormatInt(ctx.Event.UserID, 10) + "1"
	return skillCD.Load(limitID)
}
func cdcheck2(ctx *zero.Ctx) *rate.Limiter {
	limitID := strconv.FormatInt(ctx.Event.GroupID, 10) + strconv.FormatInt(ctx.Event.UserID, 10) + "2"
	return skillCD.Load(limitID)
}
func cdcheck3(ctx *zero.Ctx) *rate.Limiter {
	limitID := strconv.FormatInt(ctx.Event.GroupID, 10) + strconv.FormatInt(ctx.Event.UserID, 10) + "3"
	return skillCD.Load(limitID)
}
func iscding(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("你的技能现在正在CD中"))
}
func iscding2(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("打灭，禁止离婚  (你的技能正在CD中)"))
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
	ok, err := 民政局.开门时间(gid)
	if err != nil {
		ctx.SendChain(message.Text("群状态查询失败\n[error]", err))
		return false
	}
	if ok {
		return true // 重置后也全是单身
	}
	// 获取用户信息
	uid := ctx.Event.UserID
	uidtarget, uidstatus, err := 民政局.查户口(gid, uid)
	switch {
	case uidstatus == "错":
		ctx.SendChain(message.Text("用户状态查询失败\n[error]", err))
		return false
	case uidstatus != "单" && (uidtarget.Target == 0 || uidtarget.User == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的你是单身贵族噢"))
		return false
	case (uidstatus == "攻" && uidtarget.Target == fiancee) ||
		(uidstatus == "受" && uidtarget.User == fiancee):
		ctx.SendChain(message.Text("笨蛋！你们已经在一起了！"))
		return false
	case uidstatus == "攻": // 如果如为攻
		ctx.SendChain(message.Text("笨蛋~你家里还有个吃白饭的w"))
		return false
	case uidstatus == "受": // 如果为受
		ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		return false
	}
	fianceeinfo, fianceestatus, err := 民政局.查户口(gid, fiancee)
	switch {
	case fianceestatus == "错":
		ctx.SendChain(message.Text("对象状态查询失败\n[error]", err))
	case fianceestatus == "单": // 如果为单身狗
		return true
	case fianceestatus != "单" && (fianceeinfo.Target == 0 || fianceeinfo.User == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的ta是单身贵族噢"))
	case fianceestatus == "攻": // 如果如为攻
		ctx.SendChain(message.Text("他有别的女人了，你该放下了"))
	case fianceestatus == "受": // 如果为受
		ctx.SendChain(message.Text("ta被别人娶了，你来晚力"))
	}
	return false
}

// 注入判断 是否满足小三要求
func checkcp(ctx *zero.Ctx) bool {
	// 得先判断用户是否存在才行在，再重置
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的target好像不存在？"))
		return false
	}
	// 判断是否需要重置
	gid := ctx.Event.GroupID
	ok, err := 民政局.开门时间(gid)
	if err != nil {
		ctx.SendChain(message.Text("群状态查询失败\n[error]", err))
		return false
	}
	if ok {
		ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
		return false // 重置后也全是单身
	}
	uid := ctx.Event.UserID
	fianceeinfo, fianceestatus, err := 民政局.查户口(gid, fiancee)
	switch {
	case fianceestatus == "错":
		ctx.SendChain(message.Text("对象状态查询失败\n[error]", err))
		return false
	case fianceestatus == "单": // 如果为单身狗
		if fiancee == uid {
			return true
		}
		ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
		return false
	case fianceestatus != "单" && (fianceeinfo.Target == 0 || fianceeinfo.User == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的ta是单身贵族噢"))
		return false
	case (fianceestatus == "攻" && fianceeinfo.Target == fiancee) ||
		(fianceestatus == "受" && fianceeinfo.User == fiancee):
		ctx.SendChain(message.Text("笨蛋！你们已经在一起了！"))
		return false
	}
	// 获取用户信息
	uidtarget, uidstatus, err := 民政局.查户口(gid, uid)
	switch {
	case uidstatus == "错":
		ctx.SendChain(message.Text("用户状态查询失败\n[error]", err))
	case uidstatus == "单": // 如果为单身狗
		return true
	case uidstatus != "单" && (uidtarget.Target == 0 || uidtarget.User == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的你是单身贵族噢"))
	case uidstatus == "攻": // 如果如为攻
		ctx.SendChain(message.Text("打灭，不给纳小妾！"))
	case uidstatus == "受": // 如果为受
		ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
	}
	return false
}
