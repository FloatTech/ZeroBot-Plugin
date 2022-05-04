package tarot

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitcode.net/shudorcl/zbp-tarot/-/raw/master/"

type Card struct {
	Name string `json:"name"`
	Info struct {
		Description        string `json:"description"`
		ReverseDescription string `json:"reverseDescription"`
		ImgURL             string `json:"imgUrl"`
	} `json:"info"`
}
type CardSet = map[string]Card

var cardMap = make(CardSet, 256)
var reasons = []string{"您抽到的是~\n", "锵锵锵，塔罗牌的预言是~\n", "诶，让我看看您抽到了~\n"}
var position = []string{"正位", "逆位"}

func init() {
	engine := control.Register("tarot", &control.Options{
		DisableOnDefault: false,
		Help: "塔罗牌\n" +
			"- 抽塔罗牌\n",
		// TODO 抽X张塔罗牌 解塔罗牌[牌名]
		PublicDataFolder: "Tarot",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnFullMatch("抽塔罗牌", ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			tarotPath := engine.DataFolder() + "tarots.json"
			data, err := file.GetLazyData(tarotPath, true, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &cardMap)
			if err != nil {
				panic(err)
			}
			log.Printf("[tarot]读取%d张塔罗牌", len(cardMap))
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		i := rand.Intn(22)
		p := rand.Intn(2)
		card := cardMap[(strconv.Itoa(i))]
		name := card.Name
		var info string
		if p == 0 {
			info = card.Info.Description
		} else {
			info = card.Info.ReverseDescription
		}
		if id := ctx.SendChain(
			message.At(ctx.Event.UserID),
			message.Text(reasons[rand.Intn(len(reasons))], position[p], " 的 ", name, "\n"),
			message.Image(fmt.Sprintf(bed+"MajorArcana/%d.png", i)),
			message.Text("\n其意义为：", info),
		); id.ID() == 0 {
			ctx.SendChain(message.Text("ERROR:可能被风控了"))
		}
	})
}
