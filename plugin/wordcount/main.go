// Package wordcount 聊天热词
package wordcount

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fumiama/jieba"
	"github.com/golang/freetype"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"github.com/wcharczuk/go-chart/v2"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	re        = regexp.MustCompile(`^[一-龥]+$`)
	stopwords []string
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "聊天热词",
		Help:             "- 热词 [消息数目]|热词 1000",
		PublicDataFolder: "WordCount",
	})
	cachePath := engine.DataFolder() + "cache/"
	dat, err := file.GetLazyData("data/Chat/dict.txt", control.Md5File, true)
	if err != nil {
		panic(err)
	}
	seg, err := jieba.LoadDictionary(bytes.NewReader(dat))
	if err != nil {
		panic(err)
	}
	_ = os.RemoveAll(cachePath)
	_ = os.MkdirAll(cachePath, 0755)
	engine.OnRegex(`^热词\s?(\d*)$`, zero.OnlyGroup, fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		_, err := engine.GetLazyData("stopwords.txt", false)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		data, err := os.ReadFile(engine.DataFolder() + "stopwords.txt")
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		stopwords = strings.Split(strings.ReplaceAll(binary.BytesToString(data), "\r", ""), "\n")
		sort.Strings(stopwords)
		logrus.Infoln("[wordcount]加载", len(stopwords), "条停用词")
		return true
	})).Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			_, err := file.GetLazyData(text.FontFile, control.Md5File, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			b, err := os.ReadFile(text.FontFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			font, err := freetype.ParseFont(b)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}

			ctx.SendChain(message.Text("少女祈祷中..."))
			p, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			if p > 10000 {
				p = 10000
			}
			if p == 0 {
				p = 1000
			}
			gid := ctx.Event.GroupID
			group := ctx.GetGroupInfo(gid, false)
			if group.MemberCount == 0 {
				ctx.SendChain(message.Text(zero.BotConfig.NickName[0], "未加入", group.Name, "(", gid, "),无法获得热词呢"))
				return
			}
			today := time.Now().Format("20060102")
			drawedFile := fmt.Sprintf("%s%d%s%dwordCount.png", cachePath, gid, today, p)
			if file.IsExist(drawedFile) {
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				return
			}
			messageMap := make(map[string]int, 256)
			h := ctx.GetGroupMessageHistory(gid, 0, p, false)
			h.Get("messages").ForEach(func(_, msgObj gjson.Result) bool {
				tex := strings.TrimSpace(message.ParseMessageFromString(msgObj.Get("raw_message").Str).ExtractPlainText())
				if tex != "" {
					words := seg.Cut(tex, true)
					for _, word := range words {
						word = strings.TrimSpace(word)
						i := sort.SearchStrings(stopwords, word)
						if re.MatchString(word) && (i >= len(stopwords) || stopwords[i] != word) {
							messageMap[word]++
						}
					}
				}
				return true
			})

			wc := rankByWordCount(messageMap)
			if len(wc) > 20 {
				wc = wc[:20]
			}
			// 绘图
			if len(wc) == 0 {
				ctx.SendChain(message.Text("ERROR: 历史消息为空或者无法获得历史消息"))
				return
			}
			bars := make([]chart.Value, len(wc))
			for i, v := range wc {
				bars[i] = chart.Value{
					Value: float64(v.Value),
					Label: v.Key,
				}
			}
			graph := chart.BarChart{
				Font:  font,
				Title: fmt.Sprintf("%s(%d)在%s号的%d条消息的热词top20", group.Name, gid, time.Now().Format("2006-01-02"), p),
				Background: chart.Style{
					Padding: chart.Box{
						Top: 40,
					},
				},
				Height:   500,
				BarWidth: 25,
				Bars:     bars,
			}
			f, err := os.Create(drawedFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = graph.Render(chart.PNG, f)
			_ = f.Close()
			if err != nil {
				_ = os.Remove(drawedFile)
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})
}

func rankByWordCount(wordFrequencies map[string]int) pairlist {
	pl := make(pairlist, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type pair struct {
	Key   string
	Value int
}

type pairlist []pair

func (p pairlist) Len() int           { return len(p) }
func (p pairlist) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
