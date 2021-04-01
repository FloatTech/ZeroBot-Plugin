package setutime

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

var limit = rate.NewManager(time.Minute*1, 3)

func init() { // 插件主体
	zero.OnRegex(`^点歌(.*)$`).SetBlock(true).SetPriority(50).
		Handle(func(ctx *zero.Ctx) {
			if limit.Load(ctx.Event.UserID).Acquire() == false {
				ctx.Send("请稍后重试0x0...")
				return
			}
			// 调用网易云 API
			var api = "http://music.163.com/api/search/pc"

			client := &http.Client{}

			// 包装请求参数
			data := url.Values{}
			data.Set("offset", "0")
			data.Set("total", "true")
			data.Set("limit", "9")
			data.Set("type", "1")
			data.Set("s", ctx.State["regex_matched"].([]string)[1])
			fromData := strings.NewReader(data.Encode())

			// 网络请求
			req, err := http.NewRequest("POST", api, fromData)
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
			resp, err := client.Do(req)
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}

			if code := resp.StatusCode; code != 200 {
				// 如果返回不是200则立刻抛出错误
				ctx.Send(fmt.Sprintf("ERROR: code %d", code))
				return
			}
			content := gjson.ParseBytes(body).Get("result.songs.0")

			// 发送搜索结果
			ctx.Send(
				fmt.Sprintf(
					"[CQ:music,type=%s,url=%s,audio=%s,title=%s,content=%s,image=%s]",
					"custom",
					fmt.Sprintf("http://y.music.163.com/m/song?id=%d", content.Get("id").Int()),
					fmt.Sprintf("http://music.163.com/song/media/outer/url?id=%d.mp3", content.Get("id").Int()),
					content.Get("name").Str,
					content.Get("artists.0.name").Str,
					content.Get("album.blurPicUrl").Str,
				),
			)
			return
		})
}
