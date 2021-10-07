// Package qingyunke 基于青云客接口的聊天对话功能
package qingyunke

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/control"
)

var (
	prio   = 100
	bucket = rate.NewManager(time.Minute, 20) // 青云客接口回复
	engine *zero.Engine
)

func init() { // 插件主体
	engine = control.Register("qingyunke", &control.Options{
		DisableOnDefault: false,
		Help: "青云客\n" +
			"- @Bot 任意文本(任意一句话回复)",
	})
	// 回复 匹配中文、英文、数字、空格但不包括下划线等符号
	engine.OnRegex("^([\u4E00-\u9FA5A-Za-z0-9\\s]{1,30})", zero.OnlyToMe).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			if !bucket.Load(ctx.Event.UserID).Acquire() {
				// 频繁触发，不回复
				return
			}
			msg := ctx.State["regex_matched"].([]string)[1]
			// 调用青云客接口
			reply, err := getMessage(msg)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 挑出 face 表情
			reg := regexp.MustCompile(`\{face:(\d+)\}(.*)`)
			faceReply := -1
			var textReply string
			if reg.MatchString(reply) {
				faceReply, _ = strconv.Atoi(reg.FindStringSubmatch(reply)[1])
				textReply = reg.FindStringSubmatch(reply)[2]
			} else {
				textReply = reply
			}
			textReply = strings.ReplaceAll(textReply, "菲菲", zero.BotConfig.NickName[0])
			textReply = strings.ReplaceAll(textReply, "{br}", "\n")
			// 回复
			time.Sleep(time.Second * 1)
			if ctx.Event.MessageType == "group" {
				if faceReply != -1 {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(textReply), message.Face(faceReply))
				} else {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(textReply))
				}
			}
			if ctx.Event.MessageType == "private" {
				if faceReply != -1 {
					ctx.SendChain(message.Text(textReply), message.Face(faceReply))
				} else {
					ctx.SendChain(message.Text(textReply))
				}
			}
		})
	// TODO: 待优化
	/*
		zero.OnRegex("CQ:image,file=|CQ:face,id=", zero.OnlyToMe, switchQYK()).SetBlock(false).SetPriority(prio).
			Handle(func(ctx *zero.Ctx) {
				imageURL := getPicture()
				time.Sleep(time.Second * 1)
				if ctx.Event.MessageType == "group" {
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image(imageURL))
				}
				if ctx.Event.MessageType == "private" {
					ctx.SendChain(message.Image(imageURL))
				}
			})
	*/
}

// 青云客数据
type dataQYK struct {
	Result  int    `json:"result"`
	Content string `json:"content"`
}

// 青云客取消息
func getMessage(msg string) (string, error) {
	url := "http://api.qingyunke.com/api.php"
	key := "free"
	appid := "0"
	url = fmt.Sprintf(url+"?key=%s&appid=%s&msg=%s", key, appid, msg)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	// 自定义Header
	req.Header.Set("User-Agent", getAgent())
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "api.qingyunke.com")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println(string(bytes))
	var dataQYK dataQYK
	if err := json.Unmarshal(bytes, &dataQYK); err != nil {
		return "", err
	}
	return dataQYK.Content, nil
}

func getAgent() string {
	agent := [...]string{
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
		"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
		"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
		"User-Agent,Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"User-Agent, Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
		"User-Agent,Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	len1 := len(agent)
	return agent[r.Intn(len1)]
}
