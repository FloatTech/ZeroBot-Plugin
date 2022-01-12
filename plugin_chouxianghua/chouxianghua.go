// Package chouxianghua 抽象话转化
package chouxianghua

import (
	control "github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const prio = 10

func init() {
	control.Register("chouxianghua", &control.Options{
		DisableOnDefault: false,
		Help:             "抽象话\n- 抽象翻译xxx",
	}).OnRegex("^抽象翻译((\\s|[\\r\\n]|[\\p{Han}\\p{P}A-Za-z0-9])+)$").SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			r := cx(ctx.State["regex_matched"].([]string)[1])
			ctx.SendChain(message.Text(r))
		})
}

func cx(s string) (r string) {
	h := []rune(s)
	for i := 0; i < len(h); i++ {
		if i < len(h)-1 {
			e := getEmojiByPronun(getPronunByDWord(h[i], h[i+1]))
			if e != "" {
				r += e
				i++
				continue
			}
		}
		e := getEmojiByPronun(getPinyinByWord(string(h[i])))
		if e != "" {
			r += e
			continue
		}
		r += string(h[i])
	}
	return
}
