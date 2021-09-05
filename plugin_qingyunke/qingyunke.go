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
)

var prio = -1
var poke = rate.NewManager(time.Minute, 20) // 青云客接口回复
var enable = true

func init() { // 插件主体
	// 被喊名字
	zero.OnRegex("(^.{1,30}$)", zero.OnlyToMe, atriSwitch()).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			switch {
			case poke.Load(ctx.Event.UserID).Acquire():
				time.Sleep(time.Second * 1)
				msg := ctx.State["regex_matched"].([]string)[1]
				fmt.Println(msg)
				reply := getMessage(msg)
				reg := regexp.MustCompile(`\{face:(\d+)\}(.*)`)
				faceReply := -1
				var textReply string
				if reg.MatchString(reply) {
					fmt.Println(reg.FindStringSubmatch(reply))
					faceReply, _ = strconv.Atoi(reg.FindStringSubmatch(reply)[1])
					textReply = reg.FindStringSubmatch(reply)[2]
				} else {
					textReply = reply
				}
				textReply = strings.Replace(textReply, "菲菲", "椛椛", -1)
				if ctx.Event.DetailType == "group" {
					if faceReply != -1 {
						ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(textReply), message.Face(faceReply))
					} else {
						ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(textReply))
					}
				}
				if ctx.Event.DetailType == "private" {
					if faceReply != -1 {
						ctx.SendChain(message.Text(textReply), message.Face(faceReply))
					} else {
						ctx.SendChain(message.Text(textReply))
					}
				}

			default:
				//频繁触发，不回复
			}
		})
	zero.OnRegex("CQ:image,file=|CQ:face,id=", zero.OnlyToMe, atriSwitch()).SetBlock(false).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			imageUrl := getPicture()
			time.Sleep(time.Second * 1)
			if ctx.Event.DetailType == "group" {
				ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image(imageUrl))
			}
			if ctx.Event.DetailType == "private" {
				ctx.SendChain(message.Image(imageUrl))
			}
		})

	zero.OnFullMatch("开启自动回复", zero.SuperUserPermission).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			enable = true
			ctx.SendChain(message.Text("自动回复开启"))
		})
	zero.OnFullMatch("关闭自动回复", zero.SuperUserPermission).SetBlock(true).SetPriority(prio).
		Handle(func(ctx *zero.Ctx) {
			enable = false
			ctx.SendChain(message.Text("自动回复关闭"))
		})

	// 群空调
}

type QYData struct {
	Result  int    `json:"result"`
	Content string `json:"content"`
}

//青云客取消息
func getMessage(msg string) string {
	url := "http://api.qingyunke.com/api.php"
	key := "free"
	appid := "0"
	//msg := "早上好"
	url = fmt.Sprintf(url+"?key=%s&appid=%s&msg=%s", key, appid, msg)
	fmt.Println(url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("new.request", err)
	}
	// 自定义Header
	req.Header.Set("User-Agent", getAgent())
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "api.qingyunke.com")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http.get.url", err)
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ioutil.ReadAll", err)
	}
	fmt.Println(string(bytes))
	var QYData QYData
	if err := json.Unmarshal(bytes, &QYData); err != nil {
		fmt.Println("json transform", err)
	}
	return QYData.Content
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

func atriSwitch() zero.Rule {
	return func(ctx *zero.Ctx) bool {
		return enable
	}
}
