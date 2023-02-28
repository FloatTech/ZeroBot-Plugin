package main

import (
	"bytes"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

func init() {
	k32 := windows.NewLazySystemDLL("kernel32.dll")
	getstdhandle := k32.NewProc("GetStdHandle")
	magic := -10
	h, _, err := getstdhandle.Call(uintptr(magic)) // STD_INPUT_HANDLE = ((DWORD)-10)
	if int(h) == 0 || int(h) == -1 {
		panic(err)
	}
	magic--
	h, _, err = k32.NewProc("SetConsoleMode").Call(h, uintptr(0x02a7)) // 禁用快速编辑
	if h == 0 {
		panic(err)
	}
	h, _, err = getstdhandle.Call(uintptr(magic)) // STD_OUTPUT_HANDLE = ((DWORD)-11)
	if int(h) == 0 || int(h) == -1 {
		panic(err)
	}
	h, _, err = k32.NewProc("SetConsoleMode").Call(h, uintptr(0x001f)) // 启用VT100
	if h == 0 {
		panic(err)
	}
	// windows 带颜色 log 自定义格式
	logrus.SetFormatter(&LogFormat{})
}

const (
	colorCodePanic = "\x1b[1;31m" // color.Style{color.Bold, color.Red}.String()
	colorCodeFatal = "\x1b[1;31m" // color.Style{color.Bold, color.Red}.String()
	colorCodeError = "\x1b[31m"   // color.Style{color.Red}.String()
	colorCodeWarn  = "\x1b[33m"   // color.Style{color.Yellow}.String()
	colorCodeInfo  = "\x1b[37m"   // color.Style{color.White}.String()
	colorCodeDebug = "\x1b[32m"   // color.Style{color.Green}.String()
	colorCodeTrace = "\x1b[36m"   // color.Style{color.Cyan}.String()
	colorReset     = "\x1b[0m"
)

// LogFormat specialize for zbp
type LogFormat struct{}

// Format implements logrus.Formatter
func (f LogFormat) Format(entry *logrus.Entry) ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.WriteByte('[')
	buf.WriteString(getLogLevelColorCode(entry.Level))
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	buf.WriteString(colorReset)
	buf.WriteString("] ")
	buf.WriteString(entry.Message)
	buf.WriteString(" \n")

	return buf.Bytes(), nil
}

// getLogLevelColorCode 获取日志等级对应色彩code
func getLogLevelColorCode(level logrus.Level) string {
	switch level {
	case logrus.PanicLevel:
		return colorCodePanic
	case logrus.FatalLevel:
		return colorCodeFatal
	case logrus.ErrorLevel:
		return colorCodeError
	case logrus.WarnLevel:
		return colorCodeWarn
	case logrus.InfoLevel:
		return colorCodeInfo
	case logrus.DebugLevel:
		return colorCodeDebug
	case logrus.TraceLevel:
		return colorCodeTrace

	default:
		return colorCodeInfo
	}
}
