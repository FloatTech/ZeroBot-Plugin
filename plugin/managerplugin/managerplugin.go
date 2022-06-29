// Package managerplugin 自定义群管插件
package managerplugin

import (
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/math"
)

func init() {
	engine := control.Register("managerplugin", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help: "自定义的群管插件\n" +
			" - 开启全员禁言 群号\n" +
			" - 解除全员禁言 群号\n" +
			" - 踢出并拉黑 QQ号\n" +
			" - 踢出(并拉黑)等级为[1-100]的人" +
			" - 反\"XX召唤术\"\n",
		PrivateDataFolder: "managerplugin",
	})
	go func() {
		db.DBPath = engine.DataFolder() + "managerplugin.db"
		err := db.Open(time.Hour * 24)
		if err != nil {
			panic(err)
		}
		err = db.Create("blacklist", &blacklist{})
		if err != nil {
			panic(err)
		}
		err = db.Create("groupban", &groupban{})
		if err != nil {
			panic(err)
		}
	}()
	// 指定开启某群全群禁言 Usage: 开启全员禁言123456
	engine.OnRegex(`^开启全员禁言.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
				true,
			)
			ctx.SendChain(message.Text("全员自闭开始"))
		})
	// 指定解除某群全群禁言 Usage: 解除全员禁言123456
	engine.OnRegex(`^解除全员禁言.*?(\d+)`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SetGroupWholeBan(
				math.Str2Int64(ctx.State["regex_matched"].([]string)[1]),
				false,
			)
			ctx.SendChain(message.Text("全员自闭结束"))
		})
	engine.OnRegex(`^踢出并拉黑.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			gid := ctx.Event.GroupID
			ctx.SetGroupKick(gid, uid, true)
			nickname := ctx.CardOrNickName(uid)
			err := writeblacklist(gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
			}
			ctx.SendChain(message.Text("已将", nickname, "流放到边界外~"))
		})
	engine.OnRegex(`^加入黑名单.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			gid := ctx.Event.GroupID
			nickname := ctx.CardOrNickName(uid)
			err := writeblacklist(gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
			}
			ctx.SendChain(message.Text("已将", nickname, "加入黑名单~"))
		})
	engine.OnRegex(`^移出黑名单.*?(\d+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			uid := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			gid := ctx.Event.GroupID
			nickname := ctx.CardOrNickName(uid)
			err := deleteblacklist(gid, uid)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
			}
			ctx.SendChain(message.Text("已将", nickname, "移出黑名单~"))
		})
	engine.OnFullMatch("查看黑名单列表", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			blacklist, err := readblacklist(gid)
			if err != nil {
				ctx.SendChain(message.Text("黑名单列表是空的~"))
				return
			}
			ctx.SendChain(message.Text(blacklist))
		})
	engine.OnRegex(`踢出等级为([0-9]+{3})的人`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			setlevel := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if setlevel > 100 {
				return
			}
			gid := ctx.Event.GroupID
			banlist := ctx.GetGroupMemberListNoCache(gid).Array()
			for _, ban := range banlist {
				banid := ban.Get("user_id").Int()
				banlevel := ban.Get("level").Int()
				if banlevel == setlevel {
					ctx.SetGroupKick(gid, banid, false)
				}
			}
		})
	engine.OnRegex(`踢出并拉黑等级为([0-9]+{3})的人`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			setlevel := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
			if setlevel > 100 {
				return
			}
			gid := ctx.Event.GroupID
			banlist := ctx.GetGroupMemberListNoCache(gid).Array()
			for _, ban := range banlist {
				banid := ban.Get("user_id").Int()
				banlevel := ban.Get("level").Int()
				if banlevel == setlevel {
					ctx.SetGroupKick(gid, banid, true)
				}
			}
		})
	engine.OnRegex(`^\[CQ:xml`, zero.OnlyGroup, zero.KeywordRule("serviceID=\"60\"")).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			nickname := ctx.CardOrNickName(ctx.Event.UserID)
			ctx.SetGroupKick(ctx.Event.GroupID, ctx.Event.UserID, false)
			ctx.SetGroupBan(ctx.Event.GroupID, ctx.Event.UserID, 7*24*60*60)
			ctx.SendChain(message.ReplyWithMessage(ctx.Event.MessageID, message.Text("检测到 ["+nickname+"]("+strconv.FormatInt(ctx.Event.UserID, 10)+") 发送了干扰性消息,已处理"))...)
			ctx.DeleteMessage(message.NewMessageIDFromInteger(ctx.Event.MessageID.(int64)))
		})
}

func writeblacklist(groupid, userid int64) (err error) {
	b := blacklist{
		GrpID:  groupid,
		UserID: userid,
	}
	err = db.Insert("blacklist", &b)
	if err != nil {
		return
	}
	return
}

func readblacklist(groupid int64) (bl []any, err error) {
	var b blacklist
	err = db.FindFor("blacklist", b, "GROUP BY gid", func() error {
		var bl []any
		bl = append(bl, b.UserID, "\n")
		return nil
	})
	if err != nil {
		return
	}
	return
}

func deleteblacklist(groupid, userid int64) (err error) {
	err = db.Del("blacklist", "WHERE gid = "+strconv.FormatInt(groupid, 10)+" and WHERE uid ="+strconv.FormatInt(userid, 10))
	if err != nil {
		return
	}
	return
}
