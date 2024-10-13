// Package chatcount 聊天时长统计
package chatcount

import (
	"fmt"
	"image"
	"net/http"
	"strconv"
	"sync"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/rendercard"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
)

const (
	rankSize = 10
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
			remindTime, remindFlag := ctdb.updateChatTime(ctx.Event.GroupID, ctx.Event.UserID)
			if remindFlag {
				ctx.SendChain(message.At(ctx.Event.UserID), message.Text(fmt.Sprintf("BOT提醒：你今天已经水群%d分钟了！", remindTime)))
			}
		})

	engine.OnPrefix(`查询水群`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		name := ctx.NickName()
		todayTime, todayMessage, totalTime, totalMessage := ctdb.getChatTime(ctx.Event.GroupID, ctx.Event.UserID)
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(fmt.Sprintf("%s今天水了%d分%d秒，发了%d条消息；总计水了%d分%d秒，发了%d条消息。", name, todayTime/60, todayTime%60, todayMessage, totalTime/60, totalTime%60, totalMessage)))
	})
	engine.OnFullMatch("查看水群排名", zero.OnlyGroup).Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			chatTimeList := ctdb.getChatRank(ctx.Event.GroupID)
			if len(chatTimeList) == 0 {
				ctx.SendChain(message.Text("ERROR: 没有水群数据"))
				return
			}
			rankinfo := make([]*rendercard.RankInfo, len(chatTimeList))

			wg := &sync.WaitGroup{}
			wg.Add(len(chatTimeList))
			for i := 0; i < len(chatTimeList) && i < rankSize; i++ {
				go func(i int) {
					defer wg.Done()
					resp, err := http.Get("https://q4.qlogo.cn/g?b=qq&nk=" + strconv.FormatInt(chatTimeList[i].UserID, 10) + "&s=100")
					if err != nil {
						return
					}
					defer resp.Body.Close()
					img, _, err := image.Decode(resp.Body)
					if err != nil {
						return
					}
					rankinfo[i] = &rendercard.RankInfo{
						TopLeftText:    ctx.CardOrNickName(chatTimeList[i].UserID),
						BottomLeftText: "消息数: " + strconv.FormatInt(chatTimeList[i].TodayMessage, 10) + " 条",
						RightText:      strconv.FormatInt(chatTimeList[i].TodayTime/60, 10) + "分" + strconv.FormatInt(chatTimeList[i].TodayTime%60, 10) + "秒",
						Avatar:         img,
					}
				}(i)
			}
			wg.Wait()
			fontbyte, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			img, err := rendercard.DrawRankingCard(fontbyte, "今日水群排行榜", rankinfo)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendimg, err := imgfactory.ToBytes(img)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.ImageBytes(sendimg)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
