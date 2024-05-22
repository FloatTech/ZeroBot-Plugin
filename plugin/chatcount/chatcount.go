// Package chatcount 聊天时长统计
package chatcount

import (
	"fmt"
	"strconv"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	rankSize = 15
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "聊天时长统计",
		Help:              "- 查询水群@xxx\n- 查看水群排名",
		PrivateDataFolder: "chatcount",
	})
	go func() {
		ctdb = initialize(engine.DataFolder() + "chatcount.db")
	}()
	engine.OnMessage(zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			todayTime, remindFlag := ctdb.updateChatTime(ctx.Event.GroupID, ctx.Event.UserID)
			if remindFlag {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("BOT提醒：你今天已经水群%d分钟了！", todayTime)))
			}
		})

	engine.OnPrefix(`查询水群`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		todayTime, totalTime := ctdb.getChatTime(ctx.Event.GroupID, ctx.Event.UserID)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("%s今天水了%d分钟，总计%d分钟。", name, todayTime, totalTime)))
	})
	engine.OnFullMatch("查看水群排名", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			text := strings.Builder{}
			text.WriteString("今日水群排行榜:\n")
			chatTimeList := ctdb.getChatRank(ctx.Event.GroupID)
			for i := 0; i < len(chatTimeList) && i < rankSize; i++ {
				text.WriteString("第")
				text.WriteString(strconv.Itoa(i + 1))
				text.WriteString("名:")
				text.WriteString(ctx.CardOrNickName(chatTimeList[i].UserID))
				text.WriteString(" - ")
				text.WriteString(strconv.FormatInt(chatTimeList[i].TodayTime/60, 10))
				text.WriteString("分钟\n")
			}
			ctx.SendChain(message.Text(text.String()))
		})

}
