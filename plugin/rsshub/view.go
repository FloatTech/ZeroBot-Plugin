package rsshub

import (
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/rsshub/domain"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
	"time"
)

const (
	rssHubPushErrMsg = "RssHub推送错误"
)

// formatRssToTextMsg 格式化RssClientView为文本消息
func formatRssToTextMsg(view *domain.RssClientView) (msg []string) {
	msg = make([]string, 0)
	// rssChannel信息
	msgStr := fmt.Sprintf("%s\n更新时间:%v\n", view.Source.Title, view.Source.UpdatedParsed.Format(time.ANSIC))
	msg = append(msg, msgStr)
	// rssItem信息
	for _, item := range view.Contents {
		contentStr := fmt.Sprintf("%s\n%s\n", item.Title, item.Link)
		// Date为空时不显示
		if !item.Date.IsZero() {
			contentStr += fmt.Sprintf("更新时间：%v\n", item.Date.Format(time.ANSIC))
		}
		msg = append(msg, contentStr)
	}
	return
}

func formatRssToMsg(view *domain.RssClientView) ([]message.Message, error) {
	fv := make([]message.Message, len(view.Contents)+1)
	// 订阅源头图
	toastPic, err := text.RenderToBase64(fmt.Sprintf("%s\n\n\n更新时间:%v\n", view.Source.Title, view.Source.UpdatedParsed.Format(time.ANSIC)), text.SakuraFontFile, 800, 40)
	if err != nil {
		return nil, err
	}
	fv[0] = message.Message{message.Image("base64://" + binary.BytesToString(toastPic))}
	// 元素信息
	for idx, item := range view.Contents {
		itemMessage := message.Message{}
		contentStr := fmt.Sprintf("%s\n\n\n", item.Title)
		// Date为空时不显示
		if !item.Date.IsZero() {
			contentStr += fmt.Sprintf("更新时间：\n%v\n", item.Date.Format(time.ANSIC))
		}
		var content []byte
		content, err = text.RenderToBase64(contentStr, text.SakuraFontFile, 800, 40)
		if err != nil {
			logrus.WithError(err).Error("RssHub订阅姬渲染图片失败")
			continue
		}
		itemMessage = append(itemMessage, message.Image("base64://"+binary.BytesToString(content)), message.Text(fmt.Sprintf("\n%s", item.Link)))
		fv[idx+1] = itemMessage
	}
	return fv, nil
}

// formatRssToPicMsg 格式化RssClientView为图片消息
//func formatRssToPicMsg(view *domain.RssClientView) (content []byte, err error) {
//	msg := fmt.Sprintf("【%s】更新时间:%v\n", view.Source.Title, view.Source.UpdatedParsed.Format(time.ANSIC))
//	// rssItem信息
//	for _, item := range view.Contents {
//		contentStr := fmt.Sprintf("标题：%s\n链接：%s\n", item.Title, item.Link)
//		if !item.Date.IsZero() {
//			contentStr += fmt.Sprintf("更新时间：%v\n", item.Date.Format(time.ANSIC))
//		}
//		msg += contentStr
//	}
//	content, err = text.RenderToBase64(msg, text.FontFile, 800, 20)
//	if err != nil {
//		return
//	}
//	return
//}

// fakeSenderForwardNode 伪造一个发送者为RssHub订阅姬的消息节点，传入userID是为了减少ws io
func fakeSenderForwardNode(userID int64, msgs ...message.MessageSegment) message.MessageSegment {
	return message.CustomNode(
		"RssHub订阅姬",
		userID,
		msgs)
}
