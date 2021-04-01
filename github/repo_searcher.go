package github

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	zero.OnRegex(`>G\s(.*)`).SetBlock(true).SetPriority(0).
		Handle(func(ctx *zero.Ctx) {
			api, _ := url.Parse("https://api.github.com/search/repositories")
			params := url.Values{}
			params.Set("q", ctx.State["regex_matched"].([]string)[1])
			api.RawQuery = params.Encode()
			link := api.String()

			client := &http.Client{}

			req, err := http.NewRequest("GET", link, nil)
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
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
			count := gjson.ParseBytes(body).Get("total_count").Int()
			if count == 0 {
				ctx.Send("没有找到这样的仓库")
				return
			}
			repo := gjson.ParseBytes(body).Get("items.0")
			language := repo.Get("language").Str
			if language == "" {
				language = "None"
			}
			license := strings.ToUpper(repo.Get("license.key").Str)
			if license == "" {
				license = "None"
			}
			id := ctx.Send(fmt.Sprintf(
				"%s: \nDescription: %s\nStar/Fork/Issue: %d/%d/%d\nLanguage: %s\nLicense: %s\nLast pushed: %s\nJump: %s",
				repo.Get("full_name").Str,
				repo.Get("description").Str,
				repo.Get("watchers").Int(),
				repo.Get("forks").Int(),
				repo.Get("open_issues").Int(),
				language,
				license,
				repo.Get("updated_at").Str,
				repo.Get("html_url").Str,
			))
			if id == 0 {
				ctx.Send("ERROR: 可能被风控，发送失败")
			}
		})
}
