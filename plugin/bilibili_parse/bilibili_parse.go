// Package bilibiliparse b站链接解析
package bilibiliparse

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/web"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 视频信息
type videoData struct {
	Code int `json:"code"`
	Data struct {
		Pic    string `json:"pic"`
		Title  string `json:"title"`
		Rights struct {
			IsCooperation int `json:"is_cooperation"`
		} `json:"rights"`
		Owner struct {
			Mid  int    `json:"mid"`
			Name string `json:"name"`
		} `json:"owner"`
		Stat struct {
			View     int `json:"view"`
			Danmaku  int `json:"danmaku"`
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

// 专栏信息
type readData struct {
	Code int `json:"code"`
	Data struct {
		Stats struct {
			View     int `json:"view"`
			Favorite int `json:"favorite"`
			Like     int `json:"like"`
			Reply    int `json:"reply"`
			Share    int `json:"share"`
			Coin     int `json:"coin"`
		} `json:"stats"`
		Title           string   `json:"title"`
		Mid             int      `json:"mid"`
		AuthorName      string   `json:"author_name"`
		OriginImageUrls []string `json:"origin_image_urls"`
	} `json:"data"`
}

// 直播间初始化信息
type liveRoomData struct {
	Code int `json:"code"`
	Data struct {
		UID        int `json:"uid"`
		LiveStatus int `json:"live_status"`
	} `json:"data"`
}

// 直播间信息
type liveData struct {
	Code int `json:"code"`
	Data struct {
		RoomStatus int    `json:"roomStatus"`
		URL        string `json:"url"`
		Title      string `json:"title"`
		Cover      string `json:"cover"`
		Online     int    `json:"online"`
	} `json:"data"`
}

// 用户信息
type ownerData struct {
	Code int `json:"code"`
	Data struct {
		Card struct {
			Fans int    `json:"fans"`
			Name string `json:"name"`
		} `json:"card"`
	} `json:"data"`
}

// 常量
const (
	videoInfoAPI = "https://api.bilibili.com/x/web-interface/view?"
	readInfoAPI  = "https://api.bilibili.com/x/article/viewinfo?"
	liveInfoAPI  = "https://api.live.bilibili.com/room/v1/Room/getRoomInfoOld?"
	liveRoomAPI  = "https://api.live.bilibili.com/room/v1/Room/room_init?"
	cardInfoAPI  = "https://api.bilibili.com/x/web-interface/card?"
	videoorigin  = "https://www.bilibili.com/video/"
	readorigin   = "https://www.bilibili.com/read/cv"
)

var (
	videoreg      = regexp.MustCompile(`https://www.bilibili.com/video/([0-9a-zA-Z]+)`)
	readmobilereg = regexp.MustCompile(`https://www.bilibili.com/read/mobile/([0-9]+)`)
	readreg       = regexp.MustCompile(`https://www.bilibili.com/read/cv([0-9]+)`)
	livereg       = regexp.MustCompile(`https://live.bilibili.com/([0-9]+)`)
	limit         = ctxext.NewLimiterManager(time.Second*10, 1)
)

// 插件主体
func init() {
	en := control.Register("bilibiliparse", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "b站视频链接解析\n" +
			"- https://www.bilibili.com/video/BV1xx411c7BF | https://www.bilibili.com/video/av1605 | https://b23.tv/I8uzWCA ",
	})
	en.OnRegex(`(av[0-9]+|BV[0-9a-zA-Z]{10}|cv[0-9]+){1}`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			id := ctx.State["regex_matched"].([]string)[1]
			m, err := parse(id)
			if err != nil {
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`https://www.bilibili.com/video/([0-9a-zA-Z]+)`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			id := ctx.State["regex_matched"].([]string)[1]
			m, err := parse(id)
			if err != nil {
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`https://www.bilibili.com/read/cv([0-9]+)`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			m, err := parseread(url)
			if err != nil {
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`https://live.bilibili.com/([0-9]+)`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			m, err := parselive(url)
			if err != nil {
				return
			}
			ctx.Send(m)
		})
	en.OnRegex(`(https://b23.tv/[0-9a-zA-Z]+)`).SetBlock(true).Limit(limit.LimitByGroup).
		Handle(func(ctx *zero.Ctx) {
			url := ctx.State["regex_matched"].([]string)[1]
			realurl, err := getrealurl(url)
			if err != nil {
				return
			}
			id, check := checkurl(realurl)
			var m message.Message
			switch check {
			case 0:
				m, err = parse(id)
				if err != nil {
					return
				}
			case 1:
				m, err = parseread(id)
				if err != nil {
					return
				}
			case 2:
				m, err = parselive(id)
				if err != nil {
					return
				}
			}
			ctx.Send(m)
		})
}

// parseVideo 解析视频数据
func parsevideo(id string) (m message.Message, err error) {
	data, err := web.GetData(videoInfoAPI + id)
	if err != nil {
		return
	}
	var vd videoData
	err = json.Unmarshal(data, &vd)
	if err != nil || vd.Code != 0 {
		return
	}
	m = make(message.Message, 0, 16)
	m = append(m, message.Text("标题: ", vd.Data.Title, "\n"))
	if vd.Data.Rights.IsCooperation == 1 {
		for i := 0; i < len(vd.Data.Staff); i++ {
			m = append(m, message.Text(vd.Data.Staff[i].Title, ": ", vd.Data.Staff[i].Name, ", 粉丝: ", row(vd.Data.Staff[i].Follower), "\n"))
		}
	} else {
		od, err := getcard(vd.Data.Owner.Mid)
		if err != nil {
			return m, err
		}
		m = append(m, message.Text("UP主: ", vd.Data.Owner.Name, ", 粉丝: ", row(od.Data.Card.Fans), "\n"))
	}
	m = append(m, message.Text("播放: ", row(vd.Data.Stat.View), ", 弹幕: ", row(vd.Data.Stat.Danmaku), "\n"),
		message.Image(vd.Data.Pic),
		message.Text("\n点赞: ", row(vd.Data.Stat.Like), ", 投币: ", row(vd.Data.Stat.Coin), "\n收藏: ", row(vd.Data.Stat.Favorite), ", 分享: ", row(vd.Data.Stat.Share), "\n", videoorigin, id))
	return
}

// parseread 解析专栏数据
func parseread(id string) (m message.Message, err error) {
	data, err := web.GetData(readInfoAPI + "id=" + id)
	if err != nil {
		return
	}
	var rd readData
	err = json.Unmarshal(data, &rd)
	if err != nil || rd.Code != 0 {
		return
	}
	od, err := getcard(rd.Data.Mid)
	if err != nil || od.Code != 0 {
		return
	}
	m = make(message.Message, 0, 3)
	m = append(m, message.Text("标题: ", rd.Data.Title, "\nUP主: ", rd.Data.AuthorName, ", 粉丝: ", row(od.Data.Card.Fans), "\n阅读: ", row(rd.Data.Stats.View), ", 喜欢: ", row(rd.Data.Stats.Like), "\n"),
		message.Image(rd.Data.OriginImageUrls[0]),
		message.Text("\n投币: ", row(rd.Data.Stats.Coin), ", 收藏: ", row(rd.Data.Stats.Favorite), "\n转发: ", row(rd.Data.Stats.Share), ", 评论: ", row(rd.Data.Stats.Reply), "\n", readorigin, id))
	return
}

// parselive 解析直播间数据
func parselive(id string) (m message.Message, err error) {
	data, err := web.GetData(liveRoomAPI + "id=" + id)
	if err != nil {
		return
	}
	var lrd liveRoomData
	err = json.Unmarshal(data, &lrd)
	if err != nil || lrd.Code != 0 {
		return
	}
	od, err := getcard(lrd.Data.UID)
	if err != nil || od.Code != 0 {
		return
	}
	data, err = web.GetData(liveInfoAPI + "mid=" + strconv.FormatInt(int64(lrd.Data.UID), 10))
	if err != nil {
		return
	}
	var ld liveData
	err = json.Unmarshal(data, &ld)
	if err != nil || ld.Code != 0 {
		return
	}
	m = make(message.Message, 0, 8)
	if ld.Data.RoomStatus != 1 {
		m = append(m, message.Text("没有该房间哦~"))
		return
	}
	m = append(m, message.Text("标题: ", ld.Data.Title, "\n主播: ", od.Data.Card.Name, ", 粉丝: ", row(od.Data.Card.Fans)))
	switch lrd.Data.LiveStatus {
	case 0:
		m = append(m, message.Text("\n状态: 未开播\n"))
	case 1:
		m = append(m, message.Text("\n状态: 直播中", ", 人气: ", row(ld.Data.Online), "\n"))
	case 2:
		m = append(m, message.Text("\n状态: 轮播中\n"))
	}
	m = append(m, message.Image(ld.Data.Cover), message.Text("\n", ld.Data.URL))
	return
}

// parse 判断id属于那种类型
func parse(id string) (m message.Message, err error) {
	switch id[:2] {
	case "av":
		m, err = parsevideo("aid=" + id[2:])
		if err != nil {
			return
		}
	case "BV":
		m, err = parsevideo("bvid=" + id[2:])
		if err != nil {
			return
		}
	case "cv":
		m, err = parseread(id[2:])
		if err != nil {
			return
		}
	}
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

// checkurl 检查链接属于哪个类型
func checkurl(url string) (id string, check int) {
	switch true {
	// 视频
	case videoreg.MatchString(url):
		return videoreg.FindStringSubmatch(url)[1], 0
	// 专栏
	case readreg.MatchString(url):
		return readreg.FindStringSubmatch(url)[1], 1
	case readmobilereg.MatchString(url):
		return readmobilereg.FindStringSubmatch(url)[1], 1
	// 直播
	case livereg.MatchString(url):
		return livereg.FindStringSubmatch(url)[1], 2
	}
	return
}

// getcard 获取个人信息
func getcard(mid int) (od ownerData, err error) {
	data, err := web.GetData(cardInfoAPI + "mid=" + strconv.Itoa(mid))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &od)
	return
}

// row 美化数据
func row(res int) string {
	if res/10000 != 0 {
		return strconv.FormatFloat(float64(res)/10000, 'f', 2, 64) + "万"
	}
	return strconv.Itoa(res)
}
