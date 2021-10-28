package b14coder

import (
	base14 "github.com/fumiama/go-base16384"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

func init() {
	en := control.Register("base16384", &control.Options{
		DisableOnDefault: false,
		Help: "base16384加解密\n" +
			"- 加密xxx\n- 解密xxx",
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
	en.OnRegex("^解密([\u4e00-\u9fa5]*[\u3d01-\u3d06]?)$").SetBlock(true).ThirdPriority().
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es, err := base14.UTF82utf16be(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(base14.DecodeString(es)))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}
