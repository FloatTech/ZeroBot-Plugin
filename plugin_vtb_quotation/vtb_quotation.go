// Package vtbquotation vtb经典语录
package vtbquotation

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/logoove/sqlite" // use sql
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/txt2img"

	"github.com/FloatTech/ZeroBot-Plugin/order"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
)

const (
	regStr = ".*/(.*)"
	dbpath = "data/VtbQuotation/"
	dbfile = dbpath + "vtb.db"
)

var engine = control.Register("vtbquotation", order.PrioVtbQuotation, &control.Options{
	DisableOnDefault: false,
	Help:             "vtbkeyboard.moe\n- vtb语录\n- 随机vtb",
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
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln("[vtb]:", err)
				return
			}
			defer db.Close()
			defer cancel()
			firstStepImageBytes, err := txt2img.RenderToBase64(db.GetAllFirstCategoryMessage(), txt2img.FontFile, 400, 20)
			if err != nil {
				log.Errorln("[vtb]:", err)
			}
			if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			// 步骤0，1，2，依次选择3个类别
			step := 0
			// 错误次数
			errorCount := 0
			for {
				select {
				case e := <-echo: // 接收到需要复读的消息
					// 错误次数达到3次，结束命令
					if errorCount >= 3 {
						ctx.SendChain(message.Reply(e.MessageID), message.Text("输入错误太多,请重新发指令"))
						return
					}
					switch step {
					case 0:
						firstIndex, err = strconv.Atoi(e.RawMessage)
						// log.Println(fmt.Sprintf("当前在第%d步", step))
						// log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号,三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							secondStepMessage := db.GetAllSecondCategoryMessageByFirstIndex(firstIndex)
							// log.Println(secondStepMessage)
							if secondStepMessage == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								firstStepImageBytes, err := txt2img.RenderToBase64(db.GetAllFirstCategoryMessage(), txt2img.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								errorCount++
							} else {
								secondStepMessageBytes, err := txt2img.RenderToBase64(secondStepMessage, txt2img.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								step++
							}
						}
					case 1:
						secondIndex, err = strconv.Atoi(e.RawMessage)
						// log.Println(fmt.Sprintf("当前在第%d步", step))
						// log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							thirdStepMessage := db.GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(firstIndex, secondIndex)
							// log.Println(thirdStepMessage)
							if thirdStepMessage == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								secondStepMessageBytes, err := txt2img.RenderToBase64(db.GetAllSecondCategoryMessageByFirstIndex(firstIndex), txt2img.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								errorCount++
							} else {
								thirdStepMessageBytes, err := txt2img.RenderToBase64(thirdStepMessage, txt2img.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(thirdStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								step++
							}
						}
					case 2:
						thirdIndex, err = strconv.Atoi(e.RawMessage)
						// log.Println(fmt.Sprintf("当前在第%d步", step))
						// log.Println(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(e.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							tc := db.GetThirdCategory(firstIndex, secondIndex, thirdIndex)
							reg := regexp.MustCompile(regStr)
							recURL := tc.ThirdCategoryPath
							if recURL == "" {
								ctx.SendChain(message.Reply(e.MessageID), message.Text("没有内容请重新选择，三次输入错误，指令可退出重输"))
								firstStepImageBytes, err := txt2img.RenderToBase64(db.GetAllFirstCategoryMessage(), txt2img.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								errorCount++
								step = 1
							} else {
								if reg.MatchString(recURL) {
									// log.Println(reg.FindStringSubmatch(recordUrl)[1])
									// log.Println(url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]))
									recURL = strings.ReplaceAll(recURL, reg.FindStringSubmatch(recURL)[1], url.QueryEscape(reg.FindStringSubmatch(recURL)[1]))
									recURL = strings.ReplaceAll(recURL, "+", "%20")
									// log.Println(recordUrl)
								}
								ctx.SendChain(message.Reply(e.MessageID), message.Text("请欣赏《"+tc.ThirdCategoryName+"》"))
								ctx.SendChain(message.Record(recURL))
								return
							}
						}
					default:
						return
					}
				case <-time.After(time.Second * 60):
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("vtb语录指令过期"))
					return
				}
			}
		})
	engine.OnFullMatch("随机vtb").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := model.Open(dbfile)
			if err != nil {
				log.Errorln("[vtb]:", err)
				return
			}
			tc := db.RandomVtb()
			fc := db.GetFirstCategoryByFirstUID(tc.FirstCategoryUID)
			if (tc != model.ThirdCategory{}) && (fc != model.FirstCategory{}) {
				reg := regexp.MustCompile(regStr)
				recURL := tc.ThirdCategoryPath
				if reg.MatchString(recURL) {
					recURL = strings.ReplaceAll(recURL, reg.FindStringSubmatch(recURL)[1], url.QueryEscape(reg.FindStringSubmatch(recURL)[1]))
					recURL = strings.ReplaceAll(recURL, "+", "%20")
				}
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请欣赏"+fc.FirstCategoryName+"的《"+tc.ThirdCategoryName+"》"))
				ctx.SendChain(message.Record(recURL))
			}
			db.Close()
		})
}
