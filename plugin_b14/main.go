package b14coder

import (
	"unsafe"

	base14 "github.com/fumiama/go-base16384"
	tea "github.com/fumiama/gofastTEA"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

func init() {
	en := control.Register("base16384", &control.Options{
		DisableOnDefault: false,
		Help: "base16384加解密\n" +
			"- 加密xxx\n- 解密xxx\n- 用yyy加密xxx\n- 用yyy解密xxx",
	})
	en.OnRegex(`^加密(.*)`).SetBlock(true).ThirdPriority().
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es, err := base14.UTF16be2utf8(base14.EncodeString(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex("^解密([\u4e00-\u8e00]*[\u3d01-\u3d06]?)$").SetBlock(true).ThirdPriority().
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es, err := base14.UTF82utf16be(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(base14.DecodeString(es)))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	en.OnRegex(`^用(.*)加密(.*)`).SetBlock(true).ThirdPriority().
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := getea(key)
			es, err := base14.UTF16be2utf8(base14.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex("^用(.*)解密([\u4e00-\u8e00]*[\u3d01-\u3d06]?)$").SetBlock(true).ThirdPriority().
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := getea(key)
			es, err := base14.UTF82utf16be(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(base14.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}

// Slice is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may
// change in a later release.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func getea(key string) tea.TEA {
	kr := []rune(key)
	if len(kr) > 4 {
		kr = kr[:4]
	} else {
		for len(kr) < 4 {
			kr = append(kr, rune(4-len(kr)))
		}
	}
	skr := *(*slice)(unsafe.Pointer(&kr))
	skr.Len *= 4
	skr.Cap *= 4
	return tea.NewTeaCipher(*(*[]byte)(unsafe.Pointer(&skr)))
}
