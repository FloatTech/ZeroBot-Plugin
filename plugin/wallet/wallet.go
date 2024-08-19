// Package wallet 钱包
package wallet

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/binary"
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
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "钱包",
		Help: "- 查看钱包排名\n" +
			"- 设置硬币名称XX\n" +
			"- 管理钱包余额[+金额|-金额][@xxx]\n" +
			"- 查看我的钱包|查看钱包余额[@xxx]\n" +
			"- 钱包转账[金额][@xxx]\n" +
			"注：仅超级用户能“管理钱包余额”\n",
		PrivateDataFolder: "wallet",
	})
	cachePath := en.DataFolder() + "cache/"
	coinNameFile := en.DataFolder() + "coin_name.txt"
	go func() {
		_ = os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
		// 更改硬币名称
		var coinName string
		if file.IsExist(coinNameFile) {
			content, err := os.ReadFile(coinNameFile)
			if err != nil {
				panic(err)
			}
			coinName = binary.BytesToString(content)
		} else {
			// 旧版本数据
			coinName = "ATRI币"
		}
		wallet.SetWalletName(coinName)
	}()

	en.OnFullMatch("查看钱包排名", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
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
				ctx.SendChain(message.Text("ERROR: 当前没人获取过", wallet.GetWalletName()))
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
				Title: wallet.GetWalletName() + "排名(1天只刷新1次)",
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
	en.OnPrefix("设置硬币名称", zero.OnlyToMe, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			coinName := strings.TrimSpace(ctx.State["args"].(string))
			err := os.WriteFile(coinNameFile, binary.StringToBytes(coinName), 0644)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			wallet.SetWalletName(coinName)
			ctx.SendChain(message.Text("记住啦~"))
		})

	en.OnPrefix(`管理钱包余额`, zero.SuperUserPermission).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			param := strings.TrimSpace(ctx.State["args"].(string))

			// 捕获修改的金额
			re := regexp.MustCompile(`^[+-]?\d+$`)
			amount, err := strconv.Atoi(re.FindString(param))
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("输入的金额异常"))
				return
			}

			// 捕获用户QQ号，只支持@事件
			var uidStr string
			if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
				uidStr = ctx.Event.Message[1].Data["qq"]
			} else {
				// 没at就修改自己的钱包
				uidStr = strconv.FormatInt(ctx.Event.UserID, 10)
			}

			uidInt, err := strconv.ParseInt(uidStr, 10, 64)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("QQ号处理失败"))
				return
			}
			if amount+wallet.GetWalletOf(uidInt) < 0 {
				ctx.SendChain(message.Text("管理失败:对方钱包余额不足，扣款失败"))
				return
			}
			err = wallet.InsertWalletOf(uidInt, amount)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:管理失败，钱包坏掉了:\n", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("钱包余额修改成功，已修改用户:", uidStr, "的钱包，修改金额为：", amount))
		})

	// 保留用户习惯,兼容旧语法“查看我的钱包”
	en.OnPrefixGroup([]string{`查看钱包余额`, `查看我的钱包`}).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			param := ctx.State["args"].(string)
			var uidStr string
			if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
				uidStr = ctx.Event.Message[1].Data["qq"]
			} else if param == "" {
				uidStr = strconv.FormatInt(ctx.Event.UserID, 10)
			}
			uidInt, err := strconv.ParseInt(uidStr, 10, 64)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("QQ号处理失败"))
				return
			}
			money := wallet.GetWalletOf(uidInt)
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("QQ号：", uidStr, "，的钱包有", money, wallet.GetWalletName()))
		})

	en.OnPrefix(`钱包转账`, zero.OnlyGroup).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			param := strings.TrimSpace(ctx.State["args"].(string))

			// 捕获修改的金额,amount扣款金额恒正（要注意符号）
			re := regexp.MustCompile(`^[+]?\d+$`)
			amount, err := strconv.Atoi(re.FindString(param))
			if err != nil || amount <= 0 {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("输入额异常，请检查金额或at是否正常"))
				return
			}

			// 捕获用户QQ号，只支持@事件
			var uidStr string
			if len(ctx.Event.Message) > 1 && ctx.Event.Message[1].Type == "at" {
				uidStr = ctx.Event.Message[1].Data["qq"]
			} else {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("获取被转方信息失败"))
				return
			}

			uidInt, err := strconv.ParseInt(uidStr, 10, 64)
			if err != nil {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("QQ号处理失败"))
				return
			}

			// 开始转账流程
			if amount > wallet.GetWalletOf(ctx.Event.UserID) {
				ctx.SendChain(message.Text("[ERROR]:钱包余额不足，转账失败"))
				return
			}

			err = wallet.InsertWalletOf(ctx.Event.UserID, -amount)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:转账失败，扣款异常:\n", err))
				return
			}

			err = wallet.InsertWalletOf(uidInt, amount)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR]:转账失败，转账时银行被打劫:\n", err))
				return
			}
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("转账成功:成功给"), message.At(uidInt), message.Text(",转账:", amount, wallet.GetWalletName()))
		})
}
