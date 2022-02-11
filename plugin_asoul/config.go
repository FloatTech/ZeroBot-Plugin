// Package asoul 相关功能
package asoul

import (
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/jinzhu/gorm"
)

const (
	datapath = "data/vtbs/"
	dbfile   = datapath + "info.db"
)

const (
	diana = 672328094
	ava   = 672346917
	kira  = 672353429
	queen = 672342685
	carol = 351609538
)

type follows struct {
	TS   int `json:"ts"`
	Code int `json:"code"`
	Card struct {
		Mid        string `json:"mid"`
		Name       string `json:"name"`
		Sex        string `json:"sex"`
		Face       string `json:"face"`
		Regtime    int    `json:"regtime"`
		Birthday   string `json:"birthday"`
		Sign       string `json:"sign"`
		Attentions []int  `json:"attentions"`
		Fans       int    `json:"fans"`
		Friend     int    `json:"friend"`
		Attention  int    `json:"attention"`
		LevelInfo  struct {
			NextExp      int `json:"next_exp"`
			CurrentLevel int `json:"current_level"`
			CurrentMin   int `json:"current_min"`
			CurrentExp   int `json:"current_exp"`
		} `json:"level_info"`
		Pendant struct {
			Pid    int    `json:"pid"`
			Name   string `json:"name"`
			Image  string `json:"image"`
			Expire int    `json:"expire"`
		} `json:"pendant"`
	} `json:"card"`
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

type vdInfo struct {
	Code int `json:"code"`
	Data struct {
		List struct {
			Vlist []struct {
				Pic     string `json:"pic"`
				Title   string `json:"title"`
				Created int    `json:"created"`
				Aid     int    `json:"aid"`
				Bvid    string `json:"bvid"`
			} `json:"vlist"`
		} `json:"list"`
		Page struct {
			Count int `json:"count"`
		} `json:"page"`
	} `json:"data"`
}

type vtbs struct {
	gorm.Model
	Mid   int64  `db:"mid"`
	Uname string `db:"uname"`
	Rid   int64  `db:"roomid"`
}

var engine = control.Register("asoul", order.AcquirePrio(), &control.Options{
	DisableOnDefault: false,
	Help: "=======asoul相关功能=======\n" +
		"- /查 [名字|uid] (查询bilibili用户关注vtb的情况)\n" +
		"- 日程表 (从asoul官号获取最新的日程表)\n" +
		"- 来点然/晚/牛/乃/狼能量 (随机推送一条对应账号的投稿)" +
		"- 粉丝信息 (发送bilibili平台asoul官号+5个小姐姐的粉丝数据)",
})
