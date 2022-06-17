// Package tarot 塔罗牌
package tarot

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
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

var cardMap = make(cardSet, 30)
var infoMap = make(map[string]cardInfo, 30)
var formationMap = make(map[string]formation, 10)

// var cardName = make([]string, 22)

func init() {
	engine := control.Register("tarot", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "塔罗牌\n" +
			"- 抽塔罗牌\n" +
			"- 抽n张塔罗牌\n" +
			"- 解塔罗牌[牌名]\n" +
			"- 塔罗牌阵[圣三角|时间之流|四要素|五牌阵|吉普赛十字|马蹄|六芒星]",
		PublicDataFolder: "Tarot",
	}).ApplySingle(ctxext.DefaultSingle)

	engine.OnRegex(`^抽(\d{1,2}张)?塔罗牌$`, ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			if len(cardMap) > 0 {
				return true
			}
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
			logrus.Infof("[tarot]读取%d张大阿尔卡纳塔罗牌", len(cardMap))
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		n := 1
		reasons := [...]string{"您抽到的是~\n", "锵锵锵，塔罗牌的预言是~\n", "诶，让我看看您抽到了~\n"}
		position := [...]string{"正位", "逆位"}
		reverse := [...]string{"", "Reverse"}
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
			i := rand.Intn(22)
			p := rand.Intn(2)
			card := cardMap[(strconv.Itoa(i))]
			name := card.Name
			if id := ctx.SendChain(
				message.Text(reasons[rand.Intn(len(reasons))], position[p], " 的 ", name, "\n"),
				message.Image(fmt.Sprintf(bed+"MajorArcana%s/%d.png", reverse[p], i))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
			return
		}
		msg := make([]message.MessageSegment, n)
		randomIntMap := make(map[int]int, 30)
		for i := range msg {
			j := rand.Intn(22)
			_, ok := randomIntMap[j]
			for ok {
				j = rand.Intn(22)
				_, ok = randomIntMap[j]
			}
			randomIntMap[j] = 0
			p := rand.Intn(2)
			card := cardMap[(strconv.Itoa(j))]
			name := card.Name
			tarotMsg := []message.MessageSegment{
				message.Text(reasons[rand.Intn(len(reasons))], position[p], " 的 ", name, "\n"),
				message.Image(fmt.Sprintf(bed+"MajorArcana%s/%d.png", reverse[p], j))}
			msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotMsg...)
		}
		ctx.SendGroupForwardMessage(ctx.Event.GroupID, msg)
	})

	engine.OnRegex(`^解塔罗牌\s?(.*)`, ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			if len(cardMap) > 0 {
				for _, card := range cardMap {
					infoMapKey := strings.Split(card.Name, "(")[0]
					infoMap[infoMapKey] = card.cardInfo
					// 可以拿来显示大阿尔卡纳列表
					// cardName = append(cardName, infoMapKey)
				}
				return true
			}
			tempMap := make(cardSet, 30)
			data, err := engine.GetLazyData("tarots.json", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &tempMap)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}

			for _, card := range tempMap {
				infoMapKey := strings.Split(card.Name, "(")[0]
				infoMap[infoMapKey] = card.cardInfo
				// 可以拿来显示大阿尔卡纳列表
				// cardName = append(cardName, infoMapKey)
			}
			return true
		},
	)).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
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
	engine.OnRegex(`^塔罗牌阵\s?(.*)`, ctxext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			if len(cardMap) == 0 {
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
				logrus.Infof("[tarot]读取%d张大阿尔卡纳塔罗牌", len(cardMap))
			}
			data, err := engine.GetLazyData("formation.json", true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			err = json.Unmarshal(data, &formationMap)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return false
			}
			logrus.Infof("[tarot]读取%d组塔罗牌阵", len(formationMap))
			return true
		})).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		info, ok := formationMap[match]
		position := [...]string{"正位", "逆位"}
		reverse := [...]string{"", "Reverse"}
		if ok {
			msg := make([]message.MessageSegment, info.CardsNum)
			randomIntMap := make(map[int]int, 30)
			for i := range msg {
				j := rand.Intn(22)
				_, ok := randomIntMap[j]
				for ok {
					j = rand.Intn(22)
					_, ok = randomIntMap[j]
				}
				randomIntMap[j] = 0
				p := rand.Intn(2)
				card := cardMap[(strconv.Itoa(j))]
				name := card.Name
				tarotMsg := []message.MessageSegment{
					message.Text(info.Represent[0][i], ":", position[p], " 的 ", name, "\n"),
					message.Image(fmt.Sprintf(bed+"MajorArcana%s/%d.png", reverse[p], j))}
				msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotMsg...)
			}
			ctx.SendGroupForwardMessage(ctx.Event.GroupID, msg)
		} else {
			ctx.SendChain(message.Text("没有找到", match, "噢~"))
		}
	})
}
