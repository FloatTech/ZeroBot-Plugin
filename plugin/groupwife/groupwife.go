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
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func init() {
	engine := control.Register("groupwife", &control.Options{
		DisableOnDefault: true,
		Help: "群老婆\n" +
			"-哪个群友是我老婆",
	})
	engine.OnFullMatchGroup([]string{"哪个群友是我老婆", "哪位群友是我老婆", "今天谁是我老婆"}, zero.OnlyGroup).SetBlock(false).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			list := ctx.CallAction("get_group_member_list", zero.Params{
				"group_id": ctx.Event.GroupID,
				"no_cache": false,
			}).Data
			temp := list.Array()
			sort.SliceStable(temp, func(i, j int) bool {
				return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
			})
			temp = temp[math.Max(0, len(temp)-30):]
			uid := ctx.Event.UserID
			now := time.Now()
			s := md5.Sum(helper.StringToBytes(fmt.Sprintf("%d%d%d%d", uid, now.Year(), now.Month(), now.Day())))
			r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(s[:]))))
			who := temp[r.Intn(len(temp))]
			wifename := who.Get("card").Str
			wifeid := who.Get("user_id").Int()
			if wifename == "" {
				wifename = who.Get("nickname").Str
			}
			avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%d&s=640,cache=0]", wifeid)
			msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群友老婆是%s\n【%s】(%d)哒！", uid, avtar, wifename, wifeid)
			msg = message.UnescapeCQCodeText(msg)
			ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
		})
}
