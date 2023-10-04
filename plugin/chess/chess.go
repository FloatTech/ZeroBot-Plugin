// Package chess 国际象棋
package chess

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const helpString = `- 参与/创建一盘游戏：「下棋」(chess)
- 参与/创建一盘盲棋：「盲棋」(blind)
- 投降认输：「认输」 (resign)
- 请求、接受和棋：「和棋」 (draw)
- 走棋：!Nxf3 中英文感叹号均可，格式请参考“代数记谱法”(Algebraic notation)
- 中断对局：「中断」 (abort)（仅群主/管理员有效）
- 查看等级分排行榜：「排行榜」(ranking)
- 查看自己的等级分：「等级分」(rate)
- 清空等级分：「清空等级分 QQ号」(.clean.rate) （仅超管有效）`

var (
	limit       = ctxext.NewLimiterManager(time.Microsecond*2500, 1)
	tempFileDir string
	engine      = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "国际象棋",
		Help:              helpString,
		PrivateDataFolder: "chess",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("有操作正在执行, 请稍后再试..."),
				),
			)
		}),
	))
)

func init() {
	// 初始化临时文件夹
	tempFileDir = path.Join(engine.DataFolder(), "temp")
	err := os.MkdirAll(tempFileDir, 0750)
	if err != nil {
		panic(err)
	}
	// 初始化数据库
	dbFilePath := engine.DataFolder() + "chess.db"
	initDatabase(dbFilePath)
	// 注册指令
	engine.OnFullMatchGroup([]string{"下棋", "chess"}, zero.OnlyGroup).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.Sender == nil {
				return
			}
			userUin := ctx.Event.UserID
			userName := ctx.Event.Sender.NickName
			groupCode := ctx.Event.GroupID
			replyMessage, err := game(groupCode, userUin, userName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnFullMatchGroup([]string{"认输", "resign"}, zero.OnlyGroup).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			groupCode := ctx.Event.GroupID
			replyMessage, err := resign(groupCode, userUin)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnFullMatchGroup([]string{"和棋", "draw"}, zero.OnlyGroup).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			groupCode := ctx.Event.GroupID
			replyMessage, err := draw(groupCode, userUin)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnFullMatchGroup([]string{"中断", "abort"}, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			groupCode := ctx.Event.GroupID
			replyMessage, err := abort(groupCode)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnFullMatchGroup([]string{"盲棋", "blind"}, zero.OnlyGroup).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.Sender == nil {
				return
			}
			userUin := ctx.Event.UserID
			userName := ctx.Event.Sender.NickName
			groupCode := ctx.Event.GroupID
			replyMessage, err := blindfold(groupCode, userUin, userName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnRegex("^[!|！]([0-8]|[R|N|B|Q|K|O|a-h|x]|[-|=|+])+$", zero.OnlyGroup).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			groupCode := ctx.Event.GroupID
			userMsgStr := ctx.State["regex_matched"].([]string)[0]
			moveStr := strings.TrimPrefix(strings.TrimPrefix(userMsgStr, "！"), "!")
			replyMessage, err := play(groupCode, userUin, moveStr)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnFullMatchGroup([]string{"排行榜", "ranking"}).SetBlock(true).Limit(limit.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			replyMessage, err := getRanking()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnFullMatchGroup([]string{"等级分", "rate"}).SetBlock(true).Limit(limit.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			if ctx.Event.Sender == nil {
				return
			}
			userUin := ctx.Event.UserID
			userName := ctx.Event.Sender.NickName
			replyMessage, err := rate(userUin, userName)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})

	engine.OnPrefixGroup([]string{"清空等级分", ".clean.rate"}, zero.SuperUserPermission).SetBlock(true).
		Limit(limit.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			playerUin, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64)
			if err != nil || playerUin <= 0 {
				ctx.Send(fmt.Sprintf("解析失败「%s」不是正确的 QQ 号。", args))
				return
			}
			replyMessage, err := cleanUserRate(playerUin)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(replyMessage)
		})
}
