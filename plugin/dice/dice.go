package dice

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	engine := control.Register("dice", &control.Options{
		DisableOnDefault:  true,
		Help:              "Dice! beta for zb ",
		PrivateDataFolder: "dice",
	})
	engine.OnRegex(`.rd([\\u4e00-\\u9fa5]|[a-z]|[A-Z]| )+|[0-9]+`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			temp := ctx.State["regex_matched"].([]string)[1]
			math, _ := strconv.Atoi(ctx.State["regex_matched"].([]string)[2])
			r := rand.Intn(100) + 1
			switch {
			case r < math && r/2 < math/2 && r/5 < math/5:
				win := "极难成功"
				msg := fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, temp, r, math, win)
				ctx.Send(msg)
			case r < math && r/2 < math/2:
				win := "困难成功"
				msg := fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, temp, r, math, win)
				ctx.Send(msg)
			case r < math:
				win := "成功"
				msg := fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, temp, r, math, win)
				ctx.Send(msg)
			default:
				win := "失败"
				msg := fmt.Sprintf("%s进行%s检定:\nD100=%d/%d %s", nickname, temp, r, math, win)
				ctx.Send(msg)
			}
		})
}
