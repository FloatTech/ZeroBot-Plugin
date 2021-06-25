package fensi

import (
	"encoding/json"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
	"strconv"
	"time"
)



func init() {
	zero.OnRegex(`^/粉丝 (.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			uid := searchapi(keyword).Data.Result[0].Mid
			fensijson := fensiapi(strconv.Itoa(uid))
			ctx.SendChain(message.Text(
				"uid: ", fensijson.Mid, "\n",
				"名字: ", fensijson.Uname, "\n",
				"当前粉丝数: ", fensijson.Follower, "\n",
				"24h涨粉数: ", fensijson.Rise, "\n",
				"视频投稿数: ", fensijson.Video, "\n",
				"直播间id: ", fensijson.Roomid, "\n",
				"舰队: ", fensijson.GuardNum, "\n",
				"直播总排名: ", fensijson.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", uid, "\n",
				"数据获取时间: ", timeStamp(),
			))
	})
}

func init() {
	zero.OnRegex(`^/info (.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			uid := searchapi(keyword).Data.Result[0].Mid
			fensijson := fensiapi(strconv.Itoa(uid))
			ctx.SendChain(message.Text(
				"uid: ", fensijson.Mid, "\n",
				"名字: ", fensijson.Uname, "\n",
				"当前粉丝数: ", fensijson.Follower, "\n",
				"24h涨粉数: ", fensijson.Rise, "\n",
				"视频投稿数: ", fensijson.Video, "\n",
				"直播间id: ", fensijson.Roomid, "\n",
				"舰队: ", fensijson.GuardNum, "\n",
				"直播总排名: ", fensijson.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", uid, "\n",
				"数据获取时间: ", timeStamp(),
			))
		})
}

func fensiapi(uid string) *follower {

	url := "https://api.vtbs.moe/v1/detail/" + uid
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &follower{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}


func timeStamp() string{
	unixtime := time.Now().Unix()
	timestamp := time.Unix(unixtime, 0)
	return timestamp.Format("2006-01-02 15:04:05")
}
