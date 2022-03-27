package bilibili

import (
	"encoding/json"
	"errors"
	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
)

// 搜索api：通过把触发指令传入的昵称找出uid返回
func search(keyword string) (*gjson.Result, error) {
	searchURL := "http://api.bilibili.com/x/web-interface/search/type?search_type=bili_user&&user_type=1&keyword=" + keyword
	data, err := web.GetData(searchURL)
	if err != nil {
		return nil, err
	}
	json := gjson.ParseBytes(data)
	if json.Get("data.numResults").Int() == 0 {
		return nil, errors.New("查无此人")
	}
	return &json, nil
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
func fansapi(uid string) (*follower, error) {
	fanURL := "https://api.vtbs.moe/v1/detail/" + uid
	data, err := web.GetData(fanURL)
	if err != nil {
		return nil, err
	}
	result := &follower{}
	if err = json.Unmarshal(data, result); err != nil {
		return nil, err
	}
	return result, nil
}

func followings(uid string) (*gjson.Result, error) {
	followingURL := "https://api.bilibili.com/x/relation/same/followings?vmid=" + uid
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, followingURL, nil)
	if err != nil {
		return nil, err
	}
	c := vdb.getBilibiliCookie()
	req.Header.Add("cookie", c.Value)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	json := gjson.ParseBytes(body)
	if json.Get("code").Int() == -101 {
		return &json, errors.New("查关注需要设置b站cookie，请发送命令设置cookie，例如\"设置b站cookie SESSDATA=82da790d,1663822823,06ecf*31\"")
	}
	if json.Get("code").Int() != 0 {
		return &json, errors.New(json.Get("message").String())
	}
	return &json, nil
}

type userinfo struct {
	Name       string
	Mid        int64
	Face       string
	Fans       int64
	Attentions []int64
}

type medal struct {
	Uname            string
	Mid              int64
	MedalName        string
	Level            int64
	MedalColorBorder int64
	MedalColorStart  int64
	MedalColorEnd    int64
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
		return result, err
	}
	cardStr := binary.BytesToString(data)
	result.Name = gjson.Get(cardStr, "card.name").String()
	result.Face = gjson.Get(cardStr, "card.face").String()
	result.Mid = gjson.Get(cardStr, "card.mid").Int()
	result.Fans = gjson.Get(cardStr, "card.fans").Int()
	attention := make([]int64, 0)
	gjson.Get(cardStr, "card.attentions").ForEach(func(key, value gjson.Result) bool {
		attention = append(attention, value.Int())
		return true
	})
	result.Attentions = attention
	return result, nil
}

// 获得牌子
func medalwall(uid string) (result []medal, err error) {
	medalwallURL := "https://api.live.bilibili.com/xlive/web-ucenter/user/MedalWall?target_id=" + uid
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, medalwallURL, nil)
	if err != nil {
		return nil, err
	}
	c := vdb.getBilibiliCookie()
	req.Header.Add("cookie", c.Value)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return result, err
	}
	medalStr := binary.BytesToString(data)
	if gjson.Get(medalStr, "code").Int() == -101 {
		return result, errors.New("查牌子需要设置b站cookie，请发送命令设置cookie，例如\"设置b站cookie SESSDATA=82da790d,1663822823,06ecf*31\"")
	}
	if gjson.Get(medalStr, "code").Int() != 0 {
		err = errors.New(gjson.Get(medalStr, "message").String())
	}
	gjson.Get(medalStr, "data.list").ForEach(func(key, value gjson.Result) bool {
		m := medal{}
		m.Uname = value.Get("target_name").String()
		m.Mid = value.Get("medal_info.target_id").Int()
		m.MedalName = value.Get("medal_info.medal_name").String()
		m.Level = value.Get("medal_info.level").Int()
		m.MedalColorBorder = value.Get("medal_info.medal_color_border").Int()
		m.MedalColorStart = value.Get("medal_info.medal_color_start").Int()
		m.MedalColorEnd = value.Get("medal_info.medal_color_end").Int()
		result = append(result, m)
		return true
	})
	return
}
