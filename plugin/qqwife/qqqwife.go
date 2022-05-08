// Package qqwife 娶群友
package qqwife

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
)

var (
	qqwifegroup [][]int64
)

func init() {
	engine := control.Register("qqwife", &control.Options{
		DisableOnDefault: false,
		Help: "一天一夫一妻制群老婆\n" +
			"-娶群友",
	})
	engine.OnFullMatchGroup([]string{"娶群友"}, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			//凌晨清空数据
			now := time.Now()
			if now.Hour() == 0 && now.Minute() == 0 {
				qqwifegroup = [][]int64{}
			}
			// 无缓存获取群员列表
			temp := ctx.GetThisGroupMemberListNoCache().Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-10):]
			//先判断是否已经娶过或者被娶
			uid := ctx.Event.UserID
			for i, v := range qqwifegroup {
				for j, v2 := range v {
					if v2 == uid {
						if j == 0 {
							//如果在别的群有了老婆则不许娶
							for k := 0; k < len(temp); k++ {
								if qqwifegroup[i][1] == temp[k].Get("user_id").Int() {
									avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", qqwifegroup[i][1])
									msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群老婆是%s\n【%s】(%d)哒！", uid, avtar, ctx.CardOrNickName(qqwifegroup[i][1]), qqwifegroup[i][1])
									msg = message.UnescapeCQCodeText(msg)
									ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
									return
								}
							}
							ctx.SendGroupMessage(ctx.Event.GroupID, message.Text("你已经在别群有老婆了！渣男！"))
						} else {
							avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", qqwifegroup[i][0])
							msg := fmt.Sprintf("[CQ:at,qq=%d]今天你被娶了,群老公是%s\n【%s】(%d)哒！", uid, avtar, ctx.CardOrNickName(qqwifegroup[i][0]), qqwifegroup[i][0])
							msg = message.UnescapeCQCodeText(msg)
							ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
						}
						return
					}
				}
			}
			//将已经娶过的人剔除
			var qqgrouplist = []int64{}
			var alive = true
			var k = 0
			if len(qqwifegroup) == 0 {
				for ; k < len(temp); k++ {
					qqgrouplist = append(qqgrouplist, temp[k].Get("user_id").Int())
				}
			} else {
				for ; k < len(temp); k++ {
					for _, v := range qqwifegroup {
						for _, v2 := range v {
							if v2 == temp[k].Get("user_id").Int() {
								alive = false
								continue
							}
						}
						if !alive {
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
			qqwifegroup = append(qqwifegroup, []int64{uid, who})
			avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", who)
			msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群老婆是%s\n【%s】(%d)哒！", uid, avtar, ctx.CardOrNickName(who), who)
			msg = message.UnescapeCQCodeText(msg)
			ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
		})
}
