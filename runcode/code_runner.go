package runcode

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	RunAllow := true
	table := map[string][2]string{
		"py2":        {"0", "py"},
		"ruby":       {"1", "rb"},
		"rb":         {"1", "rb"},
		"php":        {"3", "php"},
		"javascript": {"4", "js"},
		"js":         {"4", "js"},
		"node.js":    {"4", "js"},
		"scala":      {"5", "scala"},
		"go":         {"6", "go"},
		"c":          {"7", "c"},
		"c++":        {"7", "cpp"},
		"cpp":        {"7", "cpp"},
		"java":       {"8", "java"},
		"rust":       {"9", "rs"},
		"rs":         {"9", "rs"},
		"c#":         {"10", "cs"},
		"cs":         {"10", "cs"},
		"csharp":     {"10", "cs"},
		"shell":      {"10", "sh"},
		"bash":       {"10", "sh"},
		"erlang":     {"12", "erl"},
		"perl":       {"14", "pl"},
		"python":     {"15", "py3"},
		"py":         {"15", "py3"},
		"swift":      {"16", "swift"},
		"lua":        {"17", "lua"},
		"pascal":     {"18", "pas"},
		"kotlin":     {"19", "kt"},
		"kt":         {"19", "kt"},
		"r":          {"80", "r"},
		"vb":         {"84", "vb"},
		"typescript": {"1010", "ts"},
		"ts":         {"1010", "ts"},
	}

	zero.OnFullMatch(">runcode help").SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text(
				"使用说明: ", "\n",
				">runcode [language] [code block]", "\n",
				"支持语种: ", "\n",
				"Go || Python || C/C++ || C# || Java || Lua ", "\n",
				"JavaScript || TypeScript || PHP || Shell ", "\n",
				"Kotlin  || Rust || Erlang || Ruby || Swift ", "\n",
				"R || VB || Py2 || Perl || Pascal || Scala ", "\n",
			))
		})
	zero.OnFullMatch(">runcode on", zero.AdminPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			RunAllow = true
			ctx.SendChain(
				message.Text("> ", ctx.Event.Sender.NickName, "\n"),
				message.Text("在线运行代码功能已启用"),
			)
		})

	zero.OnFullMatch(">runcode off", zero.AdminPermission).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			RunAllow = false
			ctx.SendChain(
				message.Text("> ", ctx.Event.Sender.NickName, "\n"),
				message.Text("在线运行代码功能已禁用"),
			)
		})

	zero.OnRegex(`>runcode\s(.+?)\s([\s\S]+)`).SetBlock(true).SecondPriority().
		Handle(func(ctx *zero.Ctx) {
			language := ctx.State["regex_matched"].([]string)[1]
			language = strings.ToLower(language)
			if runType, exist := table[language]; !exist {
				// 不支持语言
				ctx.SendChain(
					message.Text("> ", ctx.Event.Sender.NickName, "\n"),
					message.Text("语言不是受支持的编程语种呢~"),
				)
				return
			} else {
				if RunAllow == false {
					// 运行代码被禁用
					ctx.SendChain(
						message.Text("> ", ctx.Event.Sender.NickName, "\n"),
						message.Text("在线运行代码功能已被禁用"),
					)
					return
				}
				// 执行运行
				block := ctx.State["regex_matched"].([]string)[2]
				block = message.UnescapeCQCodeText(block)
				if output, err := runCode(block, runType); err != nil {
					// 运行失败
					ctx.SendChain(
						message.Text("> ", ctx.Event.Sender.NickName, "\n"),
						message.Text("ERROR: ", err),
					)
					return
				} else {
					// 运行成功
					ctx.SendChain(
						message.Text("> ", ctx.Event.Sender.NickName, "\n"),
						message.Text(output),
					)
					return
				}
			}
		})
}

func runCode(code string, runType [2]string) (string, error) {
	// 对菜鸟api发送数据并返回结果
	api := "https://tool.runoob.com/compile2.php"

	header := http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded; charset=UTF-8"},
		"Referer":      []string{"https://c.runoob.com/"},
		"User-Agent":   []string{"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:87.0) Gecko/20100101 Firefox/87.0"},
	}

	val := url.Values{
		"code":     []string{code},
		"token":    []string{"4381fe197827ec87cbac9552f14ec62a"},
		"stdin":    []string{""},
		"language": []string{runType[0]},
		"fileext":  []string{runType[1]},
	}
	// 发送请求
	client := &http.Client{
		Timeout: time.Duration(6 * time.Second),
	}
	request, _ := http.NewRequest("POST", api, strings.NewReader(val.Encode()))
	request.Header = header
	body, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer body.Body.Close()
	if body.StatusCode != http.StatusOK {
		return "", fmt.Errorf("code %d", body.StatusCode)
	}
	res, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return "", err
	}
	// 结果处理
	content := gjson.ParseBytes(res)
	if e := content.Get("errors").Str; e != "\n\n" {
		return "", fmt.Errorf(e)
	}
	output := content.Get("output").Str
	return output[:len(output)-1], nil
}
