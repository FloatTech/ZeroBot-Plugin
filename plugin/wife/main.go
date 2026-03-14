// Package wife 抽老婆
package wife

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	cards  = []string{}
	re     = regexp.MustCompile(`^\[(.*?)\](.*)\..*$`)
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help:             "- 抽老婆",
		Brief:            "从老婆库抽每日老婆",
		PublicDataFolder: "Wife",
	}).ApplySingle(ctxext.DefaultSingle)
	getJSON = fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			data, err := engine.GetLazyData("wife.json", true)
			if err != nil {
				logrus.Warnf("[wife] 远程同步 wife.json 失败: %v，正在尝试读取本地缓存...", err)
				data, err = engine.GetLazyData("wife.json", false)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: 无法获取老婆库数据（同步及本地读取均失败）: ", err))
					return false
				}
			}
			err = json.Unmarshal(data, &cards)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: 老婆库格式解析失败: ", err))
				return false
			}
			logrus.Infof("[wife] 已成功加载 %d 个老婆", len(cards))
			return true
		},
	)
)

func card2name(card string) (string, string) {
	match := re.FindStringSubmatch(card)
	if len(match) >= 3 {
		return match[1], match[2]
	}
	return "", ""
}

func init() {
	_ = os.MkdirAll(engine.DataFolder()+"wives", 0755)
	engine.OnFullMatch("抽老婆", getJSON).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			card := cards[fcext.RandSenderPerDayN(ctx.Event.UserID, len(cards))]
			data, err := engine.GetLazyData("wives/"+card, true)
			var msgText string
			work, name := card2name(card)
			if work != "" && name != "" {
				msgText = fmt.Sprintf("今天的二次元老婆是~来自【%s】的【%s】哒", work, name)
			} else {
				msgText = fmt.Sprintf("今天的二次元老婆是~【%s】哒", card)
			}
			if err != nil {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text(msgText, "\n【图片下载失败: ", err, "】"),
				)
				return
			}
			if id := ctx.SendChain(
				message.At(ctx.Event.UserID),
				message.Text(msgText),
				message.ImageBytes(data),
			); id.ID() == 0 {
				ctx.SendChain(
					message.At(ctx.Event.UserID),
					message.Text(msgText, "\n【图片发送失败, 多半是被夹了，请联系维护者】"),
				)
			}
		})
}
