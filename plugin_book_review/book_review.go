package plugin_book_review

import (
	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_book_review/model"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const dbpath = "data/BookReview/"
const dbfile = dbpath + "bookreview.db"

var (
	blank = `\n\s*\n`
	//blank = `\n`
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
				logrus.Errorln(err)
				return
			}
			defer db.Close()
			BookReviewList := db.GetBookReviewByKeyword(ctx.State["regex_matched"].([]string)[1])
			ctx.SendChain(message.Text(BookReviewList.BookReview))
			/*
				var m message.Message
				for _,v := range BookReviewList{
					m = append(m,
						message.CustomNode(
							ctx.Event.Sender.NickName,
							ctx.Event.UserID,
							v.BookReview,
						))

					ctx.SendChain(message.Text(v.BookReview))
				}


				ctx.SendGroupForwardMessage(
					ctx.Event.GroupID,
					m,
				)

			*/

		})

	engine.OnFullMatch("随机书评").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {

			db, err := model.Open(dbfile)
			if err != nil {
				logrus.Errorln(err)
				return
			}
			defer db.Close()
			br := db.GetRandomBookReview()
			// 合并消息屡次出错暂时不用
			ctx.SendChain(message.Text(br.BookReview))
			/*

				var m message.Message
				m = append(m,
					message.CustomNode(
						ctx.Event.Sender.NickName,
						ctx.Event.UserID,
						s,

				}

				log.Println("合并消息的数量为:",len(m))
				ctx.SendGroupForwardMessage(
					ctx.Event.GroupID,
					m,
				)

			*/
		})

}
