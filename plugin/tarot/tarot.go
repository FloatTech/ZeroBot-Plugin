// Package tarot 塔罗牌
package tarot

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const bed = "https://gitcode.net/shudorcl/zbp-tarot/-/raw/master/"

type cardInfo struct {
	Description        string `json:"description"`
	ReverseDescription string `json:"reverseDescription"`
	ImgURL             string `json:"imgUrl"`
}
type card struct {
	Name     string `json:"name"`
	cardInfo `json:"info"`
}

type formation struct {
	CardsNum  int        `json:"cards_num"`
	IsCut     bool       `json:"is_cut"`
	Represent [][]string `json:"represent"`
}
type cardSet = map[string]card

var cardMap = make(cardSet, 80)
var infoMap = make(map[string]cardInfo, 80)
var formationMap = make(map[string]formation, 10)

var majorArcanaName = make([]string, 0, 80)
var formationName = make([]string, 0, 10)

func init() {
	engine := control.Register("tarot", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "塔罗牌\n" +
			"- 抽[塔罗牌|大阿卡纳|小阿卡纳]\n" +
			"- 抽n张[塔罗牌|大阿卡纳|小阿卡纳]\n" +
			"- 解塔罗牌[牌名]\n" +
			"- [塔罗|大阿卡纳|小阿卡纳|混合]牌阵[圣三角|时间之流|四要素|五牌阵|吉普赛十字|马蹄|六芒星]",
		PublicDataFolder: "Tarot",
	}).ApplySingle(ctxext.DefaultSingle)

	getTarot := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		data, err := engine.GetLazyData("tarots.json", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = json.Unmarshal(data, &cardMap)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		for _, card := range cardMap {
			infoMap[card.Name] = card.cardInfo
		}
		for i := 0; i < 22; i++ {
			majorArcanaName = append(majorArcanaName, cardMap[strconv.Itoa(i)].Name)
		}
		logrus.Infof("[tarot]读取%d张塔罗牌", len(cardMap))
		formation, err := engine.GetLazyData("formation.json", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		err = json.Unmarshal(formation, &formationMap)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		for k := range formationMap {
			formationName = append(formationName, k)
		}
		logrus.Infof("[tarot]读取%d组塔罗牌阵", len(formationMap))
		return true
	})
	engine.OnRegex(`^抽(\d{1,2}张)?((塔罗牌|大阿(尔)?卡纳)|小阿(尔)?卡纳)$`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		cardType := ctx.State["regex_matched"].([]string)[2]
		n := 1
		reasons := [...]string{"您抽到的是~\n", "锵锵锵，塔罗牌的预言是~\n", "诶，让我看看您抽到了~\n"}
		position := [...]string{"『正位』", "『逆位』"}
		reverse := [...]string{"", "Reverse/"}
		start := 0
		length := 22
		if match != "" {
			var err error
			n, err = strconv.Atoi(match[:len(match)-3])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if n <= 0 {
				ctx.SendChain(message.Text("ERROR: 张数必须为正"))
				return
			}
			if n > 1 && !zero.OnlyGroup(ctx) {
				ctx.SendChain(message.Text("ERROR: 抽取多张仅支持群聊"))
				return
			}
			if n > 20 {
				ctx.SendChain(message.Text("ERROR: 抽取张数过多"))
				return
			}
		}
		if strings.Contains(cardType, "小") {
			start = 22
			length = 55
		}
		if n == 1 {
			i := rand.Intn(length) + start
			p := rand.Intn(2)
			card := cardMap[strconv.Itoa(i)]
			name := card.Name
			description := card.Description
			if p == 1 {
				description = card.ReverseDescription
			}
			if id := ctx.SendChain(
				message.Text(reasons[rand.Intn(len(reasons))], position[p], "的『", name, "』\n"),
				message.Image(bed+reverse[p]+card.ImgURL),
				message.Text("\n其释义为: ", description)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			return
		}
		msg := make(message.Message, n)
		randomIntMap := make(map[int]int, 30)
		for i := range msg {
			j := rand.Intn(length)
			_, ok := randomIntMap[j]
			for ok {
				j = rand.Intn(length)
				_, ok = randomIntMap[j]
			}
			randomIntMap[j] = 0
			p := rand.Intn(2)
			card := cardMap[strconv.Itoa(j+start)]
			name := card.Name
			description := card.Description
			if p == 1 {
				description = card.ReverseDescription
			}
			tarotMsg := message.Message{
				message.Text(position[p], "的『", name, "』\n"),
				message.Image(bed + reverse[p] + card.ImgURL),
				message.Text("\n其释义为: ", description)}
			msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotMsg...)
		}
		ctx.Send(msg)
	})

	engine.OnRegex(`^解塔罗牌\s?(.*)`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		info, ok := infoMap[match]
		if ok {
			ctx.SendChain(
				message.Image(bed+info.ImgURL),
				message.Text("\n", match, "的含义是~"),
				message.Text("\n『正位』:", info.Description),
				message.Text("\n『逆位』:", info.ReverseDescription))
		} else {
			var build strings.Builder
			build.WriteString("塔罗牌列表\n大阿尔卡纳:\n")
			build.WriteString(strings.Join(majorArcanaName[:7], " "))
			build.WriteString("\n")
			build.WriteString(strings.Join(majorArcanaName[7:14], " "))
			build.WriteString("\n")
			build.WriteString(strings.Join(majorArcanaName[14:22], " "))
			build.WriteString("\n小阿尔卡纳:\n[圣杯|星币|宝剑|权杖] [0-10|侍从|骑士|王后|国王]")
			txt := build.String()
			cardList, err := text.RenderToBase64(txt, text.FontFile, 420, 20)
			if err != nil {
				ctx.SendChain(message.Text("没有找到", match, "噢~"))
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			ctx.SendChain(message.Text("没有找到", match, "噢~"), message.Image("base64://"+binary.BytesToString(cardList)))
		}
	})
	engine.OnRegex(`^((塔罗|大阿(尔)?卡纳)|小阿(尔)?卡纳|混合)牌阵\s?(.*)`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		cardType := ctx.State["regex_matched"].([]string)[1]
		match := ctx.State["regex_matched"].([]string)[5]
		info, ok := formationMap[match]
		position := [...]string{"『正位』", "『逆位』"}
		reverse := [...]string{"", "Reverse/"}
		start, length := 0, 22
		if strings.Contains(cardType, "小") {
			start = 22
			length = 55
		} else if cardType == "混合" {
			start = 0
			length = 77
		}
		if ok {
			ctx.SendChain(message.Text("少女祈祷中..."))
			var build strings.Builder
			build.WriteString(ctx.CardOrNickName(ctx.Event.UserID))
			build.WriteString("---")
			build.WriteString(match)
			build.WriteString("\n")
			msg := make(message.Message, info.CardsNum+1)
			randomIntMap := make(map[int]int, 30)
			for i := 0; i < info.CardsNum; i++ {
				j := rand.Intn(length)
				_, ok := randomIntMap[j]
				for ok {
					j = rand.Intn(length)
					_, ok = randomIntMap[j]
				}
				randomIntMap[j] = 0
				p := rand.Intn(2)
				card := cardMap[strconv.Itoa(j+start)]
				name := card.Name
				description := card.Description
				if p == 1 {
					description = card.ReverseDescription
				}
				tarotMsg := message.Message{message.Image(bed + reverse[p] + card.ImgURL)}
				build.WriteString(info.Represent[0][i])
				build.WriteString(":")
				build.WriteString(position[p])
				build.WriteString("的『")
				build.WriteString(name)
				build.WriteString("』\n其释义为: \n")
				build.WriteString(description)
				build.WriteString("\n")
				msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotMsg...)
			}
			txt := build.String()
			formation, err := text.RenderToBase64(txt, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			msg[info.CardsNum] = ctxext.FakeSenderForwardNode(ctx, message.Message{message.Image("base64://" + binary.BytesToString(formation))}...)
			ctx.Send(msg)
		} else {
			ctx.SendChain(message.Text("没有找到", match, "噢~\n现有牌阵列表: \n", strings.Join(formationName, "\n")))
		}
	})
}
