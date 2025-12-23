package aichat

import (
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
)

const (
	bitmaprate = 0x0000ff
	bitmaptemp = 0x00ff00
	bitmapnagt = 0x010000
	bitmapnrec = 0x020000
	bitmapnrat = 0x040000
)

var (
	fastfailnorecord = false
)

type storage ctxext.Storage

func newstorage(ctx *zero.Ctx, gid int64) (storage, error) {
	s, err := ctxext.NewStorage(ctx, gid)
	return storage(s), err
}

func (s storage) rate() uint8 {
	return uint8((ctxext.Storage)(s).Get(bitmaprate))
}

func (s storage) temp() float32 {
	temp := int8((ctxext.Storage)(s).Get(bitmaptemp))
	// 处理温度参数
	if temp <= 0 {
		temp = 70 // default setting
	}
	if temp > 100 {
		temp = 100
	}
	return float32(temp) / 100
}

func (s storage) noagent() bool {
	return (ctxext.Storage)(s).GetBool(bitmapnagt)
}

func (s storage) norecord() bool {
	return (ctxext.Storage)(s).GetBool(bitmapnrec)
}

func (s storage) noreplyat() bool {
	return (ctxext.Storage)(s).GetBool(bitmapnrat)
}
