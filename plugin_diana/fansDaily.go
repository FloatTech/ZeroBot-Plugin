package plugin_diana

import (
	"github.com/robfig/cron"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type follower struct {
	Mid         int    `json:"mid"`
	Uname       string `json:"uname"`
	Video       int    `json:"video"`
	Roomid      int    `json:"roomid"`
	Rise        int    `json:"rise"`
	Follower    int    `json:"follower"`
	GuardNum    int    `json:"guardNum"`
	AreaRank    int    `json:"areaRank"`
}


// 开启日报推送
func init() {
	zero.OnFullMatch("/开启粉丝日报", zero.AdminPermission).
		Handle(func(ctx *zero.Ctx) {
			FansDaily(130591566) //群号传进去给下面发信息的函数
	})
}

// 定时任务每天晚上最后2分钟执行一次
func FansDaily(groupID int64) {
	c := cron.New()
	c.AddFunc("0 58 23 * * ?", func() { fansData(groupID) })
	c.Start()
}

// 获取数据拼接消息链并发送
func fansData(groupID int64) {
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		var (
			diana  = fensiapi("672328094")
			ava    = fensiapi("672346917")
			eileen = fensiapi("672342685")
			bella  = fensiapi("672353429")
			carol  = fensiapi("351609538")
		)
		ctx.SendGroupMessage(
			groupID,
			message.Text(
				time.Now().Format("2006-01-02"), "    Asoul全团粉丝日报如下", "\n\n",
				"uid: ", diana.Mid, "\n",
				"名字: ", diana.Uname, "\n",
				"当前粉丝数: ", diana.Follower, "\n",
				"今日涨粉数: ", diana.Rise, "\n",
				"视频投稿数: ", diana.Video, "\n",
				"直播间id: ", diana.Roomid, "\n",
				"舰队: ", diana.GuardNum, "\n",
				"直播总排名: ", diana.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", "672328094", "\n\n",

				"uid: ", ava.Mid, "\n",
				"名字: ", ava.Uname, "\n",
				"当前粉丝数: ", ava.Follower, "\n",
				"今日涨粉数: ", ava.Rise, "\n",
				"视频投稿数: ", ava.Video, "\n",
				"直播间id: ", ava.Roomid, "\n",
				"舰队: ", ava.GuardNum, "\n",
				"直播总排名: ", ava.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", "672346917", "\n\n",

				"uid: ", eileen.Mid, "\n",
				"名字: ", eileen.Uname, "\n",
				"当前粉丝数: ", eileen.Follower, "\n",
				"今日涨粉数: ", eileen.Rise, "\n",
				"视频投稿数: ", eileen.Video, "\n",
				"直播间id: ", eileen.Roomid, "\n",
				"舰队: ", eileen.GuardNum, "\n",
				"直播总排名: ", eileen.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", "672342685", "\n\n",

				"uid: ", bella.Mid, "\n",
				"名字: ", bella.Uname, "\n",
				"当前粉丝数: ", bella.Follower, "\n",
				"今日涨粉数: ", bella.Rise, "\n",
				"视频投稿数: ", bella.Video, "\n",
				"直播间id: ", bella.Roomid, "\n",
				"舰队: ", bella.GuardNum, "\n",
				"直播总排名: ", bella.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", "672353429", "\n\n",

				"uid: ", carol.Mid, "\n",
				"名字: ", carol.Uname, "\n",
				"当前粉丝数: ", carol.Follower, "\n",
				"今日涨粉数: ", carol.Rise, "\n",
				"视频投稿数: ", carol.Video, "\n",
				"直播间id: ", carol.Roomid, "\n",
				"舰队: ", carol.GuardNum, "\n",
				"直播总排名: ", carol.AreaRank, "\n",
				"数据来源: ", "https://vtbs.moe/detail/", "351609538",
				))
		return true
	})
}

// 请求api
func fensiapi(uid string) *follower {

	url := "https://api.vtbs.moe/v1/detail/" + uid
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := &follower{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		panic(err)
	}
	return result
}
