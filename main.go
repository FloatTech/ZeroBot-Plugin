package main

import (
	"fmt"

	// 注：以下插件均可通过前面加 // 注释，注释后停用并不加载插件
	// 下列插件可与 wdvxdr1123/ZeroBot v1.1.2 以上配合单独使用
	// 词库类
	_ "github.com/FloatTech/ZeroBot-Plugin/atri" // ATRI词库
	_ "github.com/FloatTech/ZeroBot-Plugin/chat" // 基础词库

	// 实用类
	_ "github.com/FloatTech/ZeroBot-Plugin/github"  // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/manager" // 群管
	_ "github.com/FloatTech/ZeroBot-Plugin/runcode" // 在线运行代码

	// 娱乐类
	_ "github.com/FloatTech/ZeroBot-Plugin/music" // 点歌
	//_ "github.com/Yiwen-Chan/ZeroBot-Plugin/randimg" //简易随机图片
	_ "github.com/FloatTech/ZeroBot-Plugin/setutime" // 涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/shindan"  // 测定

	// 以下为内置依赖，勿动
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"

	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
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
* Version 1.0.3 - 2021-05-02 18:50:40.5489203 +0800 CST
* Copyright © 2021 Kanri, DawnNights, Fumiama
* Project: https://github.com/FloatTech/ZeroBot-Plugin
========================================================
`) // 启动打印
	zero.Run(zero.Config{
		NickName:      []string{"椛椛", "ATRI", "atri", "亚托莉", "アトリ"},
		CommandPrefix: "/",
		SuperUsers:    []string{"213864964"}, // 必须修改，否则无权限
		Driver: []zero.Driver{
			&driver.WSClient{
				Url:         "ws://127.0.0.1:6700",
				AccessToken: "",
			},
		},
	})
	// 帮助
	zero.OnFullMatchGroup([]string{"help", "/help", ".help", "菜单", "帮助"}, zero.OnlyToMe).SetBlock(true).SetPriority(999).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"* OneBot + ZeroBot + Golang ", "\n",
				"* Version 1.0.3 - 2021-05-02 18:50:40.5489203 +0800 CST", "\n",
				"* Copyright © 2021 Kanri, DawnNights, Fumiama ", "\n",
				"* Project: https://github.com/FloatTech/ZeroBot-Plugin",
			))
		})
	select {}
}
