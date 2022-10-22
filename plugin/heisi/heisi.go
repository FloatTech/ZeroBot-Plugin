// Package heisi 黑丝
package heisi

import (
	"math/rand"
	"strconv"
	"unsafe"

	fbctxext "github.com/FloatTech/floatbox/ctxext"
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
	engine := control.Register("heisi", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "黑丝\n" +
			"- 来点黑丝\n- 来点白丝\n- 来点jk\n- 来点巨乳\n- 来点足控\n- 来点网红",
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
			var pic item
			switch matched {
			case "来点黑丝":
				pic = heisiPic[rand.Intn(len(heisiPic))]
			case "来点白丝":
				pic = baisiPic[rand.Intn(len(baisiPic))]
			case "来点jk":
				pic = jkPic[rand.Intn(len(jkPic))]
			case "来点巨乳":
				pic = jurPic[rand.Intn(len(jurPic))]
			case "来点足控":
				pic = zukPic[rand.Intn(len(zukPic))]
			case "来点网红":
				pic = mcnPic[rand.Intn(len(mcnPic))]
			}
			m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image(pic.String()))}
			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
			}
		})
}
