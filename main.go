// Package main ZeroBot-Plugin main file
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/FloatTech/ZeroBot-Plugin/console" // 更改控制台属性

	"github.com/FloatTech/ZeroBot-Plugin/kanban" // 打印 banner

	// 支付宝到账语音

	//"github.com/FloatTech/ZeroBot-Plugin/plugin/antiabuse"    // 违禁词
	// 触发者撤回时也自动撤回
	// 百度内容审核
	// base16384加解密
	// base64卦加解密

	// DeepDanbooru二次元图标签识别

	"github.com/FloatTech/ZeroBot-Plugin/plugin/fortune"
	// 本地老婆
	// 百度文心AI画图
	//"github.com/FloatTech/ZeroBot-Plugin/plugin/chat" // 基础词库

	//"github.com/FloatTech/ZeroBot-Plugin/plugin/sleepmanage" // 统计睡眠时间

	//"github.com/FloatTech/ZeroBot-Plugin/plugin/atri" // ATRI词库

	//"github.com/FloatTech/ZeroBot-Plugin/plugin/manager" // 群管

	_ "github.com/FloatTech/zbputils/job" // 定时指令触发器

	// 骂人

	// 人工智能回复

	// 词典匹配回复

	// 打断复读

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/process"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"

	// webctrl "github.com/FloatTech/zbputils/control/web"

	"github.com/FloatTech/ZeroBot-Plugin/kanban/banner"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

type zbpcfg struct {
	Z zero.Config        `json:"zero"`
	W []*driver.WSClient `json:"ws"`
	S []*driver.WSServer `json:"wss"`
}

var config zbpcfg

func init() {
	sus := make([]int64, 0, 16)
	// 解析命令行参数
	d := flag.Bool("d", false, "Enable debug level log and higher.")
	w := flag.Bool("w", false, "Enable warning level log and higher.")
	h := flag.Bool("h", false, "Display this help.")
	// g := flag.String("g", "127.0.0.1:3000", "Set webui url.")
	// 直接写死 AccessToken 时，请更改下面第二个参数
	token := flag.String("t", "", "Set AccessToken of WSClient.")
	// 直接写死 URL 时，请更改下面第二个参数
	url := flag.String("u", "ws://127.0.0.1:6700", "Set Url of WSClient.")
	// 默认昵称
	adana := flag.String("n", "椛椛", "Set default nickname.")
	prefix := flag.String("p", "/", "Set command prefix.")
	runcfg := flag.String("c", "", "Run from config file.")
	save := flag.String("s", "", "Save default config to file and exit.")
	late := flag.Uint("l", 233, "Response latency (ms).")
	rsz := flag.Uint("r", 4096, "Receiving buffer ring size.")
	maxpt := flag.Uint("x", 4, "Max process time (min).")
	markmsg := flag.Bool("m", false, "Don't mark message as read automatically")
	flag.BoolVar(&file.SkipOriginal, "mirror", false, "Use mirrored lazy data at first")

	flag.Parse()

	if *h {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if *d && !*w {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if *w {
		logrus.SetLevel(logrus.WarnLevel)
	}

	for _, s := range flag.Args() {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			continue
		}
		sus = append(sus, i)
	}

	// 通过代码写死的方式添加主人账号
	// sus = append(sus, 12345678)
	// sus = append(sus, 87654321)

	// 启用 webui
	// go webctrl.RunGui(*g)

	if *runcfg != "" {
		f, err := os.Open(*runcfg)
		if err != nil {
			panic(err)
		}
		config.W = make([]*driver.WSClient, 0, 2)
		err = json.NewDecoder(f).Decode(&config)
		f.Close()
		if err != nil {
			panic(err)
		}
		config.Z.Driver = make([]zero.Driver, len(config.W)+len(config.S))
		for i, w := range config.W {
			config.Z.Driver[i] = w
		}
		for i, s := range config.S {
			config.Z.Driver[i+len(config.W)] = s
		}
		logrus.Infoln("[main] 从", *runcfg, "读取配置文件")
		return
	}
	config.W = []*driver.WSClient{driver.NewWebSocketClient(*url, *token)}
	config.Z = zero.Config{
		NickName:       append([]string{*adana}, "ATRI", "atri", "亚托莉", "アトリ"),
		CommandPrefix:  *prefix,
		SuperUsers:     sus,
		RingLen:        *rsz,
		Latency:        time.Duration(*late) * time.Millisecond,
		MaxProcessTime: time.Duration(*maxpt) * time.Minute,
		MarkMessage:    !*markmsg,
		Driver:         []zero.Driver{config.W[0]},
	}

	if *save != "" {
		f, err := os.Create(*save)
		if err != nil {
			panic(err)
		}
		err = json.NewEncoder(f).Encode(&config)
		f.Close()
		if err != nil {
			panic(err)
		}
		logrus.Infoln("[main] 配置文件已保存到", *save)
		os.Exit(0)
	}
}

func main() {
	if !strings.Contains(runtime.Version(), "go1.2") { // go1.20之前版本需要全局 seed，其他插件无需再 seed
		rand.Seed(time.Now().UnixNano()) //nolint: staticcheck
	}

	// 初始化插件
	initializePlugins()

	// 帮助
	zero.OnFullMatchGroup([]string{"help", "/help", ".help", "菜单"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(banner.Banner, "\n管理发送\"/服务列表\"查看 bot 功能\n发送\"/用法name\"查看功能用法"))
		})

	zero.OnFullMatch("查看zbp公告", zero.OnlyToMe, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(strings.ReplaceAll(kanban.Kanban(), "\t", "")))
		})

	zero.RunAndBlock(&config.Z, process.GlobalInitMutex.Unlock)
}

// initializeHighPriorityPlugins 初始化所有高优先级插件
func initializeHighPriorityPlugins() {
	// 注意：这里的函数名和实际插件可能不同，需要你根据实际情况调整
	// manager.Initialize()     // 群管
	// antiabuse.Initialize()   // 违禁词
	// chat.Initialize()        // 基础词库
	// sleepmanage.Initialize() // 统计睡眠时间
	// atri.Initialize()        // ATRI词库
	// 继续添加其他高优先级插件
}

// initializeMediumPriorityPlugins 初始化所有中优先级插件
func initializeMediumPriorityPlugins() {
	// 以下是插件初始化调用，确保每个插件都实现了initialize方法
	// ahsai.Initialize()            // ahsai tts
	// aifalse.Initialize()          // 服务器监控
	// aiwife.Initialize()           // 随机老婆
	// alipayvoice.Initialize()      // 支付宝到账语音
	// autowithdraw.Initialize()     // 触发者撤回时也自动撤回
	// baiduaudit.Initialize()       // 百度内容审核
	// base16384.Initialize()        // base16384加解密
	// base64gua.Initialize()        // base64卦加解密
	// baseamasiro.Initialize()      // base天城文加解密
	// bilibili.Initialize()         // b站相关
	// bilibili.Initialize_1()       // b站相关
	// bilibili.Initialize_2()       // b站相关
	// bookreview.Initialize()       // 哀伤雪刃吧推书记录
	// chess.Initialize()            // 国际象棋
	// choose.Initialize()           // 选择困难症帮手
	// chouxianghua.Initialize()     // 说抽象话
	// chrev.Initialize()            // 英文字符翻转
	// chrev.Initialize_2()          // 英文字符翻转
	// coser.Initialize()            // 三次元小姐姐
	// cpstory.Initialize()          // cp短打
	// dailynews.Initialize()        // 今日早报
	// danbooru.Initialize()         // DeepDanbooru二次元图标签识别
	// diana.Initialize()            // 嘉心糖发病
	// dish.Initialize()             // 程序员做饭指南
	// drawlots.Initialize()         // 多功能抽签
	// driftbottle.Initialize()      // 漂流瓶
	// emojimix.Initialize()         // 合成emoji
	// event.Initialize()            // 好友申请群聊邀请事件处理
	// font.Initialize()             // 渲染任意文字到图片
	fortune.Initialize() // 运势
	// funny.Initialize()            // 笑话
	// genshin.Initialize()          // 原神抽卡
	// gif.Initialize()              // 制图
	// github.Initialize()           // 搜索GitHub仓库
	// guessmusic.Initialize()       // 猜歌
	// guessmusic.Initialize_1()     // 猜歌
	// guessmusic.Initialize_2()     // 猜歌
	// hitokoto.Initialize()         // 一言
	// hs.Initialize()               // 炉石
	// hyaku.Initialize()            // 百人一首
	// inject.Initialize()           // 注入指令
	// jandan.Initialize()           // 煎蛋网无聊图
	// jptingroom.Initialize()       // 日语听力学习材料
	// kfccrazythursday.Initialize() // 疯狂星期四
	// lolicon.Initialize()          // lolicon 随机图片
	// lolimi.Initialize()           // 桑帛云 API
	// magicprompt.Initialize()      // magicprompt吟唱提示
	// mcfish.Initialize()           // 钓鱼模拟器
	// mcfish.Initialize_1()         // 钓鱼模拟器
	// mcfish.Initialize_2()         // 钓鱼模拟器
	// mcfish.Initialize_3()         // 钓鱼模拟器
	// mcfish.Initialize_4()         // 钓鱼模拟器
	// midicreate.Initialize()       // 简易midi音乐制作
	// moegoe.Initialize()           // 日韩 VITS 模型拟声
	// moyu.Initialize()             // 摸鱼
	// moyucalendar.Initialize()     // 摸鱼人日历
	// music.Initialize()            // 点歌
	// nativesetu.Initialize()       // 本地涩图
	// nbnhhsh.Initialize()          // 拼音首字母缩写释义工具
	// nihongo.Initialize()          // 日语语法学习
	// novel.Initialize()            // 铅笔小说网搜索
	// nsfw.Initialize()             // nsfw图片识别
	// nwife.Initialize()            // 本地老婆
	// omikuji.Initialize()          // 浅草寺求签
	// poker.Initialize()            // 抽扑克
	// qqwife.Initialize()           // 一群一天一夫一妻制群老婆
	// qqwife.Initialize_1()         // 一群一天一夫一妻制群老婆
	// qqwife.Initialize_2()         // 一群一天一夫一妻制群老婆
	// qzone.Initialize()            // qq空间表白墙
	// realcugan.Initialize()        // realcugan清晰术
	// reborn.Initialize()           // 投胎
	// robbery.Initialize()          // 打劫群友的ATRI币
	// runcode.Initialize()          // 在线运行代码
	// saucenao.Initialize()         // 以图搜图
	// score.Initialize()            // 分数
	// setutime.Initialize()         // 来份涩图
	// shadiao.Initialize()          // 沙雕app
	// shadiao.Initialize_1()        // 沙雕app
	// shadiao.Initialize_2()        // 沙雕app
	// shadiao.Initialize_3()        // 沙雕app
	// shindan.Initialize()          // 测定
	// steam.Initialize()            // steam相关
	// steam.Initialize_2()          // steam相关
	// tarot.Initialize()            // 抽塔罗牌
	// tiangou.Initialize()          // 舔狗日记
	// tracemoe.Initialize()         // 搜番
	// translation.Initialize()      // 翻译
	// vitsnyaru.Initialize()        // vits猫雷
	// wallet.Initialize()           // 钱包
	// wantquotes.Initialize()       // 据意查句
	// warframeapi.Initialize()      // warframeAPI插件
	// wenxinvilg.Initialize()       // 百度文心AI画图
	// wife.Initialize()             // 抽老婆
	// wordcount.Initialize()        // 聊天热词
	// wordle.Initialize()           // 猜单词
	// ygo.Initialize()              // 游戏王相关插件
	// ygo.Initialize_1()            // 游戏王相关插件
	// ymgal.Initialize()            // 月幕galgame
	// yujn.Initialize()             // 遇见API
}

// initializeLowPriorityPlugins 初始化所有低优先级插件
func initializeLowPriorityPLugins() {
	// curse.Initialize()       // 骂人
	// aireply.Initialize()     // 人工智能回复
	// thesaurus.Initialize()   // 词典匹配回复
	// breakrepeat.Initialize() // 打断复读
	// 继续添加其他低优先级插件
}

// initializePlugins 按优先级初始化所有插件
func initializePlugins() {
	fmt.Println("Initializing high priority plugins...")
	initializeHighPriorityPlugins()
	fmt.Println("High priority plugins initialized.")

	fmt.Println("Initializing medium priority plugins...")
	initializeMediumPriorityPlugins()
	fmt.Println("Medium priority plugins initialized.")

	fmt.Println("Initializing low priority plugins...")
	initializeLowPriorityPLugins()
	fmt.Println("Low priority plugins initialized.")
}
