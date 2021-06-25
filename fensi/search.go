package fensi

import (
	"encoding/json"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
)

func init() {
	zero.OnRegex(`^/搜索 (.*)$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			searchJson := searchapi(keyword)

			if searchapi(keyword).Data.NumResults == 0 {
				ctx.Send("名字没搜到")
				return
			}

			if searchapi(keyword).Data.NumResults < 5 {
				ctx.SendChain(message.Text(
					"uid：  ", searchJson.Data.Result[0].Mid, "\n",
					"name：  ", searchJson.Data.Result[0].Uname, "\n",
					))
				return
			}

			ctx.SendChain(message.Text(
				"搜索结果很多，请尽量准确关键字，以下为你返回前5条结果", "\n\n",
				"uid1：  ", searchJson.Data.Result[0].Mid, "\n",
				"name1：  ", searchJson.Data.Result[0].Uname, "\n\n",
				"uid2：  ", searchJson.Data.Result[1].Mid, "\n",
				"name2:  ", searchJson.Data.Result[1].Uname, "\n\n",
				"uid3:  ", searchJson.Data.Result[2].Mid, "\n",
				"name3:  ", searchJson.Data.Result[2].Uname, "\n\n",
				"uid4:  ", searchJson.Data.Result[3].Mid, "\n",
				"name4:  ", searchJson.Data.Result[3].Uname, "\n\n",
				"uid5:  ", searchJson.Data.Result[4].Mid, "\n",
				"name5:  ", searchJson.Data.Result[4].Uname,
				))
	})
}

func searchapi(keyword string) *search {
	url := "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&&user_type=1&keyword=" + keyword
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &search{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}
