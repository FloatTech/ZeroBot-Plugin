// Package groupwife 群老婆
package groupwife

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/math"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var (
	me     gjson.Result
	wifeid int
	luid   int
	lwid   int
	sign   int
	wife   = make(map[int]int)
	swife  = make(map[int]int)
)

func init() {
	engine := control.Register("groupwife", &control.Options{
		DisableOnDefault: true,
		Help: "群老婆\n" +
			"-哪个群友是我老婆",
	})
	engine.OnFullMatchGroup([]string{"哪个群友是我老婆", "哪位群友是我老婆", "今天谁是我老婆"}, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			gid := int(ctx.Event.GroupID)
			uid := int(ctx.Event.UserID)
			for sign = range swife {
				if sign == gid+uid {
					wifeid = swife[luid] - gid
					wifename := ctx.CardOrNickName(int64(wifeid))
					avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", wifeid)
					msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群友老婆是%s\n【%s】(%d)哒！", uid, avtar, wifename, wifeid)
					msg = message.UnescapeCQCodeText(msg)
					ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
					ctx.SendChain(message.Text("这是标记1"))
					return
				}
			}
			for sign = range wife {
				if sign == gid+uid {
					wifeid = wife[lwid] - gid
					wifename := ctx.CardOrNickName(int64(wifeid))
					avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", wifeid)
					msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群友老婆是%s\n【%s】(%d)哒！", uid, avtar, wifename, wifeid)
					msg = message.UnescapeCQCodeText(msg)
					ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
					ctx.SendChain(message.Text("这是标记2"))
					return
				}
			}
			list := ctx.GetGroupMemberListNoCache(int64(gid))
			temp := list.Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			now := time.Now()
			s := md5.Sum(helper.StringToBytes(fmt.Sprintf("%d%d%d%d", uid, now.Year(), now.Month(), now.Day())))
			r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(s[:]))))
			rn := r.Intn(len(temp))
			who := temp[rn]
			wifeid = int(who.Get("user_id").Int())
			if wifeid == uid {
				me = who
				temp = append(temp[:rn], temp[rn:]...)
				rn = r.Intn(len(temp))
				who = temp[rn]
			}
			wifename := who.Get("card").Str
			if wifename == "" {
				wifename = who.Get("nickname").Str
			}
			avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", who.Get("user_id").Int())
			msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群友老婆是%s\n【%s】(%d)哒！", uid, avtar, wifename, wifeid)
			msg = message.UnescapeCQCodeText(msg)
			ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
			ctx.SendChain(message.Text("这是标记3"))
			luid := gid + uid
			lwid := gid + wifeid
			wife[lwid] = (luid)
			swife[luid] = (lwid)

			if len(me.Array()) != 0 {
				mlist := append(temp[:rn], me)
				temp = append(mlist, temp[rn:]...)
			}
		})
}
