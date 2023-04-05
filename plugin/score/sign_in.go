// Package score 签到
package score

import (
	"image"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/AnimeAPI/wallet"
	fcext "github.com/FloatTech/floatbox/ctxext"
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
	SCOREMAX       = 1200
	defKeyID int64 = -6
)

var (
	rankArray = [...]int{0, 10, 20, 50, 100, 200, 350, 550, 750, 1000, 1200}
	engine    = control.Register("score", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "签到",
		Help:              "- 签到\n- 获得签到背景[@xxx] | 获得签到背景\n- 设置[默认]签到预设(1~9)\n- 查看等级排名\n注:为跨群排名\n- 查看我的钱包\n- 查看钱包排名\n注:为本群排行，若群人数太多不建议使用该功能!!!",
		PrivateDataFolder: "score",
	})
	initDef = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		var defkey string
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		_ = m.Manager.Response(defKeyID)
		_ = m.Manager.GetExtra(defKeyID, &defkey)
		if defkey == "" {
			_ = m.Manager.SetExtra(defKeyID, "1")
			return true
		}
		return true
	})
	stylemap = map[string]func(a *scdata) (image.Image, error){
		"1": drawScore15,
		"2": drawScore16,
		"3": drawScore17,
		"4": drawScore17b2,
	}
)

func init() {
	cachePath := engine.DataFolder() + "cache/"
	go func() {
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
		sdb = initialize(engine.DataFolder() + "score.db")
	}()
	engine.OnRegex(`^签到\s?(\d*)$`, initDef).Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 选择key
		key := ctx.State["regex_matched"].([]string)[1]
		gid := ctx.Event.GroupID
		if gid < 0 {
			// 个人用户设为负数
			gid = -ctx.Event.UserID
		}
		if key == "" {
			m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
			_ = m.Manager.GetExtra(gid, &key)
			if key == "" {
				_ = m.Manager.GetExtra(defKeyID, &key)
			}
		}
		drawfunc, ok := stylemap[key]
		if !ok {
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
		drawimage, err := drawfunc(alldata)
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
			if id := ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + picFile)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 消息发送失败, 账号可能被风控"))
			}
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
	engine.OnRegex(`^设置(默认)?签到预设\s?(\d*)$`, zero.SuperUserPermission).Limit(ctxext.LimitByUser).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if key := ctx.State["regex_matched"].([]string)[2]; key == "" {
			ctx.SendChain(message.Text("设置失败, 数据为空"))
		} else {
			s := ctx.State["regex_matched"].([]string)[1]
			_, ok := stylemap[key]
			if !ok {
				ctx.SendChain(message.Text("ERROR: 未找到签到设定: ", key)) // 避免签到配置错误
				return
			}
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			if s != "" {
				gid = defKeyID
			}
			err := ctx.State["manager"].(*ctrl.Control[*zero.Ctx]).Manager.SetExtra(gid, key)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("设置成功, 当前", s, "预设为:", key))
		}
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
