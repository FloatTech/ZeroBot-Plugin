// Package poker 抽扑克牌
package poker

import (
	"encoding/json"
	"math/rand"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 图片来源 https://www.bilibili.com/opus/834601953403076633

var cardImgPathList []string

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "抽扑克牌",
		Help:             "- 抽扑克\n- poker",
		PublicDataFolder: "Poker",
	}).ApplySingle(ctxext.DefaultSingle)

	getImg := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		data, err := engine.GetLazyData("imgdata.json", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = json.Unmarshal(data, &cardImgPathList)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})

	engine.OnFullMatchGroup([]string{"抽扑克", "poker"}, getImg).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			randomIndex := rand.Intn(len(cardImgPathList))
			randomImgPath := cardImgPathList[randomIndex]
			imgData, err := engine.GetLazyData(randomImgPath, true)
			if err != nil {
				ctx.Send("[poker]读取扑克图片失败")
				return
			}
			ctx.Send(message.ImageBytes(imgData))
		})
}
