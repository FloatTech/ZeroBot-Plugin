//go:build windows
// +build windows

package main

import (
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {
	logrus.SetFormatter(&easy.Formatter{
		LogFormat: "[%lvl%] %msg%\n",
	})
}
