//go:build windows
// +build windows

package main

import (
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {
	// windows 不支持带颜色的 log，故自定义格式
	logrus.SetFormatter(&easy.Formatter{
		LogFormat: "[%lvl%] %msg%\n",
	})
}
