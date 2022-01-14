// Package bilibilipush b站推送
package bilibilipush

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/txt2img"
	"github.com/FloatTech/zbputils/web"
	"github.com/chromedp/chromedp"
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
	serviceName     = "bilibilipush"
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
	upMap = map[int64]string{}
)

func init() {
	go bilibiliPushDaily()
	en := control.Register(serviceName, &control.Options{
		DisableOnDefault: false,
		Help: "bilibilipush\n" +
			"- 添加订阅[uid]\n" +
			"- 取消订阅[uid]\n" +
			"- 取消动态订阅[uid]\n" +
			"- 取消直播订阅[uid]\n" +
			"- 推送列表",
	})

	en.OnRegex(`^添加订阅(\d+)$`, ctxext.UserOrGrpAdmin).SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[status]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := subscribe(buid, gid); err != nil {
			log.Errorln("[bilibilipush]:", err)
		} else {
			ctx.SendChain(message.Text("已添加" + name + "的订阅"))
		}

	})
	en.OnRegex(`^取消订阅(\d+)$`, ctxext.UserOrGrpAdmin).SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[status]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribe(buid, gid); err != nil {
			log.Errorln("[bilibilipush]:", err)
		} else {
			ctx.SendChain(message.Text("已取消" + name + "的订阅"))
		}

	})
	en.OnRegex(`^取消动态订阅(\d+)$`, ctxext.UserOrGrpAdmin).SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[status]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribeDynamic(buid, gid); err != nil {
			log.Errorln("[bilibilipush]:", err)
		} else {
			ctx.SendChain(message.Text("已取消" + name + "的动态订阅"))
		}

	})
	en.OnRegex(`^取消直播订阅(\d+)$`, ctxext.UserOrGrpAdmin).SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			status, name = checkBuid(buid)
			if status != 0 {
				msg, ok := uidErrorMsg[status]
				if !ok {
					msg = "未知错误，请私聊反馈给" + zero.BotConfig.NickName[0]
				}
				ctx.SendChain(message.Text(msg))
				return
			}
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribeLive(buid, gid); err != nil {
			log.Errorln("[bilibilipush]:", err)
		} else {
			ctx.SendChain(message.Text("已取消" + name + "的直播订阅"))
		}

	})
	en.OnFullMatch("推送列表", ctxext.UserOrGrpAdmin).SetBlock(true).SetPriority(prio).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		bpl := bdb.getAllPushByGroup(gid)
		fmt.Println(bpl)
		msg := "--------推送列表--------"
		for _, v := range bpl {
			if _, ok := upMap[v.BilibiliUID]; !ok {
				bdb.updateAllUp()
				fmt.Println(upMap)
			}
			msg += fmt.Sprintf("\nuid:%-12d 动态：", v.BilibiliUID)
			if v.DynamicDisable == 0 {
				msg += "●"
			} else {
				msg += "○"
			}
			msg += " 直播："
			if v.LiveDisable == 0 {
				msg += "●"
			} else {
				msg += "○"
			}
			msg += " up主：" + upMap[v.BilibiliUID]
		}
		data, err := txt2img.RenderToBase64(msg, txt2img.FontFile, 600, 20)
		if err != nil {
			log.Errorln("[bilibilipush]:", err)
		}
		if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(data))); id == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}

	})
}

func bilibiliPushDaily() {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for range t.C {
		log.Println("-----bilibilipush拉取推送信息-----")
		sendDynamic()
		sendLive()
	}
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
		upMap[buid] = name
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
	uids := bdb.getAllBuidByDynamic()
	for _, buid := range uids {
		cardList := getUserDynamicCard(buid)
		if len(cardList) == 0 {
			return
		}
		if t, ok := lastTime[buid]; !ok {
			lastTime[buid] = cardList[0].Get("desc.timestamp").Int()
			return
		} else {
			for i := len(cardList) - 1; i >= 0; i-- {
				ct := cardList[i].Get("desc.timestamp").Int()
				log.Println(ct, t)
				if ct > t && ct > time.Now().Unix()-600 {
					lastTime[buid] = ct
					m, ok := control.Lookup(serviceName)
					if ok {
						groupList := bdb.getAllGroupByBuidAndDynamic(buid)
						cId := cardList[i].Get("desc.dynamic_id").String()
						cType := cardList[i].Get("desc.type").Int()
						cName := cardList[i].Get("desc.user_profile.info.uname").String()
						screenshotFile := cachePath + cId + ".png"
						initDynamicScreenshot(cId)
						var msg []message.MessageSegment
						msg = append(msg, message.Text(cName+typeMsg[cType]))
						msg = append(msg, message.Image("file:///"+file.BOTPATH+"/"+screenshotFile))
						msg = append(msg, message.Text(tURL+cId))

						zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
							for _, gid := range groupList {
								if m.IsEnabledIn(gid) {
									if gid > 0 {
										ctx.SendGroupMessage(gid, msg)
									} else if gid < 0 {
										ctx.SendPrivateMessage(-gid, msg)
									} else {
										log.Errorln("[bilibilipush]:gid为0")
									}
								}
							}
							return true
						})

					}
				}
			}
		}
	}

}

func sendLive() {
	uids := bdb.getAllBuidByLive()
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
		log.Println("newStatus:", newStatus, "oldStatus", oldStatus)
		if newStatus != oldStatus && newStatus == 1 {
			liveStatus[key.Int()] = newStatus
			m, ok := control.Lookup(serviceName)
			if ok {
				groupList := bdb.getAllGroupByBuidAndLive(key.Int())
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
				var msg []message.MessageSegment
				msg = append(msg, message.Text(lName+" 正在直播:\n"))
				msg = append(msg, message.Text(lTitle))
				msg = append(msg, message.Image(lCover))
				msg = append(msg, message.Text(lURL))
				zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
					for _, gid := range groupList {
						if m.IsEnabledIn(gid) {
							if gid > 0 {
								ctx.SendGroupMessage(gid, msg)
							} else if gid < 0 {
								ctx.SendPrivateMessage(-gid, msg)
							} else {
								log.Errorln("[bilibilipush]:gid为0")
							}
						}
					}
					return true
				})

			}

		}
		return true
	})
}

func initDynamicScreenshot(dynamicId string) {
	screenshotFile := cachePath + dynamicId + ".png"
	if file.IsNotExist(screenshotFile) {
		var imageBuf []byte
		dynamicURL := tURL + dynamicId
		ctx, cancel := chromedp.NewContext(context.Background(), chromedp.WithDebugf(log.Printf))
		defer cancel()

		if err := chromedp.Run(ctx, chromedp.Tasks{
			chromedp.Navigate(dynamicURL),
			chromedp.SetAttributeValue(`div.unlogin-popover-avatar`, "style", "display:none;", chromedp.ByQuery),
			chromedp.SetAttributeValue(`div.bb-comment`, "style", "display:none;", chromedp.ByQuery),
			chromedp.Screenshot(`.card`, &imageBuf, chromedp.NodeVisible, chromedp.ByQuery),
		}); err != nil {
			log.Errorln("[bilibilipush]:", err)
		}

		if err := ioutil.WriteFile(screenshotFile, imageBuf, 0644); err != nil {
			log.Errorln("[bilibilipush]:", err)
		}
	}
}
