// Package coser images
package coser

import (
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/tidwall/gjson"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/AnimeAPI/setu"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

var (
	ua       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36"
	coserURL = "http://ovooa.com/API/cosplay/api.php"
)

func init() {
	p, err := setu.NewPool(setu.DefaultPoolDir,
		func(s string) (string, error) {
			if s != "coser" {
				return "", errors.New("invalid call")
			}
			typ := setu.DefaultPoolDir + "/" + "coser"
			if file.IsNotExist(typ) {
				err := os.MkdirAll(typ, 0755)
				if err != nil {
					return "", err
				}
			}
			data, err := web.RequestDataWith(web.NewDefaultClient(), coserURL, "GET", "", ua, nil)
			if err != nil {
				return "", err
			}
			arr := gjson.Get(helper.BytesToString(data), "data.data").Array()
			if len(arr) == 0 {
				return "", errors.New("data is empty")
			}
			pic := arr[rand.Intn(len(arr))]
			return pic.String(), nil
		}, web.GetData, time.Minute)
	if err != nil {
		panic(err)
	}
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "三次元coser",
		Help:             "- coser",
	}).ApplySingle(ctxext.DefaultSingle).OnFullMatch("coser").SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			pic, err := p.Roll("coser")
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.Send(message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+file.BOTPATH+"/"+pic))}).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
		})
}
