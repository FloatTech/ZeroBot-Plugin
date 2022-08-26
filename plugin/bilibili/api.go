package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	"github.com/tidwall/gjson"
)

var (
	errNeedCookie = errors.New("该api需要设置b站cookie，请发送命令设置cookie，例如\"设置b站cookie SESSDATA=82da790d,1663822823,06ecf*31\"")
)

// searchUser 查找b站用户
func searchUser(keyword string) (r []searchResult, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(searchUserURL, keyword), nil)
	if err != nil {
		return
	}
	err = reflushBilibiliCookie()
	if err != nil {
		return
	}
	req.Header.Add("cookie", cfg.BilibiliCookie)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		s := fmt.Sprintf("status code: %d", res.StatusCode)
		err = errors.New(s)
		return
	}
	var sd searchData
	err = json.NewDecoder(res.Body).Decode(&sd)
	if err != nil {
		return
	}
	r = sd.Data.Result
	return
}

// getVtbDetail 查找vtb信息
func getVtbDetail(uid string) (result vtbDetail, err error) {
	data, err := web.GetData(fmt.Sprintf(vtbDetailURL, uid))
	if err != nil {
		return
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return
	}
	return
}

// getMemberCard 获取b站个人详情
func getMemberCard(uid interface{}) (result memberCard, err error) {
	data, err := web.GetData(fmt.Sprintf(memberCardURL, uid))
	if err != nil {
		return
	}
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("card").Raw), &result)
	if err != nil {
		return
	}
	return
}

// getMedalwall 用b站uid获得牌子
func getMedalwall(uid string) (result []medal, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf(medalwallURL, uid), nil)
	if err != nil {
		return
	}
	err = reflushBilibiliCookie()
	if err != nil {
		return
	}
	req.Header.Add("cookie", cfg.BilibiliCookie)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	var md medalData
	err = json.NewDecoder(res.Body).Decode(&md)
	if err != nil {
		return
	}
	if md.Code == -101 {
		err = errNeedCookie
		return
	}
	if md.Code != 0 {
		err = errors.New(md.Message)
	}
	result = md.Data.List
	return
}

// getArticleInfo 用id查专栏信息
func getArticleInfo(id string) (card Card, err error) {
	var data []byte
	data, err = web.GetData(fmt.Sprintf(articleInfoURL, id))
	if err != nil {
		return
	}
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("data").Raw), &card)
	return
}

// getLiveRoomInfo 用直播间id查直播间信息
func getLiveRoomInfo(roomID string) (card roomCard, err error) {
	var data []byte
	data, err = web.GetData(fmt.Sprintf(liveRoomInfoURL, roomID))
	if err != nil {
		return
	}
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("data").Raw), &card)
	return
}

// getVideoInfo 用av或bv查视频信息
func getVideoInfo(id string) (card Card, err error) {
	var data []byte
	_, err = strconv.Atoi(id)
	if err == nil {
		data, err = web.GetData(fmt.Sprintf(videoInfoURL, id, ""))
	} else {
		data, err = web.GetData(fmt.Sprintf(videoInfoURL, "", id))
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("data").Raw), &card)
	return
}
