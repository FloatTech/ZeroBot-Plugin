package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"

	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/chat"
	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/github"
	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/manager"
	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/music"
	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/runcode"
	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/shindan"
	//_ "github.com/Yiwen-Chan/ZeroBot-Plugin/setutime"
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zero][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)
}

func main() {
	var declare = `====================[ZeroBot-Plugin]====================
* OneBot + ZeroBot + Golang
* Copyright © 2021 Kanri, DawnNights, All Rights Reserved
* Project: https://github.com/Yiwen-Chan/ZeroBot-Plugin
========================================================`
	fmt.Println(declare) // 启动打印
	zero.Run(zero.Config{
		NickName:      []string{"椛椛"},
		CommandPrefix: "/",
		SuperUsers:    []string{"825111790", "213864964"}, // 必须修改，否则无权限
		Driver: []zero.Driver{
			driver.NewWebSocketClient("127.0.0.1", "6700", ""),
		},
	})
	// 帮助
	zero.OnFullMatchGroup([]string{"/help", ".help", "菜单", "帮助"}, zero.OnlyToMe).SetBlock(true).SetPriority(999).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(
				message.Text(declare),
			)
		})
	select {}
}
