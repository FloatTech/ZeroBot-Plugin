package aireply

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type XiaoAiReply struct{}

// DealQuestion 把椛椛替换为小爱
func (*XiaoAiReply) DealQuestion(preMsg string) (msg string) {
	msg = strings.ReplaceAll(preMsg, zero.BotConfig.NickName[0], xiaoaiBotName)
	return msg
}

// GetReply 取得回复消息
func (*XiaoAiReply) GetReply(msg string) (reply string) {
	u := fmt.Sprintf(xiaoaiURL, url.QueryEscape(msg))
	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Errorln("[aireply-xiaoai]:", err)
		return ""
	}
	// 自定义Header
	req.Header.Set("User-Agent", getAgent())
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "81.70.100.130")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("[aireply-xiaoai]:", err)
		return
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln("[aireply-xiaoai]:", err)
		return
	}
	reply = helper.BytesToString(bytes)
	log.Println("reply:", reply)
	return
}

// DealReply 处理回复消息
func (*XiaoAiReply) DealReply(reply string) (textReply string, faceReply int) {
	textReply = strings.ReplaceAll(reply, xiaoaiBotName, zero.BotConfig.NickName[0])
	textReply = strings.ReplaceAll(textReply, "小米智能助理", "电子宠物")
	faceReply = -1
	return
}
