// Package vtbquotation vtb经典语录
package vtbquotation

import (
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
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/web"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/vtb_quotation/model"
)

const regStr = ".*/(.*)"
const recordRe = "(\\.mp3|\\.wav|\\.wma|\\.mpa|\\.ram|\\.ra|\\.aac|\\.aif|\\.m4a|\\.tsa)"

var (
	re = regexp.MustCompile(recordRe)
)

func init() {
	engine := control.Register("vtbquotation", &control.Options{
		DisableOnDefault: false,
		Help:             "vtbkeyboard.moe\n- vtb语录\n- 随机vtb\n- 更新vtb\n",
		PublicDataFolder: "VtbQuotation",
	})
	dbfile := engine.DataFolder() + "vtb.db"
	storePath := engine.DataFolder() + "store/"
	getdb := ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		err := os.MkdirAll(storePath, 0755)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		_, err = file.GetLazyData(dbfile, false, false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		return true
	})

	engine.OnFullMatch("vtb语录", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var firstIndex int
			var secondIndex int
			var thirdIndex int
			echo, cancel := ctx.FutureEvent("message",
				ctx.CheckSession()). // 只复读开启复读模式的人的消息
				Repeat()             // 不断监听复读
			db, err := model.Open(dbfile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			defer db.Close()
			defer cancel()
			r, err := db.GetAllFirstCategoryMessage()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			firstStepImageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
			// 步骤0，1，2，依次选择3个类别
			step := 0
			// 错误次数
			errorCount := 0
			for {
				select {
				case c := <-echo: // 接收到需要复读的消息
					// 错误次数达到3次，结束命令
					if errorCount >= 3 {
						ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("输入错误太多,请重新发指令"))
						return
					}
					switch step {
					case 0:
						firstIndex, err = strconv.Atoi(c.Event.RawMessage)
						// log.Debugln(fmt.Sprintf("当前在第%d步", step))
						// log.Debugln(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("请输入正确的序号,三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							secondStepMessage, err := db.GetAllSecondCategoryMessageByFirstIndex(firstIndex)
							if err != nil {
								ctx.SendChain(message.Text("ERROR:", err))
								return
							}
							// log.Debugln(secondStepMessage)
							if secondStepMessage == "" {
								ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								r, err := db.GetAllFirstCategoryMessage()
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								firstStepImageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR:可能被风控了"))
								}
								errorCount++
							} else {
								secondStepMessageBytes, err := text.RenderToBase64(secondStepMessage, text.FontFile, 400, 20)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR:可能被风控了"))
								}
								step++
							}
						}
					case 1:
						secondIndex, err = strconv.Atoi(c.Event.RawMessage)
						// log.Debugln(fmt.Sprintf("当前在第%d步", step))
						// log.Debugln(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							thirdStepMessage, err := db.GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(firstIndex, secondIndex)
							if err != nil {
								ctx.SendChain(message.Text("ERROR:", err))
								return
							}
							// log.Debugln(thirdStepMessage)
							if thirdStepMessage == "" {
								ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
								r, err := db.GetAllSecondCategoryMessageByFirstIndex(firstIndex)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								secondStepMessageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR:可能被风控了"))
								}
								errorCount++
							} else {
								thirdStepMessageBytes, err := text.RenderToBase64(thirdStepMessage, text.FontFile, 400, 20)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(thirdStepMessageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR:可能被风控了"))
								}
								step++
							}
						}
					case 2:
						thirdIndex, err = strconv.Atoi(c.Event.RawMessage)
						// log.Debugln(fmt.Sprintf("当前在第%d步", step))
						// log.Debugln(fmt.Sprintf("firstIndex:%d,secondIndex:%d,thirdIndex:%d", firstIndex, secondIndex, thirdIndex))
						if err != nil {
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
							errorCount++
						} else {
							tc := db.GetThirdCategory(firstIndex, secondIndex, thirdIndex)
							reg := regexp.MustCompile(regStr)
							recURL := tc.ThirdCategoryPath
							if recURL == "" {
								ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("没有内容请重新选择，三次输入错误，指令可退出重输"))
								r, err := db.GetAllFirstCategoryMessage()
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								firstStepImageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
								if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
									ctx.SendChain(message.Text("ERROR:可能被风控了"))
								}
								errorCount++
								step = 1
							} else {
								if reg.MatchString(recURL) {
									// log.Debugln(reg.FindStringSubmatch(recordUrl)[1])
									// log.Debugln(url.QueryEscape(reg.FindStringSubmatch(recordUrl)[1]))
									recURL = strings.ReplaceAll(recURL, reg.FindStringSubmatch(recURL)[1], url.QueryEscape(reg.FindStringSubmatch(recURL)[1]))
									recURL = strings.ReplaceAll(recURL, "+", "%20")
									// log.Debugln(recordUrl)
								}
								ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("请欣赏《"+tc.ThirdCategoryName+"》"))

								if !re.MatchString(recURL) {
									ctx.SendChain(message.Text("ERROR:文件格式不匹配"))
									return
								}
								format := re.FindStringSubmatch(recURL)[1]
								recordFile := storePath + fmt.Sprintf("%d-%d-%d", firstIndex, secondIndex, thirdIndex) + format
								if file.IsExist(recordFile) {
									ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
									return
								}
								err = initRecord(recordFile, recURL)
								if err != nil {
									ctx.SendChain(message.Text("ERROR:", err))
									return
								}
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
	engine.OnFullMatch("随机vtb", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			db, err := model.Open(dbfile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			defer db.Close()
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
					ctx.SendChain(message.Text("ERROR:文件格式不匹配"))
					return
				}
				format := re.FindStringSubmatch(recURL)[1]
				recordFile := storePath + fmt.Sprintf("%d-%d-%d", fc.FirstCategoryIndex, tc.SecondCategoryIndex, tc.ThirdCategoryIndex) + format
				if file.IsExist(recordFile) {
					ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
					return
				}
				err = initRecord(recordFile, recURL)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
			}
		})
	engine.OnFullMatch("更新vtb", zero.SuperUserPermission, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send("少女祈祷中......")
			db := model.Initialize(dbfile)
			if db != nil {
				vl, err := db.GetVtbList()
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				for _, v := range vl {
					err = db.StoreVtb(v)
					if err != nil {
						ctx.SendChain(message.Text("ERROR:", err))
						return
					}
				}
				err = db.Close()
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
			}
			ctx.Send("vtb数据库已更新")
		})
}

func initRecord(recordFile, recordURL string) error {
	if file.IsNotExist(recordFile) {
		client := web.NewTLS12Client()
		req, _ := http.NewRequest("GET", recordURL, nil)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		err = os.WriteFile(recordFile, data, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}
