package utils

import (
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"time"
)

func RandText(text ...[]string) message.MessageSegment {
	length := len(text)
	rand.Seed(time.Now().UnixNano())
	return message.Text(text[rand.Intn(length)])
}

func Suiji() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(30)
}
