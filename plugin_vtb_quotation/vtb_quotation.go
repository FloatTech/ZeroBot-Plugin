package plugin_vtb_quotation

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
)

var (
	regStr = ".*/(.*)"
)

var engine = control.Register("vtbquotation", &control.Options{
	DisableOnDefault: false,
	Help: "vtb语录\n" +
		"随机vtb\n",
})

func init() {
	engine.OnFullMatch("vtb语录").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var firstIndex int
			var secondIndex int
			var thirdIndex int
			echo, cancel := ctx.FutureEvent("message",
				ctx.CheckSession()). // 只复读开启复读模式的人的消息
				Repeat()             // 不断监听复读
			db, err := gorm.Open("sqlite3", "data/VtbQuotation/vtb.db")
			if err != nil {
				panic("failed to connect database")
			}
			defer db.Close()
			firstStepMessage := model.GetAllFirstCategoryMessage(db)
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(firstStepMessage))
			//步骤1，2，3，依次选择3个类别
			step := 1
			//错误次数
			errorCount := 0
			for {
				select {
				case e := <-echo: // 接收到需要复读的消息
					//错误次数达到3次，结束命令
					if errorCount == 3 {
						ctx.SendChain(message.Reply(e.MessageID), message.Text("输入错误太多,请重新发指令"))
						cancel()
						return
					}
					if step == 1 {
						firstIndex, err = strconv.Atoi(e.RawMessage)
						log.Println(fmt.Sprintf("当前在第%d步", step))
						log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号,三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							SecondStepMessage := model.GetAllSecondCategoryMessageByFirstIndex(db, firstIndex)
							log.Println(SecondStepMessage)
							if SecondStepMessage == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								ctx.SendChain(message.Reply(e.MessageID), message.Text(model.GetAllFirstCategoryMessage(db)))
								errorCount++
							} else {
								ctx.SendChain(message.Reply(e.MessageID), message.Text(SecondStepMessage))
								step++
							}
						}
					} else if step == 2 {
						secondIndex, err = strconv.Atoi(e.RawMessage)
						log.Println(fmt.Sprintf("当前在第%d步", step))
						log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							ThirdStepMessage := model.GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(db, firstIndex, secondIndex)
							log.Println(ThirdStepMessage)
							if ThirdStepMessage == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								ctx.SendChain(message.Reply(e.MessageID), message.Text(model.GetAllSecondCategoryMessageByFirstIndex(db, firstIndex)))
								errorCount++
							} else {
								ctx.SendChain(message.Reply(e.MessageID), message.Text(ThirdStepMessage))
								step++
							}
						}
					} else if step == 3 {
						thirdIndex, err = strconv.Atoi(e.RawMessage)
						log.Println(fmt.Sprintf("当前在第%d步", step))
						log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							tc := model.GetThirdCategory(db, firstIndex, secondIndex, thirdIndex)
							reg := regexp.MustCompile(regStr)
							recordUrl := tc.ThirdCategoryPath
							if recordUrl == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("没有内容请重新选择，三次输入错误，指令可退出重输"))
								ctx.SendChain(message.Reply(e.MessageID), message.Text(model.GetAllFirstCategoryMessage(db)))
								errorCount++
								step = 1
							} else {
								if reg.MatchString(recordUrl) {
									log.Println(reg.FindStringSubmatch(recordUrl)[1])
									log.Println(url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]))
									recordUrl = strings.Replace(recordUrl, reg.FindStringSubmatch(recordUrl)[1], url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]), -1)
									recordUrl = strings.Replace(recordUrl, "+", "%20", -1)
									log.Println(recordUrl)
								}
								ctx.SendChain(message.Reply(e.MessageID), message.Text("请欣赏《"+tc.ThirdCategoryName+"》"))
								ctx.SendChain(message.Record(recordUrl))
								cancel()
								return
							}
						}
					}
				case <-time.After(time.Second * 60):
					cancel()
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("vtb语录指令过期"))
					return
				}
			}
		})
	engine.OnFullMatch("随机vtb").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := gorm.Open("sqlite3", "data/VtbQuotation/vtb.db")
			if err != nil {
				panic("failed to connect database")
			}
			defer db.Close()
			tc := model.RandomVtb(db)
			fc := model.GetFirstCategoryByFirstUid(db, tc.FirstCategoryUid)
			if (tc != model.ThirdCategory{}) && (fc != model.FirstCategory{}) {
				reg := regexp.MustCompile(regStr)
				recordUrl := tc.ThirdCategoryPath
				if reg.MatchString(recordUrl) {
					log.Println(reg.FindStringSubmatch(recordUrl)[1])
					log.Println(url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]))
					recordUrl = strings.Replace(recordUrl, reg.FindStringSubmatch(recordUrl)[1], url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]), -1)
					recordUrl = strings.Replace(recordUrl, "+", "%20", -1)
					log.Println(recordUrl)
				}
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请欣赏"+fc.FirstCategoryName+"的《"+tc.ThirdCategoryName+"》"))
				ctx.SendChain(message.Record(recordUrl))
			}
		})
}
