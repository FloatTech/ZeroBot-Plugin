// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品
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
)

type 婚姻登记 struct {
	sync.Mutex
	mp map[int64]map[int64]int64
}

func 新登记处() (db 婚姻登记) {
	db.mp = make(map[int64]map[int64]int64, 64)
	return
}

func (db *婚姻登记) 重置() {
	db.Lock()
	defer db.Unlock()
	for k := range db.mp {
		delete(db.mp, k)
	}
}

func (db *婚姻登记) 有登记在(gid int64) (ok bool) {
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

func (db *婚姻登记) 登记情况(gid int64, ctx *zero.Ctx) string {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return ""
	}
	return binary.BytesToString(binary.NewWriterF(func(w *binary.Writer) {
		w.WriteString("群老公←———→群老婆\n-----------")
		for husband, wife := range mp {
			if husband > 0 {
				w.WriteByte('\n')
				w.WriteString(ctx.CardOrNickName(husband))
				w.WriteString(" & ")
				w.WriteString(ctx.CardOrNickName(wife))
			}
		}
	}))
}

func (db *婚姻登记) 有妻子(gid, uid int64) (ok bool) {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return
	}
	_, ok = mp[uid]
	return
}

func (db *婚姻登记) 查询妻子(gid, uid int64) (wife int64) {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return
	}
	return mp[uid]
}

func (db *婚姻登记) 有丈夫(gid, uid int64) (ok bool) {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return
	}
	_, ok = mp[-uid]
	return
}

func (db *婚姻登记) 查询丈夫(gid, uid int64) (husband int64) {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		return
	}
	return mp[-uid]
}

func (db *婚姻登记) 登记(gid, wife, husband int64) {
	db.Lock()
	defer db.Unlock()
	mp, ok := db.mp[gid]
	if !ok {
		mp = make(map[int64]int64, 32)
		db.mp[gid] = mp
	}
	// 绑定CP
	mp[husband] = wife
	mp[-wife] = husband
}

var (
	民政局      = 新登记处()
	lastdate time.Time
	sendtext = [...][]string{
		{
			"今天你向ta表白了, ta羞涩的点了点头同意了!\n",
			"你对ta说“以我之名, 冠你指间, 一天相伴, 一天相随”. ta捂着嘴点了点头\n\n",
		},
		{
			"今天你向ta表白了, ta毫无感情的拒绝了你\n",
			"今天你向ta表白了, ta对你说“你是一个非常好的人”\n",
			"今天你向ta表白了, ta给了你一个拥抱后擦肩而过\n",
		},
	}
)

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一群一天一夫一妻制群老婆\n" +
			"- 娶群友\n" +
			"- 娶[老婆QQ号|@老婆QQ]\n(注:单身专属技能,CD24H,不跨天刷新)\n" +
			"- 群老婆列表",
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if time.Now().Day() != lastdate.Day() {
				民政局.重置()
				// 更新时间
				lastdate = time.Now()
			}
			// 先判断是否已经娶过或者被娶
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			// 如果娶过
			wife := 民政局.查询妻子(gid, uid)
			if wife > 0 {
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你的群老婆是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(wife, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(wife), "]",
						"(", wife, ")哒",
					),
				)
				return
			}
			// 如果被娶过
			husband := 民政局.查询丈夫(gid, uid)
			if husband > 0 {
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你被娶了, 群老公是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(husband, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(husband), "]",
						"(", husband, ")哒",
					),
				)
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
				if 民政局.有妻子(gid, usr) || 民政局.有丈夫(gid, usr) {
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
			wife = qqgrouplist[rand.Intn(len(qqgrouplist))]
			if wife == uid { // 如果是自己
				ctx.SendChain(message.Text("噢, 此时此刻你还是一只单身狗, 等待下一次情缘吧"))
				return
			}
			// 绑定CP
			民政局.登记(gid, wife, uid)
			// 输出结果
			ctx.SendChain(
				message.At(uid),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(wife, 10)+"&s=640").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(wife), "]",
					"(", wife, ")哒",
				),
			)
		})
	// 单身狗专属技能
	engine.OnRegex(`^娶(\d+|\[CQ:at,qq=(\d+)\])`, zero.OnlyGroup, checkdog).SetBlock(true).
		Limit(ctxext.NewLimiterManager(time.Hour*12, 1).LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			if err != nil {
				fiancee, _ = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			}
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			if rand.Intn(2) == 1 {
				民政局.登记(gid, fiancee, uid)
				// 输出结果
				ctx.SendChain(
					message.Text(sendtext[0][rand.Intn(len(sendtext[0]))]),
					message.At(uid),
					message.Text("今天你的群老婆是"),
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
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if !民政局.有登记在(ctx.Event.GroupID) {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			ctx.SendChain(message.Text(民政局.登记情况(ctx.Event.GroupID, ctx)))
		})
}

// 注入判断 是否为单身狗
func checkdog(ctx *zero.Ctx) bool {
	fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
	// fmt.Println("1:", fiancee)
	if err != nil {
		fiancee, _ = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
		// fmt.Println("2:", fiancee)
	}
	gid := ctx.Event.GroupID
	uid := ctx.Event.UserID
	if uid == fiancee {
		ctx.SendChain(message.Text("今日获得成就：自恋狂"))
		return false
	}
	// 如果用户娶过
	if o := 民政局.查询妻子(gid, uid); o > 0 {
		switch o {
		case fiancee:
			ctx.SendChain(message.Text("笨蛋~你明明已经娶了啊w"))
		default:
			ctx.SendChain(message.Text("笨蛋~你家里还有个吃白饭的w"))
		}
		return false
	}
	// 如果用户被娶过
	if o := 民政局.查询丈夫(gid, uid); o > 0 {
		switch o {
		case fiancee:
			ctx.SendChain(message.Text("笨蛋~你明明已经嫁给他了啊w"))
		default:
			ctx.SendChain(message.Text("该是0就是0, 当0有什么不好"))
		}
		return false
	}
	// 如果未婚妻娶过
	if o := 民政局.查询妻子(gid, fiancee); o > 0 {
		switch o {
		case uid:
			ctx.SendChain(message.Text("笨蛋~你明明已经嫁给他了啊w"))
		default:
			ctx.SendChain(message.Text("他有别的女人了, 你该放下了"))
		}
		return false
	}
	// 如果未婚妻被娶过
	if o := 民政局.查询丈夫(gid, fiancee); o > 0 {
		switch o {
		case uid:
			ctx.SendChain(message.Text("笨蛋~你明明已经娶了啊w"))
		default:
			ctx.SendChain(message.Text("这是一个纯爱的世界, 拒绝NTR"))
		}
		return false
	}
	return true
}
