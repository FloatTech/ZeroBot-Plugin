package tarot

import (
	"fmt"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitcode.net/shudorcl/zbp-tarot/-/raw/master/"


var tarotData gjson.Result
var reasons = []string{"您抽到的是~\n", "锵锵锵，塔罗牌的预言是~\n", "诶，让我看看您抽到了~\n"}
var position = []string{"正位", "逆位"}
var description = []string{"description", "reverseDescription"}

func init() {
	engine := control.Register("tarot", &control.Options{
		DisableOnDefault: false,
		Help: "塔罗牌\n" +
			"- 抽塔罗牌\n",
		// TODO 抽X张塔罗牌 解塔罗牌[牌名]
		PrivateDataFolder: "tarot",
	})
	go func() {
		data, err := web.GetData(bed + "tarots.json")
		if err != nil {
			panic(err)
		}
		tarotData = gjson.ParseBytes(data)
		logrus.Infoln("[tarot]加载tarot成功")
	}()
	engine.OnFullMatch("抽塔罗牌").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i := rand.Intn(22)
		p := rand.Intn(2)
		card := tarotData.Get(strconv.Itoa(i))
		name := card.Get("name").String()
		info := card.Get("info").Map()
		text1 := reasons[rand.Intn(len(reasons))]
		text1 += position[p] + " 的 " + name + "\n"
		text2 := "其意义为：" + info[description[p]].String()
		ctx.SendChain(
			message.Text(text1),
			message.Image(fmt.Sprintf(bed+"MajorArcana/%d.png", i)),
			message.Text(text2),
		)
	})
}
