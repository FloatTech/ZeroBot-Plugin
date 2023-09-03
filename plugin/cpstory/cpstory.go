// Package cpstory cp短打
package cpstory

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "cp短打", // 这里也许有更好的名字
		Help:             "- 组cp[@xxx][@xxx]\n- 磕cp大老师 雪乃",
		PublicDataFolder: "CpStory",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = engine.DataFolder() + "cp.db"
		// os.RemoveAll(dbpath)
		_, _ = engine.GetLazyData("cp.db", true)
		err := db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Create("cp_story", &cpstory{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		n, err := db.Count("cp_story")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		logrus.Printf("[cpstory]读取%d条故事", n)
		return true
	})

	engine.OnRegex("^组cp.*?(\\d+).*?(\\d+)", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		cs := getRandomCpStory()
		gong := ctx.CardOrNickName(math.Str2Int64(ctx.State["regex_matched"].([]string)[1]))
		shou := ctx.CardOrNickName(math.Str2Int64(ctx.State["regex_matched"].([]string)[2]))
		text := strings.ReplaceAll(cs.Story, "<攻>", gong)
		text = strings.ReplaceAll(text, "<受>", shou)
		text = strings.ReplaceAll(text, cs.Gong, gong)
		text = strings.ReplaceAll(text, cs.Shou, shou)
		ctx.SendChain(message.Text(text))
	})
	engine.OnPrefix("磕cp", getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
