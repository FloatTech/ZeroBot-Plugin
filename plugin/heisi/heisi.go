// Package heisi 黑丝
package heisi

import (
	"errors"
	"math/rand"
	"os"
	"strconv"
	"time"
	"unsafe"

	"github.com/FloatTech/AnimeAPI/setu"
	fbctxext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	heisiPic []item
	baisiPic []item
	jkPic    []item
	jurPic   []item
	zukPic   []item
	mcnPic   []item
	fileList = [...]string{"heisi.bin", "baisi.bin", "jk.bin", "jur.bin", "zuk.bin", "mcn.bin"}
)

func init() { // 插件主体
	p, err := setu.NewPool(setu.DefaultPoolDir,
		func(s string) (string, error) {
			if s != "黑丝" && s != "白丝" && s != "jk" && s != "巨乳" && s != "足控" && s != "网红" {
				return "", errors.New("invalid call")
			}
			typ := setu.DefaultPoolDir + "/" + s
			if file.IsNotExist(typ) {
				err := os.MkdirAll(typ, 0755)
				if err != nil {
					return "", err
				}
			}
			var pic item
			switch s {
			case "黑丝":
				pic = heisiPic[rand.Intn(len(heisiPic))]
			case "白丝":
				pic = baisiPic[rand.Intn(len(baisiPic))]
			case "jk":
				pic = jkPic[rand.Intn(len(jkPic))]
			case "巨乳":
				pic = jurPic[rand.Intn(len(jurPic))]
			case "足控":
				pic = zukPic[rand.Intn(len(zukPic))]
			case "网红":
				pic = mcnPic[rand.Intn(len(mcnPic))]
			}
			return pic.String(), nil
		}, func(s string) ([]byte, error) {
			return web.RequestDataWith(web.NewTLS12Client(), s, "GET", "http://hs.heisiwu.com/", web.RandUA(), nil)
		}, time.Minute)
	if err != nil {
		panic(err)
	}

	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "黑丝",
		Help:             "- 来点黑丝\n- 来点白丝\n- 来点jk\n- 来点巨乳\n- 来点足控\n- 来点网红",
		PublicDataFolder: "Heisi",
	})

	engine.OnFullMatchGroup([]string{"来点黑丝", "来点白丝", "来点jk", "来点巨乳", "来点足控", "来点网红"}, zero.OnlyGroup, fbctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		for i, filePath := range fileList {
			data, err := engine.GetLazyData(filePath, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
			if len(data)%10 != 0 {
				ctx.SendChain(message.Text("ERROR: invalid data " + strconv.Itoa(i)))
				return false
			}
			s := (*slice)(unsafe.Pointer(&data))
			s.len /= 10
			s.cap /= 10
			switch i {
			case 0:
				heisiPic = *(*[]item)(unsafe.Pointer(s))
			case 1:
				baisiPic = *(*[]item)(unsafe.Pointer(s))
			case 2:
				jkPic = *(*[]item)(unsafe.Pointer(s))
			case 3:
				jurPic = *(*[]item)(unsafe.Pointer(s))
			case 4:
				zukPic = *(*[]item)(unsafe.Pointer(s))
			case 5:
				mcnPic = *(*[]item)(unsafe.Pointer(s))
			}
		}
		return true
	})).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			matched := ctx.State["matched"].(string)
			pic, err := p.Roll(matched[3*2:])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+file.BOTPATH+"/"+pic))}
			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
		})
}
