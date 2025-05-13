// Package wordcount 聊天热词
package wordcount

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/go-ego/gse"
	"github.com/golang/freetype"
	"github.com/wcharczuk/go-chart/v2"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	stopwords           map[string]struct{}
	wordcountDataFolder string
	seg                 gse.Segmenter
)

// 保存聊天消息的时间与内容到json
type MessageRecord struct {
	Time int64  `json:"time"`
	Text string `json:"text"`
}

func appendJSONLine(filePath string, record MessageRecord) {
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := json.Marshal(record)
	if err != nil {
		return
	}
	// 错误处理
	if _, err := f.Write(data); err != nil {
		return
	}
	if _, err := f.WriteString("\n"); err != nil {
		return
	}
}

func loadStopwords() {
	stopwords = make(map[string]struct{})
	data, err := os.ReadFile(wordcountDataFolder + "stopwords.txt")
	if err != nil {
		return
	}
	for _, w := range strings.Split(strings.ReplaceAll(string(data), "\r", ""), "\n") {
		w = strings.TrimSpace(w)
		if w != "" {
			stopwords[w] = struct{}{}
		}
	}
}

func loadCustomDicts() {
	err := seg.LoadDictEmbed("zh_s")
	if err != nil {
		fmt.Println("加载内置词典失败:", err)
	} else {
		fmt.Println("成功加载内置词典")
	}
}

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "聊天热词",
		Help:             "- 热词 | 历史热词",
		PublicDataFolder: "WordCount",
	})
	wordcountDataFolder = engine.DataFolder()
	_ = os.MkdirAll(wordcountDataFolder+"cache/", 0755)

	// 加载 stopwords.txt（如不存在）
	_, err := engine.GetLazyData("stopwords.txt", false)
	if err != nil {
		fmt.Println("下载 stopwords.txt 失败：", err)
	}

	loadStopwords()
	loadCustomDicts()

	engine.OnMessage(zero.OnlyGroup).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			today := time.Now().Format("20060102")
			groupFolder := fmt.Sprintf("%s/messages/%d/", wordcountDataFolder, gid)
			_ = os.MkdirAll(groupFolder, 0755)
			filePath := fmt.Sprintf("%s%s.json", groupFolder, today)

			textContent := strings.TrimSpace(message.ParseMessageFromString(ctx.Event.RawMessage).ExtractPlainText())
			if textContent == "" {
				return
			}
			record := MessageRecord{Time: time.Now().Unix(), Text: textContent}
			appendJSONLine(filePath, record)
		})

	engine.OnRegex(`^热词$`, zero.OnlyGroup).
		Handle(func(ctx *zero.Ctx) {
			_, _ = file.GetLazyData(text.FontFile, control.Md5File, true)
			b, _ := os.ReadFile(text.FontFile)
			font, _ := freetype.ParseFont(b)

			ctx.SendChain(message.Text("开始统计中..."))
			gid := ctx.Event.GroupID

			baseFolder := fmt.Sprintf("%s/messages/%d/", wordcountDataFolder, gid)
			today := time.Now().Format("20060102")
			filePath := fmt.Sprintf("%s%s.json", baseFolder, today)
			if !file.IsExist(filePath) {
				ctx.SendChain(message.Text("ERROR: 今日无聊天记录"))
				return
			}
			content, _ := os.ReadFile(filePath)
			messages := []string{}
			for _, line := range strings.Split(string(content), "\n") {
				if strings.TrimSpace(line) == "" {
					continue
				}
				var rec MessageRecord
				if err := json.Unmarshal([]byte(line), &rec); err == nil {
					messages = append(messages, rec.Text)
				}
			}
			if len(messages) == 0 {
				ctx.SendChain(message.Text("ERROR: 今日无有效聊天记录"))
				return
			}

			// 跳过stopword和2个字以下的词
			messageMap := make(map[string]int)

			for _, msg := range messages {
				text := strings.TrimSpace(msg)
				if text == "" {
					continue
				}

				segments := seg.Segment([]byte(text))
				words := gse.ToSlice(segments, true)

				for _, word := range words {
					// 跳过停用词
					if _, isStopword := stopwords[word]; isStopword {
						continue
					}
					// 跳过所有单字词
					if len([]rune(word)) < 2 {
						continue
					}

					messageMap[word]++
				}
			}

			wc := rankByWordCount(messageMap)
			if len(wc) > 20 {
				wc = wc[:20]
			}

			bars := make([]chart.Value, len(wc))
			for i, v := range wc {
				bars[i] = chart.Value{Value: float64(v.Value), Label: v.Key}
			}

			drawedFile := fmt.Sprintf("%s%d%swordCount.png", wordcountDataFolder+"cache/", gid, today)
			graph := chart.BarChart{
				Font:       font,
				Title:      "热词TOP20 - 今日",
				Background: chart.Style{Padding: chart.Box{Top: 40}},
				Height:     500,
				BarWidth:   35,
				Bars:       bars,
			}
			f, _ := os.Create(drawedFile)
			_ = graph.Render(chart.PNG, f)
			_ = f.Close()
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})

	//历史所有热词
	engine.OnRegex(`^(历史热词)$`, zero.OnlyGroup).
		Handle(func(ctx *zero.Ctx) {
			// 加载字体
			_, _ = file.GetLazyData(text.FontFile, control.Md5File, true)
			b, _ := os.ReadFile(text.FontFile)
			font, _ := freetype.ParseFont(b)

			ctx.SendChain(message.Text("开始统计历史热词中..."))
			gid := ctx.Event.GroupID

			baseFolder := fmt.Sprintf("%s/messages/%d/", wordcountDataFolder, gid)
			files, _ := os.ReadDir(baseFolder)

			messages := []string{}
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
					content, _ := os.ReadFile(baseFolder + f.Name())
					for _, line := range strings.Split(string(content), "\n") {
						if strings.TrimSpace(line) == "" {
							continue
						}
						var rec MessageRecord
						if err := json.Unmarshal([]byte(line), &rec); err == nil {
							messages = append(messages, rec.Text)
						}
					}
				}
			}

			if len(messages) == 0 {
				ctx.SendChain(message.Text("ERROR: 没有历史聊天记录"))
				return
			}

			// 跳过stopword和2个字以下的词
			messageMap := make(map[string]int)

			for _, msg := range messages {
				text := strings.TrimSpace(msg)
				if text == "" {
					continue
				}

				segments := seg.Segment([]byte(text))
				words := gse.ToSlice(segments, true)

				for _, word := range words {
					if _, isStopword := stopwords[word]; isStopword {
						continue
					}

					if len([]rune(word)) < 2 {
						continue
					}

					messageMap[word]++
				}
			}

			wc := rankByWordCount(messageMap)
			if len(wc) > 20 {
				wc = wc[:20]
			}

			bars := make([]chart.Value, len(wc))
			for i, v := range wc {
				bars[i] = chart.Value{Value: float64(v.Value), Label: v.Key}
			}

			drawedFile := fmt.Sprintf("%s%d_historyWordCount.png", wordcountDataFolder+"cache/", gid)
			graph := chart.BarChart{
				Font:       font,
				Title:      "热词TOP20 - 历史",
				Background: chart.Style{Padding: chart.Box{Top: 40}},
				Height:     500,
				BarWidth:   35,
				Bars:       bars,
			}
			f, _ := os.Create(drawedFile)
			_ = graph.Render(chart.PNG, f)
			_ = f.Close()
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + drawedFile))
		})

}

type pair struct {
	Key   string
	Value int
}

type pairlist []pair

func (p pairlist) Len() int           { return len(p) }
func (p pairlist) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p pairlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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
