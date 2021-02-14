package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"

	_ "bot/manager"
	_ "bot/music"
	_ "bot/setutime"
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zero][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)
}

func main() {
	fmt.Printf(`
====================[ZeroBot-Plugin]====================
* OneBot + ZeroBot + Golang
* Copyright © 2018-2020 Kanri, All Rights Reserved
* Project: https://github.com/Yiwen-Chan/ZeroBot-Plugin
========================================================
`)
	zero.Run(zero.Option{
		Host:          "127.0.0.1",
		Port:          "6700",
		AccessToken:   "",
		NickName:      []string{"椛椛"},
		CommandPrefix: "/",
		SuperUsers:    []string{"825111790"},
	})
	select {}
}
