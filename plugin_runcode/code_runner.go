// Package runcode 基于 https://tool.runoob.com 的在线运行代码
package runcode

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"

	"github.com/tidwall/gjson"
)

var (
	templates = map[string]string{
		"py2":        "print 'Hello World!'",
		"ruby":       "puts \"Hello World!\";",
		"rb":         "puts \"Hello World!\";",
		"php":        "<?php\n\techo 'Hello World!';\n?>",
		"javascript": "console.log(\"Hello World!\");",
		"js":         "console.log(\"Hello World!\");",
		"node.js":    "console.log(\"Hello World!\");",
		"scala":      "object Main {\n  def main(args:Array[String])\n  {\n    println(\"Hello World!\")\n  }\n\t\t\n}",
		"go":         "package main\n\nimport \"fmt\"\n\nfunc main() {\n   fmt.Println(\"Hello, World!\")\n}",
		"c":          "#include <stdio.h>\n\nint main()\n{\n   printf(\"Hello, World! \n\");\n   return 0;\n}",
		"c++":        "#include <iostream>\nusing namespace std;\n\nint main()\n{\n   cout << \"Hello World\";\n   return 0;\n}",
		"cpp":        "#include <iostream>\nusing namespace std;\n\nint main()\n{\n   cout << \"Hello World\";\n   return 0;\n}",
		"java":       "public class HelloWorld {\n    public static void main(String []args) {\n       System.out.println(\"Hello World!\");\n    }\n}",
		"rust":       "fn main() {\n    println!(\"Hello World!\");\n}",
		"rs":         "fn main() {\n    println!(\"Hello World!\");\n}",
		"c#":         "using System;\nnamespace HelloWorldApplication\n{\n   class HelloWorld\n   {\n      static void Main(string[] args)\n      {\n         Console.WriteLine(\"Hello World!\");\n      }\n   }\n}",
		"cs":         "using System;\nnamespace HelloWorldApplication\n{\n   class HelloWorld\n   {\n      static void Main(string[] args)\n      {\n         Console.WriteLine(\"Hello World!\");\n      }\n   }\n}",
		"csharp":     "using System;\nnamespace HelloWorldApplication\n{\n   class HelloWorld\n   {\n      static void Main(string[] args)\n      {\n         Console.WriteLine(\"Hello World!\");\n      }\n   }\n}",
		"shell":      "echo 'Hello World!'",
		"bash":       "echo 'Hello World!'",
		"erlang":     "% escript will ignore the first line\n\nmain(_) ->\n    io:format(\"Hello World!~n\").",
		"perl":       "print \"Hello, World!\n\";",
		"python":     "print(\"Hello, World!\")",
		"py":         "print(\"Hello, World!\")",
		"swift":      "var myString = \"Hello, World!\"\nprint(myString)",
		"lua":        "var myString = \"Hello, World!\"\nprint(myString)",
		"pascal":     "runcode Hello;\nbegin\n  writeln ('Hello, world!')\nend.",
		"kotlin":     "fun main(args : Array<String>){\n    println(\"Hello World!\")\n}",
		"kt":         "fun main(args : Array<String>){\n    println(\"Hello World!\")\n}",
		"r":          "myString <- \"Hello, World!\"\nprint ( myString)",
		"vb":         "Module Module1\n\n    Sub Main()\n        Console.WriteLine(\"Hello World!\")\n    End Sub\n\nEnd Module",
		"typescript": "const hello : string = \"Hello World!\"\nconsole.log(hello)",
		"ts":         "const hello : string = \"Hello World!\"\nconsole.log(hello)",
	}
	table = map[string][2]string{
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
)

func init() {
	control.Register("runcode", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "在线代码运行: \n" +
			">runcode [language] [code block]\n" +
			"模板查看: \n" +
			">runcode [language] help\n" +
			"支持语种: \n" +
			"Go || Python || C/C++ || C# || Java || Lua \n" +
			"JavaScript || TypeScript || PHP || Shell \n" +
			"Kotlin  || Rust || Erlang || Ruby || Swift \n" +
			"R || VB || Py2 || Perl || Pascal || Scala",
	}).ApplySingle(ctxext.DefaultSingle).OnRegex(`^>runcode\s(.+?)\s([\s\S]+)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			language := ctx.State["regex_matched"].([]string)[1]
			language = strings.ToLower(language)
			if runType, exist := table[language]; !exist {
				// 不支持语言
				ctx.SendChain(
					message.Text("> ", ctx.Event.Sender.NickName, "\n"),
					message.Text("语言不是受支持的编程语种呢~"),
				)
			} else {
				// 执行运行
				block := ctx.State["regex_matched"].([]string)[2]
				block = message.UnescapeCQCodeText(block)
				if block == "help" {
					// 输出模板
					ctx.SendChain(
						message.Text("> ", ctx.Event.Sender.NickName, "  ", language, "-template:\n"),
						message.Text(
							">runcode ", language, "\n",
							templates[language],
						),
					)
				} else {
					if output, err := runCode(block, runType); err != nil {
						// 运行失败
						ctx.SendChain(
							message.Text("> ", ctx.Event.Sender.NickName, "\n"),
							message.Text("ERROR: ", err),
						)
					} else {
						// 运行成功
						ctx.SendChain(
							message.Text("> ", ctx.Event.Sender.NickName, "\n"),
							message.Text(output),
						)
					}
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
		Timeout: 15 * time.Second,
	}
	request, _ := http.NewRequest("POST", api, strings.NewReader(val.Encode()))
	request.Header = header
	body, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer body.Body.Close()
	if body.StatusCode != http.StatusOK {
		return "", errors.New("code not 200")
	}
	res, err := ioutil.ReadAll(body.Body)
	if err != nil {
		return "", err
	}
	// 结果处理
	content := gjson.ParseBytes(res)
	if e := content.Get("errors").Str; e != "\n\n" {
		return "", errors.New(cutTooLong(clearNewLineSuffix(e)))
	}
	output := content.Get("output").Str

	return cutTooLong(clearNewLineSuffix(output)), nil
}

// 清除末尾多余的换行符
func clearNewLineSuffix(text string) string {
	for strings.HasSuffix(text, "\n") {
		text = text[:len(text)-1]
	}
	return text
}

// 截断过长文本
func cutTooLong(text string) string {
	temp := []rune(text)
	count := 0
	for i := range temp {
		switch {
		case temp[i] == 13 && i < len(temp)-1 && temp[i+1] == 10:
			// 匹配 \r\n 跳过，等 \n 自己加
		case temp[i] == 10:
			count++
		case temp[i] == 13:
			count++
		}
		if count > 30 || i > 1000 {
			temp = append(temp[:i-1], []rune("\n............\n............")...)
			break
		}
	}
	return string(temp)
}
