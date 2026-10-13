// Package tarot 塔罗牌
package tarot

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/chat"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/fumiama/deepinfra"
	"github.com/fumiama/deepinfra/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

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

type drawResult struct {
	Name        string
	Position    string
	Description string
	Represent   string
}

type drawResults []drawResult

type cardSet = map[string]card

var (
	cardMap         = make(cardSet, 80)
	infoMap         = make(map[string]cardInfo, 80)
	formationMap    = make(map[string]formation, 10)
	majorArcanaName = make([]string, 0, 80)
	formationName   = make([]string, 0, 10)
	reverse         = [...]string{"", "Reverse/"}
	arcanaType      = [...]string{"MajorArcana", "MinorArcana"}
	minorArcanaType = [...]string{"Cups", "Pentacles", "Swords", "Wands"}
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "塔罗牌",
		Help: "- 抽[塔罗牌|大阿卡纳|小阿卡纳] [询问的事情]\n" +
			"- 抽n张[塔罗牌|大阿卡纳|小阿卡纳]\n" +
			"- 解塔罗牌[牌名]\n" +
			"- [塔罗|大阿卡纳|小阿卡纳|混合]牌阵[圣三角|时间之流|四要素|五牌阵|吉普赛十字|马蹄|六芒星] [询问的事情]",
		PublicDataFolder: "Tarot",
	}).ApplySingle(ctxext.DefaultSingle)

	for _, r := range reverse {
		for _, at := range arcanaType {
			if at == "MinorArcana" {
				for _, mat := range minorArcanaType {
					cachePath := path.Join(engine.DataFolder(), r, at, mat)
					err := os.MkdirAll(cachePath, 0755)
					if err != nil {
						panic(err)
					}
				}
			} else {
				cachePath := path.Join(engine.DataFolder(), r, at)
				err := os.MkdirAll(cachePath, 0755)
				if err != nil {
					panic(err)
				}
			}
		}
	}
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
	engine.OnRegex(`^抽(\d{1,2}张)?((塔罗牌|大阿(尔)?卡纳)|小阿(尔)?卡纳)\s?(.*)$`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		cardType := ctx.State["regex_matched"].([]string)[2]
		question := strings.TrimSpace(ctx.State["regex_matched"].([]string)[6])
		withQuestion := match == "" && question != ""
		n := 1
		reasons := [...]string{"您抽到的是~\n", "锵锵锵，塔罗牌的预言是~\n", "诶，让我看看您抽到了~\n"}
		position := [...]string{"『正位』", "『逆位』"}
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
			imgurl := reverse[p] + card.ImgURL
			data, err := engine.GetLazyData(imgurl, true)
			if err != nil {
				// ctx.SendChain(message.Text("ERROR: ", err))
				logrus.Infof("[tarot]获取图片失败: %v", err)
			} else {
				ctx.SendChain(message.ImageBytes(data))
			}
			ctx.SendChain(message.Text(reasons[rand.Intn(len(reasons))], position[p], "的『", name, "』\n其释义为: ", description))
			if !withQuestion {
				return
			}
			if !chat.EnsureConfig(ctx) {
				ctx.SendChain(message.Text("塔罗解析失败: 无法读取 AI 聊天配置"))
				return
			}
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			stor, err := chat.NewStorage(ctx, gid)
			if err != nil {
				ctx.SendChain(message.Text("塔罗解析失败: ", errors.Wrap(err, "读取 AI 聊天温度配置失败")))
				return
			}
			reply, err := drawResults{{
				Name:        name,
				Position:    position[p],
				Description: description,
			}}.analyze(question, "", stor.Temp())
			if err != nil {
				logrus.Warnln("[tarot]大模型解析失败:", err)
				ctx.SendChain(message.Text("塔罗解析失败: ", err))
				return
			}
			if reply == "" {
				ctx.SendChain(message.Text("塔罗解析失败: 大模型返回为空"))
				return
			}
			if id := ctx.Send(makeNodeMessage(
				reply,
				ctx.CardOrNickName(ctx.Event.UserID),
				ctx.Event.UserID,
			)).ID(); id == 0 {
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
			imgurl := reverse[p] + card.ImgURL
			tarotmsg := message.Message{message.Text(reasons[rand.Intn(len(reasons))], position[p], "的『", name, "』\n")}
			var imgmsg message.Segment
			var err error
			data, err := engine.GetLazyData(imgurl, true)
			if err != nil {
				// ctx.SendChain(message.Text("ERROR: ", err))
				logrus.Infof("[tarot]获取图片失败: %v", err)
				// return
			} else {
				imgmsg = message.ImageBytes(data)
				tarotmsg = append(tarotmsg, imgmsg)
			}
			tarotmsg = append(tarotmsg, message.Text("\n其释义为: ", description))
			msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotmsg...)
		}
		if id := ctx.Send(msg).ID(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})

	engine.OnRegex(`^解塔罗牌\s?(.*)`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		match := ctx.State["regex_matched"].([]string)[1]
		info, ok := infoMap[match]
		if ok {
			imgurl := info.ImgURL
			var tarotmsg message.Message
			data, err := engine.GetLazyData(imgurl, true)
			if err != nil {
				// ctx.SendChain(message.Text("ERROR: ", err))
				logrus.Infof("[tarot]获取图片失败: %v", err)
				// return
			} else {
				imgmsg := message.ImageBytes(data)
				tarotmsg = append(tarotmsg, imgmsg)
			}
			tarotmsg = append(tarotmsg, message.Text("\n", match, "的含义是~\n『正位』:", info.Description, "\n『逆位』:", info.ReverseDescription))
			if id := ctx.Send(tarotmsg).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			return
		}
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("没有找到", match, "噢~"), message.Image("base64://"+binary.BytesToString(cardList)))
	})
	engine.OnRegex(`^((塔罗|大阿(尔)?卡纳)|小阿(尔)?卡纳|混合)牌阵\s?(.*)`, getTarot).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		cardType := ctx.State["regex_matched"].([]string)[1]
		rawMatch := strings.TrimSpace(ctx.State["regex_matched"].([]string)[5])
		var match string
		for name := range formationMap {
			if strings.HasPrefix(rawMatch, name) && len(name) > len(match) {
				match = name
			}
		}
		question := strings.TrimSpace(strings.TrimPrefix(rawMatch, match))
		_, ok := formationMap[match]
		info := formationMap[match]
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
			results := make(drawResults, 0, info.CardsNum)
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
				var tarotmsg message.Message
				imgurl := reverse[p] + card.ImgURL
				var imgmsg message.Segment
				var err error
				data, err := engine.GetLazyData(imgurl, true)
				if err != nil {
					// ctx.SendChain(message.Text("ERROR: ", err))
					logrus.Infof("[tarot]获取图片失败: %v", err)
					// return
				} else {
					imgmsg = message.ImageBytes(data)
					tarotmsg = append(tarotmsg, imgmsg)
				}
				build.WriteString(info.Represent[0][i])
				build.WriteString(":")
				build.WriteString(position[p])
				build.WriteString("的『")
				build.WriteString(name)
				build.WriteString("』\n其释义为: \n")
				build.WriteString(description)
				build.WriteString("\n")
				results = append(results, drawResult{
					Name:        name,
					Position:    position[p],
					Description: description,
					Represent:   info.Represent[0][i],
				})
				msg[i] = ctxext.FakeSenderForwardNode(ctx, tarotmsg...)
			}
			txt := build.String()
			formation, err := text.RenderToBase64(txt, text.FontFile, 420, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			msg[info.CardsNum] = ctxext.FakeSenderForwardNode(ctx, message.Message{message.Image("base64://" + binary.BytesToString(formation))}...)
			if id := ctx.Send(msg).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			if question == "" {
				return
			}

			if !chat.EnsureConfig(ctx) {
				ctx.SendChain(message.Text("塔罗解析失败: 无法读取 AI 聊天配置"))
				return
			}
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
			}
			stor, err := chat.NewStorage(ctx, gid)
			if err != nil {
				ctx.SendChain(message.Text("塔罗解析失败: ", errors.Wrap(err, "读取 AI 聊天温度配置失败")))
				return
			}
			reply, err := results.analyze(question, match, stor.Temp())
			if err != nil {
				logrus.Warnln("[tarot]大模型解析失败:", err)
				ctx.SendChain(message.Text("塔罗解析失败: ", err))
				return
			}
			if reply == "" {
				ctx.SendChain(message.Text("塔罗解析失败: 大模型返回为空"))
				return
			}
			if id := ctx.Send(makeNodeMessage(
				reply,
				ctx.CardOrNickName(ctx.Event.UserID),
				ctx.Event.UserID,
			)).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		} else {
			ctx.SendChain(message.Text("没有找到", rawMatch, "噢~\n现有牌阵列表: \n", strings.Join(formationName, "\n")))
		}
	})
}

func (draws drawResults) analyze(question, formationName string, temperature float32) (string, error) {
	var build strings.Builder
	build.WriteString("你是一位谨慎的塔罗牌解读者。请围绕用户的问题和本次牌面解读，先说明牌面，再给出综合建议，字数控制在300-500字。")
	build.WriteString("不要把占卜结果表述为确定事实。\n")
	build.WriteString("用户问题: ")
	build.WriteString(question)
	build.WriteByte('\n')
	if formationName != "" {
		build.WriteString("牌阵: ")
		build.WriteString(formationName)
		build.WriteByte('\n')
	}
	build.WriteString("牌面:\n")
	for i, draw := range draws {
		build.WriteString(strconv.Itoa(i + 1))
		build.WriteString(". ")
		if draw.Represent != "" {
			build.WriteString("牌位: ")
			build.WriteString(draw.Represent)
			build.WriteString("; ")
		}
		build.WriteString("牌: ")
		build.WriteString(draw.Name)
		build.WriteString("; 方位: ")
		build.WriteString(draw.Position)
		build.WriteString("; 固定释义: ")
		build.WriteString(draw.Description)
		build.WriteByte('\n')
	}
	cfg, err := chat.CurrentConfig()
	if err != nil {
		return "", errors.New("读取 AI 配置失败")
	}
	topp, maxn := cfg.MParams()
	mod, err := cfg.Type.Protocol(cfg.ModelName, temperature, topp, maxn, cfg.ReasoningEffort, cfg.Separator)
	if err != nil {
		return "", errors.Wrap(err, "创建 AI 模型协议失败")
	}

	api := deepinfra.NewAPI(cfg.API, string(cfg.Key))
	api.SetHTTPClient(&http.Client{Timeout: 30 * time.Second})
	data, err := api.Request(mod.User(model.NewContentText(build.String())))
	if err != nil {
		return "", errors.New("请求 AI 模型失败，请检查接口、密钥及网络")
	}
	return strings.TrimSpace(data), nil
}

func makeNodeMessage(reply, nickname string, userID int64) message.Message {
	chunks := splitTextChunks("塔罗解析:\n"+reply, 1000)
	msg := make(message.Message, 0, len(chunks))
	for _, chunk := range chunks {
		msg = append(msg, message.CustomNode(nickname, userID, message.Message{message.Text(chunk)}))
	}
	return msg
}

func splitTextChunks(txt string, maxRunes int) []string {
	runes := []rune(txt)
	if maxRunes <= 0 || len(runes) <= maxRunes {
		return []string{txt}
	}
	chunks := make([]string, 0, (len(runes)+maxRunes-1)/maxRunes)
	for len(runes) > maxRunes {
		chunks = append(chunks, string(runes[:maxRunes]))
		runes = runes[maxRunes:]
	}
	if len(runes) > 0 {
		chunks = append(chunks, string(runes))
	}
	return chunks
}
