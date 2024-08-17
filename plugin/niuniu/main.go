// Package niuniu 牛牛大作战
package niuniu

import (
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/shopspring/decimal"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/exp/rand"
	"math"
	"sort"
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
		var m []userInfo
		for _, info := range niuniuList {
			if info.Length > 0 {
				m = append(m, *info)
			}
		}
		if m == nil {
			ctx.SendChain(message.Text("暂时没有男孩子哦"))
			return
		}
		var messages strings.Builder
		messages.WriteString("牛子长度排行\n")
		userInfos := sortUsersByLength(m)
		for i, user := range userInfos {
			messages.WriteString(fmt.Sprintf("第%d名    id:%s    长度:%.2fcm\n", i+1,
				ctx.CardOrNickName(user.UID), user.Length))
		}
		ctx.SendChain(message.Text(messages.String()))
	})
	en.OnFullMatch("牛子深度排行", zero.OnlyGroup, getdb).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		gid := ctx.Event.GroupID
		niuniuList, err := db.readAllTable(gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		var m []userInfo
		for _, info := range niuniuList {
			if info.Length <= 0 {
				m = append(m, *info)
			}
		}
		if m == nil {
			ctx.SendChain(message.Text("暂时没有女孩子哦"))
			return
		}
		var messages strings.Builder
		userInfos := sortUsersByNegativeLength(m)
		messages.WriteString("牛牛深度排行榜\n")
		for i, user := range userInfos {
			messages.WriteString(fmt.Sprintf("第%d名    id:%s    长度:%.2fcm\n", i+1,
				ctx.CardOrNickName(user.UID), user.Length))
		}
		ctx.SendChain(message.Text(messages.String()))
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
		result.WriteString(fmt.Sprintf("\n📛%s<%s>的牛牛信息\n⭕性别:%s\n⭕%s度:%.2fcm\n⭕ ",
			ctx.CardOrNickName(uid), strconv.FormatInt(uid, 10),
			sex, sexLong, niuniu))
		switch {
		case niuniu <= -100:
			result.WriteString("wtf？你已经进化成魅魔了！魅魔在击剑时有20%的几率消耗自身长度吞噬对方牛牛呢。")
		case niuniu <= -50:
			result.WriteString("嗯....好像已经穿过了身体吧..从另一面来看也可以算是凸出来的吧?")
		case niuniu <= -25:
			result.WriteString(randomChoice([]string{
				"这名女生，你的身体很健康哦！",
				"WOW,真的凹进去了好多呢！",
				"你已经是我们女孩子的一员啦！",
			}))
		case niuniu <= -10:
			result.WriteString(randomChoice([]string{
				"你已经是一名女生了呢，",
				"从女生的角度来说，你发育良好(,",
				"你醒啦？你已经是一名女孩子啦！",
				"唔...可以放进去一根手指了都...",
			}))
		case niuniu <= 0:
			result.WriteString(randomChoice([]string{
				"安了安了，不要伤心嘛，做女生有什么不好的啊。",
				"不哭不哭，摸摸头，虽然很难再长出来，但是请不要伤心啦啊！",
				"加油加油！我看好你哦！",
				"你醒啦？你现在已经是一名女孩子啦！",
			}))
		case niuniu <= 10:
			result.WriteString(randomChoice([]string{
				"你行不行啊？细狗！",
				"虽然短，但是小小的也很可爱呢。",
				"像一只蚕宝宝。",
				"长大了。",
			}))
		case niuniu <= 25:
			result.WriteString(randomChoice([]string{
				"唔...没话说",
				"已经很长了呢！",
			}))
		case niuniu <= 50:
			result.WriteString(randomChoice([]string{
				"话说这种真的有可能吗？",
				"厚礼谢！",
			}))
		case niuniu <= 100:
			result.WriteString(randomChoice([]string{
				"已经突破天际了嘛...",
				"唔...这玩意应该不会变得比我高吧？",
				"你这个长度会死人的...！",
				"你马上要进化成牛头人了！！",
				"你是什么怪物，不要过来啊！！",
			}))
		case niuniu > 100:
			result.WriteString("惊世骇俗！你已经进化成牛头人了！牛头人在击剑时有20%的几率消耗自身长度吞噬对方牛牛呢。")
		}
		ctx.SendChain(message.At(uid), message.Text(result.String()))
	})
	en.OnFullMatchGroup([]string{"dj", "打胶"}, zero.OnlyGroup,
		getdb).SetBlock(true).Limit(func(ctx *zero.Ctx) *rate.Limiter {
		lt := dajiaoLimiter.Load(fmt.Sprintf("dj%s%s", strconv.FormatInt(ctx.Event.GroupID, 10),
			strconv.FormatInt(ctx.Event.UserID, 10)))
		return lt
	}, func(ctx *zero.Ctx) {
		lt := dajiaoLimiter.Load(fmt.Sprintf("dj%s%s", strconv.FormatInt(ctx.Event.GroupID, 10),
			strconv.FormatInt(ctx.Event.UserID, 10)))
		timePass := lt.AcquireTime().Second()
		messages1 := []string{
			fmt.Sprintf("才过去了%ds时间,你就又要打🦶了，身体受得住吗", timePass),
			fmt.Sprintf("不行不行，你的身体会受不了的，歇%ds再来吧", 90-timePass),
			fmt.Sprintf("休息一下吧，会炸膛的！%ds后再来吧", 90-timePass),
			fmt.Sprintf("打咩哟，你的牛牛会爆炸的，休息%ds再来吧", 90-timePass),
		}
		ctx.SendChain(message.Text(randomChoice(messages1)))
	}).Handle(func(ctx *zero.Ctx) {
		// 获取群号和用户ID
		gid := ctx.Event.GroupID
		uid := ctx.Event.UserID
		niuniu, err := db.findniuniu(gid, uid)
		if err != nil {
			ctx.SendChain(message.Text("请先注册牛牛！"))
			return
		}
		probability := rand.Intn(100 + 1)
		reduce := math.Abs(hitGlue(decimal.NewFromFloat(niuniu)))
		switch {
		case probability <= 40:
			niuniu += reduce
			r := rand.Intn(2)
			ctx.SendChain(message.Text([]string{
				fmt.Sprintf("你嘿咻嘿咻一下，促进了牛牛发育，牛牛增加%.2fcm了呢！", reduce),
				fmt.Sprintf("你打了个舒服痛快的🦶呐，牛牛增加了%.2fcm呢！", reduce),
			}[r]))
		case probability <= 60:
			r := rand.Intn(2)
			ctx.SendChain(message.Text([]string{
				"你打了个🦶，但是什么变化也没有，好奇怪捏~",
				"你的牛牛刚开始变长了，可过了一会又回来了，什么变化也没有，好奇怪捏~",
			}[r]))
		default:
			niuniu -= reduce
			r := rand.Intn(3)
			if niuniu < 0 {
				ctx.SendChain(message.Text([]string{
					fmt.Sprintf("哦吼！？看来你的牛牛凹进去了%.2fcm呢！", reduce),
					fmt.Sprintf("你突发恶疾！你的牛牛凹进去了%.2fcm！", reduce),
					fmt.Sprintf("笑死，你因为打🦶过度导致牛牛凹进去了%.2fcm！🤣🤣🤣", reduce),
				}[r]))
			} else {
				ctx.SendChain(message.Text([]string{
					fmt.Sprintf("阿哦，你过度打🦶，牛牛缩短%.2fcm了呢！", reduce),
					fmt.Sprintf("你的牛牛变长了很多，你很激动地继续打🦶，然后牛牛缩短了%.2fcm呢！", reduce),
					fmt.Sprintf("小打怡情，大打伤身，强打灰飞烟灭！你过度打🦶，牛牛缩短了%.2fcm捏！", reduce),
				}[r]))
			}
		}
		u := userInfo{
			UID:    uid,
			Length: niuniu,
			ID:     1,
		}
		if err = db.insertniuniu(u, gid); err != nil {
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
		long := db.randomLong().InexactFloat64()
		u := userInfo{
			UID:    uid,
			Length: long,
			ID:     1,
		}
		//添加数据进入表
		err := db.insertniuniu(u, gid)
		if err != nil {
			err = db.createGIDTable(gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			err = db.insertniuniu(u, gid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
		}
		ctx.SendChain(message.Reply(ctx.Event.GroupID),
			message.Text("注册成功,你的牛牛现在有", u.Length, "cm"))
	})
	en.OnRegex(`jj\[CQ:at,qq=([0-9]+)\].*`, getdb,
		zero.OnlyGroup).SetBlock(true).Limit(func(ctx *zero.Ctx) *rate.Limiter {
		lt := jjLimiter.Load(fmt.Sprintf("jj%s%s", strconv.FormatInt(ctx.Event.GroupID, 10),
			strconv.FormatInt(ctx.Event.UserID, 10)))
		return lt
	}, func(ctx *zero.Ctx) {
		lt := jjLimiter.Load(fmt.Sprintf("jj%s%s", strconv.FormatInt(ctx.Event.GroupID, 10),
			strconv.FormatInt(ctx.Event.UserID, 10)))
		timePass := lt.AcquireTime().Second()
		if lt.Acquire() {
			ctx.SendChain(message.Text(randomChoice([]string{
				fmt.Sprintf("才过去了%ds时间,你就又要击剑了，真是饥渴难耐啊", timePass),
				fmt.Sprintf("不行不行，你的身体会受不了的，歇%ds再来吧", 150-timePass),
				fmt.Sprintf("你这种男同就应该被送去集中营！等待%ds再来吧", 150-timePass),
				fmt.Sprintf("打咩哟！你的牛牛会炸的，休息%ds再来吧", 150-timePass),
			})))
		}
	},
	).Handle(func(ctx *zero.Ctx) {
		adduser, err := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
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
		if myniuniu == adduserniuniu {
			ctx.SendChain(message.Text("你要和谁🤺？你自己吗？"))
			return
		}
		fencingResult, f, f1 := fencing(myniuniu, adduserniuniu)
		err = db.insertniuniu(userInfo{UID: uid, Length: f}, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		err = db.insertniuniu(userInfo{UID: adduser, Length: f1}, gid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		ctx.SendChain(message.At(uid), message.Text(fencingResult))
	})
	en.OnFullMatch("注销牛牛", getdb, zero.OnlyGroup).SetBlock(false).Handle(func(ctx *zero.Ctx) {
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

// 牛子深度
func sortUsersByNegativeLength(users []userInfo) []userInfo {
	sort.Slice(users, func(i, j int) bool {
		return math.Abs(users[i].Length) > math.Abs(users[j].Length)
	})
	return users
}

// 牛子长度
func sortUsersByLength(users []userInfo) []userInfo {
	sort.Slice(users, func(i, j int) bool {
		return users[i].Length > users[j].Length
	})
	return users
}
