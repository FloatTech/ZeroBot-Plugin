// Package scale 叔叔的AI二次元图片放大
package scale

import (
	"os"
	"strconv"
	"time"

	"github.com/FloatTech/AnimeAPI/nsfw"
	"github.com/FloatTech/AnimeAPI/scale"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

const cachedir = "data/scale/"

func init() {
	_ = os.RemoveAll(cachedir)
	err := os.MkdirAll(cachedir, 0755)
	if err != nil {
		panic(err)
	}
	engine := control.Register("scale", order.PrioScale, &control.Options{
		DisableOnDefault: false,
		Help:             "叔叔的AI二次元图片放大\n- 放大图片[图片]",
	})
	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"放大图片"}, zero.OnlyGroup, ctxext.CmdMatch, ctxext.MustGiven, getPara).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["image_url"].([]string)
			if len(url) > 0 {
				ctx.SendChain(message.Text("少女祈祷中..."))
				p, err := nsfw.Classify(url[0])
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				if p[0].Drawings < 0.1 || p[0].Neutral > 0.8 {
					ctx.SendChain(message.Text("请发送二次元图片!"))
					return
				}
				paras := ctx.State["scale_paras"].([2]int)
				data, err := scale.Get(url[0], paras[0], paras[1], 2)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				n := cachedir + strconv.Itoa(int(ctx.Event.UserID))
				f, err := os.Create(n)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				_, _ = f.Write(data)
				_ = f.Close()
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + n))
			}
		})
}

func getPara(ctx *zero.Ctx) bool {
	next := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID))
	recv, cancel := next.Repeat()
	i := 0
	paras := [2]int{}
	ctx.SendChain(message.Text("请输入模型序号\n0.", scale.Models[0], "\n1.", scale.Models[1], "\n2.", scale.Models[2], "\n3.", scale.Models[3], "\n4.", scale.Models[4]))
	for {
		select {
		case <-time.After(time.Second * 120):
			return false
		case e := <-recv:
			msg := e.Message.ExtractPlainText()
			num, err := strconv.Atoi(msg)
			if err != nil {
				ctx.SendChain(message.Text("请输入数字!"))
				continue
			}
			switch i {
			case 0:
				if num < 0 || num > 4 {
					ctx.SendChain(message.Text("模型序号非法!"))
					continue
				}
				paras[0] = num
				ctx.SendChain(message.Text("请输入放大倍数(2-4)"))
			case 1:
				if num < 2 || num > 4 {
					ctx.SendChain(message.Text("放大倍数非法!"))
					continue
				}
				cancel()
				paras[1] = num
				ctx.State["scale_paras"] = paras
				return true
			}
			i++
		}
	}
}
