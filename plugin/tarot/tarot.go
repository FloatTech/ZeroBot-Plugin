package tarot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitcode.net/shudorcl/zbp-tarot/-/raw/master/"

type card struct {
	Name string `json:"name"`
	Info struct {
		Description        string `json:"description"`
		ReverseDescription string `json:"reverseDescription"`
		ImgURL             string `json:"imgUrl"`
	} `json:"info"`
}
type cardset = map[string]card

var cardMap = make(cardset, 256)
var reasons = [...]string{"您抽到的是~\n", "锵锵锵，塔罗牌的预言是~\n", "诶，让我看看您抽到了~\n"}
var position = [...]string{"正位", "逆位"}

func init() {
	engine := control.Register("tarot", &control.Options{
		DisableOnDefault: false,
		Help: "塔罗牌\n" +
			"- 抽塔罗牌\n" +
			"- 抽n张塔罗牌",
		// TODO 抽X张塔罗牌 解塔罗牌[牌名]
		PublicDataFolder: "Tarot",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnRegex(`^抽(\d{1,2}张)?塔罗牌$`, ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			data, err := engine.GetLazyData("tarots.json", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &cardMap)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			logrus.Infof("[tarot]读取%d张塔罗牌", len(cardMap))
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		n := 1
		if match != "" {
			var err error
			n, err = strconv.Atoi(match[:len(match)-3])
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if n <= 0 {
				ctx.SendChain(message.Text("ERROR:张数必须为正"))
				return
			}
			if n > 1 && !zero.OnlyGroup(ctx) {
				ctx.SendChain(message.Text("ERROR:抽取多张仅支持群聊"))
				return
			}
			if n > 20 {
				ctx.SendChain(message.Text("ERROR:抽取张数过多"))
				return
			}
		}
		if n == 1 {
			if id := ctx.Send(randTarot()); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
			return
		}
		msg := make([]message.MessageSegment, n)
		for i := range msg {
			msg[i] = ctxext.FakeSenderForwardNode(ctx, randTarot()...)
		}
		ctx.SendGroupForwardMessage(ctx.Event.GroupID, msg)
		return
	})
}

func randTarot() []message.MessageSegment {
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
	return []message.MessageSegment{
		message.Text(reasons[rand.Intn(len(reasons))], position[p], " 的 ", name, "\n"),
		message.Image(fmt.Sprintf(bed+"MajorArcana/%d.png", i)),
		message.Text("\n其意义为: ", info),
	}
}
