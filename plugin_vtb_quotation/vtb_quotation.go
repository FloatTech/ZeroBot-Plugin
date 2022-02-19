// Package vtbquotation vtb经典语录
package vtbquotation

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/fumiama/sqlite3" // use sql
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/text"

	"github.com/FloatTech/zbputils/control/order"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
)

const regStr = ".*/(.*)"
const recordRe = "(\\.mp3|\\.wav|\\.wma|\\.mpa|\\.ram|\\.ra|\\.aac|\\.aif|\\.m4a|\\.tsa)"

var (
	re = regexp.MustCompile(recordRe)
)

func init() {
	engine := control.Register("vtbquotation", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "vtbkeyboard.moe\n- vtb语录\n- 随机vtb\n- 更新vtb\n",
		PublicDataFolder: "VtbQuotation",
	})
	dbfile := engine.DataFolder() + "vtb.db"
	storePath := engine.DataFolder() + "store/"
	go func() {
		defer order.DoneOnExit()()
		err := os.MkdirAll(storePath, 0755)
		if err != nil {
			panic(err)
		}
		_, _ = file.GetLazyData(dbfile, false, false)
	}()
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
			firstStepImageBytes, err := text.RenderToBase64(db.GetAllFirstCategoryMessage(), text.FontFile, 400, 20)
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
								firstStepImageBytes, err := text.RenderToBase64(db.GetAllFirstCategoryMessage(), text.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								errorCount++
							} else {
								secondStepMessageBytes, err := text.RenderToBase64(secondStepMessage, text.FontFile, 400, 20)
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
								secondStepMessageBytes, err := text.RenderToBase64(db.GetAllSecondCategoryMessageByFirstIndex(firstIndex), text.FontFile, 400, 20)
								if err != nil {
									log.Errorln("[vtb]:", err)
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR: 可能被风控了"))
								}
								errorCount++
							} else {
								thirdStepMessageBytes, err := text.RenderToBase64(thirdStepMessage, text.FontFile, 400, 20)
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
								firstStepImageBytes, err := text.RenderToBase64(db.GetAllFirstCategoryMessage(), text.FontFile, 400, 20)
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

								if !re.MatchString(recURL) {
									log.Errorln("[vtb]:文件格式不匹配")
									return
								}
								format := re.FindStringSubmatch(recURL)[1]
								recordFile := storePath + fmt.Sprintf("%d-%d-%d", firstIndex, secondIndex, thirdIndex) + format
								if file.IsExist(recordFile) {
									ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
									return
								}
								initRecord(recordFile, recURL)
								ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
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
				if !re.MatchString(recURL) {
					log.Errorln("[vtb]:文件格式不匹配")
					return
				}
				format := re.FindStringSubmatch(recURL)[1]
				recordFile := storePath + fmt.Sprintf("%d-%d-%d", fc.FirstCategoryIndex, tc.SecondCategoryIndex, tc.ThirdCategoryIndex) + format
				if file.IsExist(recordFile) {
					ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
					return
				}
				initRecord(recordFile, recURL)
				ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
			}
			db.Close()
		})
	engine.OnFullMatch("更新vtb", zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			db := model.Initialize(dbfile)
			if db != nil {
				for _, v := range db.GetVtbList() {
					db.StoreVtb(v)
				}
				err := db.Close()
				if err != nil {
					log.Errorln("[vtb/cron]", err)
				}
			}
			ctx.Send("vtb数据库已更新")
		})
}

func initRecord(recordFile, recordURL string) {
	if file.IsNotExist(recordFile) {
		transport := http.Transport{
			TLSClientConfig: &tls.Config{
				MaxVersion: tls.VersionTLS12,
			},
		}
		client := &http.Client{
			Transport: &transport,
		}
		req, _ := http.NewRequest("GET", recordURL, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
		resp, err := client.Do(req)
		if err != nil {
			log.Errorln("[vtb]:", err)
			return
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorln("[vtb]:", err)
			return
		}
		err = os.WriteFile(recordFile, data, 0666)
		if err != nil {
			log.Errorln("[vtb]:", err)
		}
	}
}
