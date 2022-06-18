package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/kanban" // 在最前打印 banner

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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chat" // 基础词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/sleep_manage" // 统计睡眠时间

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/atri" // ATRI词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager" // 群管

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus" // 词典匹配回复

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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false"       // 服务器监控
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"         // 随机老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/b14"            // base16384加解密
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baidu"          // 百度一下
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"       // 查询b站用户信息
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili_parse" // b站视频链接解析
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/book_review"    // 哀伤雪刃吧推书记录
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cangtoushi"     // 藏头诗
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"         // 选择困难症帮手
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua"   // 说抽象话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"          // 三次元小姐姐
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cpstory"        // cp短打
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"       // DeepDanbooru二次元图标签识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/diana"          // 嘉心糖发病
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drift_bottle"   // 漂流瓶
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/emojimix"       // 合成emoji
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/epidemic"       // 城市疫情查询
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"           // 渲染任意文字到图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/fortune"        // 运势
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/funny"          // 笑话
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/genshin"        // 原神抽卡
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"            // 制图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"         // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hs"             // 炉石
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hyaku"          // 百人一首
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/image_finder"   // 关键字搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/inject"         // 注入指令
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jandan"         // 煎蛋网无聊图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/juejuezi"       // 绝绝子生成器
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"        // lolicon 随机图片
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"     // 简易midi音乐制作
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"           // 摸鱼
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu_calendar"  // 摸鱼人日历
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"          // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativesetu"     // 本地涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativewife"     // 本地老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"        // 拼音首字母缩写释义工具
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/novel"          // 铅笔小说网搜索
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nsfw"           // nsfw图片识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/omikuji"        // 浅草寺求签
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qqwife"         // 一群一天一夫一妻制群老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/reborn"         // 投胎
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"        // 在线运行代码
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"       // 以图搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/scale"          // 叔叔的AI二次元图片放大
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/score"          // 分数
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"       // 来份涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shadiao"        // 沙雕app
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shindan"        // 测定
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"          // 抽塔罗牌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"        // 舔狗日记
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"       // 搜番
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"    // 翻译
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/vtb_quotation"  // vtb语录
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wangyiyun"      // 网易云音乐热评
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/word_count"     // 聊天热词
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wordle"         // 猜单词
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"          // 月幕galgame
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/zaobao"         // 早报

	// _ "github.com/FloatTech/ZeroBot-Plugin/plugin/wtf"            // 鬼东西
	// _ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili_push"  // b站推送

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
	"github.com/FloatTech/zbputils/process"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

func init() {
	sus := make([]int64, 0, 16)
	// 解析命令行参数
	d := flag.Bool("d", false, "Enable debug level log and higher.")
	w := flag.Bool("w", false, "Enable warning level log and higher.")
	h := flag.Bool("h", false, "Display this help.")
	// 直接写死 AccessToken 时，请更改下面第二个参数
	token := flag.String("t", "", "Set AccessToken of WSClient.")
	// 直接写死 URL 时，请更改下面第二个参数
	url := flag.String("u", "ws://127.0.0.1:6700", "Set Url of WSClient.")
	// 默认昵称
	adana := flag.String("n", "咕咕咕", "Set default nickname.")
	prefix := flag.String("p", "/", "Set command prefix.")
	runcfg := flag.String("c", "", "Run from config file.")
	save := flag.String("s", "", "Save default config to file and exit.")

	flag.Parse()

	if *h {
		kanban.PrintBanner()
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
		config.Z.Driver = make([]zero.Driver, len(config.W))
		for i, w := range config.W {
			config.Z.Driver[i] = w
		}
		logrus.Infoln("[main] 从", *runcfg, "读取配置文件")
		return
	}

	config.W = []*driver.WSClient{driver.NewWebSocketClient(*url, *token)}
	config.Z = zero.Config{
		NickName:      append([]string{*adana}, "ATRI", "atri", "亚托莉", "アトリ"),
		CommandPrefix: *prefix,
		SuperUsers:    sus,
		Driver:        []zero.Driver{config.W[0]},
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
	rand.Seed(time.Now().UnixNano()) // 全局 seed，其他插件无需再 seed
	// 帮助
	zero.OnFullMatchGroup([]string{"/help", ".help", "菜单"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(kanban.Banner, "\n可发送\"/服务列表\"查看 bot 功能"))
		})
	zero.OnFullMatch("查看zbp公告", zero.OnlyToMe, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(kanban.Kanban()))
		})
	zero.RunAndBlock(config.Z, process.GlobalInitMutex.Unlock)
}
