// Package gif 制图
package gif

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

var (
	cmds = []string{"搓", "冲", "摸", "拍", "丢", "吃", "敲", "啃", "蹭", "爬", "撕",
		"灰度", "上翻", "下翻", "左翻", "右翻", "反色", "浮雕", "打码", "负片"}
	botpath, _ = os.Getwd()
	datapath   = botpath + "/data/gif/"
)

func init() { // 插件主体
	_ = os.RemoveAll(datapath) // 清除缓存图片
	err := os.MkdirAll(datapath, 0755)
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UnixNano()) // 设置种子
	control.Register("gif", order.PrioGIF, &control.Options{
		DisableOnDefault: false,
		Help:             "制图\n- " + strings.Join(cmds, "\n- "),
	}).ApplySingle(ctxext.DefaultSingle).OnRegex(`^(` + strings.Join(cmds, "|") + `)\D*?(\[CQ:(image\,file=([0-9a-zA-Z]{32}).*|at.+?(\d{5,11}))\].*|(\d+))$`).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		c := newContext(ctx.Event.UserID)
		list := ctx.State["regex_matched"].([]string)
		c.prepareLogos(list[4]+list[5]+list[6], strconv.FormatInt(ctx.Event.UserID, 10))
		var picurl string
		switch list[1] {
		case "爬":
			picurl = c.pa()
		case "摸":
			picurl = c.mo()
		case "吃":
			picurl = c.chi()
		case "啃":
			picurl = c.ken()
		case "蹭":
			picurl = c.ceng()
		case "敲":
			picurl = c.qiao()
		case "搓":
			picurl = c.cuo()
		case "拍":
			picurl = c.pai()
		case "丢":
			picurl = c.diu()
		case "撕":
			picurl = c.si()
		case "冲":
			picurl = c.chong()
		default:
			picurl = c.other(list[1]) // "灰度", "上翻", "下翻", "左翻", "右翻", "反色", "倒放", "浮雕", "打码", "负片"
		}
		ctx.SendChain(message.Image(picurl))
	})
}
