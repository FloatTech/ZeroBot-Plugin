// Package ahsai AH Soft フリーテキスト音声合成 demo API
package ahsai

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	ahsaitts "github.com/fumiama/ahsai"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	namelist = [...]string{"伊織弓鶴", "紲星あかり", "結月ゆかり", "京町セイカ", "東北きりたん", "東北イタコ", "ついなちゃん標準語", "ついなちゃん関西弁", "音街ウナ", "琴葉茜", "吉田くん", "民安ともえ", "桜乃そら", "月読アイ", "琴葉葵", "東北ずん子", "月読ショウタ", "水奈瀬コウ"}
	namesort = func() []string {
		nl := namelist[:]
		sort.Strings(nl)
		return nl
	}()
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "フリーテキスト音声合成",
		Help:              "- 使[伊織弓鶴|紲星あかり|結月ゆかり|京町セイカ|東北きりたん|東北イタコ|ついなちゃん標準語|ついなちゃん関西弁|音街ウナ|琴葉茜|吉田くん|民安ともえ|桜乃そら|月読アイ|琴葉葵|東北ずん子|月読ショウタ|水奈瀬コウ]说(日语)",
		PrivateDataFolder: "ahsai",
	})
	cachePath := engine.DataFolder() + "cache/"
	_ = os.RemoveAll(cachePath)
	_ = os.MkdirAll(cachePath, 0755)
	engine.OnRegex("^使(.{0,10})说([A-Za-z\\s\\d\u3005\u3040-\u30ff\u4e00-\u9fff\uff11-\uff19\uff21-\uff3a\uff41-\uff5a\uff66-\uff9d\\pP]+)$", selectName).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("少女祈祷中..."))
		uid := ctx.Event.UserID
		today := time.Now().Format("20060102150405")
		ahsaiFile := cachePath + strconv.FormatInt(uid, 10) + today + "ahsai.wav"
		s := ahsaitts.NewSpeaker()
		err := s.SetName(ctx.State["ahsainame"].(string))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		u, err := s.Speak(ctx.State["ahsaitext"].(string))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		err = ahsaitts.SaveOggToFile(u, ahsaiFile)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Record("file:///" + file.BOTPATH + "/" + ahsaiFile))
	})
}

func selectName(ctx *zero.Ctx) bool {
	regexMatched := ctx.State["regex_matched"].([]string)
	ctx.State["ahsaitext"] = regexMatched[2]
	name := regexMatched[1]
	index := sort.SearchStrings(namesort, name)
	if index < len(namelist) && namesort[index] == name {
		ctx.State["ahsainame"] = name
		return true
	}
	speaktext := ""
	for i, v := range namelist {
		speaktext += fmt.Sprintf("%d. %s\n", i, v)
	}
	ctx.SendChain(message.Text("输入的音源为空, 请输入音源序号\n", speaktext))
	next, cancel := zero.NewFutureEvent("message", 999, false, ctx.CheckSession(), zero.RegexRule(`\d{0,2}`)).Repeat()
	defer cancel()
	for {
		select {
		case <-time.After(time.Second * 10):
			ctx.State["ahsainame"] = namelist[rand.Intn(len(namelist))]
			ctx.SendChain(message.Text("时间太久啦！", zero.BotConfig.NickName[0], "帮你选择", ctx.State["ahsainame"]))
			return true
		case c := <-next:
			msg := c.Event.Message.ExtractPlainText()
			num, _ := strconv.Atoi(msg)
			if num < 0 || num >= len(namelist) {
				ctx.SendChain(message.Text("序号非法!"))
				continue
			}
			ctx.State["ahsainame"] = namelist[num]
			return true
		}
	}
}
