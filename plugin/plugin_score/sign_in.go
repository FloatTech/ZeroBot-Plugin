// Package score 签到，答题得分
package score

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fogleman/gg"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/FloatTech/zbputils/web"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	backgroundURL = "https://iw233.cn/API/pc.php?type=json"
	referer       = "https://iw233.cn/main.html"
	ua            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	signinMax     = 1
	// SCOREMAX 分数上限定为120
	SCOREMAX = 120
)

var levelArray = [...]int{0, 1, 2, 5, 10, 20, 35, 55, 75, 100, 120}

func init() {
	engine := control.Register("score", order.AcquirePrio(), &control.Options{
		DisableOnDefault:  false,
		Help:              "签到得分\n- 签到\n- 获得签到背景[@xxx] | 获得签到背景",
		PrivateDataFolder: "score",
	})
	cachePath := engine.DataFolder() + "cache/"
	go func() {
		defer order.DoneOnExit()()
		os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
		_, err = file.GetLazyData(text.BoldFontFile, false, true)
		if err != nil {
			panic(err)
		}
		_, err = file.GetLazyData(text.FontFile, false, true)
		if err != nil {
			panic(err)
		}
		sdb = initialize(engine.DataFolder() + "score.db")
		log.Println("[score]加载score数据库")
	}()
	engine.OnFullMatch("签到", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			now := time.Now()
			today := now.Format("20060102")
			si := sdb.GetSignInByUID(uid)
			siUpdateTimeStr := si.UpdatedAt.Format("20060102")
			if siUpdateTimeStr != today {
				_ = sdb.InsertOrUpdateSignInCountByUID(uid, 0)
			}

			drawedFile := cachePath + strconv.FormatInt(uid, 10) + today + "signin.png"
			if si.Count >= signinMax && siUpdateTimeStr == today {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("今天你已经签到过了！"))
				if file.IsExist(drawedFile) {
					ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				}
				return
			}

			picFile := cachePath + strconv.FormatInt(uid, 10) + today + ".png"
			initPic(picFile)

			_ = sdb.InsertOrUpdateSignInCountByUID(uid, si.Count+1)

			back, err := gg.LoadImage(picFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}

			// 避免图片过大，最大 1280*720
			back = img.Limit(back, 1280, 720)

			canvas := gg.NewContext(back.Bounds().Size().X, int(float64(back.Bounds().Size().Y)*1.7))
			canvas.SetRGB(1, 1, 1)
			canvas.Clear()
			canvas.DrawImage(back, 0, 0)

			monthWord := now.Format("01/02")
			hourWord := getHourWord(now)
			if err = canvas.LoadFontFace(text.BoldFontFile, float64(back.Bounds().Size().X)*0.1); err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			canvas.SetRGB(0, 0, 0)
			canvas.DrawString(hourWord, float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.2)
			canvas.DrawString(monthWord, float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*1.2)
			nickName := ctxext.CardOrNickName(ctx, uid)
			if err = canvas.LoadFontFace(text.FontFile, float64(back.Bounds().Size().X)*0.04); err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			add := 1
			canvas.DrawString(nickName+fmt.Sprintf(" 小熊饼干+%d", add), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.3)
			score := sdb.GetScoreByUID(uid).Score
			score += add
			if score > SCOREMAX {
				score = SCOREMAX
				ctx.SendChain(message.At(uid), message.Text("你获得的小熊饼干已经达到上限"))
			}
			_ = sdb.InsertOrUpdateScoreByUID(uid, score)
			level := getLevel(score)
			canvas.DrawString("当前小熊饼干:"+strconv.FormatInt(int64(score), 10), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.4)
			canvas.DrawString("LEVEL:"+strconv.FormatInt(int64(level), 10), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.5)
			canvas.DrawRectangle(float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.55, float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*0.1)
			canvas.SetRGB255(150, 150, 150)
			canvas.Fill()
			var nextLevelScore int
			if level < 10 {
				nextLevelScore = levelArray[level+1]
			} else {
				nextLevelScore = SCOREMAX
			}
			canvas.SetRGB255(0, 0, 0)
			canvas.DrawRectangle(float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.55, float64(back.Bounds().Size().X)*0.6*float64(score)/float64(nextLevelScore), float64(back.Bounds().Size().Y)*0.1)
			canvas.SetRGB255(102, 102, 102)
			canvas.Fill()
			canvas.DrawString(fmt.Sprintf("%d/%d", score, nextLevelScore), float64(back.Bounds().Size().X)*0.75, float64(back.Bounds().Size().Y)*1.62)

			f, err := os.Create(drawedFile)
			if err != nil {
				log.Errorln("[score]", err)
				data, cl := writer.ToBytes(canvas.Image())
				ctx.SendChain(message.ImageBytes(data))
				cl()
				return
			}
			_, err = writer.WriteTo(canvas.Image(), f)
			_ = f.Close()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})
	engine.OnPrefix("获得签到背景", zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			var uidStr string
			if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
				uidStr = ctx.Event.Message[1].Data["qq"]
			} else if param == "" {
				uidStr = strconv.FormatInt(ctx.Event.UserID, 10)
			}
			picFile := cachePath + uidStr + time.Now().Format("20060102") + ".png"
			if file.IsNotExist(picFile) {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请先签到！"))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + picFile))
		})
}

func getHourWord(t time.Time) string {
	switch {
	case 6 <= t.Hour() && t.Hour() < 12:
		return "早上好"
	case 12 <= t.Hour() && t.Hour() < 14:
		return "中午好"
	case 14 <= t.Hour() && t.Hour() < 19:
		return "下午好"
	case 19 <= t.Hour() && t.Hour() < 24:
		return "晚上好"
	case 0 <= t.Hour() && t.Hour() < 6:
		return "凌晨好"
	default:
		return ""
	}
}

func getLevel(count int) int {
	for k, v := range levelArray {
		if count == v {
			return k
		} else if count < v {
			return k - 1
		}
	}
	return -1
}

func initPic(picFile string) {
	if file.IsNotExist(picFile) {
		data, err := web.ReqWith(backgroundURL, "GET", referer, ua)
		if err != nil {
			log.Errorln("[score]", err)
		}
		picURL := gjson.Get(string(data), "pic").String()
		data, err = web.ReqWith(picURL, "GET", "", ua)
		if err != nil {
			log.Errorln("[score]", err)
		}
		err = os.WriteFile(picFile, data, 0666)
		if err != nil {
			log.Errorln("[score]", err)
		}
	}
}
