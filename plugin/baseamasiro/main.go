// Package baseamasiro base天城文 与 tea 加解密
package baseamasiro

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
		Brief:            "天城文加解密",
		Help:             "- 天城文加密xxx\n- 天城文解密xxx\n- 天城文用yyy加密xxx\n- 天城文用yyy解密xxx",
	})
	en.OnRegex(`^天城文加密\s*(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := unibase2n.BaseDevanagari.EncodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex(`^天城文解密\s*([ऀ-ॿ]+[০-৫]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := unibase2n.BaseDevanagari.DecodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	en.OnRegex(`^天城文用(.+)加密\s*(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF16BE2UTF8(unibase2n.BaseDevanagari.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex(`^天城文用(.+)解密\s*([ऀ-ॿ]+[০-৫]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := crypto.GetTEA(key)
			es, err := unibase2n.UTF82UTF16BE(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(unibase2n.BaseDevanagari.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}
