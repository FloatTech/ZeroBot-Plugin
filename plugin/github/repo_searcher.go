// Package github GitHub 仓库搜索
package github

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/tidwall/gjson"
)

func init() { // 插件主体
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "GitHub仓库搜索",
		Help: "- >github [xxx]\n" +
			"- >github -p [xxx]",
	}).OnRegex(`^>github\s(-.{1,10}? )?(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 发送请求
			header := http.Header{
				"User-Agent": []string{"User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"},
			}
			api, _ := url.Parse("https://api.github.com/search/repositories")
			api.RawQuery = url.Values{
				"q": []string{ctx.State["regex_matched"].([]string)[2]},
			}.Encode()
			body, err := netGet(api.String(), header)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			// 解析请求
			info := gjson.ParseBytes(body)
			if info.Get("total_count").Int() == 0 {
				ctx.SendChain(message.Text("ERROR: 没有找到这样的仓库"))
				return
			}
			repo := info.Get("items.0")
			// 发送结果
			switch ctx.State["regex_matched"].([]string)[1] {
			case "-p ": // 图片模式
				ctx.SendChain(
					message.Image(
						"https://opengraph.githubassets.com/0/"+repo.Get("full_name").Str,
					).Add("cache", 0),
				)
			case "-t ": // 文字模式
				ctx.SendChain(
					message.Text(
						repo.Get("full_name").Str, "\n",
						"Description: ",
						repo.Get("description").Str, "\n",
						"Star/Fork/Issue: ",
						repo.Get("watchers").Int(), "/", repo.Get("forks").Int(), "/", repo.Get("open_issues").Int(), "\n",
						"Language: ",
						notnull(repo.Get("language").Str), "\n",
						"License: ",
						notnull(strings.ToUpper(repo.Get("license.key").Str)), "\n",
						"Last pushed: ",
						repo.Get("pushed_at").Str, "\n",
						"Jump: ",
						repo.Get("html_url").Str, "\n",
					),
				)
			default: // 文字模式
				ctx.SendChain(
					message.Text(
						repo.Get("full_name").Str, "\n",
						"Description: ",
						repo.Get("description").Str, "\n",
						"Star/Fork/Issue: ",
						repo.Get("watchers").Int(), "/", repo.Get("forks").Int(), "/", repo.Get("open_issues").Int(), "\n",
						"Language: ",
						notnull(repo.Get("language").Str), "\n",
						"License: ",
						notnull(strings.ToUpper(repo.Get("license.key").Str)), "\n",
						"Last pushed: ",
						repo.Get("pushed_at").Str, "\n",
						"Jump: ",
						repo.Get("html_url").Str, "\n",
					),
					message.Image(
						"https://opengraph.githubassets.com/0/"+repo.Get("full_name").Str,
					).Add("cache", 0),
				)
			}
		})
}

// notnull 如果传入文本为空，则返回默认值

func notnull(text string) string {
	if text == "" {
		return "None"
	}
	return text
}

// netGet 返回请求结果
func netGet(dest string, header http.Header) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", dest, nil)
	if err != nil {
		return nil, err
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if code := resp.StatusCode; code != 200 {
		// 如果返回不是200则立刻抛出错误
		errmsg := fmt.Sprintf("code %d", code)
		return nil, errors.New(errmsg)
	}
	return body, nil
}
