package job

import (
	"unsafe"

	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type matcherinstance struct {
	m *zero.Matcher
}

func getmatcher(m control.Matcher) *zero.Matcher {
	return (*matcherinstance)(unsafe.Pointer(&m)).m
}
