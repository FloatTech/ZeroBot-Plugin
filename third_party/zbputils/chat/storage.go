package chat

import (
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/zbputils/ctxext"
)

// Bitmap constants for storage
const (
	BitmapRate = 0x0000ff
	BitmapTemp = 0x00ff00
	BitmapNagt = 0x010000
	BitmapNrec = 0x020000
	BitmapNrat = 0x040000
)

// Storage wraps ctxext.Storage for chat-specific storage operations
type Storage ctxext.Storage

// NewStorage creates a new Storage instance
func NewStorage(ctx *zero.Ctx, gid int64) (Storage, error) {
	s, err := ctxext.NewStorage(ctx, gid)
	return Storage(s), err
}

// Rate returns the rate value from storage
func (s Storage) Rate() uint8 {
	return uint8((ctxext.Storage)(s).Get(BitmapRate))
}

// Temp returns the temperature value from storage
func (s Storage) Temp() float32 {
	temp := int8((ctxext.Storage)(s).Get(BitmapTemp))
	// 处理温度参数
	if temp <= 0 {
		temp = 70 // default setting
	}
	if temp > 100 {
		temp = 100
	}
	return float32(temp) / 100
}

// NoAgent returns whether the agent is disabled
func (s Storage) NoAgent() bool {
	return (ctxext.Storage)(s).GetBool(BitmapNagt)
}

// NoRecord returns whether recording is disabled
func (s Storage) NoRecord() bool {
	return (ctxext.Storage)(s).GetBool(BitmapNrec)
}

// NoReplyAt returns whether replying with @ is disabled
func (s Storage) NoReplyAt() bool {
	return (ctxext.Storage)(s).GetBool(BitmapNrat)
}
