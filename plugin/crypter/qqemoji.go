// Package crypter QQ表情加解密
package crypter

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	emojiZeroID = 297
	emojiOneID  = 424
)

func encodeQQEmoji(text string) message.Message {
	if text == "" {
		return message.Message{message.Text("请输入要加密的文本")}
	}

	var bin strings.Builder
	for _, b := range []byte(text) {
		fmt.Fprintf(&bin, "%08b", b)
	}

	s := bin.String()
	msg := make(message.Message, 0, len(s))
	for _, bit := range s {
		if bit == '0' {
			msg = append(msg, message.Face(emojiZeroID))
		} else {
			msg = append(msg, message.Face(emojiOneID))
		}
	}
	return msg
}

func decodeQQEmoji(faceIDs []int) string {
	var bin strings.Builder
	for _, id := range faceIDs {
		if id == emojiZeroID {
			bin.WriteByte('0')
		} else if id == emojiOneID {
			bin.WriteByte('1')
		}
	}
	binary := bin.String()
	if len(binary) == 0 || len(binary)%8 != 0 {
		return "QQ表情密文格式错误"
	}

	data := make([]byte, len(binary)/8)
	for i := range data {
		for j := 0; j < 8; j++ {
			if binary[i*8+j] == '1' {
				data[i] |= 1 << (7 - j)
			}
		}
	}

	if !utf8.Valid(data) {
		return "QQ表情解密失败：结果不是有效文本"
	}
	return string(data)
}
