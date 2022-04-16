// Package wordcount 聊天热词
package wordcount

import (
	"fmt"
	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/golang/freetype"
	"github.com/sirupsen/logrus"
	"github.com/wcharczuk/go-chart/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	re = regexp.MustCompile(`^[一-龥]+$`)
)

func init() {
	engine := control.Register("wordcount", &control.Options{
		DisableOnDefault: false,
		Help: "聊天热词\n" +
			"- 热词[群号]",
		PublicDataFolder: "WordCount",
	})
	cachePath := engine.DataFolder() + "cache/"
	stopwordsMap := make(map[string]int)
	go func() {
		_ = os.RemoveAll(cachePath)
		err := os.MkdirAll(cachePath, 0755)
		if err != nil {
			panic(err)
		}
		_, _ = file.GetLazyData(engine.DataFolder()+"stopwords.txt", false, false)
		data, err := os.ReadFile(engine.DataFolder() + "stopwords.txt")
		if err != nil {
			panic(err)
		}
		for _, v := range strings.Split(binary.BytesToString(data), "\r\n") {
			stopwordsMap[v] = 1
		}
		logrus.Infoln("[wordcount]加载", len(stopwordsMap), "条停用词")
	}()
	engine.OnRegex(`^热词\s?(\d*)$`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			gid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			if gid == 0 {
				gid = ctx.Event.GroupID
			}
			group := ctx.GetGroupInfo(gid, false)
			if group.MemberCount == 0 {
				ctx.SendChain(message.Text(fmt.Sprintf("%s未加入%s(%d),无法获得热词呢", zero.BotConfig.NickName[0], group.Name, gid)))
				return
			}
			today := time.Now().Format("20060102")
			drawedFile := cachePath + strconv.FormatInt(ctx.Event.GroupID, 10) + today + "wordCount.png"
			if file.IsExist(drawedFile) {
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
				return
			}
			messageMap := make(map[string]int)
			h := ctx.CallAction("get_group_msg_history", zero.Params{"group_id": gid}).Data
			messageSeq := h.Get("messages.0.message_seq").Int()
			for i := 0; i < 50 && messageSeq != 0; i++ {
				if i != 0 {
					h = ctx.CallAction("get_group_msg_history", zero.Params{"group_id": ctx.Event.GroupID, "message_seq": messageSeq}).Data
				}
				for _, v := range h.Get("messages.#.message").Array() {
					tex := strings.TrimSpace(message.ParseMessage(binary.StringToBytes(v.Raw)).ExtractPlainText())
					if tex == "" {
						continue
					}
					for _, t := range ctx.GetWordSlices(tex).Get("slices").Array() {
						tex := strings.TrimSpace(t.Str)
						if _, ok := stopwordsMap[tex]; !ok && re.MatchString(tex) {
							messageMap[tex]++
						}
					}
				}
				messageSeq = h.Get("messages.0.message_seq").Int()
			}
			wc := rankByWordCount(messageMap)
			if len(wc) > 20 {
				wc = wc[:20]
			}
			// 绘图
			if len(wc) == 0 {
				ctx.SendChain(message.Text("ERROR:历史消息为空或者无法获得历史消息"))
				return
			}
			_, err := file.GetLazyData(text.FontFile, false, true)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			b, err := os.ReadFile(text.FontFile)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			font, err := freetype.ParseFont(b)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
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
				Title: time.Now().Format("2006-01-02") + "热词top20",
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
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			err = graph.Render(chart.PNG, f)
			_ = f.Close()
			if err != nil {
				_ = os.Remove(drawedFile)
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))

		})
}

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
