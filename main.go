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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/antiabuse" // 违禁词

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chat" // 基础词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/sleepmanage" // 统计睡眠时间

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/atri" // ATRI词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager" // 群管

	_ "github.com/FloatTech/zbputils/job" // 定时指令触发器

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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ahsai"            // ahsai tts
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aifalse"          // 服务器监控
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aipaint"          // ai绘图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"           // 随机老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/alipayvoice"      // 支付宝到账语音
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/autowithdraw"     // 触发者撤回时也自动撤回
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baidu"            // 百度一下
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baiduaudit"       // 百度内容审核
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base16384"        // base16384加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/base64gua"        // base64卦加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baseamasiro"      // base天城文加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"         // b站相关
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bookreview"       // 哀伤雪刃吧推书记录
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cangtoushi"       // 藏头诗
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chess"            // 国际象棋
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"           // 选择困难症帮手
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua"     // 说抽象话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chrev"            // 英文字符翻转
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"            // 三次元小姐姐
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cpstory"          // cp短打
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dailynews"        // 今日早报
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"         // DeepDanbooru二次元图标签识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/diana"            // 嘉心糖发病
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dish"             // 程序员做饭指南
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drawlots"         // 多功能抽签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/dress"            // 女装
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/driftbottle"      // 漂流瓶
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/emojimix"         // 合成emoji
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/event"            // 好友申请群聊邀请事件处理
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"             // 渲染任意文字到图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/fortune"          // 运势
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/funny"            // 笑话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/genshin"          // 原神抽卡
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"              // 制图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"           // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/guessmusic"       // 猜歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/heisi"            // 黑丝
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hitokoto"         // 一言
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hs"               // 炉石
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hyaku"            // 百人一首
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/imgfinder"        // 关键字搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/inject"           // 注入指令
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jandan"           // 煎蛋网无聊图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jiami"            // 兽语加密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jptingroom"       // 日语听力学习材料
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/juejuezi"         // 绝绝子生成器
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/kfccrazythursday" // 疯狂星期四
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"          // lolicon 随机图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/magicprompt"      // magicprompt吟唱提示
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/mcfish"           // 钓鱼模拟器
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"       // 简易midi音乐制作
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moegoe"           // 日韩 VITS 模型拟声
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"             // 摸鱼
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyucalendar"     // 摸鱼人日历
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"            // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativesetu"       // 本地涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"          // 拼音首字母缩写释义工具
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nihongo"          // 日语语法学习
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/novel"            // 铅笔小说网搜索
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nsfw"             // nsfw图片识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nwife"            // 本地老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/omikuji"          // 浅草寺求签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qqwife"           // 一群一天一夫一妻制群老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/quan"             // QQ权重查询
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qzone"            // qq空间表白墙
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/realcugan"        // realcugan清晰术
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/reborn"           // 投胎
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"          // 在线运行代码
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"         // 以图搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/scale"            // 叔叔的AI二次元图片放大
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/score"            // 分数
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"         // 来份涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shadiao"          // 沙雕app
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shindan"          // 测定
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/steam"            // steam相关
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"            // 抽塔罗牌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"          // 舔狗日记
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"         // 搜番
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"      // 翻译
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/vitsnyaru"        // vits猫雷
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/vtbmusic"         // vtb点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/vtbquotation"     // vtb语录
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wallet"           // 钱包
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wangyiyun"        // 网易云音乐热评
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wantquotes"       // 据意查句
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/warframeapi"      // warframeAPI插件
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenben"           // 文本指令大全
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenxinvilg"       // 百度文心AI画图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wife"             // 抽老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wordcount"        // 聊天热词
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wordle"           // 猜单词
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ygo"              // 游戏王相关插件
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"            // 月幕galgame

	// _ "github.com/FloatTech/ZeroBot-Plugin/plugin/wtf"           // 鬼东西

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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/curse" // 骂人

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply" // 人工智能回复

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus" // 词典匹配回复

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/breakrepeat" // 打断复读

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
	adana := flag.String("n", "青喵", "Set default nickname.")
	prefix := flag.String("p", "/", "Set command prefix.")
	runcfg := flag.String("c", "", "Run from config file.")
	save := flag.String("s", "", "Save default config to file and exit.")
	late := flag.Uint("l", 233, "Response latency (ms).")
	rsz := flag.Uint("r", 4096, "Receiving buffer ring size.")
	maxpt := flag.Uint("x", 4, "Max process time (min).")
	markmsg := flag.Bool("m", false, "Don't mark message as read automatically")

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
	// sus = append(sus, 2243363336)
	// sus = append(sus, 928003739)

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
