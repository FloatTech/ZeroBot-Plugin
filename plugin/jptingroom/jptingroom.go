// Package jptingroom 日语听力学习材料
package jptingroom

import (
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "日语听力学习材料",
		Help: "- 随机日语听力\n" +
			"- 随机日语歌曲\n" +
			"- 日语听力 xxx\n" +
			"- 日语歌曲 xxx\n",
		PublicDataFolder: "Jptingroom",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.DBPath = engine.DataFolder() + "item.db"
		_, err := engine.GetLazyData("item.db", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = db.Create("item", &item{})
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		n, err := db.Count("item")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		log.Infof("[jptingroom]读取%d条日语听力材料", n)
		return true
	})
	// 开启
	engine.OnFullMatchGroup([]string{"随机日语听力", "随机日语歌曲"}, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			matched := ctx.State["matched"].(string)
			var t item
			switch matched {
			case "随机日语听力":
				t = getRandomAudioByCategory("tingli")
			case "随机日语歌曲":
				t = getRandomAudioByCategory("gequ")
			default:
			}
			if t.AudioURL == "" {
				ctx.SendChain(message.Text("未能找到相关材料"))
				return
			}
			ctx.SendChain(message.Record(t.AudioURL))
			content := t.Title + "\n\n" + t.Content
			data, err := text.RenderToBase64(content, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
	engine.OnRegex(`日语(听力|歌曲)\s?([一-龥A-Za-z0-9ぁ-んァ-ヶ]{1,50})$`, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			var t item
			switch regexMatched[1] {
			case "听力":
				t = getRandomAudioByCategoryAndKeyword("tingli", regexMatched[2])
			case "歌曲":
				t = getRandomAudioByCategoryAndKeyword("gequ", regexMatched[2])
			default:
			}
			if t.AudioURL == "" {
				ctx.SendChain(message.Text("未能找到相关材料"))
				return
			}
			content := t.Title + "\n\n" + t.Content
			data, err := text.RenderToBase64(content, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
