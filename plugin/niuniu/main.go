// Package niuniu 牛牛大作战
package niuniu

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/RomiChan/syncx"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type lastLength struct {
	TimeLimit time.Time
	Count     int
	Length    float64
}

type propsCount struct {
	Count     int
	TimeLimit time.Time
}

var (
	en = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "牛牛大作战",
		Help: "- 打胶\n" +
			"- 使用[道具名称]打胶\n" +
			"- jj@xxx\n" +
			"- 使用[道具名称]jj@xxx\n" +
			"- 注册牛牛\n" +
			"- 赎牛牛(cd:45分钟)\n" +
			"- 牛牛商店\n" +
			"- 牛牛背包\n" +
			"- 注销牛牛\n" +
			"- 查看我的牛牛\n" +
			"- 牛子长度排行\n" +
			"- 牛子深度排行\n",
		PrivateDataFolder: "niuniu",
	})
	dajiaoLimiter = rate.NewManager[string](time.Second*90, 1)
	jjLimiter     = rate.NewManager[string](time.Second*150, 1)
	jjCount       = syncx.Map[string, *lastLength]{}
	prop          = syncx.Map[string, *propsCount]{}
)

func init() {
	en.OnFullMatch("牛牛背包", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		niu, err := db.findNiuNiu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("你还没有牛牛呢快去注册一个吧！"))
			return
		}
		ctx.SendChain(message.Text("当前牛牛背包如下",
			"\n伟哥:", niu.WeiGe,
			"\n媚药:", niu.Philter,
			"\n击剑神器:", niu.Artifact,
			"\n击剑神稽:", niu.ShenJi))
	})
	en.OnFullMatch("牛牛商店", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID

		if _, err := db.findNiuNiu(gid, uid); err != nil {
			ctx.SendChain(message.Text("你还没有牛牛呢快去注册一个吧！"))
			return
		}

		var messages message.Message
		messages = append(messages, ctxext.FakeSenderForwardNode(ctx, message.Text("牛牛商店当前售卖的物品如下")))
		messages = append(messages,
			ctxext.FakeSenderForwardNode(ctx,
				message.Text("商品1\n商品名:伟哥\n商品价格:300ATRI币\n商品描述:可以让你打胶每次都增长，有效5次")))
		messages = append(messages,
			ctxext.FakeSenderForwardNode(ctx,
				message.Text("商品2\n商品名:媚药\n商品价格:300ATRI币\n商品描述:可以让你打胶每次都减少，有效5次")))
		messages = append(messages,
			ctxext.FakeSenderForwardNode(ctx,
				message.Text("商品3\n商品名:击剑神器\n商品价格:500ATRI币\n商品描述:可以让你每次击剑都立于不败之地，有效2次")))
		messages = append(messages,
			ctxext.FakeSenderForwardNode(ctx,
				message.Text("商品4\n商品名:击剑神稽\n商品价格:500ATRI币\n商品描述:可以让你每次击剑都失败，有效2次")))

		if id := ctx.Send(messages).ID(); id == 0 {
			ctx.Send(message.Text("发送商店失败"))
			return
		}

		ctx.SendChain(message.Text("输入对应序号进行购买商品"))
		recv, cancel := zero.NewFutureEvent("message", 999, false, zero.CheckUser(uid), zero.CheckGroup(gid), zero.RegexRule(`^(\d+)$`)).Repeat()
		defer cancel()
		timer := time.NewTimer(120 * time.Second)
		answer := ""
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				ctx.SendChain(message.At(uid), message.Text("超时,已自动取消"))
				return
			case r := <-recv:
				answer = r.Event.Message.String()
				n, err := strconv.Atoi(answer)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}

				info, err := db.findNiuNiu(gid, uid)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}

				u, money, err := purchaseItem(n, info)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}

				if wallet.GetWalletOf(uid) < money {
					ctx.SendChain(message.Text("你还没有足够的ATRI币呢,不能购买"))
					return
				}

				if err = wallet.InsertWalletOf(uid, -money); err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}

				if err = db.insertNiuNiu(u, gid); err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}

				ctx.SendChain(message.Text("购买成功!"))
				return
			}
		}
	})
	en.OnFullMatch("赎牛牛", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		last, ok := jjCount.Load(fmt.Sprintf("%d_%d", gid, uid))

		if !ok {
			ctx.SendChain(message.Text("你还没有被厥呢"))
			return
		}

		if time.Since(last.TimeLimit) > time.Minute*45 {
			ctx.SendChain(message.Text("时间已经过期了,牛牛已被收回!"))
			jjCount.Delete(fmt.Sprintf("%d_%d", gid, uid))
			return
		}

		if last.Count < 6 {
			ctx.SendChain(message.Text("你还没有被厥够6次呢,不能赎牛牛"))
			return
		}

		money := wallet.GetWalletOf(uid)
		if money < 150 {
			ctx.SendChain(message.Text("赎牛牛需要150ATRI币，快去赚钱吧"))
			return
		}

		if err := wallet.InsertWalletOf(uid, -150); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		niuniu, err := db.findNiuNiu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		niuniu.Length = last.Length

		if err = db.insertNiuNiu(&niuniu, gid); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		jjCount.Delete(fmt.Sprintf("%d_%d", gid, uid))
		ctx.SendChain(message.At(uid), message.Text(fmt.Sprintf("恭喜你!成功赎回牛牛,当前长度为:%.2fcm", last.Length)))
	})
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
		messages.WriteString("牛子长度排行榜\n")
		for i, user := range m.sort(true) {
			messages.WriteString(fmt.Sprintf("第%d名  id:%s  长度:%.2fcm\n", i+1,
				ctx.CardOrNickName(user.UID), user.Length))
		}
		msg := ctxext.FakeSenderForwardNode(ctx, message.Text(&messages))
		if id := ctx.Send(message.Message{msg}).ID(); id == 0 {
			ctx.Send(message.Text("发送排行失败"))
		}
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
		for i, user := range m.sort(false) {
			messages.WriteString(fmt.Sprintf("第%d名  id:%s  长度:%.2fcm\n", i+1,
				ctx.CardOrNickName(user.UID), user.Length))
		}
		msg := ctxext.FakeSenderForwardNode(ctx, message.Text(&messages))
		if id := ctx.Send(message.Message{msg}).ID(); id == 0 {
			ctx.Send(message.Text("发送排行失败"))
		}
	})
	en.OnFullMatch("查看我的牛牛", getdb, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		i, err := db.findNiuNiu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("你还没有牛牛呢不能查看!"))
			return
		}
		niuniu := i.Length
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
	en.OnRegex(`^(?:.*使用(.*))??打胶$`, zero.OnlyGroup,
		getdb).SetBlock(true).Limit(func(ctx *zero.Ctx) *rate.Limiter {
		lt := dajiaoLimiter.Load(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
		ctx.State["dajiao_last_touch"] = lt.LastTouch()
		return lt
	}, func(ctx *zero.Ctx) {
		timePass := int(time.Since(time.Unix(ctx.State["dajiao_last_touch"].(int64), 0)).Seconds())
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
		t := fmt.Sprintf("%d_%d", gid, uid)
		fiancee := ctx.State["regex_matched"].([]string)
		updateMap(t, false)
		niuniu, err := db.findNiuNiu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("请先注册牛牛！"))
			dajiaoLimiter.Delete(fmt.Sprintf("%d_%d", gid, uid))
			return
		}
		messages, u, err := processNiuniuAction(t, &niuniu, fiancee[1])
		if err != nil {
			ctx.SendChain(message.Text(err))
			return
		}
		ctx.SendChain(message.Text(messages))
		if err = db.insertNiuNiu(&u, gid); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
	})
	en.OnFullMatch("注册牛牛", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		if _, err := db.findNiuNiu(gid, uid); err == nil {
			ctx.SendChain(message.Text("你已经注册过了"))
			return
		}
		// 获取初始长度
		long := db.randLength()
		u := userInfo{
			UID:       uid,
			Length:    long,
			UserCount: 0,
		}
		// 添加数据进入表
		if err := db.insertNiuNiu(&u, gid); err != nil {
			if err = db.createGIDTable(gid); err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}

			if err = db.insertNiuNiu(&u, gid); err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
		}
		ctx.SendChain(message.Reply(ctx.Event.GroupID),
			message.Text("注册成功,你的牛牛现在有", u.Length, "cm"))
	})
	en.OnRegex(`^(?:.*使用(.*))??jj\s?(\[CQ:at,(?:\S*,)?qq=(\d+)(?:,\S*)?\]|(\d+))$`, getdb,
		zero.OnlyGroup).SetBlock(true).Limit(func(ctx *zero.Ctx) *rate.Limiter {
		lt := jjLimiter.Load(fmt.Sprintf("%d_%d", ctx.Event.GroupID, ctx.Event.UserID))
		ctx.State["jj_last_touch"] = lt.LastTouch()
		return lt
	}, func(ctx *zero.Ctx) {
		timePass := int(time.Since(time.Unix(ctx.State["jj_last_touch"].(int64), 0)).Seconds())
		ctx.SendChain(message.Text(randomChoice([]string{
			fmt.Sprintf("才过去了%ds时间,你就又要击剑了，真是饥渴难耐啊", timePass),
			fmt.Sprintf("不行不行，你的身体会受不了的，歇%ds再来吧", 150-timePass),
			fmt.Sprintf("你这种男同就应该被送去集中营！等待%ds再来吧", 150-timePass),
			fmt.Sprintf("打咩哟！你的牛牛会炸的，休息%ds再来吧", 150-timePass),
		})))
	},
	).Handle(func(ctx *zero.Ctx) {
		fiancee := ctx.State["regex_matched"].([]string)
		adduser, err := strconv.ParseInt(fiancee[3]+fiancee[4], 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		t := fmt.Sprintf("%d_%d", gid, uid)
		updateMap(t, false)
		myniuniu, err := db.findNiuNiu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("你还没有牛牛快去注册一个吧!"))
			jjLimiter.Delete(t)
			return
		}
		adduserniuniu, err := db.findNiuNiu(gid, adduser)
		if err != nil {
			ctx.SendChain(message.At(uid), message.Text("对方还没有牛牛呢，不能🤺"))
			jjLimiter.Delete(t)
			return
		}
		if uid == adduser {
			ctx.SendChain(message.Text("你要和谁🤺？你自己吗？"))
			jjLimiter.Delete(t)
			return
		}
		fencingResult, f1, u, err := processJJuAction(&myniuniu, &adduserniuniu, t, fiancee[1])
		if err != nil {
			ctx.SendChain(message.Text(err))
			return
		}

		if err = db.insertNiuNiu(&u, gid); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		if err = db.insertNiuNiu(&userInfo{UID: adduser, Length: f1}, gid); err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}

		ctx.SendChain(message.At(uid), message.Text(" ", fencingResult))
		j := fmt.Sprintf("%d_%d", gid, adduser)
		count, ok := jjCount.Load(j)
		var c lastLength
		// 按照第一次jj时的时间计算，超过45分钟则重置
		if !ok {
			c = lastLength{
				TimeLimit: time.Now(),
				Count:     1,
				Length:    adduserniuniu.Length,
			}
		} else {
			c = lastLength{
				TimeLimit: c.TimeLimit,
				Count:     count.Count + 1,
				Length:    count.Length,
			}
			if time.Since(c.TimeLimit) > time.Minute*45 {
				c = lastLength{
					TimeLimit: time.Now(),
					Count:     1,
					Length:    adduserniuniu.Length,
				}
			}
		}

		jjCount.Store(j, &c)
		if c.Count > 5 {
			ctx.SendChain(message.Text(randomChoice([]string{fmt.Sprintf("你们太厉害了，对方已经被你们打了%d次了，你们可以继续找他🤺", c.Count),
				"你们不要再找ta🤺啦！"})))
			// 保证只发生一次
			if c.Count < 7 {
				id := ctx.SendPrivateMessage(adduser,
					message.Text(fmt.Sprintf("你在%d群里已经被厥冒烟了，快去群里赎回你原本的牛牛!\n发送:`赎牛牛`即可！", gid)))
				if id == 0 {
					ctx.SendChain(message.At(adduser), message.Text("快发送`赎牛牛`来赎回你原本的牛牛!"))
				}
			}
		}
	})
	en.OnFullMatch("注销牛牛", getdb, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		gid := ctx.Event.GroupID
		_, err := db.findNiuNiu(gid, uid)
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

func updateMap(t string, d bool) {
	value, ok := prop.Load(t)
	if value == nil {
		return
	}
	if !d {
		if time.Since(value.TimeLimit) > time.Minute*8 {
			prop.Delete(t)
		}
		return
	}
	if ok {
		prop.Store(t, &propsCount{
			Count:     value.Count + 1,
			TimeLimit: value.TimeLimit,
		})
	} else {
		prop.Store(t, &propsCount{
			Count:     1,
			TimeLimit: time.Now(),
		})
	}
	if time.Since(value.TimeLimit) > time.Minute*8 {
		prop.Delete(t)
	}
}
