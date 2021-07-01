package plugin_bilibili

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnRegex(`^>bili info\s?(.{1,25})$`).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			res, err := uid(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			id := res.Get("data.result.0.mid").Int()
			// 获取详情
			api := fmt.Sprintf("https://api.vtbs.moe/v1/detail/%d", id)
			resp, err := http.Get(api)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				ctx.SendChain(message.Text("ERROR: code ", resp.StatusCode))
				return
			}
			data, _ := ioutil.ReadAll(resp.Body)
			json := gjson.ParseBytes(data)
			ctx.SendChain(message.Text(
				"uid: ", json.Get("mid").Int(), "\n",
				"名字: ", json.Get("uname").Str, "\n",
				"当前粉丝数: ", json.Get("follower").Int(), "\n",
				"24h涨粉数: ", json.Get("rise").Int(), "\n",
				"视频投稿数: ", json.Get("video").Int(), "\n",
				"直播间id: ", json.Get("roomid").Int(), "\n",
				"舰队: ", json.Get("guardNum").Int(), "\n",
				"直播总排名: ", json.Get("areaRank").Int(), "\n",
				"数据来源: ", "https://vtbs.moe/detail/", uid, "\n",
				"数据获取时间: ", time.Now().Format("2006-01-02 15:04:05"),
			))
		})
}

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
