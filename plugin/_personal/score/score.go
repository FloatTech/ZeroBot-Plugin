// Package score 签到系统
package score

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/process"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	"github.com/disintegration/imaging"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 数据库

	"github.com/FloatTech/AnimeAPI/wallet"
	names "github.com/FloatTech/ZeroBot-Plugin/plugin/dataSystem"
	sql "github.com/FloatTech/sqlite"

	// 图片输出
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/zbputils/img/text"
)

type score struct {
	db *sql.Sqlite
	sync.RWMutex
}

// 用户数据信息
type userdata struct {
	Uid        int64  // `Userid`
	UserName   string // `User`
	UpdatedAt  int64  // `签到时间`
	Continuous int    // `连续签到次数`
	Level      int    // `决斗者等级`
}

const (
	backgroundURL = "https://iw233.cn/api.php?sort=pc"
	referer       = "https://weibo.com/"
	scoreMax      = 1200
)

var (
	scoredata = &score{
		db: &sql.Sqlite{},
	}
	/************************************10*****20******30*****40*****50******60*****70*****80******90**************/
	/*************************2******10*****20******40*****70*****110******160******220***290*****370*******460***************/
	levelrank = [...]string{"新手", "青铜", "白银", "黄金", "白金Ⅲ", "白金Ⅱ", "白金Ⅰ", "传奇Ⅲ", "传奇Ⅱ", "传奇Ⅰ", "决斗王"}
	engine    = control.Register("score", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "签到",
		PrivateDataFolder: "ygoscore",
		Help:              "- 签到\n",
	})
	cachePath = engine.DataFolder() + "cache/"
)

func init() {
	go func() {
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
	}()
	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		scoredata.db.DBPath = engine.DataFolder() + "score.db"
		err := scoredata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		err = scoredata.db.Create("score", &userdata{})
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})

	engine.OnFullMatchGroup([]string{"签到", "打卡"}, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		userinfo := scoredata.getData(uid)
		userinfo.Uid = uid
		newName := names.GetNameOf(uid) //更新昵称
		if newName != "" && newName != userinfo.UserName {
			userinfo.UserName = newName
		} else if userinfo.UserName == "" {
			userinfo.UserName = ctx.CardOrNickName(uid)
		}
		lasttime := time.Unix(userinfo.UpdatedAt, 0)
		// 判断是否已经签到过了
		if time.Now().Format("2006/01/02") == lasttime.Format("2006/01/02") {
			score := wallet.GetWalletOf(uid)
			data, err := drawimagePro(&userinfo, score, 0)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				data, err = drawimage(&userinfo, score, 0)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
			}
			ctx.SendChain(message.Text("今天已经签到过了"))
			ctx.SendChain(message.ImageBytes(data))
			return
		}
		// 更新数据
		add := 1
		subtime := time.Since(lasttime).Hours()
		if subtime > 48 {
			userinfo.Continuous = 1
		} else {
			userinfo.Continuous += 1
			add = int(math.Min(5, float64(userinfo.Continuous)))
		}
		userinfo.UpdatedAt = time.Now().Unix()
		if userinfo.Level < scoreMax {
			userinfo.Level += add
		}
		if err := scoredata.setData(userinfo); err != nil {
			ctx.SendChain(message.Text("[ERROR]:签到记录失败。", err))
			return
		}
		level, _ := getLevel(userinfo.Level)
		if err := wallet.InsertWalletOf(uid, add+level*5); err != nil {
			ctx.SendChain(message.Text("[ERROR]:货币记录失败。", err))
			return
		}
		score := wallet.GetWalletOf(uid)
		// 生成签到图片
		data, err := drawimagePro(&userinfo, score, add)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			data, err = drawimage(&userinfo, score, add)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
		}
		ctx.SendChain(message.ImageBytes(data))
	})
	engine.OnRegex(`^\/修改(\s*(\[CQ:at,qq=)?(\d+).*)?信息\s*(.*)`, zero.AdminPermission, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		changeuser := ctx.State["regex_matched"].([]string)[3]
		data := ctx.State["regex_matched"].([]string)[4]
		uid := ctx.Event.UserID
		changeData := make(map[string]string, 10)
		infoList := strings.Split(data, " ")
		if len(infoList) == 1 {
			ctx.SendChain(message.Text("[ERROR]:", "请输入正确的参数"))
			return
		}
		for _, manager := range infoList {
			infoData := strings.Split(manager, ":")
			if len(infoData) > 1 {
				changeData[infoData[0]] = infoData[1]
			}
		}
		if changeuser != "" {
			uid, _ = strconv.ParseInt(changeuser, 10, 64)
		}
		userinfo := scoredata.getData(uid)
		userinfo.Uid = uid
		for dataName, value := range changeData {
			switch dataName {
			case "签到时间":
				now, err := time.Parse("2006/01/02", value)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				userinfo.UpdatedAt = now.Unix()
			case "签到次数":
				times, err := strconv.Atoi(value)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				userinfo.Continuous = times
			case "等级":
				level, err := strconv.Atoi(value)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:", err))
					return
				}
				userinfo.Level = level
			}
		}
		err := scoredata.db.Insert("score", &userinfo)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
}

// 获取签到数据
func (sdb *score) getData(uid int64) (userinfo userdata) {
	sdb.Lock()
	defer sdb.Unlock()
	_ = sdb.db.Find("score", &userinfo, "where uid = "+strconv.FormatInt(uid, 10))
	return
}

// 保存签到数据
func (sdb *score) setData(userinfo userdata) error {
	sdb.Lock()
	defer sdb.Unlock()
	return sdb.db.Insert("score", &userinfo)

}

// 下载图片
func initPic(picFile string, uid int64) (avatar []byte, err error) {
	avatar, err = web.GetData("http://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(uid, 10) + "&s=640")
	if err != nil {
		return nil, err
	}
	if file.IsExist(picFile) {
		return avatar, nil
	}
	defer process.SleepAbout1sTo2s()
	data, err := web.GetData("https://img.moehu.org/pic.php?id=yu-gi-oh")
	if err != nil {
		return nil, err
	}
	return avatar, os.WriteFile(picFile, data, 0644)
}
func drawimagePro(userinfo *userdata, score, add int) (data []byte, err error) {
	picFile := cachePath + time.Now().Format("20060102_") + strconv.FormatInt(userinfo.Uid, 10) + ".png"
	getAvatar, err := initPic(picFile, userinfo.Uid)
	if err != nil {
		return
	}
	back, err := gg.LoadImage(picFile)
	if err != nil {
		return
	}
	imgDX := back.Bounds().Dx()
	imgDY := back.Bounds().Dy()
	backDX := 1500

	imgDW := backDX - 100
	scale := float64(imgDW) / float64(imgDX)
	imgDH := int(float64(imgDY) * scale)
	back = imgfactory.Size(back, imgDW, imgDH).Image()

	backDY := imgDH + 500
	canvas := gg.NewContext(backDX, backDY)
	// 放置毛玻璃背景
	backBlurW := float64(imgDW) * (float64(backDY) / float64(imgDH))
	canvas.DrawImageAnchored(imaging.Blur(imgfactory.Size(back, int(backBlurW), backDY).Image(), 8), backDX/2, backDY/2, 0.5, 0.5)
	canvas.DrawRectangle(1, 1, float64(backDX), float64(backDY))
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(255, 255, 255, 100)
	canvas.StrokePreserve()
	canvas.SetRGBA255(255, 255, 255, 140)
	canvas.Fill()
	// 信息框
	canvas.DrawRoundedRectangle(20, 20, 1500-20-20, 450-20, (450-20)/5)
	canvas.SetLineWidth(6)
	canvas.SetDash(20.0, 10.0, 0)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Stroke()
	// 放置头像
	avatar, _, err := image.Decode(bytes.NewReader(getAvatar))
	if err != nil {
		return
	}
	avatarf := imgfactory.Size(avatar, 300, 300)
	canvas.DrawCircle(50+float64(avatarf.W())/2, 50+float64(avatarf.H())/2, float64(avatarf.W())/2+2)
	canvas.SetLineWidth(3)
	canvas.SetDash()
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Stroke()
	canvas.DrawImage(avatarf.Circle(0).Image(), 50, 50)
	// 放置昵称
	canvas.SetRGB(0, 0, 0)
	fontSize := 150.0
	_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
		return
	}
	nameW, nameH := canvas.MeasureString(userinfo.UserName)
	if nameW > float64(backDX)/3 { // 如果文字超过长度了，比列缩小字体
		scale := (float64(backDX) / 3) / nameW
		fontSize = fontSize * scale
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
		return
	}
	canvas.DrawStringAnchored(userinfo.UserName, float64(backDX)/2, 50+nameH/2, 0.5, 0.5)

	// level
	if err = canvas.LoadFontFace(text.BoldFontFile, 72); err != nil {
		return
	}
	level, nextLevelScore := getLevel(userinfo.Level)
	if level == -1 {
		err = errors.New("计算等级出现了问题")
		return
	}
	levelX := float64(backDX) * 4 / 5
	canvas.DrawRoundedRectangle(levelX, 50, 200, 200, 200/5)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 100)
	canvas.StrokePreserve()
	canvas.SetRGBA255(255, 255, 255, 100)
	canvas.Fill()
	canvas.DrawRoundedRectangle(levelX, 50, 200, 100, 200/5)
	canvas.SetLineWidth(3)
	canvas.SetRGBA255(0, 0, 0, 100)
	canvas.StrokePreserve()
	canvas.SetRGBA255(255, 255, 255, 100)
	canvas.Fill()
	canvas.SetRGBA255(0, 0, 0, 255)
	canvas.DrawStringAnchored(levelrank[level], levelX+100, 50+50, 0.5, 0.5)
	canvas.DrawStringAnchored(fmt.Sprintf("LV%d", level), levelX+100, 50+100+50, 0.5, 0.5)

	if add == 0 {
		canvas.DrawString(fmt.Sprintf("已连签 %d 天    总资产: %d", userinfo.Continuous, score), 350, 350)
	} else {
		canvas.DrawString(fmt.Sprintf("连签 %d 天 总资产( +%d ) : %d", userinfo.Continuous, add+level*5, score), 350, 350)
	}
	// 绘制等级进度条
	if err = canvas.LoadFontFace(text.BoldFontFile, 50); err != nil {
		return
	}
	_, textH := canvas.MeasureString("/")
	switch {
	case userinfo.Level < scoreMax && add == 0:
		canvas.DrawStringAnchored(fmt.Sprintf("%d/%d", userinfo.Level, nextLevelScore), float64(backDX)/2, 455-textH, 0.5, 0.5)
	case userinfo.Level < scoreMax:
		canvas.DrawStringAnchored(fmt.Sprintf("(%d+%d)/%d", userinfo.Level-add, add, nextLevelScore), float64(backDX)/2, 455-textH, 0.5, 0.5)
	default:
		canvas.DrawStringAnchored("Max/Max", float64(backDX)/2, 455-textH, 0.5, 0.5)

	}
	// 创建彩虹条
	grad := gg.NewLinearGradient(0, 500, 1500, 300)
	grad.AddColorStop(0, color.RGBA{G: 255, A: 255})
	grad.AddColorStop(0.25, color.RGBA{B: 255, A: 255})
	grad.AddColorStop(0.5, color.RGBA{R: 255, A: 255})
	grad.AddColorStop(0.75, color.RGBA{B: 255, A: 255})
	grad.AddColorStop(1, color.RGBA{G: 255, A: 255})
	canvas.SetStrokeStyle(grad)
	canvas.SetLineWidth(7)
	// 设置长度
	gradMax := 1300.0
	LevelLength := gradMax * (float64(userinfo.Level) / float64(nextLevelScore))
	canvas.MoveTo((float64(backDX)-LevelLength)/2, 450)
	canvas.LineTo((float64(backDX)+LevelLength)/2, 450)
	canvas.ClosePath()
	canvas.Stroke()
	// 放置图片
	canvas.DrawImageAnchored(back, backDX/2, imgDH/2+475, 0.5, 0.5)
	// 生成图片
	return imgfactory.ToBytes(canvas.Image())
}

// 绘制图片
func drawimage(userinfo *userdata, score, add int) (data []byte, err error) {
	/***********获取头像***********/
	backX := 500
	backY := 500
	uid := strconv.FormatInt(userinfo.Uid, 10)
	data, err = web.GetData("http://q4.qlogo.cn/g?b=qq&nk=" + uid + "&s=640&cache=0")
	if err != nil {
		return
	}
	back, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}
	back = imgfactory.Size(back, backX, backY).Image()
	/***********设置图片的大小和底色***********/
	canvas := gg.NewContext(1500, 500)
	canvas.SetRGB(1, 1, 1)
	canvas.Clear()

	/***********放置头像***********/
	canvas.DrawImage(back, 0, 0)

	/***********写入用户信息***********/
	fontSize := 50.0
	_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
		return
	}
	canvas.SetRGB(0, 0, 0)
	length, h := canvas.MeasureString(uid)
	// 用户名字和QQ号
	n, _ := canvas.MeasureString(userinfo.UserName)
	canvas.DrawString(userinfo.UserName, 550, 130-h)
	canvas.DrawRoundedRectangle(600+n-length*0.1, 130-h*2.5, length*1.2, h*2, fontSize*0.2)
	canvas.SetRGB255(221, 221, 221)
	canvas.Fill()
	canvas.SetRGB(0, 0, 0)
	canvas.DrawString(uid, 600+n, 130-h)
	// 填如签到数据
	level, nextLevelScore := getLevel(userinfo.Level)
	if level == -1 {
		err = errors.New("计算等级出现了问题")
		return
	}
	if add == 0 {
		canvas.DrawString(fmt.Sprintf("决斗者等级:LV%d", level), 550, 240-h)
		canvas.DrawString("等级阶段: "+levelrank[level], 1030, 240-h)
		canvas.DrawString(fmt.Sprintf("已连续签到 %d 天", userinfo.Continuous), 550, 320-h)
	} else {
		if userinfo.Level < scoreMax {
			canvas.DrawString(fmt.Sprintf("经验 +1,ATRI币 +%d", add), 550, 240-h)
		} else {
			canvas.DrawString(fmt.Sprintf("签到ATRI币 + %d", add), 550, 240-h)
		}
		canvas.DrawString(fmt.Sprintf("决斗者等级:LV%d", level), 1000, 240-h)
		canvas.DrawString(fmt.Sprintf("已连续签到 %d 天", userinfo.Continuous), 550, 320-h)
	}
	// ATRI币详情
	canvas.DrawString(fmt.Sprintf("当前总ATRI币:%d", score), 550, 500-h)
	// 更新时间
	canvas.DrawString("更新日期:"+time.Unix(userinfo.UpdatedAt, 0).Format("01/02"), 1050, 500-h)
	// 绘制等级进度条
	canvas.DrawRectangle(550, 350-h, 900, 80)
	canvas.SetRGB255(150, 150, 150)
	canvas.Fill()
	canvas.SetRGB255(0, 0, 0)
	canvas.DrawRectangle(550, 350-h, 900*float64(userinfo.Level)/float64(nextLevelScore), 80)
	canvas.SetRGB255(102, 102, 102)
	canvas.Fill()
	canvas.DrawString(fmt.Sprintf("%d/%d", userinfo.Level, nextLevelScore), 1250, 320-h)
	// 生成图片
	return imgfactory.ToBytes(canvas.Image())
}

func getLevel(count int) (int, int) {
	switch {
	case count < 2:
		return 0, 2
	case count > scoreMax:
		return len(levelrank) - 1, scoreMax
	default:
		for k, i := 1, 10; i <= scoreMax; i += (k * 10) * scoreMax / 460 {
			if count < i {
				return k, i
			}
			k++
		}
	}
	return -1, -1
}
