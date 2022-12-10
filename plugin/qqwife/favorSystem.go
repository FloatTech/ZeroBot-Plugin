package qqwife

import (
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/math"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 画图
	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/zbputils/img/text"

	// 货币系统
	"github.com/FloatTech/AnimeAPI/wallet"
)

// 好感度系统
type favorability struct {
	Userinfo string // 记录用户
	Favor    int    // 好感度
}

func init() {
	// 好感度系统
	engine.OnRegex(`^查好感度\s*(\[CQ:at,qq=)?(\d+)`, zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			fiancee, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			uid := ctx.Event.UserID
			favor, err := 民政局.查好感度(uid, fiancee)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			// 输出结果
			ctx.SendChain(
				message.At(uid),
				message.Text("\n当前你们好感度为", favor),
			)
		})
	// 礼物系统
	engine.OnRegex(`^买礼物给\s?(\[CQ:at,qq=(\d+)\]|(\d+))`, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			fiancee := ctx.State["regex_matched"].([]string)
			gay, _ := strconv.ParseInt(fiancee[2]+fiancee[3], 10, 64)
			if gay == uid {
				ctx.Send(message.ReplyWithMessage(ctx.Event.MessageID, message.At(uid), message.Text("你想给自己买什么礼物呢?")))
				return
			}
			// 获取CD
			groupInfo, err := 民政局.查看设置(gid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			ok, err := 民政局.判断CD(gid, uid, "买礼物", groupInfo.CDtime)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:", err))
				return
			}
			if !ok {
				ctx.SendChain(message.Text("舔狗，今天你已经送过礼物了。"))
				return
			}
			// 获取好感度
			favor, err := 民政局.查好感度(uid, gay)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:好感度库发生问题力\n", err))
				return
			}
			// 对接小熊饼干
			walletinfo := wallet.GetWalletOf(uid)
			if walletinfo < 1 {
				ctx.SendChain(message.Text("你钱包没钱啦！"))
				return
			}
			moneyToFavor := rand.Intn(math.Min(walletinfo, 100))
			// 计算钱对应的好感值
			newFavor := 1
			if favor > 50 {
				newFavor = moneyToFavor % 10 // 礼物厌倦
			} else {
				newFavor += rand.Intn(moneyToFavor)
			}
			// 随机对方心情
			mood := rand.Intn(2)
			if mood == 0 {
				newFavor = -newFavor
			}
			// 记录结果
			err = wallet.InsertWalletOf(uid, -moneyToFavor)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:钱包坏掉力:\n", err))
				return
			}
			lastfavor, err := 民政局.更新好感度(uid, gay, newFavor)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:好感度数据库发生问题力\n", err))
				return
			}
			// 写入CD
			err = 民政局.记录CD(gid, uid, "买礼物")
			if err != nil {
				ctx.SendChain(message.At(uid), message.Text("[ERROR]:你的技能CD记录失败\n", err))
			}
			// 输出结果
			if mood == 0 {
				ctx.SendChain(message.Text("你花了", moneyToFavor, "ATRI币买了一件女装送给了ta,ta很不喜欢,你们的好感度降低至", lastfavor))
			} else {
				ctx.SendChain(message.Text("你花了", moneyToFavor, "ATRI币买了一件女装送给了ta,ta很喜欢,你们的好感度升至", lastfavor))
			}
		})
	engine.OnFullMatch("好感度列表", zero.OnlyGroup, getdb).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			uid := ctx.Event.UserID
			fianceeInfo, err := 民政局.getGroupFavorability(uid)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
				return
			}
			/***********设置图片的大小和底色***********/
			number := len(fianceeInfo)
			if number > 10 {
				number = 10
			}
			fontSize := 50.0
			canvas := gg.NewContext(1150, int(170+(50+70)*float64(number)))
			canvas.SetRGB(1, 1, 1) // 白色
			canvas.Clear()
			/***********下载字体***********/
			_, err = file.GetLazyData(text.BoldFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
			}
			/***********设置字体颜色为黑色***********/
			canvas.SetRGB(0, 0, 0)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize*2); err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
				return
			}
			sl, h := canvas.MeasureString("你的好感度排行列表")
			/***********绘制标题***********/
			canvas.DrawString("你的好感度排行列表", (1100-sl)/2, 100) // 放置在中间位置
			canvas.DrawString("————————————————————", 0, 160)
			/***********设置字体大小,并获取字体高度用来定位***********/
			if err = canvas.LoadFontFace(text.BoldFontFile, fontSize); err != nil {
				ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
				return
			}
			i := 0
			for _, info := range fianceeInfo {
				if i > 9 {
					break
				}
				if info.Userinfo == "" {
					continue
				}
				fianceID, err := strconv.ParseInt(info.Userinfo, 10, 64)
				if err != nil {
					ctx.SendChain(message.Text("[ERROR]:ERROR: ", err))
					return
				}
				if fianceID == 0 {
					continue
				}
				userName := ctx.CardOrNickName(fianceID)
				canvas.SetRGB255(0, 0, 0)
				canvas.DrawString(userName+"("+info.Userinfo+")", 10, float64(180+(50+70)*i))
				canvas.DrawString(strconv.Itoa(info.Favor), 1020, float64(180+60+(50+70)*i))
				canvas.DrawRectangle(10, float64(180+60+(50+70)*i)-h/2, 1000, 50)
				canvas.SetRGB255(150, 150, 150)
				canvas.Fill()
				canvas.SetRGB255(0, 0, 0)
				canvas.DrawRectangle(10, float64(180+60+(50+70)*i)-h/2, float64(info.Favor)*10, 50)
				canvas.SetRGB255(231, 27, 100)
				canvas.Fill()
				i++
			}
			data, cl := writer.ToBytes(canvas.Image())
			ctx.SendChain(message.ImageBytes(data))
			cl()
		})
}

func (sql *婚姻登记) 查好感度(uid, target int64) (int, error) {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create("favorability", &favorability{})
	if err != nil {
		return 0, err
	}
	info := favorability{}
	uidstr := strconv.FormatInt(uid, 10)
	targstr := strconv.FormatInt(target, 10)
	_ = sql.db.Find("favorability", &info, "where Userinfo glob '*"+uidstr+"+"+targstr+"*'")
	return info.Favor, nil
}

// 获取好感度数据组
type favorList []favorability

func (s favorList) Len() int {
	return len(s)
}
func (s favorList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s favorList) Less(i, j int) bool {
	return s[i].Favor > s[j].Favor
}
func (sql *婚姻登记) getGroupFavorability(uid int64) (list favorList, err error) {
	uidStr := strconv.FormatInt(uid, 10)
	sql.RLock()
	defer sql.RUnlock()
	info := favorability{}
	err = sql.db.FindFor("favorability", &info, "where Userinfo glob '*"+uidStr+"*'", func() error {
		var target string
		userList := strings.Split(info.Userinfo, "+")
		switch {
		case len(userList) == 0:
			return errors.New("好感度系统数据存在错误")
		case userList[0] == uidStr:
			target = userList[1]
		default:
			target = userList[0]
		}
		list = append(list, favorability{
			Userinfo: target,
			Favor:    info.Favor,
		})
		return nil
	})
	sort.Sort(list)
	return
}

// 设置好感度 正增负减
func (sql *婚姻登记) 更新好感度(uid, target int64, score int) (favor int, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create("favorability", &favorability{})
	if err != nil {
		return
	}
	info := favorability{}
	uidstr := strconv.FormatInt(uid, 10)
	targstr := strconv.FormatInt(target, 10)
	_ = sql.db.Find("favorability", &info, "where Userinfo glob '*"+uidstr+"+"+targstr+"*'")
	info.Userinfo = uidstr + "+" + targstr + "+" + uidstr
	info.Favor += score
	if info.Favor > 100 {
		info.Favor = 100
	} else if info.Favor < 0 {
		info.Favor = 0
	}
	err = sql.db.Insert("favorability", &info)
	return info.Favor, err
}
