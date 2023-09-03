// Package vtbquotation vtb经典语录
package vtbquotation

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/vtbquotation/model"
)

var reg = regexp.MustCompile(".*/(.*)")

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "vtb语录",
		Help:             "- vtb语录\n- 随机vtb\n- 更新vtb\n来源: vtbkeyboard.moe",
		PublicDataFolder: "VtbQuotation",
	})
	dbfile := engine.DataFolder() + "vtb.db"
	storePath := engine.DataFolder() + "store/"
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		err := os.MkdirAll(storePath, 0755)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		_, err = engine.GetLazyData("vtb.db", false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})

	engine.OnFullMatch("vtb语录", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			indexs := [3]int{}
			echo, cancel := ctx.FutureEvent("message",
				ctx.CheckSession()). // 只复读开启复读模式的人的消息
				Repeat()             // 不断监听复读
			db, err := model.Open(dbfile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer db.Close()
			defer cancel()
			r, err := db.GetAllFirstCategoryMessage()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			firstStepImageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
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
				case c := <-echo: // 接收到需要复读的消息
					// 错误次数达到3次，结束命令
					if errorCount >= 3 {
						ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("输入错误太多,请重新发指令"))
						return
					}
					msg := c.Event.Message.ExtractPlainText()
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("请输入正确的序号，三次输入错误，指令可退出重输"))
						errorCount++
						continue
					}
					switch step {
					case 0:
						indexs[0] = num
						secondStepMessage, err := db.GetAllSecondCategoryMessageByFirstIndex(indexs[0])
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						if secondStepMessage == "" {
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
							r, err := db.GetAllFirstCategoryMessage()
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							firstStepImageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR: 可能被风控了"))
							}
							errorCount++
						} else {
							secondStepMessageBytes, err := text.RenderToBase64(secondStepMessage, text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR: 可能被风控了"))
							}
							step++
						}

					case 1:
						indexs[1] = num
						thirdStepMessage, err := db.GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(indexs[0], indexs[1])
						if err != nil {
							ctx.SendChain(message.Text("ERROR: ", err))
							return
						}
						if thirdStepMessage == "" {
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("你选择的序号没有内容，请重新选择，三次输入错误，指令可退出重输"))
							r, err := db.GetAllSecondCategoryMessageByFirstIndex(indexs[0])
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							secondStepMessageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(secondStepMessageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR: 可能被风控了"))
							}
							errorCount++
						} else {
							thirdStepMessageBytes, err := text.RenderToBase64(thirdStepMessage, text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(thirdStepMessageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR: 可能被风控了"))
							}
							step++
						}
					case 2:
						indexs[2] = num
						tc := db.GetThirdCategory(indexs[0], indexs[1], indexs[2])
						recURL := tc.ThirdCategoryPath
						if recURL == "" {
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("没有内容请重新选择，三次输入错误，指令可退出重输"))
							r, err := db.GetAllFirstCategoryMessage()
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							firstStepImageBytes, err := text.RenderToBase64(r, text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(firstStepImageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR: 可能被风控了"))
							}
							errorCount++
							step = 1
						} else {
							if reg.MatchString(recURL) {
								recURL = strings.ReplaceAll(recURL, reg.FindStringSubmatch(recURL)[1], url.QueryEscape(reg.FindStringSubmatch(recURL)[1]))
								recURL = strings.ReplaceAll(recURL, "+", "%20")
							}
							ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("请欣赏《"+tc.ThirdCategoryName+"》"))
							recordFile := storePath + fmt.Sprintf("%d-%d-%d", indexs[0], indexs[1], indexs[2]) + path.Ext(recURL)
							if file.IsExist(recordFile) {
								ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
								return
							}
							err = initRecord(recordFile, recURL)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
							return
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
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer db.Close()
			tc := db.RandomVtb()
			fc := db.GetFirstCategoryByFirstUID(tc.FirstCategoryUID)
			if (tc != model.ThirdCategory{}) && (fc != model.FirstCategory{}) {
				recURL := tc.ThirdCategoryPath
				if reg.MatchString(recURL) {
					recURL = strings.ReplaceAll(recURL, reg.FindStringSubmatch(recURL)[1], url.QueryEscape(reg.FindStringSubmatch(recURL)[1]))
					recURL = strings.ReplaceAll(recURL, "+", "%20")
				}
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请欣赏"+fc.FirstCategoryName+"的《"+tc.ThirdCategoryName+"》"))
				recordFile := storePath + fmt.Sprintf("%d-%d-%d", fc.FirstCategoryIndex, tc.SecondCategoryIndex, tc.ThirdCategoryIndex) + path.Ext(recURL)
				if file.IsExist(recordFile) {
					ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + recordFile))
					return
				}
				err = initRecord(recordFile, recURL)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
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
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				for _, v := range vl {
					err = db.StoreVtb(v)
					if err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
				}
				err = db.Close()
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
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
