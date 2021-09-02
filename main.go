package main

import (
	"fmt"
	"os"
	"strings"

	// 注：以下插件均可通过前面加 // 注释，注释后停用并不加载插件
	// 下列插件可与 wdvxdr1123/ZeroBot v1.1.2 以上配合单独使用
	// 词库类
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_atri" // ATRI词库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_chat" // 基础词库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_qingyunke"   //青云客

	// 实用类
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_github"  // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_manager" // 群管
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_runcode" // 在线运行代码

	// 娱乐类
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_ai_false"  // 服务器监控
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_hs"        // 炉石
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_minecraft" // MCSManager
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_music"     // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_omikuji"   // 浅草寺求签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_shindan"   // 测定
        _ "github.com/FloatTech/ZeroBot-Plugin/plugin_personal" //个人私货

	// b站相关
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili" // 查询b站用户信息
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_diana"    // 嘉心糖发病

	// 二次元图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_acgimage"     // 随机图片与AI点评
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_aiwife"       // 随机老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_image_finder" // 关键字搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_lolicon"      // lolicon 随机图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_saucenao"     // 以图搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_setutime"     // 来份涩图

	// 以下为内置依赖，勿动
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

var (
	contents = []string{
		"* OneBot + ZeroBot + Golang ",
		"* Version 1.1.0 - 2021-08-06 23:36:29 +0800 CST",
		"* Copyright © 2020 - 2021  Kanri, DawnNights, Fumiama, Suika",
		"* Project: https://github.com/FloatTech/ZeroBot-Plugin",
	}
	banner = strings.Join(contents, "\n")
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zero][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)
}

func main() {
	fmt.Print(
		"\n======================[ZeroBot-Plugin]======================",
		"\n", banner, "\n",
		"============================================================\n",
	) // 启动打印
	zero.Run(zero.Config{
		NickName:      []string{"椛椛", "ATRI", "atri", "亚托莉", "アトリ"},
		CommandPrefix: "/",

		// SuperUsers 某些功能需要主人权限，可通过以下两种方式修改
		// []string{}：通过代码写死的方式添加主人账号
		// os.Args[1:]：通过命令行参数的方式添加主人账号
		SuperUsers: append([]string{"1156544355","1846515440","1776620359"}, os.Args[1:]...),

		Driver: []zero.Driver{
			&driver.WSClient{
				// OneBot 正向WS 默认使用 6700 端口
				Url:         "ws://127.0.0.1:6700",
				AccessToken: "",
			},
		},
	})

	// 帮助
	zero.OnFullMatchGroup([]string{"help", "/help", ".help", "菜单", "帮助"}, zero.OnlyToMe).SetBlock(true).SetPriority(999).
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(banner)
		})
	select {}
}
