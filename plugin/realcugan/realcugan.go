// Package realcugan Real-CUGAN清晰术
package realcugan

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"strings"

	hf "github.com/FloatTech/AnimeAPI/huggingface"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	realcuganRepo = "shichen1231/Real-CUGAN"
)

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "Real-CUGAN清晰术",
		Help:              "- 清晰术(双重吟唱|三重吟唱|四重吟唱)(强力术式|中等术式|弱术式|不变式|原式)[图片]",
		PrivateDataFolder: "realcugan",
	})
	engine.OnPrefix("清晰术", zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			realcuganURL := fmt.Sprintf(hf.HTTPSPredictPath, realcuganRepo)
			for _, url := range ctx.State["image_url"].([]string) {
				imgdata, err := web.GetData(url)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				img, _, err := image.Decode(bytes.NewReader(imgdata))
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				// 初始化参数
				var (
					fashu = ctx.Event.Message.ExtractPlainText()
					scale = 2
					con   = "conservative"
				)
				switch {
				case strings.Contains(fashu, "双重吟唱"):
					scale = 2
				case strings.Contains(fashu, "三重吟唱") && img.Bounds().Dx()*img.Bounds().Dy() < 400000:
					scale = 3
				case strings.Contains(fashu, "四重吟唱") && img.Bounds().Dx()*img.Bounds().Dy() < 400000:
					scale = 4
				}
				switch {
				case strings.Contains(fashu, "强力术式"):
					con = "denoise3x"
				case strings.Contains(fashu, "中等术式"):
					con = "no-denoise"
					if scale == 2 {
						con = "denoise2x"
					}
				case strings.Contains(fashu, "弱术式"):
					con = "no-denoise"
					if scale == 2 {
						con = "denoise1x"
					}
				case strings.Contains(fashu, "不变式"):
					con = "no-denoise"
				case strings.Contains(fashu, "原式"):
					con = "conservative"
				}
				modelname := fmt.Sprintf("up%vx-latest-%v.pth", scale, con)
				encodeStr := base64.StdEncoding.EncodeToString(imgdata)
				encodeStr = "data:image/jpeg;base64," + encodeStr
				pr := hf.PushRequest{
					Data: []interface{}{encodeStr, modelname, 2},
				}
				buf := bytes.NewBuffer([]byte{})
				err = json.NewEncoder(buf).Encode(pr)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				data, err := web.PostData(realcuganURL, "application/json", buf)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				imgStr := gjson.ParseBytes(data).Get("data.0").String()
				m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text(scale, "重唱", con, "分支大清晰术!")),
					ctxext.FakeSenderForwardNode(ctx, message.Image("base64://"+strings.TrimPrefix(imgStr, "data:image/png;base64,")))}
				if id := ctx.Send(m).ID(); id == 0 {
					ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
				}
			}
		})
}
