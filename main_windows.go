package main

import (
	"bytes"
	"golang.org/x/sys/windows"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
	stdin := windows.Handle(os.Stdin.Fd())

	var mode uint32
	err := windows.GetConsoleMode(stdin, &mode)
	if err != nil {
		panic(err)
	}

	mode &^= windows.ENABLE_QUICK_EDIT_MODE // 禁用快速编辑模式
	mode |= windows.ENABLE_EXTENDED_FLAGS   // 启用扩展标志

	mode &^= windows.ENABLE_MOUSE_INPUT    // 禁用鼠标输入
	mode |= windows.ENABLE_PROCESSED_INPUT // 启用控制输入

	mode &^= windows.ENABLE_INSERT_MODE                           // 禁用插入模式
	mode |= windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT // 启用输入回显&逐行输入

	mode &^= windows.ENABLE_WINDOW_INPUT           // 禁用窗口输入
	mode &^= windows.ENABLE_VIRTUAL_TERMINAL_INPUT // 禁用虚拟终端输入

	err = windows.SetConsoleMode(stdin, mode)
	if err != nil {
		panic(err)
	}

	stdout := windows.Handle(os.Stdout.Fd())
	err = windows.GetConsoleMode(stdout, &mode)
	if err != nil {
		panic(err)
	}

	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING // 启用虚拟终端处理
	mode |= windows.ENABLE_PROCESSED_OUTPUT            // 启用处理后的输出

	err = windows.SetConsoleMode(stdout, mode)
	if err != nil {
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
