package ctxext

import (
	"errors"
	"fmt"
	"math/bits"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// ErrManagerNotFound unexpected
	ErrManagerNotFound = errors.New("manager not found")
)

// Storage is a wrapper of *ctrl.Control[*zero.Ctx].Get/SetData
type Storage int64

// NewStorage private is -uid
func NewStorage(ctx *zero.Ctx, gid int64) (Storage, error) {
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return 0, ErrManagerNotFound
	}
	x := c.GetData(gid)
	logrus.Debugln("[ctxext] NewStorage get plugin", c.Service, "grp:", gid, "val:", fmt.Sprintf("%016x", x))
	return Storage(x), nil
}

// SaveTo ...
func (s Storage) SaveTo(ctx *zero.Ctx, gid int64) error {
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return ErrManagerNotFound
	}
	logrus.Debugln("[ctxext] SaveTo plugin", c.Service, "grp:", gid, "val:", fmt.Sprintf("%016x", int64(s)))
	return c.SetData(gid, int64(s))
}

// UpdateStorage merges one setting with the latest data under the control lock.
func UpdateStorage(ctx *zero.Ctx, gid int64, change func(Storage) Storage) error {
	c, ok := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
	if !ok {
		return ErrManagerNotFound
	}
	_, err := c.UpdateData(gid, "", func(old int64) (int64, error) { return int64(change(Storage(old))), nil })
	return err
}

// Get bmp is a continuous 1's sequence like 0x00ff00
func (s Storage) Get(bmp int64) int64 {
	sft := bits.TrailingZeros64(uint64(bmp))
	x := (int64(s) >> int64(sft)) & (bmp >> int64(sft))
	logrus.Debugln("[ctxext] Storage", fmt.Sprintf("%016x", int64(s)), "get bmp", fmt.Sprintf("%016x", bmp), "sft:", sft, "val:", fmt.Sprintf("%016x", x))
	return x
}

// Set bmp is a continuous 1's sequence like 0x00ff00
func (s Storage) Set(x int64, bmp int64) Storage {
	if bmp == 0 {
		panic("cannot use bmp == 0")
	}
	sft := bits.TrailingZeros64(uint64(bmp))
	news := Storage((int64(s) & (^bmp)) | ((x << int64(sft)) & bmp))
	logrus.Debugln("[ctxext] Storage", fmt.Sprintf("%016x", int64(s)), "set bmp", fmt.Sprintf("%016x", bmp), "sft:", sft, "val:", fmt.Sprintf("%016x", x), "after set:", fmt.Sprintf("%016x", int64(news)))
	return news
}

// GetBool bmp must be 1-bit-long
func (s Storage) GetBool(bmp int64) bool {
	x := s.Get(bmp) != 0
	logrus.Debugln("[ctxext] Storage", fmt.Sprintf("%016x", int64(s)), "get bmp", fmt.Sprintf("%016x", bmp), "val:", x)
	return x
}

// NewStorageSaveBitmapHandler is a easy handler that cooperate with OnPrefix
// and extract int64 from args, save it to storage.
func NewStorageSaveBitmapHandler(bmp int64, minv, maxv int64) func(ctx *zero.Ctx) {
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
		stor, err := NewStorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		_ = stor
		err = UpdateStorage(ctx, gid, func(old Storage) Storage { return old.Set(r, bmp) })
		if err != nil {
			ctx.SendChain(message.Text("配置保存失败，原设置保留。"))
			return
		}
		ctx.SendChain(message.Text("成功"))
	}
}

// NewStorageSaveBoolHandler is a easy handler that cooperate with OnRegex
// and extract bool from args, save it to storage.
//
// Note:
//   - The first submatch must be (不)?
//   - If 不 is matched, true (1) will be set.
//   - Otherwise, false (0) will be set.
func NewStorageSaveBoolHandler(bmp int64) func(ctx *zero.Ctx) {
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
		stor, err := NewStorage(ctx, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		v := 0
		if isone {
			v = 1
		}
		_ = stor
		err = UpdateStorage(ctx, gid, func(old Storage) Storage { return old.Set(int64(v), bmp) })
		if err != nil {
			ctx.SendChain(message.Text("配置保存失败，原设置保留。"))
			return
		}
		ctx.SendChain(message.Text("成功设置为", isone))
	}
}
