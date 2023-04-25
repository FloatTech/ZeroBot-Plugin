// Package moegoe 日韩中 VITS 模型拟声
package moegoe

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
)

const (
	ygomoegoeapi = "http://127.0.0.1:25565/tts"
)

var (
	speakers = map[string]uint{
		"游城十代": 0, "十代": 0,
		"丸藤亮": 1, "亮": 1, "凯撒": 1,
		"海马濑人": 2, "海马": 2, "社长": 2,
		"爱德菲尼克斯": 3, "爱德": 3,
		"不动游星": 4, "游星": 4,
		"鬼柳京介": 5, "鬼柳": 5,
		"榊遊矢": 6, "榊游矢": 6, "游矢": 6,
	}
)

type repdata struct {
	Data DataTTS `json:"data"`
}

// DataTTS ...
type DataTTS struct {
	Model  string `json:"model"` //非必需
	ID     uint   `json:"id"`
	TTS    bool   `json:"tts_choice"` //非必需
	Text   string `json:"text"`
	Output string `json:"outputName"`
}

type respData struct {
	ID      uint   `json:"id"`
	Speaker string `json:"speaker"`
	URL     string `json:"url"`
}

func init() {
	speakerList := make([]string, 0, len(speakers))
	for speaker := range speakers {
		speakerList = append(speakerList, speaker)
	}
	en := control.Register("ygomoegoe", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏王 moegoe 模型拟声",
		Help: "- 让[xxxx]说(日语)\n" +
			"当前角色:\n游城十代;丸藤亮;海马濑人;爱德菲尼克斯;不动游星;鬼柳京介;榊遊矢;",
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnRegex("^让(" + strings.Join(speakerList, "|") + ")说([A-Za-z\\s\\d\u3005\u3040-\u30ff\u4e00-\u9fff\uff11-\uff19\uff21-\uff3a\uff41-\uff5a\uff66-\uff9d\\pP]+)$").Limit(ctxext.LimitByGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("正在尝试"))
			text := ctx.State["regex_matched"].([]string)[2]
			id := speakers[ctx.State["regex_matched"].([]string)[1]]
			urlValues := repdata{
				Data: DataTTS{
					Model:  "ygo7",
					ID:     id,
					Text:   text,
					TTS:    true,
					Output: strconv.FormatInt(ctx.Event.UserID, 10),
				},
			}
			data, err := json.Marshal(urlValues)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			resp, err := web.PostData(ygomoegoeapi, "application/json", bytes.NewReader(data))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			var reslut respData
			err = json.Unmarshal(resp, &reslut)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Record("file:///Users/liuyu.fang/Documents/Vits/MoeGoe/" + reslut.URL))
		})
}
