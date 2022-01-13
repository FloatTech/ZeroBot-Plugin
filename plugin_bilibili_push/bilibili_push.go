// Package bilibilipush b站推送
package bilibilipush

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
	"github.com/chromedp/chromedp"
	"github.com/fumiama/cron"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	ua              = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	referer         = "https://www.bilibili.com/"
	infoURL         = "https://api.bilibili.com/x/space/acc/info?mid=%d"
	userDynamicsURL = "https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/space_history?host_uid=%d&offset_dynamic_id=0&need_top=0"
	liveListURL     = "https://api.live.bilibili.com/room/v1/Room/get_status_info_by_uids"
	tURL            = "https://t.bilibili.com/"
	liveURL         = "https://live.bilibili.com/"
	prio            = 10
)

var (
	lastTime = map[int64]int64{}
	typeMsg  = map[int64]string{
		0:   "发布了新动态",
		1:   "转发了一条动态",
		2:   "发布了新动态",
		4:   "发布了新动态",
		8:   "发布了新投稿",
		16:  "发布了短视频",
		64:  "发布了新专栏",
		256: "发布了新音频",
	}
	liveStatus  = map[int64]int{}
	uidErrorMsg = map[int]string{
		0:    "输入的uid有效",
		-400: "uid不存在，注意uid不是房间号",
		-402: "uid不存在，注意uid不是房间号",
		-412: "操作过于频繁IP暂时被风控，请半小时后再尝试",
	}
)

func init() {
	bilibiliPushDaily()
	en := control.Register("bilibilipush", &control.Options{
		DisableOnDefault: true,
		Help: "bilibilipush\n" +
			"- 添加订阅[uid]\n" +
			"- 取消订阅[uid]\n" +
			"- 取消动态订阅[uid]\n" +
			"- 取消直播订阅[uid]\n",
	})
	en.OnPrefix("添加订阅", zero.AdminPermission).SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, err := strconv.ParseInt(ctx.State["args"].(string), 10, 64)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请正确输入b站的uid,它是一个数字"))
			return
		}
		name := bdb.getBilibiliUpName(buid)
		var status int
		if name == "" {
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[int(status)]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		if ctx.Event.GroupID != 0 {
			if err = subscribe(buid, ctx.Event.GroupID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已添加" + name + "的订阅"))
			}
		} else {
			if err = subscribe(buid, -ctx.Event.UserID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已添加" + name + "的订阅"))
			}
		}
	})
	en.OnPrefix("取消订阅").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, err := strconv.ParseInt(ctx.State["args"].(string), 10, 64)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请正确输入b站的uid,它是一个数字"))
			return
		}
		name := bdb.getBilibiliUpName(buid)
		var status int
		if name == "" {
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[int(status)]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		if ctx.Event.GroupID != 0 {
			if err = unsubscribe(buid, ctx.Event.GroupID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已取消" + name + "的订阅"))
			}
		} else {
			if err = unsubscribe(buid, -ctx.Event.UserID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已取消" + name + "的订阅"))
			}
		}
	})
	en.OnPrefix("取消动态订阅").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, err := strconv.ParseInt(ctx.State["args"].(string), 10, 64)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请正确输入b站的uid,它是一个数字"))
			return
		}
		name := bdb.getBilibiliUpName(buid)
		var status int
		if name == "" {
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[int(status)]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		if ctx.Event.GroupID != 0 {
			if err = unsubscribeDynamic(buid, ctx.Event.GroupID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已取消" + name + "的动态订阅"))
			}
		} else {
			if err = unsubscribeDynamic(buid, -ctx.Event.UserID); err != nil {
				log.Errorln("[bilibilipush]:", err)
				ctx.SendChain(message.Text("已取消" + name + "的动态订阅"))
			}
		}
	})
	en.OnPrefix("取消直播订阅").SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, err := strconv.ParseInt(ctx.State["args"].(string), 10, 64)
		if err != nil {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("请正确输入b站的uid"))
			return
		}
		name := bdb.getBilibiliUpName(buid)
		var status int
		if name == "" {
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[int(status)]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		if ctx.Event.GroupID != 0 {
			if err = unsubscribeLive(buid, ctx.Event.GroupID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已取消" + name + "的直播订阅"))
			}
		} else {
			if err = unsubscribeLive(buid, -ctx.Event.UserID); err != nil {
				log.Errorln("[bilibilipush]:", err)
			} else {
				ctx.SendChain(message.Text("已取消" + name + "的直播订阅"))
			}
		}
	})
}

func bilibiliPushDaily() {
	c := cron.New()
	_, err := c.AddFunc("* * * * *", sendDynamic)
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	_, err = c.AddFunc("* * * * *", sendLive)
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	log.Println("开启bilibilipush定时任务")
	c.Start()
}

func checkBuid(buid int64) (status int, name string) {
	data, err := web.ReqWith(fmt.Sprintf(infoURL, buid), "GET", referer, ua)
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	status = int(gjson.Get(helper.BytesToString(data), "code").Int())
	name = gjson.Get(helper.BytesToString(data), "data.name").String()
	if status == 0 {
		bdb.insertBilibiliUp(buid, name)
	}
	return
}

// subscribe 订阅
func subscribe(buid, groupid int64) (err error) {
	bpMap := map[string]interface{}{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"live_disable":    0,
		"dynamic_disable": 0,
	}
	err = bdb.insertOrUpdateLiveAndDynamic(bpMap)
	return
}

// unsubscribe 取消订阅
func unsubscribe(buid, groupid int64) (err error) {
	bpMap := map[string]interface{}{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"live_disable":    1,
		"dynamic_disable": 1,
	}
	err = bdb.insertOrUpdateLiveAndDynamic(bpMap)
	return
}

func unsubscribeDynamic(buid, groupid int64) (err error) {
	bpMap := map[string]interface{}{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"dynamic_disable": 1,
	}
	err = bdb.insertOrUpdateLiveAndDynamic(bpMap)
	return
}

func unsubscribeLive(buid, groupid int64) (err error) {
	bpMap := map[string]interface{}{
		"bilibili_uid": buid,
		"group_id":     groupid,
		"live_disable": 1,
	}
	err = bdb.insertOrUpdateLiveAndDynamic(bpMap)
	return
}

func getUserDynamicCard(buid int64) (cardList []gjson.Result) {
	data, err := web.ReqWith(fmt.Sprintf(userDynamicsURL, buid), "GET", referer, ua)
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	cardList = gjson.Get(helper.BytesToString(data), "data.cards").Array()
	return
}

func getLiveList(uids ...int64) string {
	m := make(map[string]interface{})
	m["uids"] = uids
	b, _ := json.Marshal(m)
	client := &http.Client{}
	// 提交请求
	request, err := http.NewRequest("POST", liveListURL, bytes.NewBuffer(b))
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	request.Header.Add("Referer", referer)
	request.Header.Add("User-Agent", ua)
	response, err := client.Do(request)
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Errorln("[bilibilipush]:", err)
	}
	return helper.BytesToString(data)
}

func sendDynamic() {
	time.Sleep(time.Second * 10)
	uids := bdb.getAllBuid()
	for _, buid := range uids {
		cardList := getUserDynamicCard(buid)
		log.Println("已获得", len(cardList))
		if len(cardList) == 0 {
			return
		}
		if t, ok := lastTime[buid]; !ok {
			lastTime[buid] = cardList[0].Get("desc.timestamp").Int()
			return
		} else {
			for i, v := range cardList {
				if i >= 5 {
					break
				}
				ct := v.Get("desc.timestamp").Int()
				log.Println(ct, t)
				if ct > t && ct > time.Now().Unix()-600 {
					cId := v.Get("desc.dynamic_id").String()
					data := helper.BytesToString(getDynamicScreenshot(cId))
					log.Println(data)
					cType := v.Get("desc.type").Int()
					log.Println("cType", cType)
					cName := v.Get("desc.user_profile.info.uname").String()
					log.Println("card", v.Get("card").String())
					cDesc := v.Get("card.user.description").String()
					log.Println(fmt.Sprintf("%s%s\n%s\n%s", cName, typeMsg[cType], tURL+cId, cDesc))
				}
			}
		}
	}

}

func sendLive() {
	time.Sleep(time.Second * 10)
	uids := bdb.getAllBuid()
	gjson.Get(getLiveList(uids...), "data").ForEach(func(key, value gjson.Result) bool {
		newStatus := int(value.Get("live_status").Int())
		if newStatus == 2 {
			newStatus = 0
		}
		if _, ok := liveStatus[key.Int()]; !ok {
			liveStatus[key.Int()] = newStatus
			return true
		}
		oldStatus := liveStatus[key.Int()]
		if newStatus != oldStatus && newStatus == 1 {
			roomId := value.Get("short_id").Int()
			if roomId == 0 {
				roomId = value.Get("room_id").Int()
			}
			lURL := liveURL + strconv.FormatInt(roomId, 10)
			lName := value.Get("uname").String()
			lTitle := value.Get("title").String()
			lCover := value.Get("cover_from_user").String()
			if lCover == "" {
				lCover = value.Get("keyframe").String()
			}
			fmt.Printf("%s 正在直播:\n%s\n%s\n%s", lName, lTitle, lCover, lURL)
		}
		return true
	})
}

func getDynamicScreenshot(burl string) (imageBuf []byte) {
	// Start Chrome
	// Remove the 2nd param if you don't need debug information logged
	filename := burl + ".png"
	burl = tURL + burl
	ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
	defer cancel()

	// Run Tasks
	// List of actions to run in sequence (which also fills our image buffer)
	if err := chromedp.Run(ctx, chromedp.Tasks{
		chromedp.Navigate(burl),
		chromedp.SetAttributeValue(`div.unlogin-popover-avatar`, "style", "display:none;", chromedp.ByQuery),
		chromedp.SetAttributeValue(`div.bb-comment`, "style", "display:none;", chromedp.ByQuery),
		chromedp.Screenshot(`.card`, &imageBuf, chromedp.NodeVisible, chromedp.ByQuery),
	}); err != nil {
		log.Fatal(err)
	}

	// Write our image to file
	if err := ioutil.WriteFile(filename, imageBuf, 0644); err != nil {
		log.Fatal(err)
	}
	return imageBuf
}
