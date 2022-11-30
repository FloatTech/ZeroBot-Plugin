// Package main ZeroBot-Plugin main file
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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/private" // 自建词库

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager" // 群管

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus" // 词典匹配回复

	_ "github.com/FloatTech/zbputils/job" // 定时指令触发器

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wallet" // 钱包系统
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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/baidufanyi" // 百度翻译
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/gamesystem" // 游戏系统
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/nekoneko"   // nekoneko
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/ygocombo"   // 游戏王combo记录器
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/_personal/ygoscore"   // 游戏王combo记录器
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aipaint"              // ai绘图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/alipayvoice"          // 支付宝到账语音
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"             // b站相关
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"             // DeepDanbooru二次元图标签识别
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/epidemic"             // 城市疫情查询
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/gif"                  // 制图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"               // 搜索GitHub仓库
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/guessmusic"           // 猜歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/image_finder"         // 关键字搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jikipedia"            // 小鸡词典
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moegoe"               // 日韩 VITS 模型拟声
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"                // 点歌
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"              // 拼音首字母缩写释义工具
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qqwife"               // 一群一天一夫一妻制群老婆
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"              // 在线运行代码
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"             // 以图搜图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"             // 来份涩图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"             // 搜番
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wenxinAI"             // AI画图
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wordle"               // 猜单词
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ygo"                  // 游戏王相关插件
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"                // 月幕galgame

	//                              ^^^^                               //
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
	"github.com/FloatTech/floatbox/process"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"github.com/wdvxdr1123/ZeroBot/message"
	// -----------------------以上为内置依赖，勿动------------------------ //
)

type zbpcfg struct {
	Z zero.Config        `json:"zero"`
	W []*driver.WSClient `json:"ws"`
}

var config zbpcfg

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
	adana := flag.String("n", "椛椛", "Set default nickname.")
	prefix := flag.String("p", "/", "Set command prefix.")
	runcfg := flag.String("c", "", "Run from config file.")
	save := flag.String("s", "", "Save default config to file and exit.")
	late := flag.Uint("l", 233, "Response latency (ms).")
	rsz := flag.Uint("r", 4096, "Receiving buffer ring size.")
	maxpt := flag.Uint("x", 4, "Max process time (min).")

	flag.Parse()

	if *h {
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
		NickName:       append([]string{*adana}, "ATRI", "atri", "亚托莉", "アトリ"),
		CommandPrefix:  *prefix,
		SuperUsers:     sus,
		RingLen:        *rsz,
		Latency:        time.Duration(*late) * time.Millisecond,
		MaxProcessTime: time.Duration(*maxpt) * time.Minute,
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
	rand.Seed(time.Now().UnixNano()) // 全局 seed，其他插件无需再 seed
	// 帮助
	zero.OnFullMatchGroup([]string{"/help", ".help", "菜单"}, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(kanban.Banner,
				"\n————————————\n基础指令:\n",
				"- 查看zbp公告\n- /服务列表\n- /用法[插件名称]\n- /启用[插件名称]\n- /禁用[插件名称]",
			))
		})
	zero.OnFullMatch("查看zbp公告", zero.OnlyToMe, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(kanban.Kanban()))
		})
	zero.RunAndBlock(&config.Z, process.GlobalInitMutex.Unlock)
}
