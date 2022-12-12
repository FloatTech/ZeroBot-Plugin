// Package score 签到，答题得分
package score

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/golang/freetype"
	log "github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 货币系统
	"github.com/FloatTech/AnimeAPI/wallet"
)

const (
	backgroundURL = "https://img.moehu.org/pic.php?id=pc"
	signinMax     = 1
	// SCOREMAX 分数上限定为1200
	SCOREMAX = 1200
)

var (
	rankArray = [...]int{0, 10, 20, 50, 100, 200, 350, 550, 750, 1000, 1200}
	engine    = control.Register("score", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "签到",
		Help:              "- 签到\n- 获得签到背景[@xxx] | 获得签到背景\n- 查看等级排名\n注:为跨群排名\n- 查看我的钱包\n- 查看钱包排名\n注:为本群排行，若群人数太多不建议使用该功能!!!",
		PrivateDataFolder: "score",
	})
)

func init() {
	cachePath := engine.DataFolder() + "cache/"
	go func() {
		_ = os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
		sdb = initialize(engine.DataFolder() + "score.db")
	}()
	zero.OnFullMatch("查看我的钱包").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)
		ctx.SendChain(message.At(uid), message.Text("你的钱包当前有", money, "ATRI币"))
	})
	engine.OnFullMatch("签到").Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			now := time.Now()
			today := now.Format("20060102")
			// 签到图片
			drawedFile := cachePath + strconv.FormatInt(uid, 10) + today + "signin.png"
			picFile := cachePath + strconv.FormatInt(uid, 10) + today + ".png"
			// 获取签到时间
			si := sdb.GetSignInByUID(uid)
			siUpdateTimeStr := si.UpdatedAt.Format("20060102")
			switch {
			case si.Count >= signinMax && siUpdateTimeStr == today:
				// 如果签到时间是今天
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("今天你已经签到过了！"))
				if file.IsExist(drawedFile) {
					ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				}
				return
			case siUpdateTimeStr != today:
				// 如果是跨天签到就清数据
				err := sdb.InsertOrUpdateSignInCountByUID(uid, 0)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
			}
			// 更新签到次数
			err := sdb.InsertOrUpdateSignInCountByUID(uid, si.Count+1)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 更新经验
			level := sdb.GetScoreByUID(uid).Score + 1
			if level > SCOREMAX {
				level = SCOREMAX
				ctx.SendChain(message.At(uid), message.Text("你的等级已经达到上限"))
			}
			err = sdb.InsertOrUpdateScoreByUID(uid, level)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 更新钱包
			rank := getrank(level)
			add := 1 + rand.Intn(10) + rank*5 // 等级越高获得的钱越高
			err = wallet.InsertWalletOf(uid, add)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			score := wallet.GetWalletOf(uid)
			// 绘图
			err = initPic(picFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			back, err := gg.LoadImage(picFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
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
			_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if err = canvas.LoadFontFace(text.BoldFontFile, float64(back.Bounds().Size().X)*0.1); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			canvas.SetRGB(0, 0, 0)
			canvas.DrawString(hourWord, float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.2)
			canvas.DrawString(monthWord, float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*1.2)
			nickName := ctx.CardOrNickName(uid)
			_, err = file.GetLazyData(text.FontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if err = canvas.LoadFontFace(text.FontFile, float64(back.Bounds().Size().X)*0.04); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			canvas.DrawString(nickName+fmt.Sprintf(" ATRI币+%d", add), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.3)
			canvas.DrawString("当前ATRI币:"+strconv.FormatInt(int64(score), 10), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.4)
			canvas.DrawString("LEVEL:"+strconv.FormatInt(int64(rank), 10), float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.5)
			canvas.DrawRectangle(float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.55, float64(back.Bounds().Size().X)*0.6, float64(back.Bounds().Size().Y)*0.1)
			canvas.SetRGB255(150, 150, 150)
			canvas.Fill()
			var nextrankScore int
			if rank < 10 {
				nextrankScore = rankArray[rank+1]
			} else {
				nextrankScore = SCOREMAX
			}
			canvas.SetRGB255(0, 0, 0)
			canvas.DrawRectangle(float64(back.Bounds().Size().X)*0.1, float64(back.Bounds().Size().Y)*1.55, float64(back.Bounds().Size().X)*0.6*float64(level)/float64(nextrankScore), float64(back.Bounds().Size().Y)*0.1)
			canvas.SetRGB255(102, 102, 102)
			canvas.Fill()
			canvas.DrawString(fmt.Sprintf("%d/%d", level, nextrankScore), float64(back.Bounds().Size().X)*0.75, float64(back.Bounds().Size().Y)*1.62)

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
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})
	engine.OnPrefix("获得签到背景", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
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
	engine.OnFullMatch("查看等级排名", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			today := time.Now().Format("20060102")
			drawedFile := cachePath + today + "scoreRank.png"
			if file.IsExist(drawedFile) {
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				return
			}
			st, err := sdb.GetScoreRankByTopN(10)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(st) == 0 {
				ctx.SendChain(message.Text("ERROR: 目前还没有人签到过"))
				return
			}
			_, err = file.GetLazyData(text.FontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			b, err := os.ReadFile(text.FontFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			font, err := freetype.ParseFont(b)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			f, err := os.Create(drawedFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			var bars []chart.Value
			for _, v := range st {
				if v.Score != 0 {
					bars = append(bars, chart.Value{
						Label: ctx.CardOrNickName(v.UID),
						Value: float64(v.Score),
					})
				}
			}
			err = chart.BarChart{
				Font:  font,
				Title: "等级排名(1天只刷新1次)",
				Background: chart.Style{
					Padding: chart.Box{
						Top: 40,
					},
				},
				YAxis: chart.YAxis{
					Range: &chart.ContinuousRange{
						Min: 0,
						Max: math.Ceil(bars[0].Value/10) * 10,
					},
				},
				Height:   500,
				BarWidth: 50,
				Bars:     bars,
			}.Render(chart.PNG, f)
			_ = f.Close()
			if err != nil {
				_ = os.Remove(drawedFile)
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})
	engine.OnFullMatch("查看钱包排名", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := strconv.FormatInt(ctx.Event.GroupID, 10)
			today := time.Now().Format("20060102")
			drawedFile := cachePath + gid + today + "walletRank.png"
			if file.IsExist(drawedFile) {
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				return
			}
			// 无缓存获取群员列表
			temp := ctx.GetThisGroupMemberListNoCache().Array()
			usergroup := make([]int64, len(temp))
			for i, info := range temp {
				usergroup[i] = info.Get("user_id").Int()
			}
			// 获取钱包信息
			st, err := wallet.GetGroupWalletOf(true, usergroup...)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(st) == 0 {
				ctx.SendChain(message.Text("ERROR: 当前没人获取过ATRI币"))
				return
			} else if len(st) > 10 {
				st = st[:10]
			}
			_, err = file.GetLazyData(text.FontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			b, err := os.ReadFile(text.FontFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			font, err := freetype.ParseFont(b)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			f, err := os.Create(drawedFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			var bars []chart.Value
			for _, v := range st {
				if v.Money != 0 {
					bars = append(bars, chart.Value{
						Label: ctx.CardOrNickName(v.UID),
						Value: float64(v.Money),
					})
				}
			}
			err = chart.BarChart{
				Font:  font,
				Title: "ATRI币排名(1天只刷新1次)",
				Background: chart.Style{
					Padding: chart.Box{
						Top: 40,
					},
				},
				YAxis: chart.YAxis{
					Range: &chart.ContinuousRange{
						Min: 0,
						Max: math.Ceil(bars[0].Value/10) * 10,
					},
				},
				Height:   500,
				BarWidth: 50,
				Bars:     bars,
			}.Render(chart.PNG, f)
			_ = f.Close()
			if err != nil {
				_ = os.Remove(drawedFile)
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})
}

func getHourWord(t time.Time) string {
	h := t.Hour()
	switch {
	case 6 <= h && h < 12:
		return "早上好"
	case 12 <= h && h < 14:
		return "中午好"
	case 14 <= h && h < 19:
		return "下午好"
	case 19 <= h && h < 24:
		return "晚上好"
	case 0 <= h && h < 6:
		return "凌晨好"
	default:
		return ""
	}
}

func getrank(count int) int {
	for k, v := range rankArray {
		if count == v {
			return k
		} else if count < v {
			return k - 1
		}
	}
	return -1
}

func initPic(picFile string) error {
	if file.IsExist(picFile) {
		return nil
	}
	return file.DownloadTo(backgroundURL, picFile, true)
}
