// Package bilibilipush b站推送
package bilibilipush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/FloatTech/zbputils/web"
)

const (
	ua             = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	referer        = "https://www.bilibili.com/"
	infoURL        = "https://api.bilibili.com/x/space/acc/info?mid=%d"
	userDynamicURL = "https://api.vc.bilibili.com/dynamic_svr/v1/dynamic_svr/space_history?host_uid=%d&offset_dynamic_id=0&need_top=0"
	liveListURL    = "https://api.live.bilibili.com/room/v1/Room/get_status_info_by_uids"
	tURL           = "https://t.bilibili.com/"
	liveURL        = "https://live.bilibili.com/"
	serviceName    = "bilibilipush"
)

// bdb bilibili推送数据库
var bdb *bilibilipushdb

var (
	lastTime = map[int64]int64{}
	typeMsg  = map[int64]string{
		1:    "转发了一条动态",
		2:    "有图营业",
		4:    "无图营业",
		8:    "发布了新投稿",
		16:   "发布了短视频",
		64:   "发布了新专栏",
		256:  "发布了新音频",
		2048: "发布了新简报",
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
			"- 添加b站订阅[uid]\n" +
			"- 取消b站订阅[uid]\n" +
			"- 取消b站动态订阅[uid]\n" +
			"- 取消b站直播订阅[uid]\n" +
			"- b站推送列表",
		PrivateDataFolder: serviceName,
	})

	// 加载数据库
	go func() {
		dbpath := en.DataFolder()
		dbfile := dbpath + "push.db"
		bdb = initialize(dbfile)
	}()

	en.OnRegex(`^添加b站订阅\s?(\d+)$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			var err error
			status, name, err = checkBuid(buid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
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
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("已添加" + name + "的订阅"))
	})
	en.OnRegex(`^取消b站订阅\s?(\d+)$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			var err error
			status, name, err = checkBuid(buid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
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
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的订阅"))
	})
	en.OnRegex(`^取消b站动态订阅\s?(\d+)$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			var err error
			status, name, err = checkBuid(buid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
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
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的动态订阅"))
	})
	en.OnRegex(`^取消b站直播订阅\s?(\d+)$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		var name string
		var ok bool
		if name, ok = upMap[buid]; !ok {
			var status int
			var err error
			status, name, err = checkBuid(buid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
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
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的直播订阅"))
	})
	en.OnFullMatch("b站推送列表", zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		bpl := bdb.getAllPushByGroup(gid)
		fmt.Println(bpl)
		msg := "--------b站推送列表--------"
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
		data, err := text.RenderToBase64(msg, text.FontFile, 600, 20)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
			ctx.SendChain(message.Text("ERROR:可能被风控了"))
		}
	})
}

func bilibiliPushDaily() {
	t := time.NewTicker(time.Second * 60)
	defer t.Stop()
	for range t.C {
		log.Debugln("-----bilibilipush拉取推送信息-----")
		_ = sendDynamic()
		_ = sendLive()
	}
}

func checkBuid(buid int64) (status int, name string, err error) {
	data, err := web.RequestDataWith(web.NewDefaultClient(), fmt.Sprintf(infoURL, buid), "GET", referer, ua)
	if err != nil {
		return
	}
	status = int(gjson.Get(binary.BytesToString(data), "code").Int())
	name = gjson.Get(binary.BytesToString(data), "data.name").String()
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

func getUserDynamicCard(buid int64) (cardList []gjson.Result, err error) {
	data, err := web.RequestDataWith(web.NewDefaultClient(), fmt.Sprintf(userDynamicURL, buid), "GET", referer, ua)
	if err != nil {
		return
	}
	cardList = gjson.Get(binary.BytesToString(data), "data.cards").Array()
	return
}

func getLiveList(uids ...int64) (string, error) {
	m := make(map[string]interface{})
	m["uids"] = uids
	b, _ := json.Marshal(m)
	data, err := web.PostData(liveListURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	return binary.BytesToString(data), nil
}

func sendDynamic() error {
	uids := bdb.getAllBuidByDynamic()
	for _, buid := range uids {
		cardList, err := getUserDynamicCard(buid)
		if err != nil {
			return err
		}
		t, ok := lastTime[buid]
		if !ok {
			lastTime[buid] = cardList[0].Get("desc.timestamp").Int()
			return nil
		}
		for i := len(cardList) - 1; i >= 0; i-- {
			ct := cardList[i].Get("desc.timestamp").Int()
			if ct > t && ct > time.Now().Unix()-600 {
				lastTime[buid] = ct
				m, ok := control.Lookup(serviceName)
				if ok {
					groupList := bdb.getAllGroupByBuidAndDynamic(buid)
					var msg []message.MessageSegment
					cType := cardList[i].Get("desc.type").Int()
					cardStr := cardList[i].Get("card").String()
					switch cType {
					case 0:
						cName := cardList[i].Get("desc.user_profile.info.uname").String()
						cTime := time.Unix(cardList[i].Get("desc.timestamp").Int(), 0).Format("2006-01-02 15:04:05")
						msg = append(msg, message.Text(cName+"在"+cTime+typeMsg[cType]+"\n"))
					case 1:
						cName := gjson.Get(cardStr, "user.uname").String()
						msg = append(msg, message.Text(cName+typeMsg[cType]+"\n"))
						cContent := gjson.Get(cardStr, "item.content").String()
						msg = append(msg, message.Text(cContent+"\n"))
						msg = append(msg, message.Text("转发的内容：\n"))
						cOrigType := gjson.Get(cardStr, "item.orig_type").Int()
						cOrigin := gjson.Get(cardStr, "origin").String()
						switch cOrigType {
						case 1:
							cName := gjson.Get(cOrigin, "user.uname").String()
							msg = append(msg, message.Text(cName+typeMsg[cOrigType]+"\n"))
						case 2:
							cName := gjson.Get(cOrigin, "user.name").String()
							cUploadTime := time.Unix(gjson.Get(cOrigin, "item.upload_time").Int(), 0).Format("2006-01-02 15:04:05")
							msg = append(msg, message.Text(cName+"在"+cUploadTime+typeMsg[cOrigType]+"\n"))
							cDescription := gjson.Get(cOrigin, "item.description")
							msg = append(msg, message.Text(cDescription))
							if gjson.Get(cOrigin, "item.pictures.#").Int() != 0 {
								gjson.Get(cOrigin, "item.pictures").ForEach(func(_, v gjson.Result) bool {
									msg = append(msg, message.Image(v.Get("img_src").String()))
									return true
								})
							}
						case 4:
							cName := gjson.Get(cOrigin, "user.uname").String()
							cTimestamp := time.Unix(gjson.Get(cOrigin, "item.timestamp").Int(), 0).Format("2006-01-02 15:04:05")
							msg = append(msg, message.Text(cName+"在"+cTimestamp+typeMsg[cOrigType]+"\n"))
							cContent := gjson.Get(cOrigin, "item.content").String()
							msg = append(msg, message.Text(cContent+"\n"))
						case 8:
							cName := gjson.Get(cOrigin, "owner.name").String()
							cTime := time.Unix(gjson.Get(cOrigin, "pubdate").Int(), 0).Format("2006-01-02 15:04:05")
							msg = append(msg, message.Text(cName+"在"+cTime+typeMsg[cOrigType]+"\n"))
							cTitle := gjson.Get(cOrigin, "title").String()
							msg = append(msg, message.Text(cTitle))
							cPic := gjson.Get(cOrigin, "pic").String()
							msg = append(msg, message.Image(cPic))
							cDesc := gjson.Get(cOrigin, "desc").String()
							msg = append(msg, message.Text(cDesc+"\n"))
							cShareSubtitle := gjson.Get(cOrigin, "share_subtitle").String()
							msg = append(msg, message.Text(cShareSubtitle+"\n"))
							cShortLink := gjson.Get(cOrigin, "short_link").String()
							msg = append(msg, message.Text("视频链接："+cShortLink+"\n"))
						case 16:
							cName := gjson.Get(cOrigin, "user.name").String()
							cUploadTime := gjson.Get(cOrigin, "item.upload_time").String()
							msg = append(msg, message.Text(cName+"在"+cUploadTime+typeMsg[cOrigType]+"\n"))
							cDescription := gjson.Get(cOrigin, "item.description")
							msg = append(msg, message.Text(cDescription))
							cCover := gjson.Get(cOrigin, "item.cover.default").String()
							msg = append(msg, message.Image(cCover))
						case 64:
							cName := gjson.Get(cOrigin, "author.name").String()
							cPublishTime := time.Unix(gjson.Get(cOrigin, "publish_time").Int(), 0).Format("2006-01-02 15:04:05")
							msg = append(msg, message.Text(cName+"在"+cPublishTime+typeMsg[cOrigType]+"\n"))
							cTitle := gjson.Get(cOrigin, "title").String()
							msg = append(msg, message.Text(cTitle+"\n"))
							cSummary := gjson.Get(cOrigin, "summary").String()
							msg = append(msg, message.Text(cSummary))
							cBannerURL := gjson.Get(cOrigin, "banner_url").String()
							msg = append(msg, message.Image(cBannerURL))
						case 256:
							cUpper := gjson.Get(cOrigin, "upper").String()
							cTime := time.UnixMilli(gjson.Get(cOrigin, "ctime").Int()).Format("2006-01-02 15:04:05")
							msg = append(msg, message.Text(cUpper+"在"+cTime+typeMsg[cOrigType]+"\n"))
							cTitle := gjson.Get(cOrigin, "title").String()
							msg = append(msg, message.Text(cTitle))
							cCover := gjson.Get(cOrigin, "cover").String()
							msg = append(msg, message.Image(cCover))
						case 2048:
							cName := gjson.Get(cOrigin, "user.uname").String()
							msg = append(msg, message.Text(cName+typeMsg[cOrigType]+"\n"))
							cContent := gjson.Get(cOrigin, "vest.content").String()
							msg = append(msg, message.Text(cContent+"\n"))
							cTitle := gjson.Get(cOrigin, "sketch.title").String()
							msg = append(msg, message.Text(cTitle+"\n"))
							cDescText := gjson.Get(cOrigin, "sketch.desc_text").String()
							msg = append(msg, message.Text(cDescText))
							cCoverURL := gjson.Get(cOrigin, "sketch.cover_url").String()
							msg = append(msg, message.Image(cCoverURL))
							cTargetURL := gjson.Get(cOrigin, "sketch.target_url").String()
							msg = append(msg, message.Text("简报链接："+cTargetURL+"\n"))
						default:
							msg = append(msg, message.Text("未知动态类型"+strconv.FormatInt(cOrigType, 10)+"\n"))
						}
					case 2:
						cName := gjson.Get(cardStr, "user.name").String()
						cUploadTime := time.Unix(gjson.Get(cardStr, "item.upload_time").Int(), 0).Format("2006-01-02 15:04:05")
						msg = append(msg, message.Text(cName+"在"+cUploadTime+typeMsg[cType]+"\n"))
						cDescription := gjson.Get(cardStr, "item.description")
						msg = append(msg, message.Text(cDescription))
						if gjson.Get(cardStr, "item.pictures.#").Int() != 0 {
							gjson.Get(cardStr, "item.pictures").ForEach(func(_, v gjson.Result) bool {
								msg = append(msg, message.Image(v.Get("img_src").String()))
								return true
							})
						}
					case 4:
						cName := gjson.Get(cardStr, "user.uname").String()
						cTimestamp := time.Unix(gjson.Get(cardStr, "item.timestamp").Int(), 0).Format("2006-01-02 15:04:05")
						msg = append(msg, message.Text(cName+"在"+cTimestamp+typeMsg[cType]+"\n"))
						cContent := gjson.Get(cardStr, "item.content").String()
						msg = append(msg, message.Text(cContent+"\n"))
					case 8:
						cName := gjson.Get(cardStr, "owner.name").String()
						cTime := time.Unix(gjson.Get(cardStr, "ctime").Int(), 0).Format("2006-01-02 15:04:05")
						msg = append(msg, message.Text(cName+"在"+cTime+typeMsg[cType]+"\n"))
						cTitle := gjson.Get(cardStr, "title").String()
						msg = append(msg, message.Text(cTitle))
						cPic := gjson.Get(cardStr, "pic").String()
						msg = append(msg, message.Image(cPic))
						cDesc := gjson.Get(cardStr, "desc").String()
						msg = append(msg, message.Text(cDesc+"\n"))
						cShareSubtitle := gjson.Get(cardStr, "share_subtitle").String()
						msg = append(msg, message.Text(cShareSubtitle+"\n"))
						cShortLink := gjson.Get(cardStr, "short_link").String()
						msg = append(msg, message.Text("视频链接："+cShortLink+"\n"))
					case 16:
						cName := gjson.Get(cardStr, "user.name").String()
						cUploadTime := gjson.Get(cardStr, "item.upload_time").String()
						msg = append(msg, message.Text(cName+"在"+cUploadTime+typeMsg[cType]+"\n"))
						cDescription := gjson.Get(cardStr, "item.description")
						msg = append(msg, message.Text(cDescription))
						cCover := gjson.Get(cardStr, "item.cover.default").String()
						msg = append(msg, message.Image(cCover))
					case 64:
						cName := gjson.Get(cardStr, "author.name").String()
						cPublishTime := time.Unix(gjson.Get(cardStr, "publish_time").Int(), 0).Format("2006-01-02 15:04:05")
						msg = append(msg, message.Text(cName+"在"+cPublishTime+typeMsg[cType]+"\n"))
						cTitle := gjson.Get(cardStr, "title").String()
						msg = append(msg, message.Text(cTitle+"\n"))
						cSummary := gjson.Get(cardStr, "summary").String()
						msg = append(msg, message.Text(cSummary))
						cBannerURL := gjson.Get(cardStr, "banner_url").String()
						msg = append(msg, message.Image(cBannerURL))
					case 256:
						cUpper := gjson.Get(cardStr, "upper").String()
						cTime := time.UnixMilli(gjson.Get(cardStr, "ctime").Int()).Format("2006-01-02 15:04:05")
						msg = append(msg, message.Text(cUpper+"在"+cTime+typeMsg[cType]+"\n"))
						cTitle := gjson.Get(cardStr, "title").String()
						msg = append(msg, message.Text(cTitle))
						cCover := gjson.Get(cardStr, "cover").String()
						msg = append(msg, message.Image(cCover))
					case 2048:
						cName := gjson.Get(cardStr, "user.uname").String()
						msg = append(msg, message.Text(cName+typeMsg[cType]+"\n"))
						cContent := gjson.Get(cardStr, "vest.content").String()
						msg = append(msg, message.Text(cContent+"\n"))
						cTitle := gjson.Get(cardStr, "sketch.title").String()
						msg = append(msg, message.Text(cTitle+"\n"))
						cDescText := gjson.Get(cardStr, "sketch.desc_text").String()
						msg = append(msg, message.Text(cDescText))
						cCoverURL := gjson.Get(cardStr, "sketch.cover_url").String()
						msg = append(msg, message.Image(cCoverURL))
						cTargetURL := gjson.Get(cardStr, "sketch.target_url").String()
						msg = append(msg, message.Text("简报链接："+cTargetURL+"\n"))
					default:
						msg = append(msg, message.Text("未知动态类型"+strconv.FormatInt(cType, 10)+"\n"))
					}
					cID := cardList[i].Get("desc.dynamic_id").String()
					msg = append(msg, message.Text("动态链接：", tURL+cID))

					zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
						for _, gid := range groupList {
							if m.IsEnabledIn(gid) {
								switch {
								case gid > 0:
									ctx.SendGroupMessage(gid, msg)
								case gid < 0:
									ctx.SendPrivateMessage(-gid, msg)
								}
							}
						}
						return true
					})
				}
			}
		}
	}
	return nil
}

func sendLive() error {
	uids := bdb.getAllBuidByLive()
	ll, err := getLiveList(uids...)
	if err != nil {
		return err
	}
	gjson.Get(ll, "data").ForEach(func(key, value gjson.Result) bool {
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
			liveStatus[key.Int()] = newStatus
			m, ok := control.Lookup(serviceName)
			if ok {
				groupList := bdb.getAllGroupByBuidAndLive(key.Int())
				roomID := value.Get("short_id").Int()
				if roomID == 0 {
					roomID = value.Get("room_id").Int()
				}
				lURL := liveURL + strconv.FormatInt(roomID, 10)
				lName := value.Get("uname").String()
				lTitle := value.Get("title").String()
				lCover := value.Get("cover_from_user").String()
				if lCover == "" {
					lCover = value.Get("keyframe").String()
				}
				var msg []message.MessageSegment
				msg = append(msg, message.Text(lName+" 正在直播：\n"))
				msg = append(msg, message.Text(lTitle))
				msg = append(msg, message.Image(lCover))
				msg = append(msg, message.Text("直播链接：", lURL))
				zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
					for _, gid := range groupList {
						if m.IsEnabledIn(gid) {
							switch {
							case gid > 0:
								ctx.SendGroupMessage(gid, msg)
							case gid < 0:
								ctx.SendPrivateMessage(-gid, msg)
							}
						}
					}
					return true
				})
			}
		} else if newStatus != oldStatus {
			liveStatus[key.Int()] = newStatus
		}
		return true
	})
	return nil
}
