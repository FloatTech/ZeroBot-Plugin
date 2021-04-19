package main

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"

	_ "github.com/Yiwen-Chan/ZeroBot-Plugin/atri"
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
	fmt.Print(`
====================[ZeroBot-Plugin]====================
* OneBot + ZeroBot + Golang
* Version 1.0.1 - 2021-04-20 02:38:38.4181345 +0800 CST
* Copyright © 2021 Kanri, DawnNights, All Rights Reserved
* Project: https://github.com/Yiwen-Chan/ZeroBot-Plugin
========================================================
`) // 启动打印
	zero.Run(zero.Config{
		NickName:      []string{"椛椛", "ATRI", "atri", "亚托莉", "アトリ"},
		CommandPrefix: "/",
		SuperUsers:    []string{"825111790", "213864964"}, // 必须修改，否则无权限
		Driver: []zero.Driver{
			driver.NewWebSocketClient("127.0.0.1", "6700", ""),
		},
	})
	// 帮助
	zero.OnFullMatchGroup([]string{"help", "/help", ".help", "菜单", "帮助"}, zero.OnlyToMe).SetBlock(true).SetPriority(999).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* OneBot + ZeroBot + Golang ", "\n",
				"* Version 1.0.1 - 2021-04-20 02:38:38.4181345 +0800 CST", "\n",
				"* Copyright © 2021 Kanri, DawnNights, All Rights Reserved ", "\n",
				"* Project: https://github.com/Yiwen-Chan/ZeroBot-Plugin",
			))
		})

	select {}
}
