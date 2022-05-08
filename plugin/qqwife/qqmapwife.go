// Package qqwife 娶群友
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
)

var (
	qqwifegroup = make(map[int64]map[int64]int64, 50) //50个群的预算大小
	lastdate    time.Time
)

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一群一天一夫一妻制群老婆\n" +
			"-娶群友",
	})
	engine.OnFullMatch("娶群友", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if time.Now().Day() != lastdate.Day() {
				qqwifegroup = make(map[int64]map[int64]int64, 50) //跨天就重新初始化数据
			}
			//先判断是否已经娶过或者被娶
			groupid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			var status string
			//如果娶过
			userinfo, ok := qqwifegroup[groupid][uid]
			if ok {
				ctx.SendChain(
					message.At(uid),
					message.Text("今天你的群老婆是"),
					message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(userinfo, 10)+"&s=640"),
					message.Text(
						"\n",
						"[", ctx.CardOrNickName(userinfo), "]",
						"(", strconv.FormatInt(userinfo, 10), ")哒",
					),
				)
				return
			}
			//如果被娶过
			for k, v := range qqwifegroup[groupid] {
				if v == uid { //如果为0且是在本群抽的就输出
					ctx.SendChain(
						message.At(uid),
						message.Text("今天你被娶了，群老公是"),
						message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(k, 10)+"&s=640"),
						message.Text(
							"\n",
							"[", ctx.CardOrNickName(k), "]",
							"(", strconv.FormatInt(k, 10), ")哒",
						),
					)
					return
				}
			}
			// 无缓存获取群员列表
			temp := ctx.GetThisGroupMemberListNoCache().Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			//将已经娶过的人剔除
			var qqgrouplist = []int64{}
			if len(qqwifegroup) == 0 {
				for k := 0; k < len(temp); k++ {
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			} else {
				for k := 0; k < len(temp); k++ {
					_, ok := qqwifegroup[groupid][temp[k].Get("user_id").Int()]
					if !ok {
						qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
					}
				}
			}
			//没有人（只剩自己）的时候
			if len(qqgrouplist) == 0 {
				ctx.SendChain(message.Text("噢，此时此刻你还是一只单身狗，等待下一次情缘吧"))
				return
			}
			//随机抽娶
			who := qqgrouplist[rand.Intn(len(qqgrouplist))]
			if who == uid { //如果是自己
				ctx.SendChain(message.Text("噢，此时此刻你还是一只单身狗，等待下一次情缘吧"))
				return
			}
			//绑定CP
			setcp := make(map[int64]int64, 60) //初始化中间的map元素，虽然取了30个人数，以防万一翻个倍储存
			setcp[uid] = who
			qqwifegroup[groupid] = setcp
			//输出结果
			ctx.SendChain(
				message.Text(status),
				message.At(uid),
				message.Text("今天你的群老婆是"),
				message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(who, 10)+"&s=640"),
				message.Text(
					"\n",
					"[", ctx.CardOrNickName(who), "]",
					"(", strconv.FormatInt(who, 10), ")哒",
				),
			)
			//更新时间
			lastdate = time.Now()
		})
	engine.OnFullMatch("群老婆列表", zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			var cplist = []string{"群老公←———→群老婆\n--------------------------"}
			group, ok := qqwifegroup[ctx.Event.GroupID]
			if !ok {
				ctx.SendChain(message.Text("你群并没有任何的CP额"))
				return
			}
			for k, v := range group {
				if k != 0 {
					husband := ctx.CardOrNickName(k)
					wife := ctx.CardOrNickName(v)
					cplist = append(cplist, husband+" & "+wife)
				} else {
					ctx.SendChain(message.Text("你群并没有任何的CP额"))
					return
				}
			}
			ctx.SendChain(message.Text(strings.Join(cplist, "\n")))
		})
}
