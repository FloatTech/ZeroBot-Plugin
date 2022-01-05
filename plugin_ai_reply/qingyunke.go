package aireply

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type QYKReply struct{}

func (*QYKReply) DealQuestion(preMsg string) (msg string) {
	msg = strings.ReplaceAll(preMsg, zero.BotConfig.NickName[0], qykBotName)
	return msg
}

func (*QYKReply) GetReply(msg string) (reply string) {
	u := fmt.Sprintf(qykURL, url.QueryEscape(msg))
	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Errorln("[aireply-qingyunke]:", err)
		return ""
	}
	// 自定义Header
	req.Header.Set("User-Agent", getAgent())
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "api.qingyunke.com")
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("[aireply-qingyunke]:", err)
		return
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln("[aireply-qingyunke]:", err)
		return
	}
	reply = gjson.Get(helper.BytesToString(bytes), "content").String()
	log.Println("reply:", reply)
	return
}

func (*QYKReply) DealReply(reply string) (textReply string, faceReply int) {
	reg := regexp.MustCompile(`\{face:(\d+)\}(.*)`)
	faceReply = -1
	if reg.MatchString(reply) {
		faceReply, _ = strconv.Atoi(reg.FindStringSubmatch(reply)[1])
		textReply = reg.FindStringSubmatch(reply)[2]
	} else {
		textReply = reply
	}
	textReply = strings.ReplaceAll(textReply, qykBotName, zero.BotConfig.NickName[0])
	textReply = strings.ReplaceAll(textReply, "{br}", "\n")
	return
}
