// Package runcode 基于 https://tool.runoob.com 的在线运行代码
package runcode

import (
	"strings"

	"github.com/FloatTech/AnimeAPI/runoob"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var ro = runoob.NewRunOOB("066417defb80d038228de76ec581a50a")

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "在线代码运行",
		Help: ">runcode [language] [code block]\n" +
			"模板查看: \n" +
			">runcode [language] help\n" +
			"支持语种: \n" +
			"Go || Python || C/C++ || C# || Java || Lua \n" +
			"JavaScript || TypeScript || PHP || Shell \n" +
			"Kotlin  || Rust || Erlang || Ruby || Swift \n" +
			"R || VB || Py2 || Perl || Pascal || Scala",
	}).ApplySingle(ctxext.DefaultSingle).OnRegex(`^>runcode(raw)?\s(.+?)\s([\s\S]+)$`).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			israw := ctx.State["regex_matched"].([]string)[1] != ""
			language := ctx.State["regex_matched"].([]string)[2]
			language = strings.ToLower(language)
			if _, exist := runoob.LangTable[language]; !exist {
				// 不支持语言
				ctx.SendChain(
					message.Text("> ", ctx.Event.Sender.NickName, "\n"),
					message.Text("语言不是受支持的编程语种呢~"),
				)
			} else {
				// 执行运行
				block := message.UnescapeCQText(ctx.State["regex_matched"].([]string)[3])
				switch block {
				case "help":
					ctx.SendChain(
						message.Text("> ", ctx.Event.Sender.NickName, "  ", language, "-template:\n"),
						message.Text(
							">runcode ", language, "\n",
							runoob.Templates[language],
						),
					)
				default:
					if output, err := ro.Run(block, language, ""); err != nil {
						// 运行失败
						ctx.SendChain(
							message.Text("> ", ctx.Event.Sender.NickName, "\n"),
							message.Text("ERROR: ", err),
						)
					} else {
						// 运行成功
						output = cutTooLong(strings.Trim(output, "\n"))
						if israw && zero.AdminPermission(ctx) {
							ctx.SendChain(message.Text(output))
						} else {
							ctx.SendChain(
								message.Text("> ", ctx.Event.Sender.NickName, "\n"),
								message.Text(output),
							)
						}
					}
				}
			}
		})
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
