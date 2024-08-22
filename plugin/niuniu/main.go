// Package niuniu 牛牛大作战
package niuniu

import (
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/exp/rand"
	"strconv"
	"strings"
	"time"
)

var (
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "牛牛大作战",
		Help: "- 打胶\n" +
			"- 注册牛牛\n" +
			"- 注销牛牛\n" +
			"- 查看我的牛牛\n" +
			"- jj@xxx\n" +
			"- 牛子长度排行\n" +
			"- 牛子深度排行\n",
		PrivateDataFolder: "niuniu",
	})
	dajiaoLimiter = rate.NewManager[string](time.Second*90, 1)
	jjLimiter     = rate.NewManager[string](time.Second*150, 1)
)

func init() {
	en.OnFullMatch("牛子长度排行", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		niuniuList, err := db.readAllTable(gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		m := niuniuList.positive()
		if m == nil {
			ctx.SendChain(message.Text("暂时没有男孩子哦"))
			return
		}
		var messages strings.Builder
		messages.WriteString("牛子长度排行\n")
		for i, user := range niuniuList.sort(true) {
			messages.WriteString(fmt.Sprintf("第%d名  id:%s  长度:%.2fcm\n", i+1,
				ctx.CardOrNickName(user.UID), user.Length))
		}
		ctx.SendChain(message.Text(&messages))
	})
	en.OnFullMatch("牛子深度排行", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		niuniuList, err := db.readAllTable(gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		m := niuniuList.negative()
		if m == nil {
			ctx.SendChain(message.Text("暂时没有女孩子哦"))
			return
		}
		var messages strings.Builder
		messages.WriteString("牛牛深度排行榜\n")
		for i, user := range niuniuList.sort(false) {
			messages.WriteString(fmt.Sprintf("第%d名  id:%s  长度:%.2fcm\n", i+1,
				ctx.CardOrNickName(user.UID), user.Length))
		}

		ctx.SendChain(message.Text(&messages))
	})

	en.OnFullMatch("查看我的牛牛", getdb, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		niuniu, err := db.findniuniu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("你还没有牛牛呢不能查看!"))
			return
		}
		var result strings.Builder
		sexLong := "长"
		sex := "♂️"
		if niuniu < 0 {
			sexLong = "深"
			sex = "♀️"
		}
		niuniuList, err := db.readAllTable(gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		result.WriteString(fmt.Sprintf("\n📛%s<%s>的牛牛信息\n⭕性别:%s\n⭕%s度:%.2fcm\n⭕排行:%d\n⭕%s ",
			ctx.CardOrNickName(uid), strconv.FormatInt(uid, 10),
			sex, sexLong, niuniu, niuniuList.ranking(niuniu, uid), generateRandomString(niuniu)))
		ctx.SendChain(message.At(uid), message.Text(&result))
	})
	en.OnFullMatchGroup([]string{"dj", "打胶"}, zero.OnlyGroup,
		getdb).SetBlock(true).Limit(func(ctx *zero.Ctx) *rate.Limiter {
		lt := dajiaoLimiter.Load(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
		ctx.State["dajiao_last_touch"] = lt.LastTouch()
		return lt
	}, func(ctx *zero.Ctx) {
		timePass := ctx.State["dajiao_last_touch"].(int64)
		ctx.SendChain(message.Text(randomChoice([]string{
			fmt.Sprintf("才过去了%ds时间,你就又要打🦶了，身体受得住吗", timePass),
			fmt.Sprintf("不行不行，你的身体会受不了的，歇%ds再来吧", 90-timePass),
			fmt.Sprintf("休息一下吧，会炸膛的！%ds后再来吧", 90-timePass),
			fmt.Sprintf("打咩哟，你的牛牛会爆炸的，休息%ds再来吧", 90-timePass),
		})))
	}).Handle(func(ctx *zero.Ctx) {
		// 获取群号和用户ID
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		niuniu, err := db.findniuniu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("请先注册牛牛！"))
			return
		}
		messages, f := generateRandomStingTwo(niuniu)
		u := userInfo{
			UID:    uid,
			Length: f,
		}
		ctx.SendChain(message.Text(messages))
		if err = db.insertniuniu(&u, gid); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
	})
	en.OnFullMatch("注册牛牛", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		if _, err := db.findniuniu(gid, uid); err == nil {
			ctx.SendChain(message.Text("你已经注册过了"))
			return
		}
		//获取初始长度
		long := db.randLength().InexactFloat64()
		u := userInfo{
			UID:       uid,
			Length:    long,
			UserCount: 1,
		}
		//添加数据进入表
		err := db.insertniuniu(&u, gid)
		if err != nil {
			err = db.createGIDTable(gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			err = db.insertniuniu(&u, gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
		}
		ctx.SendChain(message.Reply(ctx.Event.GroupID),
			message.Text("注册成功,你的牛牛现在有", u.Length, "cm"))
	})
	en.OnRegex(`jj\[CQ:at,(?:\S*,)?qq=(\d+)(?:,\S*)?\]|(\d+)`, getdb,
		zero.OnlyGroup).SetBlock(true).Limit(func(ctx *zero.Ctx) *rate.Limiter {
		lt := jjLimiter.Load(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
		ctx.State["jj_last_touch"] = lt.LastTouch()
		return lt
	}, func(ctx *zero.Ctx) {
		timePass := ctx.State["jj_last_touch"].(int64)
		ctx.SendChain(message.Text(randomChoice([]string{
			fmt.Sprintf("才过去了%ds时间,你就又要击剑了，真是饥渴难耐啊", timePass),
			fmt.Sprintf("不行不行，你的身体会受不了的，歇%ds再来吧", 150-timePass),
			fmt.Sprintf("你这种男同就应该被送去集中营！等待%ds再来吧", 150-timePass),
			fmt.Sprintf("打咩哟！你的牛牛会炸的，休息%ds再来吧", 150-timePass),
		})))
	},
	).Handle(func(ctx *zero.Ctx) {
		fiancee := ctx.State["regex_matched"].([]string)
		adduser, _ := strconv.ParseInt(fiancee[2]+fiancee[3], 10, 64)
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		myniuniu, err := db.findniuniu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("你还没有牛牛快去注册一个吧!"))
			return
		}
		adduserniuniu, err := db.findniuniu(gid, adduser)
		if err != nil {
			ctx.SendChain(message.At(uid), message.Text("对方还没有牛牛呢，不能🤺"))
			return
		}
		if uid == adduser {
			ctx.SendChain(message.Text("你要和谁🤺？你自己吗？"))
			return
		}
		fencingResult, f, f1 := fencing(myniuniu, adduserniuniu)
		err = db.insertniuniu(&userInfo{UID: uid, Length: f}, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		err = db.insertniuniu(&userInfo{UID: adduser, Length: f1}, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.At(uid), message.Text(fencingResult))
	})
	en.OnFullMatch("注销牛牛", getdb, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		_, err := db.findniuniu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("你还没有牛牛呢，咋的你想凭空造一个啊"))
			return
		}
		err = db.deleteniuniu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("注销失败"))
			return
		}
		ctx.SendChain(message.Text("注销成功,你已经没有牛牛了"))
	})
}

func randomChoice(options []string) string {
	return options[rand.Intn(len(options))]
}
