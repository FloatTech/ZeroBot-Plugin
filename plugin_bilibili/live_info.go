package bilibili

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 查vup粉丝数据
func init() {
	engine.OnRegex(`^>vup info\s?(.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			res, err := uid(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			id := res.Get("data.result.0.mid").String()
			// 获取详情
			json := fansapi(id)
			ctx.SendChain(message.Text(
				"uid: ", json.Mid, "\n",
				"名字: ", json.Uname, "\n",
				"当前粉丝数: ", json.Follower, "\n",
				"24h涨粉数: ", json.Rise, "\n",
				"视频投稿数: ", json.Video, "\n",
				"直播间id: ", json.Roomid, "\n",
				"舰队: ", json.GuardNum, "\n",
				"直播总排名: ", json.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", json.Mid, "\n",
				"数据获取时间: ", time.Now().Format("2006-01-02 15:04:05"),
			))
		})
}

// 搜索api：通过把触发指令传入的昵称找出uid返回
func uid(keyword string) (gjson.Result, error) {
	api := "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&&user_type=1&keyword=" + keyword
	resp, err := http.Get(api)
	if err != nil {
		return gjson.Result{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return gjson.Result{}, errors.New("code not 200")
	}
	data, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	json := gjson.ParseBytes(data)
	if json.Get("data.numResults").Int() == 0 {
		return gjson.Result{}, errors.New("查无此人")
	}
	return json, nil
}
