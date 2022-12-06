// Package wallet 公用数据管理
package wallet

import (
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/golang/freetype"
	"github.com/wcharczuk/go-chart/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("DataSystem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "公用数据管理",
		Help: "- @bot 叫我[xxx]\n" +
			"- 查看我的钱包\n" +
			"- 查看钱包排名\n" +
			"注:为本群排行，若群人数太多不建议使用该功能!!!\n" +
			"\n---------主人功能---------\n" +
			"- 注销昵称 [xxx/qq号/@QQ]\n" +
			"- /钱包 [QQ号|@群友]\n" +
			"- /记录 @群友 ATRI币值\n" +
			"- /记录 @加分群友 @减分群友 ATRI币值",
	})
	cachePath := "data/wallet/cache/"
	go func() {
		_ = os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
	}()
	engine.OnRegex(`^叫我\s*([^\s]+(\s+[^\s]+)*)`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		username := ctx.State["regex_matched"].([]string)[1]
		if strings.Contains(username, "[CQ:face,id=") {
			ctx.SendChain(message.Text("昵称不支持表情包哦"))
			return
		}
		if len([]rune(username)) > 10 {
			ctx.SendChain(message.Text("昵称不得长于10个字符"))
			return
		}
		err := wallet.SetNameOf(ctx.Event.UserID, username)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("好的,", username))
	})
	engine.OnRegex(`^注销昵称(\s*\[CQ:at,qq=)?(.*[^\]$])`, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		username := ctx.State["regex_matched"].([]string)[1]
		uid, err := strconv.ParseInt(username, 10, 64)
		if err != nil {
			err = wallet.CancelNameOf(username)
		} else {
			username = wallet.GetNameOf(uid)
			err = wallet.CancelNameOf(username)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("注销成功"))
	})
	engine.OnFullMatchGroup([]string{"查看我的钱包", "/钱包"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)
		ctx.SendChain(message.At(uid), message.Text("你的钱包当前有", money, "ATRI币"))
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
			var usergroup []int64
			for _, info := range temp {
				usergroup = append(usergroup, info.Get("user_id").Int())
			}
			// 获取钱包信息
			st, err := wallet.GetGroupWalletOf(usergroup, true)
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
	engine.OnRegex(`/钱包(\s*\[CQ:at,qq=)?(\d+)`, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
		ctx.SendChain(message.Text(ctx.CardOrNickName(uid), "的钱包当前有", wallet.GetWalletOf(uid), "ATRI币"))
	})
	engine.OnRegex(`^\/记录\s*\[CQ:at,qq=(\d+)(.*\[CQ:at,qq=(\d+))?.*[^(-?\d+)$](-?\d+)`, zero.SuperUserPermission, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		adduser, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		devuser, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[3], 10, 64)
		score, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[4])
		// 第一个人记录
		err := wallet.InsertWalletOf(adduser, score)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		switch {
		case score > 0:
			ctx.SendChain(message.At(adduser), message.Text("你获取ATRI币:", score))
		case score < 0:
			ctx.SendChain(message.At(adduser), message.Text("你失去ATRI币:", score))
		}
		// 第二个人记录
		if devuser == 0 {
			return
		}
		err = wallet.InsertWalletOf(devuser, -score)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		switch {
		case -score > 0:
			ctx.SendChain(message.At(devuser), message.Text("你获取ATRI币:", score))
		case -score < 0:
			ctx.SendChain(message.At(devuser), message.Text("你失去ATRI币:", score))
		}
	})
}
