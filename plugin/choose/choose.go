// Package choose 选择困难症帮手
package choose

import (
	"math/rand"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "选择困难症帮手",
		Help: "例: 选择可口可乐还是百事可乐\n" +
			"选择肯德基还是麦当劳还是必胜客",
	})
	engine.OnPrefix("选择").SetBlock(true).Handle(handle)
}
func handle(ctx *zero.Ctx) {
	rawOptions := strings.Split(ctx.State["args"].(string), "还是")
	var options = make([]string, 0)
	for count, option := range rawOptions {
		options = append(options, strconv.Itoa(count+1)+", "+option)
	}
	result := rawOptions[rand.Intn(len(rawOptions))]
	name := ctx.Event.Sender.NickName
	ctx.SendChain(message.Text("> ", name, "\n",
		"你的选项有:", "\n",
		strings.Join(options, "\n"), "\n",
		"你最终会选: ", result,
	))
}
