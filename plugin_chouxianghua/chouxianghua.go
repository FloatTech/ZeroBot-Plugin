package chouxianghua

import (
	"github.com/FloatTech/ZeroBot-Plugin/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	prio = 10
)

var (
	engine = control.Register("chouxianghua", &control.Options{
		DisableOnDefault: false,
		Help:             "抽象话\n- 抽象翻译\n",
	})
)

func init() {
	engine.OnRegex("^抽象翻译([\u4E00-\u9FA5A-Za-z0-9]{1,25})$").SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			cxresult := chouXiang(ctx.State["regex_matched"].([]string)[1])
			ctx.SendChain(message.Text(cxresult))
		})
}

func chouXiang(s string) (cxresult string) {
	h := []rune(s)
	for i := 0; i < len(h); i++ {
		if i < len(h)-1 && (getEmojiByPronunciation(getPronunciationByWord(string(h[i])).Pronunciation+getPronunciationByWord(string(h[i+1])).Pronunciation).Emoji != "") {
			cxresult += getEmojiByPronunciation(getPronunciationByWord(string(h[i])).Pronunciation + getPronunciationByWord(string(h[i+1])).Pronunciation).Emoji
			i++
		} else if getEmojiByPronunciation(getPronunciationByWord(string(h[i])).Pronunciation).Emoji != "" {
			cxresult += getEmojiByPronunciation(getPronunciationByWord(string(h[i])).Pronunciation).Emoji
		} else {
			cxresult += string(h[i])
		}
	}
	return
}
