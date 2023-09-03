// Package base64gua base64卦 与 tea 加解密
package base64gua

import (
	"github.com/FloatTech/floatbox/crypto"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/fumiama/unibase2n"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "六十四卦加解密",
		Help:             "- 六十四卦加密xxx\n- 六十四卦解密xxx\n- 六十四卦用yyy加密xxx\n- 六十四卦用yyy解密xxx",
	})
	en.OnRegex(`^六十四卦加密\s*(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := unibase2n.Base64Gua.EncodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex(`^六十四卦解密\s*([䷀-䷿]+[☰☱]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := unibase2n.Base64Gua.DecodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	en.OnRegex(`^六十四卦用(.+)加密\s*(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF16BE2UTF8(unibase2n.Base64Gua.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex(`^六十四卦用(.+)解密\s*([䷀-䷿]+[☰☱]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF82UTF16BE(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(unibase2n.Base64Gua.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}
