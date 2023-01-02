// Package bilibili b站推送
package bilibili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
)

const (
	ua      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	referer = "https://www.bilibili.com/"
	infoURL = "https://api.bilibili.com/x/space/acc/info?mid=%v"
)

// bdb bilibili推送数据库
var bdb *bilibilipushdb

var (
	lastTime   = map[int64]int64{}
	liveStatus = map[int64]int{}
	upMap      = map[int64]string{}
)

func init() {
	en := control.Register("bilibilipush", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "b站推送",
		Help: "- 添加b站订阅[uid|name]\n" +
			"- 取消b站订阅[uid|name]\n" +
			"- 取消b站动态订阅[uid|name]\n" +
			"- 取消b站直播订阅[uid|name]\n" +
			"- b站推送列表\n" +
			"Tips: 需要配合job一起使用, 全局只需要设置一个, 无视响应状态推送, 下为例子\n" +
			"记录在\"@every 5m\"触发的指令)\n" +
			"拉取b站推送",
		PrivateDataFolder: "bilibilipush",
	})

	// 加载bilibili推送数据库
	dbpath := en.DataFolder()
	dbfile := dbpath + "push.db"
	bdb = initializePush(dbfile)

	en.OnRegex(`^添加[B|b]站订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		name, err := getName(buid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := subscribe(buid, gid); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已添加" + name + "的订阅"))
	})
	en.OnRegex(`^取消[B|b]站订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		name, err := getName(buid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribe(buid, gid); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的订阅"))
	})
	en.OnRegex(`^取消[B|b]站动态订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		name, err := getName(buid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		if err := unsubscribeDynamic(buid, gid); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的动态订阅"))
	})
	en.OnRegex(`^取消[B|b]站直播订阅\s?(.{1,25})$`, zero.UserOrGrpAdmin, getPara).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		buid, _ := strconv.ParseInt(ctx.State["uid"].(string), 10, 64)
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		name, err := getName(buid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if err := unsubscribeLive(buid, gid); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已取消" + name + "的直播订阅"))
	})
	en.OnRegex(`^[B|b]站推送列表$`, zero.UserOrGrpAdmin).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		if gid == 0 {
			gid = -ctx.Event.UserID
		}
		bpl := bdb.getAllPushByGroup(gid)
		msg := "--------B站推送列表--------"
		for _, v := range bpl {
			if _, ok := upMap[v.BilibiliUID]; !ok {
				bdb.updateAllUp()
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
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
	})
	en.OnRegex(`拉取[B|b]站推送$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := sendDynamic(ctx)
		if err != nil {
			ctx.SendPrivateMessage(ctx.Event.UserID, message.Text("Error: bilibilipush,", err))
		}
		err = sendLive(ctx)
		if err != nil {
			ctx.SendPrivateMessage(ctx.Event.UserID, message.Text("Error: bilibilipush,", err))
		}
	})
}

// 取得uid的名字
func getName(buid int64) (name string, err error) {
	var ok bool
	if name, ok = upMap[buid]; !ok {
		var data []byte
		data, err = web.RequestDataWith(web.NewDefaultClient(), fmt.Sprintf(infoURL, buid), "GET", referer, ua)
		if err != nil {
			return
		}
		status := int(gjson.Get(binary.BytesToString(data), "code").Int())
		if status != 0 {
			err = errors.New(gjson.Get(binary.BytesToString(data), "message").String())
			return
		}
		name = gjson.Get(binary.BytesToString(data), "data.name").String()
		bdb.insertBilibiliUp(buid, name)
		upMap[buid] = name
	}
	return
}

// subscribe 订阅
func subscribe(buid, groupid int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"live_disable":    0,
		"dynamic_disable": 0,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

// unsubscribe 取消订阅
func unsubscribe(buid, groupid int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"live_disable":    1,
		"dynamic_disable": 1,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

func unsubscribeDynamic(buid, groupid int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid":    buid,
		"group_id":        groupid,
		"dynamic_disable": 1,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

func unsubscribeLive(buid, groupid int64) (err error) {
	bpMap := map[string]any{
		"bilibili_uid": buid,
		"group_id":     groupid,
		"live_disable": 1,
	}
	return bdb.insertOrUpdateLiveAndDynamic(bpMap)
}

func getUserDynamicCard(buid int64) (cardList []gjson.Result, err error) {
	data, err := web.RequestDataWith(web.NewDefaultClient(), fmt.Sprintf(bz.SpaceHistoryURL, buid, 0), "GET", referer, ua)
	if err != nil {
		return
	}
	cardList = gjson.Get(binary.BytesToString(data), "data.cards").Array()
	return
}

func getLiveList(uids ...int64) (string, error) {
	m := make(map[string]any)
	m["uids"] = uids
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	data, err := web.PostData(bz.LiveListURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	return binary.BytesToString(data), nil
}

func sendDynamic(ctx *zero.Ctx) error {
	uids := bdb.getAllBuidByDynamic()
	for _, buid := range uids {
		time.Sleep(2 * time.Second)
		cardList, err := getUserDynamicCard(buid)
		if err != nil {
			return err
		}
		if len(cardList) == 0 {
			return errors.Errorf("%v的历史动态数为0", buid)
		}
		t, ok := lastTime[buid]
		// 第一次先记录时间,啥也不做
		if !ok {
			lastTime[buid] = cardList[0].Get("desc.timestamp").Int()
			return nil
		}
		for i := len(cardList) - 1; i >= 0; i-- {
			ct := cardList[i].Get("desc.timestamp").Int()
			if ct > t && ct > time.Now().Unix()-600 {
				lastTime[buid] = ct
				m, ok := control.Lookup("bilibilipush")
				if ok {
					groupList := bdb.getAllGroupByBuidAndDynamic(buid)
					dc, err := bz.LoadDynamicDetail(cardList[i].Raw)
					if err != nil {
						err = errors.Errorf("动态%v的解析有问题,%v", cardList[i].Get("desc.dynamic_id_str"), err)
						return err
					}
					msg, err := dynamicCard2msg(&dc)
					if err != nil {
						err = errors.Errorf("动态%v的解析有问题,%v", cardList[i].Get("desc.dynamic_id_str"), err)
						return err
					}
					for _, gid := range groupList {
						if m.IsEnabledIn(gid) {
							time.Sleep(time.Millisecond * 100)
							switch {
							case gid > 0:
								ctx.SendGroupMessage(gid, msg)
							case gid < 0:
								ctx.SendPrivateMessage(-gid, msg)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func sendLive(ctx *zero.Ctx) error {
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
			m, ok := control.Lookup("bilibilipush")
			if ok {
				groupList := bdb.getAllGroupByBuidAndLive(key.Int())
				roomID := value.Get("short_id").Int()
				if roomID == 0 {
					roomID = value.Get("room_id").Int()
				}
				lURL := bz.LiveURL + strconv.FormatInt(roomID, 10)
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
				for _, gid := range groupList {
					if m.IsEnabledIn(gid) {
						time.Sleep(time.Millisecond * 100)
						switch {
						case gid > 0:
							ctx.SendGroupMessage(gid, msg)
						case gid < 0:
							ctx.SendPrivateMessage(-gid, msg)
						}
					}
				}
			}
		} else if newStatus != oldStatus {
			liveStatus[key.Int()] = newStatus
		}
		return true
	})
	return nil
}
