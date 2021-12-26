package cpstory

import (
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/utils/math"
)

const (
	prio = 20
)

var (
	engine = control.Register("cpstory", &control.Options{
		DisableOnDefault: false,
		Help:             "cp短打\n- 组cp[@xxx][@xxx]\n- 组cp大老师 雪乃",
	})
	limit = rate.NewManager(time.Minute, 20)
)

func init() {
	engine.OnRegex("^组cp.*?(\\d+).*?(\\d+)").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		cs := getRandomCpStory()
		gong := getCardOrNickName(ctx, math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
		shou := getCardOrNickName(ctx, math.Str2Int64(ctx.State["regex_matched"].([]string)[2]))
		text := strings.ReplaceAll(cs.Story, "<攻>", gong)
		text = strings.ReplaceAll(text, "<受>", shou)
		text = strings.ReplaceAll(text, cs.Gong, gong)
		text = strings.ReplaceAll(text, cs.Shou, gong)
		ctx.SendChain(message.Text(text))
	})
	engine.OnPrefix("组cp").SetBlock(true).SetPriority(prio + 1).Handle(func(ctx *zero.Ctx) {
		cs := getRandomCpStory()
		params := strings.Split(ctx.State["args"].(string), " ")
		if len(params) < 2 {
			ctx.SendChain(message.Text(ctx.Event.MessageID), message.Text("请用空格分开两个人名"))
		} else {
			gong := params[0]
			shou := params[1]
			text := strings.ReplaceAll(cs.Story, "<攻>", gong)
			text = strings.ReplaceAll(text, "<受>", shou)
			text = strings.ReplaceAll(text, cs.Gong, gong)
			text = strings.ReplaceAll(text, cs.Shou, gong)
			ctx.SendChain(message.Text(text))
		}
	})
}

func getCardOrNickName(ctx *zero.Ctx, userId int64) (name string) {
	name = ctx.GetGroupMemberInfo(ctx.Event.GroupID, userId, false).Get("card").String()
	if name == "" {
		name = ctx.GetStrangerInfo(userId, false).Get("nickname").String()
	}
	return
}
