// Package aipaint ai绘图
package aipaint

import (
	"bytes"
	"fmt"
	"image"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.Register("img2img", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "以图绘图",
		Help: "- [ 以图绘图 | 以图生图 | 以图画图 ] xxx [图片]|@xxx|[qq号]\n" +
			"- 官方以图绘图api已失效 需要自建api 其他配置参数同ai绘图",
		PrivateDataFolder: "img2img",
	})
	datapath = file.BOTPATH + "/" + engine.DataFolder()
	engine.OnRegex(`^(以图绘图|以图生图|以图画图)[\s\S]*?(\[CQ:(image\,file=([0-9a-zA-Z]{32}).*|at.+?(\d{5,11}))\].*|(\d+))$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			c := newContext(ctx.Event.UserID)
			list := ctx.State["regex_matched"].([]string)
			err = c.prepareLogos(list[4]+list[5]+list[6], strconv.FormatInt(ctx.Event.UserID, 10))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			args := strings.TrimSuffix(strings.TrimPrefix(list[0], list[1]), list[2])
			if args == "" {
				ctx.SendChain(message.Text("ERROR: 以图绘图必须添加tag"))
				return
			}
			ctx.SendChain(message.Text("少女祈祷中..."))
			postURL := cfg.BaseURL + fmt.Sprintf(aipaintImg2ImgURL, cfg.Token, url.QueryEscape(strings.TrimSpace(strings.ReplaceAll(args, " ", "%20"))))

			f, err := os.Open(c.headimgsdir[0])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			defer f.Close()

			img, _, err := image.Decode(f)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			imageShape := ""
			switch {
			case img.Bounds().Dx() > img.Bounds().Dy():
				imageShape = "Landscape"
			case img.Bounds().Dx() == img.Bounds().Dy():
				imageShape = "Square"
			default:
				imageShape = "Portrait"
			}

			// 图片转base64
			base64Bytes, err := imgfactory.ToBase64(img)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			data, err := web.PostData(postURL+"&shape="+imageShape, "text/plain", bytes.NewReader(base64Bytes))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			sendAiImg(ctx, data, cfg.Interval)
		})
}
