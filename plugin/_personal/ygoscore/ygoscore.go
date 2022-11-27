package ygoscore

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 数据库
	sql "github.com/FloatTech/sqlite"
	// 图片输出
	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/zbputils/img"
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

var (
	scoredata = &score{
		db: &sql.Sqlite{},
	}
	levelArray = [...]int{0, 10, 20, 50, 100, 200, 350, 550, 750, 1000, 1200}
	levelrank  = [...]string{"新手", "青铜", "白银", "黄金", "白金Ⅲ", "白金Ⅱ", "白金Ⅰ", "传奇Ⅲ", "传奇Ⅱ", "传奇Ⅰ", "决斗王"}
)

const SCOREMAX = 1200

func init() {
	engine := control.Register("ygoscore", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "签到系统",
		PrivateDataFolder: "ygoscore",
		Help: "-注册决斗者 xxxx\n" +
			"-注销决斗者 @群友\n" +
			"-签到\n",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		scoredata.db.DBPath = engine.DataFolder() + "score.db"
		err := scoredata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return false
		}
		return true
	})

	engine.OnRegex(`^注册决斗者\s*([^\s]+(\s+[^\s]+)*)`, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		username := ctx.State["regex_matched"].([]string)[1]
		if strings.Contains(username, "[CQ:face,id=") {
			ctx.SendChain(message.Text("用户名不支持表情包"))
			return
		}
		lenmane := []rune(username)
		if len(lenmane) > 10 {
			ctx.SendChain(message.Text("决斗者昵称不得长于10个字符"))
			return
		}
		err := scoredata.register(ctx.Event.UserID, username)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return
		}
		ctx.SendChain(message.Text("注册成功"))
	})
	engine.OnRegex(`^注销决斗者(\s*\[CQ:at,qq=)?(\d+).*?`, zero.AdminPermission, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		adduser, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
		err := scoredata.deleteuser(adduser)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return
		}
		ctx.SendChain(message.Text("注销成功"))
	})

	engine.OnFullMatchGroup([]string{"签到", "打卡"}, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		userinfo, err := scoredata.checkuser(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return
		}
		if userinfo.UserName == "" {
			ctx.SendChain(message.Text("决斗者未注册!\n请输入“注册决斗者 xxx”进行登记(xxx为决斗者昵称)。"))
			return
		}
		lasttime := time.Unix(userinfo.UpdatedAt, 0)
		// 判断是否已经签到过了
		if time.Now().Format("2006/01/02") == lasttime.Format("2006/01/02") {
			score := wallet.GetWalletOf(uid)
			// 生成ATRI币图片
			data, cl, err := drawimage(&userinfo, score, 0)
			if err != nil {
				ctx.SendChain(message.Text("[ygoscore]error:", err))
				return
			}
			ctx.SendChain(message.Text("今天已经签到过了"))
			ctx.SendChain(message.ImageBytes(data))
			cl()
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
		if userinfo.Level < SCOREMAX {
			userinfo.Level += 1
		}
		if err := scoredata.db.Insert("score", &userinfo); err != nil {
			ctx.SendChain(message.Text("签到失败,请再次尝试或者联系bot管理员"))
			return
		}
		if err := wallet.InsertWalletOf(uid, add); err != nil {
			ctx.SendChain(message.Text("签到失败,请再次尝试或者联系bot管理员"))
			return
		}
		score := wallet.GetWalletOf(uid)
		// 生成签到图片
		data, cl, err := drawimage(&userinfo, score, add)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return
		}
		ctx.SendChain(message.ImageBytes(data))
		cl()
	})
	engine.OnRegex(`^\/修改(\s*(\[CQ:at,qq=)?(\d+).*)?信息\s*(.*)`, zero.AdminPermission, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		changeuser := ctx.State["regex_matched"].([]string)[3]
		data := ctx.State["regex_matched"].([]string)[4]
		uid := ctx.Event.UserID
		changeData := make(map[string]string, 10)
		infoList := strings.Split(data, " ")
		if len(infoList) == 1 {
			ctx.SendChain(message.Text("[ygoscore]error:", "请输入正确的参数"))
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
		userinfo, err := scoredata.checkuser(uid)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return
		}
		userinfo.Uid = uid
		for dataName, value := range changeData {
			switch dataName {
			case "昵称":
				userinfo.UserName = value
			case "签到时间":
				now, err := time.Parse("2006/01/02", value)
				if err != nil {
					ctx.SendChain(message.Text("[ygoscore]error:", err))
					return
				}
				userinfo.UpdatedAt = now.Unix()
			case "签到次数":
				times, err := strconv.Atoi(value)
				if err != nil {
					ctx.SendChain(message.Text("[ygoscore]error:", err))
					return
				}
				userinfo.Continuous = times
			case "等级":
				level, err := strconv.Atoi(value)
				if err != nil {
					ctx.SendChain(message.Text("[ygoscore]error:", err))
					return
				}
				userinfo.Level = level
			}
		}
		err = scoredata.db.Insert("score", &userinfo)
		if err != nil {
			ctx.SendChain(message.Text("[ygoscore]error:", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	})
}

// 注册用户
func (sdb *score) register(uid int64, username string) (err error) {
	sdb.Lock()
	defer sdb.Unlock()
	if err = sdb.db.Create("score", &userdata{}); err != nil {
		return
	}
	var userinfo userdata
	_ = sdb.db.Find("score", &userinfo, "where uid = "+strconv.FormatInt(uid, 10))
	if userinfo.UserName != "" {
		return errors.New("该账号已注册过了!当前昵称为:" + userinfo.UserName)
	}
	userinfo = userdata{
		Uid:      uid,
		UserName: username,
	}
	return sdb.db.Insert("score", &userinfo)
}
func (sdb *score) checkuser(uid int64) (userinfo userdata, err error) {
	sdb.Lock()
	defer sdb.Unlock()
	if err = sdb.db.Create("score", &userdata{}); err != nil {
		return
	}
	_ = sdb.db.Find("score", &userinfo, "where uid = "+strconv.FormatInt(uid, 10))
	return
}

// 绘制图片
func drawimage(userinfo *userdata, score, add int) (data []byte, cl func(), err error) {
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
	back = img.Size(back, backX, backY).Im
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
	level := getLevel(userinfo.Level)
	if add == 0 {
		canvas.DrawString(fmt.Sprintf("决斗者等级:LV%d", level), 550, 240-h)
		canvas.DrawString("等级阶段: "+levelrank[level], 1030, 240-h)
		canvas.DrawString(fmt.Sprintf("已连续签到 %d 天", userinfo.Continuous), 550, 320-h)
	} else {
		if userinfo.Level < SCOREMAX {
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
	var nextLevelScore int
	if level < 10 {
		nextLevelScore = levelArray[level+1]
	} else {
		nextLevelScore = SCOREMAX
	}
	canvas.SetRGB255(0, 0, 0)
	canvas.DrawRectangle(550, 350-h, 900*float64(userinfo.Level)/float64(nextLevelScore), 80)
	canvas.SetRGB255(102, 102, 102)
	canvas.Fill()
	canvas.DrawString(fmt.Sprintf("%d/%d", userinfo.Level, nextLevelScore), 1250, 320-h)
	// 生成图片
	data, cl = writer.ToBytes(canvas.Image())
	return
}

func (sdb *score) deleteuser(uid int64) (err error) {
	sdb.Lock()
	defer sdb.Unlock()
	uidStr := strconv.FormatInt(uid, 10)
	err = sdb.db.Find("score", &userdata{}, "where Uid = "+uidStr)
	if err == nil {
		err = sdb.db.Del("score", "where Uid = "+uidStr)
	}
	return
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
