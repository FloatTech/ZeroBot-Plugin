// Package scale 叔叔的AI二次元图片放大
package scale

import (
	"bytes"
	"image"
	"math"
	"os"
	"strconv"
	"time"

	_ "image/gif"  // import gif decoding
	_ "image/jpeg" // import jpg decoding
	_ "image/png"  // import png decoding

	_ "golang.org/x/image/webp" // import webp decoding

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/nsfw"
	"github.com/FloatTech/AnimeAPI/scale"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img"
	"github.com/FloatTech/zbputils/img/writer"
	"github.com/FloatTech/zbputils/web"
)

func init() {
	engine := control.Register("scale", &control.Options{
		DisableOnDefault:  false,
		Help:              "叔叔的AI二次元图片放大\n- 放大图片[图片]",
		PrivateDataFolder: "scale",
	}).ApplySingle(ctxext.DefaultSingle)
	cachedir := engine.DataFolder()
	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"放大图片"}, zero.OnlyGroup, zero.MustProvidePicture, getPara).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["image_url"].([]string)
			if len(url) > 0 {
				datachan := make(chan []byte, 1)
				var errsub error
				go func() {
					var d []byte
					d, errsub = web.GetData(url[0])
					datachan <- d
				}()
				ctx.SendChain(message.Text("少女祈祷中..."))

				p, err := nsfw.Classify(url[0])
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				if p.Drawings < 0.1 || p.Neutral > 0.8 {
					ctx.SendChain(message.Text("请发送二次元图片!"))
					return
				}

				data := <-datachan
				if errsub != nil {
					ctx.SendChain(message.Text("ERROR:", errsub))
					return
				}
				im, _, err := image.Decode(bytes.NewReader(data))
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				px := im.Bounds().Size().X * im.Bounds().Size().Y
				paras := ctx.State["scale_paras"].([2]int)

				if px > 512*512 {
					px = int(math.Pow(float64(px), 0.5) + 0.5)
					x := im.Bounds().Size().X * 512 / px
					y := im.Bounds().Size().Y * 512 / px
					ctx.SendChain(message.Text("图片", im.Bounds().Size().X, "x", im.Bounds().Size().Y, "过大，调整图片至", x, "x", y))
					im = img.Size(im, x, y).Im
					w := binary.SelectWriter()
					defer binary.PutWriter(w)
					_, err = writer.WriteTo(im, w)
					if err != nil {
						ctx.SendChain(message.Text("ERROR:", err))
						return
					}
					data, err = scale.Post(bytes.NewReader(w.Bytes()), paras[0], paras[1], 2)
				} else {
					data, err = scale.Get(url[0], paras[0], paras[1], 2)
				}
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
	next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
	recv, cancel := next.Repeat()
	i := 0
	paras := [2]int{}
	ctx.SendChain(message.Text("请输入模型序号\n0.", scale.Models[0], "\n1.", scale.Models[1], "\n2.", scale.Models[2], "\n3.", scale.Models[3], "\n4.", scale.Models[4]))
	for {
		select {
		case <-time.After(time.Second * 120):
			cancel()
			return false
		case c := <-recv:
			msg := c.Event.Message.ExtractPlainText()
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
