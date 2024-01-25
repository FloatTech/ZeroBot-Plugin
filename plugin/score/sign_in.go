// Package score 签到
package score

import (
	"encoding/base64"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/golang/freetype"
	"github.com/wcharczuk/go-chart/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	backgroundURL = "https://iw233.cn/api.php?sort=pc"
	referer       = "https://weibo.com/"
	signinMax     = 1
	// SCOREMAX 分数上限定为1200
	SCOREMAX = 1200
)

var (
	rankArray = [...]int{0, 10, 20, 50, 100, 200, 350, 550, 750, 1000, 1200}
	engine    = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "签到",
		Help:              "- 签到\n- 获得签到背景[@xxx] | 获得签到背景\n- 设置签到预设(0~3)\n- 查看等级排名\n注:为跨群排名\n- 查看我的钱包\n- 查看钱包排名\n注:为本群排行，若群人数太多不建议使用该功能!!!",
		PrivateDataFolder: "score",
	})
	styles = []scoredrawer{
		drawScore15,
		drawScore16,
		drawScore17,
		drawScore17b2,
	}
)

func init() {
	cachePath := engine.DataFolder() + "cache/"
	go func() {
		sdb = initialize(engine.DataFolder() + "score.db")
		ok := file.IsExist(cachePath)
		if !ok {
			err := os.MkdirAll(cachePath, 0777)
			if err != nil {
				panic(err)
			}
			return
		}
		files, err := os.ReadDir(cachePath)
		if err == nil {
			for _, f := range files {
				if !strings.Contains(f.Name(), time.Now().Format("20060102")) {
					_ = os.Remove(cachePath + f.Name())
				}
			}
		}
	}()
	engine.OnRegex(`^签到\s?(\d*)$`).Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 选择key
		key := ctx.State["regex_matched"].([]string)[1]
		gid := ctx.Event.GroupID
		if gid < 0 {
			// 个人用户设为负数
			gid = -ctx.Event.UserID
		}
		k := uint8(0)
		if key == "" {
			k = uint8(ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).GetData(gid))
		} else {
			kn, err := strconv.Atoi(key)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			k = uint8(kn)
		}
		if int(k) >= len(styles) {
			ctx.SendChain(message.Text("ERROR: 未找到签到设定: ", key))
			return
		}
		uid := ctx.Event.UserID
		today := time.Now().Format("20060102")
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
				trySendImage(drawedFile, ctx)
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
		go func() {
			err = wallet.InsertWalletOf(uid, add)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}()
		alldata := &scdata{
			drawedfile: drawedFile,
			picfile:    picFile,
			uid:        uid,
			nickname:   ctx.CardOrNickName(uid),
			inc:        add,
			score:      wallet.GetWalletOf(uid),
			level:      level,
			rank:       rank,
		}
		drawimage, err := styles[k](alldata)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		// done.
		f, err := os.Create(drawedFile)
		if err != nil {
			data, err := imgfactory.ToBytes(drawimage)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
			return
		}
		_, err = imgfactory.WriteTo(drawimage, f)
		defer f.Close()
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		trySendImage(drawedFile, ctx)
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
			trySendImage(picFile, ctx)
		})
	engine.OnFullMatch("查看等级排名", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			today := time.Now().Format("20060102")
			drawedFile := cachePath + today + "scoreRank.png"
			if file.IsExist(drawedFile) {
				trySendImage(drawedFile, ctx)
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
			trySendImage(drawedFile, ctx)
		})
	engine.OnRegex(`^设置签到预设\s*(\d+)$`, zero.SuperUserPermission).Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		key := ctx.State["regex_matched"].([]string)[1]
		kn, err := strconv.Atoi(key)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		k := uint8(kn)
		if int(k) >= len(styles) {
			ctx.SendChain(message.Text("ERROR: 未找到签到设定: ", key))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		err = ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).SetData(gid, int64(k))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("设置成功"))
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

func initPic(picFile string, uid int64) (avatar []byte, err error) {
	defer process.SleepAbout1sTo2s()
	avatar, err = web.GetData("http://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(uid, 10) + "&s=640")
	if err != nil {
		return
	}
	if file.IsExist(picFile) {
		return
	}
	url, err := bilibili.GetRealURL(backgroundURL)
	if err != nil {
		return
	}
	data, err := web.RequestDataWith(web.NewDefaultClient(), url, "", referer, "", nil)
	if err != nil {
		return
	}
	return avatar, os.WriteFile(picFile, data, 0644)
}

// 使用"file:"发送图片失败后，改用base64发送
func trySendImage(filePath string, ctx *zero.Ctx) {
	filePath = file.BOTPATH + "/" + filePath
	if id := ctx.SendChain(message.Image("file:///" + filePath)); id.ID() != 0 {
		return
	}
	imgFile, err := os.Open(filePath)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 无法打开文件", err))
		return
	}
	defer imgFile.Close()
	// 使用 base64.NewEncoder 将文件内容编码为 base64 字符串
	var encodedFileData strings.Builder
	encodedFileData.WriteString("base64://")
	encoder := base64.NewEncoder(base64.StdEncoding, &encodedFileData)
	_, err = io.Copy(encoder, imgFile)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: 无法编码文件内容", err))
		return
	}
	encoder.Close()
	drawedFileBase64 := encodedFileData.String()
	if id := ctx.SendChain(message.Image(drawedFileBase64)); id.ID() == 0 {
		ctx.SendChain(message.Text("ERROR: 无法读取图片文件", err))
		return
	}
}
