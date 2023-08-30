// Package chess 国际象棋
package chess

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/chess/service"
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
	tempFileDir string
	engine      = control.Register("chess", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "国际象棋",
		Help:             helpString,
	})
)

func init() {
	// 初始化临时文件夹
	tempFileDir = path.Join(engine.DataFolder(), "temp")
	if !file.IsExist(tempFileDir) {
		err := os.MkdirAll(tempFileDir, 0750)
		if err != nil {
			panic(err)
		}
	}
	// 初始化数据库
	dbFilePath := engine.DataFolder() + "chess.db"
	service.InitDatabase(dbFilePath)
	// 注册指令
	engine.OnFullMatchGroup([]string{"下棋", "chess"}, zero.OnlyGroup).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			userName := ctx.Event.Sender.NickName
			groupCode := ctx.Event.GroupID
			if replyMessage := Game(groupCode, userUin, userName); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnFullMatchGroup([]string{"认输", "resign"}, zero.OnlyGroup).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			groupCode := ctx.Event.GroupID
			if replyMessage := Resign(groupCode, userUin); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnFullMatchGroup([]string{"和棋", "draw"}, zero.OnlyGroup).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			groupCode := ctx.Event.GroupID
			if replyMessage := Draw(groupCode, userUin); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnFullMatchGroup([]string{"中断", "abort"}, zero.OnlyGroup, zero.AdminPermission).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			groupCode := ctx.Event.GroupID
			if replyMessage := Abort(groupCode); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnFullMatchGroup([]string{"盲棋", "blind"}, zero.OnlyGroup).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			userName := ctx.Event.Sender.NickName
			groupCode := ctx.Event.GroupID
			if replyMessage := Blindfold(groupCode, userUin, userName); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnRegex("[!|！]([0-9]|[A-Z]|[a-z]|=|-)+", zero.OnlyGroup).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			groupCode := ctx.Event.GroupID
			userMsgStr := ctx.Event.Message.ExtractPlainText()
			userMsgStr = strings.Replace(userMsgStr, "！", "!", 1)
			moveStr := userMsgStr[1:]
			if replyMessage := Play(userUin, groupCode, moveStr); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnFullMatchGroup([]string{"排行榜", "ranking"}).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if replyMessage := Ranking(); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnFullMatchGroup([]string{"等级分", "rate"}).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userUin := ctx.Event.UserID
			userName := ctx.Event.Sender.NickName
			if replyMessage := Rate(userUin, userName); len(replyMessage) >= 1 {
				ctx.Send(replyMessage)
			}
		})
	engine.OnPrefixGroup([]string{"清空等级分", ".clean.rate"}, zero.SuperUserPermission).
		SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			if playerUin, err := strconv.ParseInt(strings.TrimSpace(args), 10, 64); err == nil && playerUin > 0 {
				if replyMessage := CleanUserRate(playerUin); len(replyMessage) >= 1 {
					ctx.Send(replyMessage)
				}
			} else {
				ctx.Send(fmt.Sprintf("解析失败「%s」不是正确的 QQ 号。", args))
			}
		})
}
