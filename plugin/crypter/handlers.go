// Package crypter 处理函数
package crypter

import (
	"github.com/FloatTech/AnimeAPI/airecord"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// hou
func houEncryptHandler(ctx *zero.Ctx) {
	text := ctx.State["regex_matched"].([]string)[1]
	result := encodeHou(text)
	logrus.Infoln("[crypter] 回复内容:", result)
	recCfg := airecord.GetConfig()
	record := ctx.GetAIRecord(recCfg.ModelID, recCfg.Customgid, result)
	if record != "" {
		ctx.SendChain(message.Record(record))
	} else {
		ctx.SendChain(message.Text(result))
	}
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
