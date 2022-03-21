package bilibili

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 查vup粉丝数据
func init() {
	engine.OnRegex(`^>vup info\s?(.{1,25})$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			res, err := uid(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			id := res.Get("data.result.0.mid").String()
			// 获取详情
			fo, err := fansapi(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text(
				"uid: ", fo.Mid, "\n",
				"名字: ", fo.Uname, "\n",
				"当前粉丝数: ", fo.Follower, "\n",
				"24h涨粉数: ", fo.Rise, "\n",
				"视频投稿数: ", fo.Video, "\n",
				"直播间id: ", fo.Roomid, "\n",
				"舰队: ", fo.GuardNum, "\n",
				"直播总排名: ", fo.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", fo.Mid, "\n",
				"数据获取时间: ", time.Now().Format("2006-01-02 15:04:05"),
			))
		})
}

// 搜索api：通过把触发指令传入的昵称找出uid返回
func uid(keyword string) (*gjson.Result, error) {
	api := "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&&user_type=1&keyword=" + keyword
	data, err := web.GetData(api)
	if err != nil {
		return nil, err
	}
	json := gjson.ParseBytes(data)
	if json.Get("data.numResults").Int() == 0 {
		return nil, errors.New("查无此人")
	}
	return &json, nil
}

type follower struct {
	Mid      int    `json:"mid"`
	Uname    string `json:"uname"`
	Video    int    `json:"video"`
	Roomid   int    `json:"roomid"`
	Rise     int    `json:"rise"`
	Follower int    `json:"follower"`
	GuardNum int    `json:"guardNum"`
	AreaRank int    `json:"areaRank"`
}

// 请求api
func fansapi(uid string) (*follower, error) {
	url := "https://api.vtbs.moe/v1/detail/" + uid
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	result := &follower{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}
