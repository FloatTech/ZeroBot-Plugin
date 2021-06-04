package msgext

import (
	"github.com/wdvxdr1123/ZeroBot/message"
)

//@全体成员
func AtAll() message.MessageSegment {
	return message.MessageSegment{
		Type: "at",
		Data: map[string]string{
			"qq": "all",
		},
	}
}

//无缓存发送图片
func ImageNoCache(url string) message.MessageSegment {
	return message.MessageSegment{
		Type: "image",
		Data: map[string]string{
			"file":  url,
			"cache": "0",
		},
	}
}
