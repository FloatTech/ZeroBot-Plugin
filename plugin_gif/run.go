// Package gif 制图
package gif

import (
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

var (
	cmds = []string{"搓", "冲", "摸", "拍", "丢", "吃", "敲", "啃", "蹭", "爬", "撕",
		"灰度", "上翻", "下翻", "左翻", "右翻", "反色", "浮雕", "打码", "负片"}
	botpath, _ = os.Getwd()
	datapath   = botpath + "/data/gif/"
)

func init() { // 插件主体
	_ = os.MkdirAll(datapath, 0755)
	rand.Seed(time.Now().UnixNano()) // 设置种子
	control.Register("gif", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "制图\n- " + strings.Join(cmds, "\n- "),
	}).ApplySingle(ctxext.DefaultSingle).OnRegex(`^(` + strings.Join(cmds, "|") + `)\D*?(\[CQ:(image\,file=([0-9a-zA-Z]{32}).*|at.+?(\d{5,11}))\].*|(\d+))$`).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := newContext(ctx.Event.UserID)
		list := ctx.State["regex_matched"].([]string)
		c.prepareLogos(list[4]+list[5]+list[6], strconv.FormatInt(ctx.Event.UserID, 10))
		var picurl string
		var err error
		if len([]rune(list[1])) == 1 {
			r := reflect.ValueOf(c).MethodByName("A" + list[1]).Call(nil)
			picurl = r[0].String()
			if !r[1].IsNil() {
				err = r[1].Interface().(error)
			}
		} else {
			picurl, err = c.other(list[1]) // "灰度", "上翻", "下翻", "左翻", "右翻", "反色", "倒放", "浮雕", "打码", "负片"
		}
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Image(picurl))
	})
}
