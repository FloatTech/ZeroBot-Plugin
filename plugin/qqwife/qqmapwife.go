// Package qqwife 娶群友  基于“翻牌”和江林大佬的“群老婆”插件魔改作品
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
)

var (
	qqwifegroup = make(map[int64]map[int64]int64, 64) // 64个群的预算大小
	lastdate    time.Time
	mu          sync.Mutex
)

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一群一天一夫一妻制群老婆\n" +
			"- 娶群友\n" +
			"- 娶[老婆QQ号|@老婆QQ]\n（注：群单身狗专属,CD为24H,不跨天刷新）\n" +
			"- 群老婆列表",
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			if time.Now().Day() != lastdate.Day() {
				qqwifegroup = make(map[int64]map[int64]int64, 64) // 跨天就重新初始化数据
			}
			// 先判断是否已经娶过或者被娶
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			// 如果娶过
			wife, ok := qqwifegroup[gid][uid]
			if ok {
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
			husband, ok := qqwifegroup[gid][-uid]
			if ok {
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你被娶了，群老公是"),
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
			if len(qqwifegroup) == 0 {
				for k := 0; k < len(temp); k++ {
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			} else {
				for k := 0; k < len(temp); k++ {
					_, ok := qqwifegroup[gid][temp[k].Get("user_id").Int()]
					if ok {
						continue
					}
					_, ok = qqwifegroup[gid][-temp[k].Get("user_id").Int()]
					if ok {
						continue
					}
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			}
			// 没有人（只剩自己）的时候
			if len(qqgrouplist) == 0 {
				ctx.SendChain(message.Text("噢，此时此刻你还是一只单身狗，等待下一次情缘吧"))
				return
			}
			// 随机抽娶
			wife = qqgrouplist[rand.Intn(len(qqgrouplist))]
			if wife == uid { // 如果是自己
				ctx.SendChain(message.Text("噢，此时此刻你还是一只单身狗，等待下一次情缘吧"))
				return
			}
			// 绑定CP
			if qqwifegroup[gid] == nil {
				qqwifegroup[gid] = make(map[int64]int64, 32)
			}
			qqwifegroup[gid][uid] = wife
			qqwifegroup[gid][-wife] = uid
			// 输出结果
			ctx.SendChain(
				message.At(uid),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(wife, 10)+"&s=640,cache=0").Add("cache", 0),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(wife), "]",
					"(", wife, ")哒",
				),
			)
			// 更新时间
			lastdate = time.Now()
		})
	var singledogCD = ctxext.NewLimiterManager(time.Hour*24, 1)
	engine.OnRegex(`^娶(\d+|\[CQ:at,qq=(\d+)\])`, zero.OnlyGroup).SetBlock(true).Limit(singledogCD.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			fiancee, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			// fmt.Println("1:", fiancee)
			if err != nil {
				fiancee, _ = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
				// fmt.Println("2:", fiancee)
			}
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			if uid == fiancee {
				ctx.SendChain(message.Text("自恋狂（CD冷却开始计时"))
				return
			}
			// 如果用户娶过
			_, ok := qqwifegroup[gid][uid]
			if ok {
				ctx.SendChain(message.Text("你这渣男，居然想找小三！"))
				return
			}
			// 如果用户被娶过
			_, ok = qqwifegroup[gid][-fiancee]
			if ok {
				ctx.SendChain(message.Text("该是0就是0，当0有什么不好"))
				return
			}
			// 如果未婚妻娶过
			_, ok = qqwifegroup[gid][fiancee]
			if ok {
				ctx.SendChain(message.Text("他已经取了别的女人，你放弃吧"))
				return
			}
			// 如果未婚妻被娶过
			_, ok = qqwifegroup[gid][-fiancee]
			if ok {
				ctx.SendChain(message.Text("你这渣男，跟别人抢女人！"))
				return
			}
			if rand.Intn(2) == 1 {
				// 绑定CP
				if qqwifegroup[gid] == nil {
					qqwifegroup[gid] = make(map[int64]int64, 32)
				}
				qqwifegroup[gid][uid] = fiancee
				qqwifegroup[gid][-fiancee] = uid
				// 输出结果
				ctx.SendChain(
					message.Text("今天你向她表白了，ta羞涩的点了点头同意了！\n"),
					message.At(uid),
					message.Text("今天你的群老婆是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(fiancee, 10)+"&s=640").Add("cache", 0),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(fiancee), "]",
						"(", fiancee, ")哒",
					),
				)
			} else {
				ctx.SendChain(message.Text("今天你向她表白了，ta毫无感情的拒绝了你\n噢，此时此刻你还是一只单身狗\n别灰心，等待下一次情缘吧"))
			}
		})
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			group, ok := qqwifegroup[ctx.Event.GroupID]
			if !ok {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			if len(group) == 0 {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			cplist := make([]string, 1, len(group)+1)
			cplist[0] = "群老公←———→群老婆\n--------------------------"
			for husband, wife := range group {
				if husband > 0 {
					cplist = append(cplist, ctx.CardOrNickName(husband)+" & "+ctx.CardOrNickName(wife))
				}
			}
			ctx.SendChain(message.Text(strings.Join(cplist, "\n")))
		})
}
