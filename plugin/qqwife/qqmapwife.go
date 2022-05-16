//Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/binary"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

//nolint: asciicheck
type 婚姻登记 struct {
	sync.Mutex
	mp map[int64]map[int64]*userinfo
}

//结婚证信息
type userinfo struct {
	target     int64  //对象身份证号
	username   string //户主名称
	targetname string //对象名称
}

//nolint: asciicheck
func 新登记处() (db 婚姻登记) {
	db.mp = make(map[int64]map[int64]*userinfo, 64)
	return
}

//nolint: asciicheck
func (db *婚姻登记) 重置() {
	db.Lock()
	defer db.Unlock()
	for k := range db.mp {
		delete(db.mp, k)
	}
}

//nolint: asciicheck
func (db *婚姻登记) 办理离婚(gid, wife int64, gender int) {
	db.Lock()
	defer db.Unlock()
	switch gender{
		case 0://休妻
			delete(db.mp[gid], -wife)
		default://休夫
			delete(db.mp[gid], wife)
	}
}

//nolint: asciicheck
func (db *婚姻登记) 登记情况(gid int64) (ok bool) {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return
	}
	for range mp {
		return true
	}
	return
}

//nolint: asciicheck
func (db *婚姻登记) 花名册(ctx *zero.Ctx, gid int64) string {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return "民政局的花名册出问题了额..."
	}
	return binary.BytesToString(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteString("群老公←———→群老婆\n-----------")
		for uid, userinfo := range mp {
			if uid > 0 {
				_ = w.WriteByte('\n')
				w.WriteString(userinfo.username)
				w.WriteString(" & ")
				w.WriteString(userinfo.targetname)
			}
		}
	}))
}

//nolint: asciicheck
func (db *婚姻登记) 查户口(gid, uid int64) (userinfo *userinfo, gender int, ok bool) {
	db.Lock()
	defer db.Unlock()
	gender = 0
	mp, ok := db.mp[gid]
	if !ok {
		return
	}
	userinfo, ok = mp[uid]
	if !ok {
		gender = 1
		userinfo, ok = mp[-uid]
	}
	return
}

//nolint: asciicheck
func (db *婚姻登记) 登记(gid, uid, target int64, username, targetname string) {
	db.Lock()
	defer db.Unlock()
	_,ok := db.mp[gid]
	if !ok{
		db.mp[gid] = make(map[int64]*userinfo, 32)
	}
	//填写夫妻信息
	uidinfo := &userinfo{
		target:     target,
		username:   username,
		targetname: targetname,
	}
	targetinfo := &userinfo{
		target:     uid,
		username:   targetname,
		targetname: username,
	}
	//民政局登记数据
	db.mp[gid][uid] = uidinfo
	db.mp[gid][-target] = targetinfo
}

var (
	//nolint: asciicheck
	民政局      = 新登记处()
	skillCD  = rate.NewManager[string](time.Hour*24, 1)
	lastdate time.Time
	sendtext = [...][]string{
		{ //表白成功
			"今天你向ta表白了，ta羞涩的点了点头同意了！\n",
			"你对ta说“以我之名，冠你指间，一天相伴，一天相随”.ta捂着嘴点了点头\n\n",
		},
		{ //表白失败
			"今天你向ta表白了，ta毫无感情的拒绝了你",
			"今天你向ta表白了，ta对你说“你是一个非常好的人”",
			"今天你向ta表白了，ta给了你一个拥抱后擦肩而过",
		},
		{ //ntr成功
			"你处心积虑的接近ta，ta最终选择跟随你\n",
		},
	}
)

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一群一天一夫一妻制群老婆\n（每天凌晨刷新CP）\n" +
			"- 娶群友\n- 群老婆列表\n" +
			"--------------------------------\n以下技能每人只能二选一\n   CD24H，不跨天刷新\n--------------------------------\n" +
			"- (娶|嫁)@对方QQ\n- 当[对方Q号|@对方QQ]的小三\n",
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if time.Now().Day() != lastdate.Day() {
				民政局.重置()
				// 更新时间
				lastdate = time.Now()
			}
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			targetinfo, status, ok := 民政局.查户口(gid, uid)
			if ok {
				switch status {
				case 0: //娶过别人
					ctx.SendChain(
						message.At(uid),
						message.Text("今天你的群老婆是"),
						message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(targetinfo.target, 10)+"&s=640").Add("cache", 0),
						message.Text(
							"\n",
							"[", targetinfo.targetname, "]",
							"(", targetinfo.target, ")哒",
						),
					)
				default: //嫁给别人
					ctx.SendChain(
						message.At(uid),
						message.Text("今天你的群老公是"),
						message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(targetinfo.target, 10)+"&s=640").Add("cache", 0),
						message.Text(
							"\n",
							"[", targetinfo.targetname, "]",
							"(", targetinfo.target, ")哒",
						),
					)
				}
				return
			}
			//  无缓存获取群员列表
			temp := ctx.GetThisGroupMemberListNoCache().Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			// 将已经娶过的人剔除
			qqgrouplist := make([]int64, 0, len(temp))
			for k := 0; k < len(temp); k++ {
				usr := temp[k].Get("user_id").Int()
				_, _, ok := 民政局.查户口(gid, usr)
				if ok {
					continue
				}
				qqgrouplist = append(qqgrouplist, usr)
			}
			// 没有人（只剩自己）的时候
			if len(qqgrouplist) == 0 {
				ctx.SendChain(message.Text("噢, 此时此刻你还是一只单身狗, 等待下一次情缘吧"))
				return
			}
			// 随机抽娶
			fiancee := qqgrouplist[rand.Intn(len(qqgrouplist))]
			if fiancee == uid { // 如果是自己
				ctx.SendChain(message.Text("噢, 此时此刻你还是一只单身狗, 等待下一次情缘吧"))
				return
			}
			//去民政局办证
			民政局.登记(ctx, gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(target))
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
	//单生狗专属技能
	engine.OnRegex(`^(娶|嫁)\[CQ:at,qq=(\d+)\]`, zero.OnlyGroup, checkdog).SetBlock(true).Limit(cdcheck, iscding).
		Handle(func(ctx *zero.Ctx) {
			choice := ctx.State["regex_matched"].([]string)[1]
			fiancee, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			uid := ctx.Event.UserID
			if uid == fiancee { //如果是自己
				ctx.SendChain(message.Text("今日获得成就：自恋狂"))
				return
			}
			if rand.Intn(2) == 1 { //二分之一的概率表白成功
				gid := ctx.Event.GroupID
				//根据技能分配0和1
				var choicetext string
				switch choice {
				case "娶":
					民政局.登记(ctx, gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(target))
					choicetext = "今天你的群老婆是"
				default:
					民政局.登记(ctx, gid, fiancee, uid, ctx.CardOrNickName(target), ctx.CardOrNickName(uid))
					choicetext = "今天你的群老公是"
				}
				//请大家吃席
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
				return
			}
			ctx.SendChain(message.Text(sendtext[1][rand.Intn(len(sendtext[1]))]))
		})
	//NTR技能
	engine.OnRegex(`^当(\[CQ:at,qq=(\d+)\] |(\d+))的小三`, zero.OnlyGroup, checkcp).SetBlock(true).Limit(cdcheck, iscding).
		Handle(func(ctx *zero.Ctx) {
			fid := ctx.State["regex_matched"].([]string)
			fiancee, _ := strconv.ParseInt(fid[2]+fid[3], 10, 64)
			if rand.Intn(10)/4 != 0 { // 十分之三的概率NTR成功
				ctx.SendChain(message.Text("你的ntr计划失败了"))
				return
			}
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			//判断target是0还是1
			choicetext := "婆"
			targetinfo, gender, _ := 民政局.查户口(gid, fiancee)
			民政局.办理离婚(gid, targetinfo.target,gender)
			if gender == 1{
				choicetext = "公"
			}
			//重新绑定CP
			switch choicetext {
			case "婆":
				民政局.登记(ctx, gid, uid, fiancee, ctx.CardOrNickName(uid), ctx.CardOrNickName(target))
				choicetext = "今天你的群老婆是"
			default:
				民政局.登记(ctx, gid, fiancee, uid, ctx.CardOrNickName(target), ctx.CardOrNickName(uid))
				choicetext = "今天你的群老公是"
			}
			// 输出结果
			ctx.SendChain(
				message.Text(sendtext[2][rand.Intn(len(sendtext[2]))]),
				message.At(uid),
				message.Text("今天你的群老"+choicetext+"是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(fiancee), "]",
					"(", fiancee, ")哒",
				),
			)
		})
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if !民政局.登记情况(ctx.Event.GroupID) {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			ctx.SendChain(message.Text(民政局.花名册(ctx, ctx.Event.GroupID)))
		})
}

//以群号和昵称为限制
func cdcheck(ctx *zero.Ctx) *rate.Limiter {
	limitID := strconv.FormatInt(ctx.Event.GroupID, 10) + strconv.FormatInt(ctx.Event.UserID, 10)
	return skillCD.Load(limitID)
}
func iscding(ctx *zero.Ctx) {
	ctx.SendChain(message.Text("你的技能现在正在CD中"))
}

// 注入判断 是否为单身狗
func checkdog(ctx *zero.Ctx) bool {
	gid := ctx.Event.GroupID
	if !民政局.登记情况(gid) {
		return true //如果没有人登记，说明全是单身
	}
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的target好像不存在？"))
		return false
	}
	uid := ctx.Event.UserID
	if uid == fiancee {
		ctx.SendChain(message.Text("今日获得成就：自恋狂"))
		return false
	}
	//获取用户info
	uidtarget, uidstatus, ok1 := 民政局.查户口(gid, uid)
	_, fianceestatus, ok2 := 民政局.查户口(gid, fiancee)
	if !ok1 && !ok2 { //必须是两个单生狗
		return true
	}
	if uidtarget.target == fiancee { //如果本就是一块
		ctx.SendChain(message.Text("笨蛋~你们明明已经在一起了啊w"))
		return false
	}
	if ok1 {
		switch uidstatus {
		case 0: //如果如为攻
			ctx.SendChain(message.Text("笨蛋~你家里还有个吃白饭的w"))
		default: //如果为受
			ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		}
		return false
	}
	if ok2 {
		switch fianceestatus {
		case 0: //如果如为攻
			ctx.SendChain(message.Text("他有别的女人了，你该放下了"))
		default: //如果为受
			ctx.SendChain(message.Text("这是一个纯爱的世界，拒绝NTR"))
		}
		return false
	}
	return true
}

//注入判断 是否满足小三要求
func checkcp(ctx *zero.Ctx) bool {
	//检查群内是否有人登记了
	gid := ctx.Event.GroupID
	if !民政局.登记情况(gid) {
		ctx.SendChain(message.Text("ta无法达成你当小三的条件"))
		return false
	}
	//检查target
	fid := ctx.State["regex_matched"].([]string)
	fiancee, err := strconv.ParseInt(fid[2]+fid[3], 10, 64)
	if err != nil {
		ctx.SendChain(message.Text("额，你的对象好像不存在?"))
		return false
	}
	//检查用户是否登记过
	uid := ctx.Event.UserID
	userinfo, uidstatus, ok := 民政局.查户口(gid, uid)
	if ok {
		if userinfo.target == fiancee { //如果本就是一块
			ctx.SendChain(message.Text("笨蛋~你们明明已经在一起了啊w"))
			return false
		}
		switch uidstatus {
		case 0: //如果如为攻
			ctx.SendChain(message.Text("抱歉，建国之后不支持后宫"))
		default: //如果为受
			ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
		}
		return false
	}
	_, _, ok = 民政局.查户口(gid, fiancee)
	if !ok {
		ctx.SendChain(message.Text("ta无法达成你当小三的条件"))
		return false
	}
	return true
}
