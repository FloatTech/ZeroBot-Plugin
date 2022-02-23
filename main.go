package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	// ---------以下插件均可通过前面加 // 注释，注释后停用并不加载插件--------- //
	// ----------------------插件优先级按顺序从高到低---------------------- //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	// ----------------------------高优先级区---------------------------- //
	// vvvvvvvvvvvvvvvvvvvvvvvvvvvv高优先级区vvvvvvvvvvvvvvvvvvvvvvvvvvvv //
	//               vvvvvvvvvvvvvv高优先级区vvvvvvvvvvvvvv               //
	//                      vvvvvvv高优先级区vvvvvvv                      //
	//                          vvvvvvvvvvvvvv                          //
	//                               vvvv                               //

	// webctrl "github.com/FloatTech/zbputils/control/web"           // web 后端控制

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_chat" // 基础词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_sleep_manage" // 统计睡眠时间

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_atri" // ATRI词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_manager" // 群管

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_thesaurus" // 词典匹配回复

	//                               ^^^^                               //
	//                          ^^^^^^^^^^^^^^                          //
	//                      ^^^^^^^高优先级区^^^^^^^                      //
	//               ^^^^^^^^^^^^^^高优先级区^^^^^^^^^^^^^^               //
	// ^^^^^^^^^^^^^^^^^^^^^^^^^^^^高优先级区^^^^^^^^^^^^^^^^^^^^^^^^^^^^ //
	// ----------------------------高优先级区---------------------------- //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	// ----------------------------中优先级区---------------------------- //
	// vvvvvvvvvvvvvvvvvvvvvvvvvvvv中优先级区vvvvvvvvvvvvvvvvvvvvvvvvvvvv //
	//               vvvvvvvvvvvvvv中优先级区vvvvvvvvvvvvvv               //
	//                      vvvvvvv中优先级区vvvvvvv                      //
	//                          vvvvvvvvvvvvvv                          //
	//                               vvvv                               //

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_acgimage"       // 随机图片与AI点评
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_ai_false"       // 服务器监控
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_aiwife"         // 随机老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_b14"            // base16384加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili"       // 查询b站用户信息
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili_parse" // b站视频链接解析
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_book_review"    // 哀伤雪刃吧推书记录
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_cangtoushi"     // 藏头诗
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_choose"         // 选择困难症帮手
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_chouxianghua"   // 说抽象话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_coser"          // 三次元小姐姐
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_cpstory"        // cp短打
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_danbooru"       // DeepDanbooru二次元图标签识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_diana"          // 嘉心糖发病
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_drift_bottle"   // 漂流瓶
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_emojimix"       // 合成emoji
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_fortune"        // 运势
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_funny"          // 笑话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_gif"            // 制图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_github"         // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_hs"             // 炉石
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_image_finder"   // 关键字搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_jandan"         // 煎蛋网无聊图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_juejuezi"       // 绝绝子生成器
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_lolicon"        // lolicon 随机图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_minecraft"      // MCSManager
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_moyu"           // 摸鱼
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_moyu_calendar"  // 摸鱼人日历
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_music"          // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_nativesetu"     // 本地涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_nativewife"     // 本地老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_nbnhhsh"        // 拼音首字母缩写释义工具
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_novel"          // 铅笔小说网搜索
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_nsfw"           // nsfw图片识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_omikuji"        // 浅草寺求签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_reborn"         // 投胎
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_runcode"        // 在线运行代码
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_saucenao"       // 以图搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_scale"          // 叔叔的AI二次元图片放大
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_score"          // 分数
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_setutime"       // 来份涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_shadiao"        // 沙雕app
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_shindan"        // 测定
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_tracemoe"       // 搜番
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_translation"    // 翻译
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation"  // vtb语录
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_wangyiyun"      // 网易云音乐热评
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_wordle"         // 猜单词

	// _ "github.com/FloatTech/ZeroBot-Plugin/plugin_wtf"            // 鬼东西
	// _ "github.com/FloatTech/ZeroBot-Plugin/plugin_bilibili_push"  // b站推送

	//                               ^^^^                               //
	//                          ^^^^^^^^^^^^^^                          //
	//                      ^^^^^^^中优先级区^^^^^^^                      //
	//               ^^^^^^^^^^^^^^中优先级区^^^^^^^^^^^^^^               //
	// ^^^^^^^^^^^^^^^^^^^^^^^^^^^^中优先级区^^^^^^^^^^^^^^^^^^^^^^^^^^^^ //
	// ----------------------------中优先级区---------------------------- //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	// ----------------------------低优先级区---------------------------- //
	// vvvvvvvvvvvvvvvvvvvvvvvvvvvv低优先级区vvvvvvvvvvvvvvvvvvvvvvvvvvvv //
	//               vvvvvvvvvvvvvv低优先级区vvvvvvvvvvvvvv               //
	//                      vvvvvvv低优先级区vvvvvvv                      //
	//                          vvvvvvvvvvvvvv                          //
	//                               vvvv                               //

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_curse" // 骂人

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin_ai_reply" // 人工智能回复

	//                               ^^^^                               //
	//                          ^^^^^^^^^^^^^^                          //
	//                      ^^^^^^^低优先级区^^^^^^^                      //
	//               ^^^^^^^^^^^^^^低优先级区^^^^^^^^^^^^^^               //
	// ^^^^^^^^^^^^^^^^^^^^^^^^^^^^低优先级区^^^^^^^^^^^^^^^^^^^^^^^^^^^^ //
	// ----------------------------低优先级区---------------------------- //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	//                                                                  //
	// -----------------------以下为内置依赖，勿动------------------------ //
	"github.com/FloatTech/zbputils/control/order"
	"github.com/fumiama/go-registry"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

var (
	contents = []string{
		"* OneBot + ZeroBot + Golang",
		"* Version 1.3.0 - 2022-02-09 14:31:34 +0800 CST",
		"* Copyright © 2020 - 2021 FloatTech. All Rights Reserved.",
		"* Project: https://github.com/FloatTech/ZeroBot-Plugin",
	}
	nicks  = []string{"ATRI", "atri", "亚托莉", "アトリ"}
	banner = strings.Join(contents, "\n")
	token  *string
	url    *string
	adana  *string
	prefix *string
	reg    = registry.NewRegReader("reilia.fumiama.top:32664", "fumiama")
)

func init() {
	// 解析命令行参数
	d := flag.Bool("d", false, "Enable debug level log and higher.")
	w := flag.Bool("w", false, "Enable warning level log and higher.")
	h := flag.Bool("h", false, "Display this help.")
	// 解析命令行参数，输入 `-g 监听地址:端口` 指定 gui 访问地址，默认 127.0.0.1:3000
	// g := flag.String("g", "127.0.0.1:3000", "Set web gui listening address.")

	// 直接写死 AccessToken 时，请更改下面第二个参数
	token = flag.String("t", "", "Set AccessToken of WSClient.")
	// 直接写死 URL 时，请更改下面第二个参数
	url = flag.String("u", "ws://127.0.0.1:6700", "Set Url of WSClient.")
	// 默认昵称
	adana = flag.String("n", "椛椛", "Set default nickname.")
	prefix = flag.String("p", "/", "Set command prefix.")

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

	// 启用 gui
	// webctrl.InitGui(*g)
}

func printBanner() {
	fmt.Print(
		"\n======================[ZeroBot-Plugin]======================",
		"\n", banner, "\n",
		"----------------------[ZeroBot-公告栏]----------------------",
		"\n", getKanban(), "\n",
		"============================================================\n",
	)
}

func getKanban() string {
	err := reg.Connect()
	if err != nil {
		return err.Error()
	}
	defer reg.Close()
	text, err := reg.Get("ZeroBot-Plugin/kanban")
	if err != nil {
		return err.Error()
	}
	return text
}

func main() {
	order.Wait()
	printBanner()
	rand.Seed(time.Now().UnixNano()) // 全局 seed，其他插件无需再 seed
	// 帮助
	zero.OnFullMatchGroup([]string{"/help", ".help", "菜单"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(banner, "\n可发送\"/服务列表\"查看 bot 功能"))
		})
	zero.OnFullMatch("查看zbp公告", zero.OnlyToMe, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(getKanban()))
		})
	zero.RunAndBlock(
		zero.Config{
			NickName:      append([]string{*adana}, nicks...),
			CommandPrefix: *prefix,
			// SuperUsers 某些功能需要主人权限，可通过以下两种方式修改
			// SuperUsers: []string{"12345678", "87654321"}, // 通过代码写死的方式添加主人账号
			SuperUsers: flag.Args(), // 通过命令行参数的方式添加主人账号
			Driver:     []zero.Driver{driver.NewWebSocketClient(*url, *token)},
		},
	)
}
