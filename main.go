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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chat" //

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/sleep_manage" //

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/atri" //

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/manager"   //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/thesaurus" //

	_ "github.com/FloatTech/zbputils/job" //

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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_false"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/aiwife"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/b14"           //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/baidu"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/bilibili"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/book_review"   //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cangtoushi"    //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/char_reverser" //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/choose"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/chouxianghua"  //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/coser"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/cpstory"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/danbooru"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/diana"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/drift_bottle"  //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/emojimix"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/epidemic"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/font"          //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/fortune"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/funny"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/genshin"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/github"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/guessmusic"    //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hs"            //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/hyaku"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/image_finder"  //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/inject"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/jandan"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/juejuezi"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/lolicon"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/midicreate"    //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu"          //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/moyu_calendar" //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/music"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativesetu"    //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nativewife"    //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nbnhhsh"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nihongo"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/novel"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/nsfw"          //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/omikuji"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/qqwife"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/reborn"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/runcode"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/saucenao"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/scale"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/score"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/setutime"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shadiao"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/shindan"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tarot"         //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tiangou"       //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/tracemoe"      //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/translation"   //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/vtb_quotation" //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wangyiyun"     //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/word_count"    //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/wordle"        //
	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ymgal"         //

	// _ "github.com/FloatTech/ZeroBot-Plugin/plugin/wtf"            //

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

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/curse" //

	_ "github.com/FloatTech/ZeroBot-Plugin/plugin/ai_reply" //

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
	adana := flag.String("n", "蔡徐坤", "Set default nickname.")
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
	sus = append(sus, 2574896927)
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
		NickName:      append([]string{*adana}, "蔡徐坤哥哥"),
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
