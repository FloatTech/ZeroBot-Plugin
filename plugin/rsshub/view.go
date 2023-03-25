package rsshub

import (
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/rsshub/domain"
	"github.com/wdvxdr1123/ZeroBot/message"
	"time"
)

// formatRssToTextMsg 格式化RssClientView为文本消息
func formatRssToTextMsg(view *domain.RssClientView) (msg []string) {
	msg = make([]string, 0)
	// rssChannel信息
	msgStr := fmt.Sprintf("【%s】更新时间:%v\n", view.Source.Title, view.Source.UpdatedParsed.Format(time.ANSIC))
	msg = append(msg, msgStr)
	// rssItem信息
	for _, item := range view.Contents {
		contentStr := fmt.Sprintf("标题：%s\n链接：%s\n更新时间：%v\n", item.Title, item.Link, item.Date.Format(time.ANSIC))
		msg = append(msg, contentStr)
	}
	return
}

// fakeSenderForwardNode 伪造一个发送者为RssHub订阅姬的消息节点，传入userID是为了减少ws io
func fakeSenderForwardNode(userID int64, msgs ...message.MessageSegment) message.MessageSegment {
	return message.CustomNode(
		"RssHub订阅姬",
		userID,
		msgs)
}
