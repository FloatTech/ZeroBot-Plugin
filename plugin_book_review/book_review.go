package plugin_book_review

import (
	log "github.com/sirupsen/logrus"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_book_review/model"
)

const dbpath = "data/BookReview/"
const dbfile = dbpath + "bookreview.db"

var (
	engine = control.Register("bookreview", &control.Options{
		DisableOnDefault: false,
		Help:             "哀伤雪刃推书记录\n- 书评[xxx]\n- 随机书评",
	})
)

func init() {
	engine.OnRegex("^书评(.{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln(err)
				return
			}
			BookReviewList := db.GetBookReviewByKeyword(ctx.State["regex_matched"].([]string)[1])
			ctx.SendChain(message.Text(BookReviewList.BookReview))
			db.Close()
		})

	engine.OnFullMatch("随机书评").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln(err)
				return
			}
			br := db.GetRandomBookReview()
			ctx.SendChain(message.Text(br.BookReview))
			db.Close()
		})

}
