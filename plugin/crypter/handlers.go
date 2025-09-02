// Package crypter 处理函数
package crypter

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// hou
func houEncryptHandler(ctx *zero.Ctx) {
	text := ctx.State["regex_matched"].([]string)[1]
	result := encodeHou(text)
	ctx.SendChain(message.Text(result))
}

func houDecryptHandler(ctx *zero.Ctx) {
	text := ctx.State["regex_matched"].([]string)[1]
	result := decodeHou(text)
	ctx.SendChain(message.Text(result))
}

// fumo
func fumoEncryptHandler(ctx *zero.Ctx) {
	text := ctx.State["regex_matched"].([]string)[1]
	result := encryptFumo(text)
	ctx.SendChain(message.Text(result))
}

func fumoDecryptHandler(ctx *zero.Ctx) {
	text := ctx.State["regex_matched"].([]string)[1]
	result := decryptFumo(text)
	ctx.SendChain(message.Text(result))
}
