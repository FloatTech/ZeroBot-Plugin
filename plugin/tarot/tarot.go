// Package tarot 塔罗牌
package tarot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/binary"
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

// var cardName = make([]string, 30)
var formationName = make([]string, 10)

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

	getTarot := ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
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
		for _, card := range cardMap {
			infoMap[card.Name] = card.cardInfo
			// 可以拿来显示塔罗牌列表
			// cardName = append(cardName, card.Name)
		}
		logrus.Infof("[tarot]读取%d张塔罗牌", len(cardMap))
		formation, err := engine.GetLazyData("formation.json", true)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return false
		}
		err = json.Unmarshal(formation, &formationMap)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
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
		position := [...]string{"正位", "逆位"}
		reverse := [...]string{"", "Reverse"}
		start := 0
		length := 22
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
		if strings.Contains(cardType, "小") {
			start = 22
			length = 55
		}
		if n == 1 {
			i := rand.Intn(length) + start
			p := rand.Intn(2)
			card := cardMap[(strconv.Itoa(i))]
			name := card.Name
			if id := ctx.SendChain(
				message.Text(reasons[rand.Intn(len(reasons))], position[p], " 的 ", name, "\n"),
				message.Image(fmt.Sprintf("%s/%s/%s", bed, reverse[p], card.ImgURL))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
			return
		}
		msg := make([]message.MessageSegment, n)
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
			card := cardMap[(strconv.Itoa(j + start))]
			name := card.Name
			tarotMsg := []message.MessageSegment{
				message.Text(reasons[rand.Intn(len(reasons))], position[p], " 的 ", name, "\n"),
				message.Image(fmt.Sprintf("%s/%s/%s", bed, reverse[p], card.ImgURL))}
			msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotMsg...)
		}
		ctx.SendGroupForwardMessage(ctx.Event.GroupID, msg)
	})

	engine.OnRegex(`^解塔罗牌\s?(.*)`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		info, ok := infoMap[match]
		if ok {
			ctx.SendChain(
				message.Image(bed+info.ImgURL),
				message.Text("\n", match, "的含义是~"),
				message.Text("\n正位:", info.Description),
				message.Text("\n逆位:", info.ReverseDescription))
		} else {
			ctx.SendChain(message.Text("没有找到", match, "噢~"))
		}
	})
	engine.OnRegex(`^((塔罗|大阿(尔)?卡纳)|小阿(尔)?卡纳|混合)牌阵\s?(.*)`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		cardType := ctx.State["regex_matched"].([]string)[1]
		match := ctx.State["regex_matched"].([]string)[5]
		info, ok := formationMap[match]
		position := [...]string{"正位", "逆位"}
		reverse := [...]string{"", "Reverse"}
		start, length := 0, 22
		if strings.Contains(cardType, "小") {
			start = 22
			length = 55
		} else if cardType == "混合" {
			start = 0
			length = 77
		}
		if ok {
			var build strings.Builder
			build.WriteString(ctx.CardOrNickName(ctx.Event.UserID))
			build.WriteString("\n")
			msg := make([]message.MessageSegment, info.CardsNum)
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
				card := cardMap[(strconv.Itoa(j + start))]
				name := card.Name
				tarotMsg := []message.MessageSegment{message.Image(fmt.Sprintf("%s/%s/%s", bed, reverse[p], card.ImgURL))}
				build.WriteString(info.Represent[0][i])
				build.WriteString(": ")
				build.WriteString(position[p])
				build.WriteString(" 的 ")
				build.WriteString(name)
				build.WriteString("\n")
				msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotMsg...)
			}
			txt := build.String()
			formation, err := text.RenderToBase64(txt, text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			// TODO 视gocq变化将牌阵信息加入转发列表中
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(formation)))
			ctx.SendGroupForwardMessage(ctx.Event.GroupID, msg)
		} else {
			ctx.SendChain(message.Text("没有找到", match, "噢~\n现有牌阵列表: ", strings.Join(formationName, " ")))
		}
	})
}
