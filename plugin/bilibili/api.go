package bilibili

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
)

var (
	errNeedCookie = errors.New("该api需要设置b站cookie，请发送命令设置cookie，例如\"设置b站cookie SESSDATA=82da790d,1663822823,06ecf*31\"")
)

type searchResult struct {
	Mid    int64  `json:"mid"`
	Uname  string `json:"uname"`
	Gender int64  `json:"gender"`
	Usign  string `json:"usign"`
	Level  int64  `json:"level"`
}

// 搜索api：通过把触发指令传入的昵称找出uid返回
func search(keyword string) (r []searchResult, err error) {
	searchURL := "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&keyword=" + keyword
	data, err := web.GetData(searchURL)
	if err != nil {
		return
	}
	j := gjson.ParseBytes(data)
	if j.Get("data.numResults").Int() == 0 {
		err = errors.New("查无此人")
		return
	}
	err = json.Unmarshal(binary.StringToBytes(j.Get("data.result").Raw), &r)
	if err != nil {
		return
	}
	return
}

type follower struct {
	Mid      int    `json:"mid"`
	Uname    string `json:"uname"`
	Video    int    `json:"video"`
	Roomid   int    `json:"roomid"`
	Rise     int    `json:"rise"`
	Follower int    `json:"follower"`
	GuardNum int    `json:"guardNum"`
	AreaRank int    `json:"areaRank"`
}

// 请求api
func fansapi(uid string) (result follower, err error) {
	fanURL := "https://api.vtbs.moe/v1/detail/" + uid
	data, err := web.GetData(fanURL)
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return
	}
	return
}

func followings(uid string) (s string, err error) {
	followingURL := "https://api.bilibili.com/x/relation/same/followings?vmid=" + uid
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, followingURL, nil)
	if err != nil {
		return
	}
	c := vdb.getBilibiliCookie()
	req.Header.Add("cookie", c.Value)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	j := gjson.ParseBytes(body)
	s = j.Get("data.list.#.uname").Raw
	if j.Get("code").Int() == -101 {
		err = errNeedCookie
		return
	}
	if j.Get("code").Int() != 0 {
		err = errors.New(j.Get("message").String())
		return
	}
	return
}

type userinfo struct {
	Name       string  `json:"name"`
	Mid        string  `json:"mid"`
	Face       string  `json:"face"`
	Fans       int64   `json:"fans"`
	Regtime    int64   `json:"regtime"`
	Attentions []int64 `json:"attentions"`
}

type medalInfo struct {
	Mid              int64  `json:"target_id"`
	MedalName        string `json:"medal_name"`
	Level            int64  `json:"level"`
	MedalColorStart  int64  `json:"medal_color_start"`
	MedalColorEnd    int64  `json:"medal_color_end"`
	MedalColorBorder int64  `json:"medal_color_border"`
}
type medal struct {
	Uname     string `json:"target_name"`
	medalInfo `json:"medal_info"`
}

type medalSlice []medal

func (m medalSlice) Len() int {
	return len(m)
}
func (m medalSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
func (m medalSlice) Less(i, j int) bool {
	return m[i].Level > m[j].Level
}

// 获取详情
func card(uid string) (result userinfo, err error) {
	cardURL := "https://account.bilibili.com/api/member/getCardByMid?mid=" + uid
	data, err := web.GetData(cardURL)
	if err != nil {
		return
	}
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("card").Raw), &result)
	if err != nil {
		return
	}
	return
}

// 获得牌子
func medalwall(uid string) (result []medal, err error) {
	medalwallURL := "https://api.live.bilibili.com/xlive/web-ucenter/user/MedalWall?target_id=" + uid
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, medalwallURL, nil)
	if err != nil {
		return
	}
	c := vdb.getBilibiliCookie()
	req.Header.Add("cookie", c.Value)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	j := gjson.ParseBytes(data)
	if j.Get("code").Int() == -101 {
		err = errNeedCookie
		return
	}
	if j.Get("code").Int() != 0 {
		err = errors.New(j.Get("message").String())
	}
	_ = json.Unmarshal(binary.StringToBytes(j.Get("data.list").Raw), &result)
	return
}
