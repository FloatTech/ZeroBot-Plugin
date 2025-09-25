package aichat

import (
	"errors"
	"math/bits"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bitmaprate = 0x0000ff
	bitmaptemp = 0x00ff00
	bitmapnagt = 0x010000
	bitmapnrec = 0x020000
)

type storage int64

func newstorage(ctx *zero.Ctx, gid int64) (storage, error) {
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return 0, errors.New("找不到 manager")
	}
	x := c.GetData(gid)
	return storage(x), nil
}

func (s storage) saveto(ctx *zero.Ctx, gid int64) error {
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return errors.New("找不到 manager")
	}
	return c.SetData(int64(s), gid)
}

func (s storage) getbybmp(bmp int64) int64 {
	sft := bits.TrailingZeros64(uint64(bmp))
	return (int64(s) & bmp) >> int64(sft)
}

func (s *storage) setbybmp(x int64, bmp int64) {
	if bmp == 0 {
		panic("cannot use bmp == 0")
	}
	sft := bits.TrailingZeros64(uint64(bmp))
	*s = storage((int64(*s) & (^bmp)) | ((x & bmp) << int64(sft)))
}

func (s storage) rate() uint8 {
	return uint8(s.getbybmp(bitmaprate))
}

func (s storage) temp() float32 {
	temp := s.getbybmp(bitmaptemp)
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
	return s.getbybmp(bitmapnagt) != 0
}

func (s storage) norecord() bool {
	return s.getbybmp(bitmapnrec) != 0
}

func newstoragebitmap(bmp int64, minv, maxv int64) func(ctx *zero.Ctx) {
	return func(ctx *zero.Ctx) {
		args := strings.TrimSpace(ctx.State["args"].(string))
		if args == "" {
			ctx.SendChain(message.Text("ERROR: empty args"))
			return
		}
		r, err := strconv.ParseInt(args, 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: parse int64 err: ", err))
			return
		}
		if r > maxv {
			r = maxv
		} else if r < minv {
			r = minv
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor, err := newstorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		stor.setbybmp(r, bmp)
		err = stor.saveto(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set data err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}

func newstoragebool(bmp int64) func(ctx *zero.Ctx) {
	if bits.OnesCount64(uint64(bmp)) != 1 {
		panic("bool bmp must be 1-bit-long")
	}
	return func(ctx *zero.Ctx) {
		args := ctx.State["regex_matched"].([]string)
		isone := args[1] == "不"
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		stor, err := newstorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		v := 0
		if isone {
			v = 1
		}
		stor.setbybmp(int64(v), bmp)
		err = stor.saveto(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: set data err: ", err))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}
