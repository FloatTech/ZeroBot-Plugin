package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	// 注：以下插件均可通过前面加 // 注释，注释后停用并不加载插件
	// 下列插件可与 wdvxdr1123/ZeroBot v1.1.2 以上配合单独使用

	// 插件控制
	//_ "github.com/FloatTech/ZeroBot-Plugin/control/web" // web 后端控制

	// 词库类
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_atri"      // ATRI词库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_chat"      // 基础词库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_qingyunke" // 青云客

	// 实用类
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_github"  // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_manager" // 群管
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_nbnhhsh" // 拼音首字母缩写释义工具
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_runcode" // 在线运行代码

	// 娱乐类
	_ "github.com/FloatTech/ZeroBot-Plugin-Gif"              // 制图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_ai_false"  // 服务器监控
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_choose"    // 选择困难症帮手
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_fortune"   // 运势
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_hs"        // 炉石
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_minecraft" // MCSManager
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_music"     // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_omikuji"   // 浅草寺求签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_reborn"    // 投胎
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_shindan"   // 测定

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
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	//_ "github.com/FloatTech/ZeroBot-Plugin/dyloader"
)

var (
	contents = []string{
		"* OneBot + ZeroBot + Golang",
		"* Version 1.1.6 - 2021-10-09 12:50:23 +0800 CST",
		"* Copyright © 2020 - 2021 Kanri, DawnNights, Fumiama, Suika",
		"* Project: https://github.com/FloatTech/ZeroBot-Plugin",
	}
	banner = strings.Join(contents, "\n")
	token  *string
	url    *string
)

func init() {
	// 解析命令行参数
	d := flag.Bool("d", false, "Enable debug level log and higher.")
	w := flag.Bool("w", false, "Enable warning level log and higher.")
	h := flag.Bool("h", false, "Display this help.")
	// 直接写死 AccessToken 时，请更改下面第二个参数
	token = flag.String("t", "", "Set AccessToken of WSClient.")
	// 直接写死 URL 时，请更改下面第二个参数
	url = flag.String("u", "ws://127.0.0.1:6700", "Set Url of WSClient.")
	flag.Parse()
	if *h {
		printBanner()
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	} else {
		if *d && !*w {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if *w {
			logrus.SetLevel(logrus.WarnLevel)
		}
	}
}

func printBanner() {
	fmt.Print(
		"\n======================[ZeroBot-Plugin]======================",
		"\n", banner, "\n",
		"============================================================\n",
	)
}

func main() {
	printBanner()
	zero.Run(zero.Config{
		NickName:      []string{"椛椛", "ATRI", "atri", "亚托莉", "アトリ"},
		CommandPrefix: "/",

		// SuperUsers 某些功能需要主人权限，可通过以下两种方式修改
		// []string{}：通过代码写死的方式添加主人账号
		// flag.Args()：通过命令行参数的方式添加主人账号
		SuperUsers: append([]string{"12345678", "87654321"}, flag.Args()...),

		Driver: []zero.Driver{
			&driver.WSClient{
				// OneBot 正向WS 默认使用 6700 端口
				Url:         *url,
				AccessToken: *token,
			},
		},
	})

	// 帮助
	zero.OnFullMatchGroup([]string{"/help", ".help", "菜单"}, zero.OnlyToMe).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			ctx.Send(banner)
		})
	select {}
}
