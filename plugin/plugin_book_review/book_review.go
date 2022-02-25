// Package bookreview 书评
package bookreview

import (
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/text"
)

func init() {
	engine := control.Register("bookreview", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "哀伤雪刃推书记录\n- 书评[xxx]\n- 随机书评",
		PublicDataFolder: "BookReview",
	})

	go func() {
		dbpath := engine.DataFolder()
		db.DBPath = dbpath + "bookreview.db"
		defer order.DoneOnExit()()
		// os.RemoveAll(dbpath)
		_, _ = file.GetLazyData(db.DBPath, false, true)
		err := db.Create("bookreview", &book{})
		if err != nil {
			panic(err)
		}
		n, err := db.Count("bookreview")
		if err != nil {
			panic(err)
		}
		log.Printf("[bookreview]读取%d条书评", n)
	}()

	// 中文、英文、数字但不包括下划线等符号
	engine.OnRegex("^书评([\u4E00-\u9FA5A-Za-z0-9]{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			b := getBookReviewByKeyword(ctx.State["regex_matched"].([]string)[1])
			data, err := text.RenderToBase64(b.BookReview, text.FontFile, 400, 20)
			if err != nil {
				log.Println("err:", err)
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnFullMatch("随机书评").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			br := getRandomBookReview()
			data, err := text.RenderToBase64(br.BookReview, text.FontFile, 400, 20)
			if err != nil {
				log.Println("err:", err)
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
