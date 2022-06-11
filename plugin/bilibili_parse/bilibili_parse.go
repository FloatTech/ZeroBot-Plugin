// Package bilibiliparse b站视频链接解析
package bilibiliparse

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type result struct {
	Data struct {
		Bvid      string `json:"bvid"`
		Aid       int    `json:"aid"`
		Copyright int    `json:"copyright"`
		Pic       string `json:"pic"`
		Title     string `json:"title"`
		Pubdate   int    `json:"pubdate"`
		Ctime     int    `json:"ctime"`
		Rights    struct {
			IsCooperation int `json:"is_cooperation"`
		} `json:"rights"`
		Owner struct {
			Mid  int    `json:"mid"`
			Name string `json:"name"`
		} `json:"owner"`
		Stat struct {
			Aid      int `json:"aid"`
			View     int `json:"view"`
			Danmaku  int `json:"danmaku"`
			Reply    int `json:"reply"`
			Favorite int `json:"favorite"`
			Coin     int `json:"coin"`
			Share    int `json:"share"`
			Like     int `json:"like"`
		} `json:"stat"`
		Staff []struct {
			Title    string `json:"title"`
			Name     string `json:"name"`
			Follower int    `json:"follower"`
		} `json:"staff"`
	} `json:"data"`
}

type owner struct {
	Data struct {
		Card struct {
			Fans int `json:"fans"`
		} `json:"card"`
	} `json:"data"`
}

const (
	videoapi = "https://api.bilibili.com/x/web-interface/view?"
	cardapi  = "http://api.bilibili.com/x/web-interface/card?"
	origin   = "https://www.bilibili.com/video/"
)

var (
	reg   = regexp.MustCompile(`https://www.bilibili.com/video/([0-9a-zA-Z]+)`)
	limit = ctxext.NewLimiterManager(time.Second*10, 1)
)

// 插件主体
func init() {
	en := control.Register("bilibiliparse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "b站视频链接解析\n" +
			"- https://www.bilibili.com/video/BV1xx411c7BF | https://www.bilibili.com/video/av1605 | https://b23.tv/I8uzWCA | https://www.bilibili.com/video/bv1xx411c7BF",
	})
	en.OnRegex(`(av[0-9]+|BV[0-9a-zA-Z]{10}){1}`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			if strings.Contains(ctx.MessageString(), "[CQ:forward") {
				return
			}
			id := ctx.State["regex_matched"].([]string)[1]
			m, err := parse(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`https://www.bilibili.com/video/([0-9a-zA-Z]+)`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			id := ctx.State["regex_matched"].([]string)[1]
			m, err := parse(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`(https://b23.tv/[0-9a-zA-Z]+)`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			realurl, err := getrealurl(url)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			m, err := parse(cuturl(realurl))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send(m)
		})
}

// parse 解析视频数据
func parse(id string) (m message.Message, err error) {
	var vid string
	switch id[:2] {
	case "av":
		vid = "aid=" + id[2:]
	case "BV":
		vid = "bvid=" + id
	}
	data, err := web.GetData(videoapi + vid)
	if err != nil {
		return
	}
	var r result
	err = json.Unmarshal(data, &r)
	if err != nil {
		return
	}
	m = make(message.Message, 0, 16)
	m = append(m, message.Text("标题: ", r.Data.Title, "\n"))
	if r.Data.Rights.IsCooperation == 1 {
		for i := 0; i < len(r.Data.Staff); i++ {
			m = append(m, message.Text(r.Data.Staff[i].Title, ": ", r.Data.Staff[i].Name, ", 粉丝: ", row(r.Data.Staff[i].Follower), "\n"))
		}
	} else {
		o, err := getcard(r.Data.Owner.Mid)
		if err != nil {
			return m, err
		}
		m = append(m, message.Text("UP主: ", r.Data.Owner.Name, ", 粉丝: ", row(o.Data.Card.Fans), "\n"))
	}
	m = append(m, message.Text("播放: ", row(r.Data.Stat.View), ", 弹幕: ", row(r.Data.Stat.Danmaku), "\n"),
		message.Image(r.Data.Pic),
		message.Text("\n点赞: ", row(r.Data.Stat.Like), ", 投币: ", row(r.Data.Stat.Coin), "\n收藏: ", row(r.Data.Stat.Favorite), ", 分享: ", row(r.Data.Stat.Share), "\n", origin, id))
	return
}

// getrealurl 获取跳转后的链接
func getrealurl(url string) (realurl string, err error) {
	data, err := http.Head(url)
	if err != nil {
		return
	}
	realurl = data.Request.URL.String()
	return
}

// cuturl 获取aid或者bvid
func cuturl(url string) (id string) {
	if !reg.MatchString(url) {
		return
	}
	return reg.FindStringSubmatch(url)[1]
}

// getcard 获取个人信息
func getcard(mid int) (o owner, err error) {
	data, err := web.GetData(cardapi + "mid=" + strconv.Itoa(mid))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &o)
	return
}

func row(res int) string {
	if res/10000 != 0 {
		return strconv.FormatFloat(float64(res)/10000, 'f', 2, 64) + "万"
	}
	return strconv.Itoa(res)
}
