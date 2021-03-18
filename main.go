package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"

	_ "bot/manager"
	_ "bot/music"
	setutime "bot/setutime"
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zero][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)

	setutime.CACHEGROUP = 868047498                       // 涩图缓冲群，必须修改
	setutime.PoolList = []string{"涩图", "二次元", "风景", "车万"} // 涩图类型，可以不修改
}

func main() {
	fmt.Printf(`
====================[ZeroBot-Plugin]====================
* OneBot + ZeroBot + Golang
* Copyright © 2018-2020 Kanri, All Rights Reserved
* Project: https://github.com/Yiwen-Chan/ZeroBot-Plugin
========================================================
`)
	zero.Run(zero.Config{
		NickName:      []string{"bot"},
		CommandPrefix: "/",
		SuperUsers:    []string{"825111790", "213864964"}, // 必须修改，否则无权限
		Driver: []zero.Driver{
			driver.NewWebSocketClient("127.0.0.1", "6700", ""),
		},
	})
	select {}
}
