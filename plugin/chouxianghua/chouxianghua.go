// Package chouxianghua 抽象话转化
package chouxianghua

import (
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "翻译为抽象话",
		Help:             "- 抽象翻译xxx",
		PublicDataFolder: "ChouXiangHua",
	})

	en.OnRegex("^抽象翻译((\\s|[\\r\\n]|[\\p{Han}\\p{P}A-Za-z0-9])+)$",
		fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
			db.DBPath = en.DataFolder() + "cxh.db"
			// os.RemoveAll(dbpath)
			_, _ = en.GetLazyData("cxh.db", true)
			err := db.Open(time.Hour)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			err = db.Create("pinyin", &pinyin{})
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			n, err := db.Count("pinyin")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			logrus.Printf("[chouxianghua]读取%d条拼音", n)
			return true
		}),
	).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			r := cx(ctx.State["regex_matched"].([]string)[1])
			ctx.SendChain(message.Text(r))
		})
}

func cx(s string) (r string) {
	h := []rune(s)
	for i := 0; i < len(h); i++ {
		if i < len(h)-1 {
			e := getEmojiByPronun(getPronunByDWord(h[i], h[i+1]))
			if e != "" {
				r += e
				i++
				continue
			}
		}
		e := getEmojiByPronun(getPinyinByWord(string(h[i])))
		if e != "" {
			r += e
			continue
		}
		r += string(h[i])
	}
	return
}
