package rsshub

import (
	"fmt"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/rsshub/domain"
)

const (
	rssHubPushErrMsg = "RssHub推送错误"
)

// formatRssViewToMessagesSlice 格式化RssClientView为消息切片
func formatRssViewToMessagesSlice(view *domain.RssClientView) ([]message.Message, error) {
	// 取前20条
	cts := view.Contents
	if len(cts) > 20 {
		cts = cts[:20]
	}
	// 2n+1条消息
	fv := make([]message.Message, len(cts)*2+1)
	// 订阅源头图
	toastPic, err := text.RenderToBase64(fmt.Sprintf("%s\n\n\n%s\n\n\n更新时间:%v\n\n\n",
		view.Source.Title, view.Source.Link, view.Source.UpdatedParsed.Local().Format(time.DateTime)),
		text.SakuraFontFile, 1200, 40)
	if err != nil {
		return nil, err
	}
	fv[0] = message.Message{message.Image("base64://" + binary.BytesToString(toastPic))}
	// 元素信息
	for idx, item := range cts {
		contentStr := fmt.Sprintf("%s\n\n\n", item.Title)
		// Date为空时不显示
		if !item.Date.IsZero() {
			contentStr += fmt.Sprintf("更新时间：\n%v\n", item.Date.Local().Format(time.DateTime))
		}
		var content []byte
		content, err = text.RenderToBase64(contentStr, text.SakuraFontFile, 1200, 40)
		if err != nil {
			logrus.WithError(err).Error("RssHub订阅姬渲染图片失败")
			continue
		}
		itemMessagePic := message.Message{message.Image("base64://" + binary.BytesToString(content))}
		fv[2*idx+1] = itemMessagePic
		fv[2*idx+2] = message.Message{message.Text(item.Link)}
	}
	return fv, nil
}

// newRssSourcesMsg Rss订阅源列表
func newRssSourcesMsg(ctx *zero.Ctx, view []*domain.RssClientView) (message.Message, error) {
	var msgSlice []message.Message
	// 生成消息
	for _, v := range view {
		if v == nil {
			continue
		}
		item, err := formatRssViewToMessagesSlice(v)
		if err != nil {
			return nil, err
		}
		msgSlice = append(msgSlice, item...)
	}
	// 伪造一个发送者为RssHub订阅姬的消息节点
	msg := make(message.Message, len(msgSlice))
	for i, item := range msgSlice {
		msg[i] = fakeSenderForwardNode(ctx.Event.SelfID, item...)
	}
	return msg, nil
}

// newRssDetailsMsg Rss订阅源详情（包含文章信息列表）
func newRssDetailsMsg(ctx *zero.Ctx, view *domain.RssClientView) (message.Message, error) {
	// 生成消息
	msgSlice, err := formatRssViewToMessagesSlice(view)
	if err != nil {
		return nil, err
	}
	// 伪造一个发送者为RssHub订阅姬的消息节点
	msg := make(message.Message, len(msgSlice))
	for i, item := range msgSlice {
		msg[i] = fakeSenderForwardNode(ctx.Event.SelfID, item...)
	}
	return msg, nil
}

// fakeSenderForwardNode 伪造一个发送者为RssHub订阅姬的消息节点
func fakeSenderForwardNode(userID int64, msgs ...message.Segment) message.Segment {
	return message.CustomNode(
		"RssHub订阅姬",
		userID,
		msgs)
}
