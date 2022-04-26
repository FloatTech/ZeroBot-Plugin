// Package groupwife 群老婆
package groupwife

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
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
	me      gjson.Result
	list    gjson.Result
	temp    []gjson.Result
	listgid int64
)

func init() {
	engine := control.Register("groupwife", &control.Options{
		DisableOnDefault: true,
		Help: "群老婆\n" +
			"-哪个群友是我老婆",
	})
	engine.OnFullMatchGroup([]string{"哪个群友是我老婆", "哪位群友是我老婆", "今天谁是我老婆"}, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if listgid == 0 {
				listgid = ctx.Event.GroupID
			}
			list = ctx.CallAction("get_group_member_list", zero.Params{
				"group_id": listgid,
				"no_cache": false,
			}).Data
			if len(temp) == 0 {
				temp = list.Array()
				sort.SliceStable(temp, func(i, j int) bool {
					return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
				})
				temp = temp[math.Max(0, len(temp)-30):]
			}
			uid := ctx.Event.UserID
			now := time.Now()
			s := md5.Sum(helper.StringToBytes(fmt.Sprintf("%d%d%d%d", uid, now.Year(), now.Month(), now.Day())))
			r := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(s[:]))))
			rn := r.Intn(len(temp))
			who := temp[rn]
			gid := ctx.Event.GroupID
			if listgid != gid {
				listgid = gid
				list = ctx.CallAction("get_group_member_list", zero.Params{
					"group_id": listgid,
					"no_cache": false,
				}).Data
				temp = list.Array()
				sort.SliceStable(temp, func(i, j int) bool {
					return temp[i].Get("last_sent_time").Int() < temp[j].Get("last_sent_time").Int()
				})
				temp = temp[math.Max(0, len(temp)-30):]
				rn = r.Intn(len(temp))
				who = temp[rn]
				listgid = gid
			}
			userid := who.Get("user_id").Int()
			if userid == uid {
				me = who
				temp = append(temp[:rn], temp[rn:]...)
				rn = r.Intn(len(temp))
				who = temp[rn]
			}
			nick := who.Get("card").Str
			if nick == "" {
				nick = who.Get("nickname").Str
			}
			avtar := fmt.Sprintf("[CQ:image,file=http://q4.qlogo.cn/g?b=qq&nk=%s&s=640,cache=0]",strconv.FormatInt(who.Get("user_id").Int(), 10))
			msg := fmt.Sprintf("[CQ:at,qq=%d]今天你的群友老婆是%s\n【%s】(%d)哒！",uid,avtar, nick, userid)
			msg = message.UnescapeCQCodeText(msg)
			ctx.SendGroupMessage(ctx.Event.GroupID, message.ParseMessageFromString(msg))
			if len(me.Array()) != 0 {
				mlist := append(temp[:rn], me)
				temp = append(mlist, temp[rn:]...)
			}
		})
	engine.OnFullMatch("换个老婆", zero.OnlyGroup,zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			list = ctx.CallAction("get_group_member_list", zero.Params{
				"group_id": ctx.Event.GroupID,
				"no_cache": true,
			}).Data
			ctx.SendChain(message.Text("换好了！"))
		})
}
