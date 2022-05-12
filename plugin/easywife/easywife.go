// Package easywife 简单本地老婆
package easywife

import (
	"os"
	"regexp"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
)

func init() {
	engine := control.Register("easywife", &control.Options{
		DisableOnDefault: false,
		Help: "本地老婆\n" +
			"抽老婆",
		PrivateDataFolder: "easywife",
	})
	cachePath := engine.DataFolder() + "wife/"
	err := os.MkdirAll(cachePath, 0755)
	if err != nil {
		panic(err)
	}
	engine.OnPrefix("抽老婆").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			wifes, err := os.ReadDir(cachePath)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			name := ctx.NickName()
			n := ctxext.RandSenderPerDayN(ctx, len(wifes))
			wn := wifes[n].Name()
			reg := regexp.MustCompile(`[^\.]+`)
			list := reg.FindAllString(wn, -1)
			ctx.SendChain(
				message.Text(name, "さんが二次元で結婚するであろうヒロインは、", "\n"),
				message.Image("file:///"+file.BOTPATH+"/"+cachePath+wn),
				message.Text("\n【", list[0], "】です！"))
		})
}
