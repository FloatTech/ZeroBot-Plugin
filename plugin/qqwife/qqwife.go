// Package qqwife 娶群友
package qqwife

import (
	"math/rand"
	"sort"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
)

var (
	qqwifegroup []qqwifeinfo
	lastdate    time.Time
)

type qqwifeinfo struct {
	UiD     int64
	Wid     int64
	Groupid int64
}

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一天一夫一妻制群老婆\n" +
			"-娶群友",
	})
	engine.OnFullMatchGroup([]string{"娶群友"}, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if time.Now().Day() != lastdate.Day() {
				qqwifegroup = []qqwifeinfo{} //如果日期不同则清空数据
			}
			//先判断是否已经娶过或者被娶
			groupid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			var status string
			for _, v := range qqwifegroup {
				if v.UiD == uid { //判定是否是1
					if v.Groupid != groupid {
						status = "（这家伙在别群有群老婆了，真是个渣男）\n" //记录渣男
					} else {
						ctx.SendChain(
							message.At(uid),
							message.Text("今天你的群老婆是"),
							message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(v.Wid, 10)+"&s=640"),
							message.Text(
								"\n",
								"[", ctx.CardOrNickName(v.Wid), "]",
								"(", strconv.FormatInt(v.Wid, 10), ")哒",
							),
						)
						return
					}
				} else if v.Wid == uid && v.Groupid == groupid { //如果为0且是在本群抽的就输出
					ctx.SendChain(
						message.At(uid),
						message.Text("今天你被娶了，群老公是"),
						message.Image("http://q4.qlogo.cn/g?b=qq&nk="+strconv.FormatInt(v.UiD, 10)+"&s=640"),
						message.Text(
							"\n",
							"[", ctx.CardOrNickName(v.UiD), "]",
							"(", strconv.FormatInt(v.UiD, 10), ")哒",
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
			temp = temp[math.Max(0, len(temp)-10):]
			//将已经娶过的人剔除
			var qqgrouplist = []int64{}
			var alive = true
			if len(qqwifegroup) == 0 {
				for k := 0; k < len(temp); k++ {
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			} else {
				for k := 0; k < len(temp); k++ {
					for _, v := range qqwifegroup {
						if v.UiD == temp[k].Get("user_id").Int() {
							alive = false
							continue
						}
					}
					if alive {
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
			newcp := qqwifeinfo{UiD: uid, Wid: who, Groupid: groupid}
			qqwifegroup = append(qqwifegroup, newcp)
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
			lastdate = time.Now()
		})
}
