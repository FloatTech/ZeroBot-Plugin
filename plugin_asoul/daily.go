package asoul

import (
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io/ioutil"
	"net/http"
	"strings"
)

func init() {
	engine.OnKeyword("日程表").
		Handle(func(ctx *zero.Ctx) {
			url := getDynamic()
			if url == "" {
				ctx.Send("值为空")
			}

			ctx.SendChain(message.Image(url))
		})
}

func getDynamic() string {
	api := "https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/space_history?host_uid=703007996"
	resp, err := http.Get(api)
	if err != nil {
		panic(err)
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json := gjson.ParseBytes(data)

	dy := json.Get("data.cards.#.card").Array()
	for i, v := range dy {
		if strings.Contains(v.Str, "日程表") {
			if strings.Contains(dy[i].Str, "img_src") {
				gi := dy[i].Str
				startStr := "\"img_src\":\""
				endStr := "\",\"img_tags"
				inurl := string([]byte(gi)[strings.Index(gi, startStr)+len(startStr) : strings.Index(gi, endStr)])
				imurl := strings.ReplaceAll(inurl, "\\", "")
				return imurl
			}
		}
	}
	return ""
}
