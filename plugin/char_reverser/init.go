// Package charreverser 英文字符反转
package charreverser

import (
	"regexp"
	"strings"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const commandRegex = `[A-z]{1}([A-z]|\s)+[A-z]{1}` // 命令正则表达式

var (
	charMap = map[rune]rune{
		'a': 'ɐ',
		'b': 'q',
		'c': 'ɔ',
		'd': 'p',
		'e': 'ǝ',
		'f': 'ɟ',
		'g': 'ƃ',
		'h': 'ɥ',
		'i': 'ᴉ',
		'j': 'ɾ',
		'k': 'ʞ',
		'l': 'l',
		'm': 'ɯ',
		'n': 'u',
		'o': 'o',
		'p': 'd',
		'q': 'b',
		'r': 'ɹ',
		's': 's',
		't': 'ʇ',
		'u': 'n',
		'v': 'ʌ',
		'w': 'ʍ',
		'x': 'x',
		'y': 'ʎ',
		'z': 'z',
		'A': '∀',
		'B': 'ᗺ',
		'C': 'Ɔ',
		'D': 'ᗡ',
		'E': 'Ǝ',
		'F': 'Ⅎ',
		'G': '⅁',
		'H': 'H',
		'I': 'I',
		'J': 'ſ',
		'K': 'ʞ',
		'L': '˥',
		'M': 'W',
		'N': 'N',
		'O': 'O',
		'P': 'Ԁ',
		'Q': 'Ò',
		'R': 'ᴚ',
		'S': 'S',
		'T': '⏊',
		'U': '∩',
		'V': 'Λ',
		'W': 'M',
		'X': 'X',
		'Y': '⅄',
		'Z': 'Z',
	}

	compiledRegex = regexp.MustCompile(commandRegex)
)

func init() {
	// 初始化engine
	engine := control.Register(
		"charreverser",
		&ctrl.Options[*zero.Ctx]{
			DisableOnDefault: false,
			Help:             "字符翻转\n -翻转 <英文字符串>",
		},
	)
	// 处理字符翻转指令
	engine.OnRegex(`翻转( )+[A-z]{1}([A-z]|\s)+[A-z]{1}`).SetBlock(true).Handle(
		func(ctx *zero.Ctx) {
			// 获取需要翻转的字符串
			results := compiledRegex.FindAllString(ctx.MessageString(), -1)
			str := results[0]

			// 将字符顺序翻转
			var tempBuilder strings.Builder
			for i := len(str) - 1; i >= 0; i-- {
				tempBuilder.WriteByte(str[i])
			}

			// 翻转字符字形
			var reversedStrBuilder strings.Builder
			for _, char := range tempBuilder.String() {
				if char != ' ' {
					reversedStrBuilder.WriteRune(charMap[char])
				} else {
					reversedStrBuilder.WriteRune(' ')
				}
			}

			// 发送翻转后的字符串
			ctx.SendChain(message.Text(reversedStrBuilder.String()))
		},
	)
}
