package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/web"
	"github.com/tidwall/gjson"
)

var (
	errNeedCookie = errors.New("该api需要设置b站cookie，请发送命令设置cookie，例如\"设置b站cookie SESSDATA=82da790d,1663822823,06ecf*31\"")
)

// searchUser 查找b站用户
func searchUser(keyword string) (r []searchResult, err error) {
	data, err := web.GetData(fmt.Sprintf(searchUserURL, keyword))
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
func getMemberCard(uid interface{}) (result MemberCard, err error) {
	data, err := web.GetData(fmt.Sprintf(MemberCardURL, uid))
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

// getArticleInfo 用id查专栏信息
func getArticleInfo(id string) (card Card, err error) {
	var data []byte
	data, err = web.GetData(fmt.Sprintf(ArticleInfoURL, id))
	if err != nil {
		return
	}
	fmt.Println(string(data))
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("data").Raw), &card)
	return
}

// getLiveRoomInfo 用直播间id查直播间信息
func getLiveRoomInfo(roomID string) (card RoomCard, err error) {
	var data []byte
	data, err = web.GetData(fmt.Sprintf(LiveRoomInfoURL, roomID))
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
		data, err = web.GetData(fmt.Sprintf(VideoInfoURL, id, ""))
	} else {
		data, err = web.GetData(fmt.Sprintf(VideoInfoURL, "", id))
	}
	if err != nil {
		return
	}
	err = json.Unmarshal(binary.StringToBytes(gjson.ParseBytes(data).Get("data").Raw), &card)
	return
}
