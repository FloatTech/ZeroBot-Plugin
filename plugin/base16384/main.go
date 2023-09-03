// Package b14coder base16384 与 tea 加解密
package b14coder

import (
	"github.com/FloatTech/floatbox/crypto"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	base14 "github.com/fumiama/go-base16384"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func init() {
	en := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "base16384加解密",
		Help:             "- 加密xxx\n- 解密xxx\n- 用yyy加密xxx\n- 用yyy解密xxx",
	})
	en.OnRegex(`^加密\s*(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := base14.EncodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex(`^解密\s*([一-踀]+[㴁-㴆]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			str := ctx.State["regex_matched"].([]string)[1]
			es := base14.DecodeString(str)
			if es != "" {
				ctx.SendChain(message.Text(es))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
	en.OnRegex(`^用(.+)加密\s*(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := crypto.GetTEA(key)
			es, err := base14.UTF16BE2UTF8(base14.Encode(t.Encrypt(helper.StringToBytes(str))))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(es)))
			} else {
				ctx.SendChain(message.Text("加密失败!"))
			}
		})
	en.OnRegex(`^用(.+)解密\s*([一-踀]+[㴁-㴆]?)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			key, str := ctx.State["regex_matched"].([]string)[1], ctx.State["regex_matched"].([]string)[2]
			t := crypto.GetTEA(key)
			es, err := base14.UTF82UTF16BE(helper.StringToBytes(str))
			if err == nil {
				ctx.SendChain(message.Text(helper.BytesToString(t.Decrypt(base14.Decode(es)))))
			} else {
				ctx.SendChain(message.Text("解密失败!"))
			}
		})
}
