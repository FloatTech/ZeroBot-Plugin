package qqwife

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 画图
	"github.com/Coloured-glaze/gg"
)

// nolint: asciicheck
// nolint: asciicheck
type 婚姻登记 struct {
	db *sql.Sqlite
	sync.RWMutex
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
	Updatetime string  // 登记时间
	CanMatch   int     // 订婚开关
	CanNtr     int     // Ntr技能开关
	CDtime     float64 // CD时间
}

// 好感度系统
type favorability struct {
	Userinfo string // 记录用户
	Favor    int    // 好感度
}

// 技能CD记录表
type cdsheet struct {
	Time    int64 // 时间
	GroupID int64 // 群号
	UserID  int64 // 用户
	ModeID  int64 // 技能类型
}

func (sql *婚姻登记) 开门时间(gid int64) (ok bool, err error) {
	sql.Lock()
	defer sql.Unlock()
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
			CanMatch:   1,
			CanNtr:     1,
			CDtime:     12,
		})
		if err == nil {
			ok = true
		}
		return
	}
	if time.Now().Format("2006/01/02") == dbinfo.Updatetime {
		return
	}
	// 开门了就拿新的花名册
	err = sql.db.Drop("group" + gidstr)
	if err != nil {
		if err = sql.db.Create("group"+gidstr, &userinfo{}); err != nil {
			return
		}
	}
	dbinfo.Updatetime = time.Now().Format("2006/01/02")
	err = sql.db.Insert("updateinfo", &dbinfo)
	if err == nil {
		ok = true
	}
	return
}

func (sql *婚姻登记) 营业模式(gid int64) (canMatch, canNtr int, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		if err = sql.db.Drop("updateinfo"); err == nil {
			err = sql.db.Create("updateinfo", &updateinfo{})
		}
		if err != nil {
			return
		}
	}
	gidstr := strconv.FormatInt(gid, 10)
	dbinfo := updateinfo{}
	err = sql.db.Find("updateinfo", &dbinfo, "where gid is "+gidstr)
	if err != nil {
		canMatch = 1
		canNtr = 1
		err = sql.db.Insert("updateinfo", &updateinfo{
			GID:      gid,
			CanMatch: canMatch,
			CanNtr:   canNtr,
			CDtime:   12,
		})
		return
	}
	canMatch = dbinfo.CanMatch
	canNtr = dbinfo.CanNtr
	return
}

func (sql *婚姻登记) 修改模式(gid int64, mode string, stauts int) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		if err = sql.db.Drop("updateinfo"); err == nil {
			err = sql.db.Create("updateinfo", &updateinfo{})
		}
		if err != nil {
			return
		}
	}
	gidstr := strconv.FormatInt(gid, 10)
	dbinfo := updateinfo{}
	err = sql.db.Find("updateinfo", &dbinfo, "where gid is "+gidstr)
	switch mode {
	case "自由恋爱":
		dbinfo.CanMatch = stauts
	case "牛头人":
		dbinfo.CanNtr = stauts
	default:
		return errors.New("错误:修改内容不匹配！")
	}
	if err != nil {
		dbinfo.GID = gid
		switch mode {
		case "自由恋爱":
			dbinfo.CanNtr = 1
		case "牛头人":
			dbinfo.CanMatch = 1
		}
		dbinfo.CDtime = 12
		err = sql.db.Insert("updateinfo", &dbinfo)
		return
	}
	err = sql.db.Insert("updateinfo", &dbinfo)
	return
}

func (sql *婚姻登记) 清理花名册(gid string) error {
	sql.Lock()
	defer sql.Unlock()
	grouplist, err := sql.db.ListTables()
	if err != nil {
		return err
	}
	if gid != "0" {
		grouplist = []string{"group" + gid}
	}
	for _, gid := range grouplist {
		if gid == "favorability" {
			continue
		}
		err = sql.db.Drop(gid)
		if err != nil || gid == "updateinfo" {
			continue
		}
		gidint, _ := strconv.ParseInt(gid, 10, 64)
		upinfo := updateinfo{
			GID:        gidint,
			Updatetime: time.Now().Format("2006/01/02"),
			CanMatch:   1,
			CanNtr:     1,
			CDtime:     12,
		}
		err = sql.db.Create("updateinfo", &updateinfo{})
		if err != nil {
			if err = sql.db.Drop("updateinfo"); err == nil {
				err = sql.db.Create("updateinfo", &updateinfo{})
			}
			if err != nil {
				return err
			}
		}
		err = sql.db.Insert("updateinfo", &upinfo)
	}
	return err
}

func (sql *婚姻登记) 查户口(gid, uid int64) (info userinfo, status string, err error) {
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
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
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
	err := sql.db.Create(gidstr, &userinfo{})
	if err != nil {
		return err
	}
	updatetime := time.Now().Format("15:04:05")
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
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
	wifestr := strconv.FormatInt(wife, 10)
	return sql.db.Del(gidstr, "where target = "+wifestr)
}

func (sql *婚姻登记) 离婚休夫(gid, husband int64) error {
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
	husbandstr := strconv.FormatInt(husband, 10)
	return sql.db.Del(gidstr, "where user = "+husbandstr)
}

func (sql *婚姻登记) 花名册(gid int64) (list [][4]string, number int, err error) {
	sql.Lock()
	defer sql.Unlock()
	gidstr := "group" + strconv.FormatInt(gid, 10)
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
	number = len(list)
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

// 获取好感度
func (sql *婚姻登记) getFavorability(uid, target int64) (favor int, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("favorability", &favorability{})
	if err != nil {
		return
	}
	info := favorability{}
	uidstr := strconv.FormatInt(uid, 10)
	targstr := strconv.FormatInt(target, 10)
	err = sql.db.Find("favorability", &info, "where Userinfo glob '*"+uidstr+"+"+targstr+"*'")
	if err != nil {
		err = sql.db.Insert("favorability", &favorability{
			Userinfo: uidstr + "+" + targstr + "+" + uidstr,
			Favor:    0,
		})
		return
	}
	favor = info.Favor
	return
}

// 获取好感度数据组
type favorList []favorability

func (s favorList) Len() int {
	return len(s)
}
func (s favorList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s favorList) Less(i, j int) bool {
	return s[i].Favor > s[j].Favor
}
func (sql *婚姻登记) getGroupFavorability(uid int64) (list favorList, err error) {
	uidStr := strconv.FormatInt(uid, 10)
	sql.RLock()
	defer sql.RUnlock()
	info := favorability{}
	err = sql.db.FindFor("favorability", &info, "where Userinfo glob '*"+uidStr+"*'", func() error {
		var target string
		userList := strings.Split(info.Userinfo, "+")
		switch {
		case len(userList) == 0:
			return errors.New("好感度系统数据存在错误")
		case userList[0] == uidStr:
			target = userList[1]
		default:
			target = userList[0]
		}
		list = append(list, favorability{
			Userinfo: target,
			Favor:    info.Favor,
		})
		return nil
	})
	sort.Sort(list)
	return
}

// 设置好感度 正增负减
func (sql *婚姻登记) setFavorability(uid, target int64, score int) (favor int, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("favorability", &favorability{})
	if err != nil {
		return
	}
	info := favorability{}
	uidstr := strconv.FormatInt(uid, 10)
	targstr := strconv.FormatInt(target, 10)
	err = sql.db.Find("favorability", &info, "where Userinfo glob '*"+uidstr+"+"+targstr+"*'")
	if err != nil {
		err = sql.db.Insert("favorability", &favorability{
			Userinfo: uidstr + "+" + targstr + "+" + uidstr,
			Favor:    score,
		})
		if err == nil {
			err = sql.db.Find("favorability", &info, "where Userinfo glob '*"+uidstr+"+"+targstr+"*'")
		}
		return info.Favor, err
	}
	info.Favor += score
	if info.Favor > 100 {
		info.Favor = 100
	} else if info.Favor < 0 {
		info.Favor = 0
	}
	err = sql.db.Insert("favorability", &info)
	return info.Favor, err
}

// 获取技能时长
func (sql *婚姻登记) getCDtime(gid int64) (skillCD float64, err error) {
	sql.Lock()
	defer sql.Unlock()
	skillCD = 12
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		if err = sql.db.Drop("updateinfo"); err == nil {
			err = sql.db.Create("updateinfo", &updateinfo{})
		}
		if err != nil {
			return
		}
	}
	gidstr := strconv.FormatInt(gid, 10)
	dbinfo := updateinfo{}
	err = sql.db.Find("updateinfo", &dbinfo, "where gid is "+gidstr)
	if err != nil {
		// 如果没有登记过就记录
		err = sql.db.Insert("updateinfo", &updateinfo{
			GID:      gid,
			CanMatch: 1,
			CanNtr:   1,
			CDtime:   12,
		})
		return
	}
	return dbinfo.CDtime, nil
}

// 设置技能时长
func (sql *婚姻登记) setCDtime(gid int64, cdTime float64) (err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("updateinfo", &updateinfo{})
	if err != nil {
		if err = sql.db.Drop("updateinfo"); err == nil {
			err = sql.db.Create("updateinfo", &updateinfo{})
		}
		if err != nil {
			return
		}
	}
	gidstr := strconv.FormatInt(gid, 10)
	dbinfo := updateinfo{}
	err = sql.db.Find("updateinfo", &dbinfo, "where gid is "+gidstr)
	if err != nil {
		// 如果没有登记过就记录
		err = sql.db.Insert("updateinfo", &updateinfo{
			GID:      gid,
			CanMatch: 1,
			CanNtr:   1,
			CDtime:   cdTime,
		})
		return
	}
	dbinfo.CDtime = cdTime
	err = sql.db.Insert("updateinfo", &dbinfo)
	return
}

// 记录CD
func (sql *婚姻登记) writeCDtime(gid, uid, mun int64) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("cdsheet", &cdsheet{})
	if err != nil {
		if err = sql.db.Drop("cdsheet"); err == nil {
			err = sql.db.Create("cdsheet", &cdsheet{})
		}
		if err != nil {
			return err
		}
	}
	err = sql.db.Insert("cdsheet", &cdsheet{
		Time:    time.Now().Unix(),
		GroupID: gid,
		UserID:  uid,
		ModeID:  mun,
	})
	return err
}

// 判断CD是否过时
func (sql *婚姻登记) compareCDtime(gid, uid, mun int64, cdtime float64) (ok bool, err error) {
	sql.Lock()
	defer sql.Unlock()
	ok = false
	err = sql.db.Create("cdsheet", &cdsheet{})
	if err != nil {
		if err = sql.db.Drop("cdsheet"); err == nil {
			err = sql.db.Create("cdsheet", &cdsheet{})
		}
		if err != nil {
			return
		}
	}
	limitID := "where GroupID is " + strconv.FormatInt(gid, 10) +
		" and UserID is " + strconv.FormatInt(uid, 10) +
		" and ModeID is " + strconv.FormatInt(mun, 10)
	exist := sql.db.CanFind("cdsheet", limitID)
	if !exist {
		return true, nil
	}
	cdinfo := cdsheet{}
	err = sql.db.Find("cdsheet", &cdinfo, limitID)
	if err != nil {
		return
	}
	getTime := time.Unix(cdinfo.Time, 0)
	if time.Since(getTime).Hours() > cdtime {
		// 如果CD已过就删除
		err = sql.db.Del("cdsheet", limitID)
		return true, err
	}
	return
}

// 注入判断 是否为单身
func checkdog(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	// 获取CD
	cdTime, err := 民政局.getCDtime(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]获取该群技能CD错误(将以CD12H计算)\n", err))
	}
	ok, err := 民政局.compareCDtime(gid, uid, 1, cdTime)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]查询用户CD状态失败,请重试\n", err))
		return false
	}
	if !ok {
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return false
	}
	// 判断是否符合条件
	stauts, _, err := 民政局.营业模式(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]", err))
		return false
	}
	if stauts == 0 {
		ctx.SendChain(message.Text("你群包分配,别在娶妻上面下功夫，好好水群"))
		return false
	}
	// 得先判断用户是否存在才行在，再重置
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的target好像不存在？"))
		return false
	}
	// 判断是否需要重置
	ok, err = 民政局.开门时间(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]群状态查询失败\n", err))
		return false
	}
	if ok {
		return true // 重置后也全是单身
	}
	// 获取用户信息
	uidtarget, uidstatus, err := 民政局.查户口(gid, uid)
	switch {
	case uidstatus == "错":
		ctx.SendChain(message.Text("[qqwife]用户状态查询失败\n", err))
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
		ctx.SendChain(message.Text("[qqwife]对象状态查询失败\n", err))
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
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	// 获取CD
	cdTime, err := 民政局.getCDtime(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]获取该群技能CD错误(将以CD12H计算)\n", err))
	}
	ok, err := 民政局.compareCDtime(gid, uid, 2, cdTime)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]查询用户CD状态失败,请重试\n", err))
		return false
	}
	if !ok {
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return false
	}
	// 判断是否符合条件
	_, stauts, err := 民政局.营业模式(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]", err))
		return false
	}
	if stauts == 0 {
		ctx.SendChain(message.Text("你群发布了牛头人禁止令，放弃吧"))
		return false
	}
	// 得先判断用户是否存在才行在，再重置
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的target好像不存在？"))
		return false
	}
	// 判断是否需要重置
	ok, err = 民政局.开门时间(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]群状态查询失败\n", err))
		return false
	}
	if ok {
		ctx.SendChain(message.Text("ta现在还是单身哦，快向ta表白吧！"))
		return false // 重置后也全是单身
	}
	fianceeinfo, fianceestatus, err := 民政局.查户口(gid, fiancee)
	switch {
	case fianceestatus == "错":
		ctx.SendChain(message.Text("[qqwife]对象状态查询失败\n", err))
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
		ctx.SendChain(message.Text("[qqwife]用户状态查询失败\n", err))
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

// 注入判断 是否满足离婚要求
func checkdivorce(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	// 获取CD
	cdTime, err := 民政局.getCDtime(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]获取该群技能CD错误(将以CD12H计算)\n", err))
	}
	ok, err := 民政局.compareCDtime(gid, uid, 4, cdTime)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]查询用户CD状态失败,请重试\n", err))
		return false
	}
	if !ok {
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return false
	}
	// 判断是否符合条件
	_, uidstatus, err := 民政局.查户口(gid, uid)
	switch uidstatus {
	case "错":
		ctx.SendChain(message.Text("[qqwife]数据库发生问题力\n", err))
		return false
	case "单":
		ctx.SendChain(message.Text("今天你还没结婚哦"))
		return false
	}
	return true
}

// 注入判断 是否满足做媒要求
func checkCondition(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	// 获取CD
	cdTime, err := 民政局.getCDtime(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]获取该群技能CD错误(将以CD12H计算)\n", err))
	}
	ok, err := 民政局.compareCDtime(gid, uid, 3, cdTime)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]查询用户CD状态失败,请重试\n", err))
		return false
	}
	if !ok {
		ctx.SendChain(message.Text("你的技能还在CD中..."))
		return false
	}
	// 得先判断用户是否存在才行在，再重置
	gayOne, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，攻方好像不存在？"))
		return false
	}
	gayZero, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，受方好像不存在？"))
		return false
	}
	if gayOne == uid || gayZero == uid {
		ctx.SendChain(message.Text("禁止自己给自己做媒!"))
		return false
	}
	if gayOne == gayZero {
		ctx.SendChain(message.Text("你这个媒人XP很怪咧，不能这样噢"))
		return false
	}
	// 判断是否需要重置
	ok, err = 民政局.开门时间(gid)
	if err != nil {
		ctx.SendChain(message.Text("[qqwife]群状态查询失败\n", err))
		return false
	}
	if ok {
		return true // 重置后也全是单身
	}
	fianceeinfo, fianceestatus, err := 民政局.查户口(gid, gayOne)
	switch {
	case fianceestatus == "错":
		ctx.SendChain(message.Text("[qqwife]对象状态查询失败\n", err))
		return false
	case fianceestatus != "单" && (fianceeinfo.Target == 0 || fianceeinfo.User == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的攻方是单身贵族噢"))
		return false
	case (fianceestatus == "攻" && fianceeinfo.Target == gayZero) ||
		(fianceestatus == "受" && fianceeinfo.User == gayZero):
		ctx.SendChain(message.Text("笨蛋！ta们已经在一起了！"))
		return false
	case fianceestatus != "单":
		ctx.SendChain(message.Text("攻方不是单身,不允许给这种人做媒!"))
		return false
	}
	// 获取用户信息
	uidtarget, uidstatus, err := 民政局.查户口(gid, gayZero)
	switch {
	case uidstatus == "错":
		ctx.SendChain(message.Text("[qqwife]用户状态查询失败\n", err))
	case uidstatus == "单": // 如果为单身狗
		return true
	case uidstatus != "单" && (uidtarget.Target == 0 || uidtarget.User == 0): // 如果是单身贵族
		ctx.SendChain(message.Text("今天的你是单身贵族噢"))
	case uidstatus != "单":
		ctx.SendChain(message.Text("受方不是单身,不允许给这种人做媒!"))
	}
	return false
}
