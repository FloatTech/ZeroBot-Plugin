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

// formatRssViewToMessagesSlice 格式化RssClientView为消息切片
func formatRssViewToMessagesSlice(view *domain.RssClientView) ([]message.Message, error) {
	// 2n+1条消息，如果太长就截短到50
	cts := view.Contents
	//if len(cts) > 20 {
	//	cts = cts[:20]
	//}
	fv := make([]message.Message, len(cts)*2+1)
	// 订阅源头图
	toastPic, err := text.RenderToBase64(fmt.Sprintf("%s\n\n\n%s\n\n\n更新时间:%v\n\n\n",
		view.Source.Title, view.Source.Link, view.Source.UpdatedParsed.Format(time.DateTime)),
		text.SakuraFontFile, 800, 40)
	if err != nil {
		return nil, err
	}
	fv[0] = message.Message{message.Image("base64://" + binary.BytesToString(toastPic))}
	// 元素信息
	for idx, item := range cts {
		contentStr := fmt.Sprintf("%s\n\n\n", item.Title)
		// Date为空时不显示
		if !item.Date.IsZero() {
			contentStr += fmt.Sprintf("更新时间：\n%v\n", item.Date.Format(time.DateTime))
		}
		var content []byte
		content, err = text.RenderToBase64(contentStr, text.SakuraFontFile, 800, 40)
		if err != nil {
			logrus.WithError(err).Error("RssHub订阅姬渲染图片失败")
			continue
		}
		itemMessagePic := message.Message{message.Image("base64://" + binary.BytesToString(content))}
		fv[2*idx+1] = itemMessagePic
		fv[2*idx+2] = message.Message{message.Text(fmt.Sprintf("%s", item.Link))}
	}
	return fv, nil
}
