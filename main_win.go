//go:build windows
// +build windows

package main

import (
	"bytes"
	"strings"

	"github.com/sirupsen/logrus"
)

func init() {
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

	buf.WriteString(getLogLevelColorCode(entry.Level))

	buf.WriteByte('[')
	buf.WriteString(strings.ToUpper(entry.Level.String()))
	buf.WriteString("] ")
	buf.WriteString(entry.Message)
	buf.WriteString(" \n")

	buf.WriteString(colorReset)

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
