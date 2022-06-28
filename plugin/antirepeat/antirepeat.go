// Package antirepeat 限制复读
package antirepeat

import (
	"strconv"
	"time"

	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/math"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type data struct {
	GrpID       int64 `db:"gid"`
	RepeatLimit int   `db:"repeatlimit"`
	BanTime     int64 `db:"bantime"`
}

var (
	db     = &sql.Sqlite{}
	limit  = make(map[int64]int, 256)
	rawMsg = make(map[int64]string, 256)
)
var (
	en = control.Register("antirepeat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Help:             "限制复读的插件，默认复读3次禁言，时长60分钟\n - 设置复读禁言次数 <次数>\n - 设置复读禁言时间 <时间> 分钟",
		PublicDataFolder: "Antirepeat",
	})
)

func init() {
	go func() {
		db.DBPath = en.DataFolder() + "antirepeat.db"
		err := db.Open(time.Hour * 24)
		if err != nil {
			panic(err)
		}
		err = db.Create("data", &data{})
		if err != nil {
			panic(err)
		}
	}()
	// 只接收群聊消息
	en.On(`message/group`, zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if zero.AdminPermission(ctx) {
				return
			}
			// 定义常用变量
			gid := ctx.Event.GroupID
			uid := ctx.Event.UserID
			raw := ctx.Event.RawMessage
			// 检查是否不是复读
			if rawMsg[gid] != raw {
				// 重置rawMsg
				rawMsg[gid] = raw
				// 重置limit
				limit[gid] = 0
				return
			}
			limit[gid]++
			// 检查是否到达limit
			dblimit, time := readdb(gid)
			if limit[gid] >= dblimit {
				ctx.SetGroupBan(gid, uid, time*60)
				ctx.SendChain(message.Text("因为你是第", limit[gid]+1, "个复读的，禁言", time, "分钟作为惩罚"))
			}
		})
	en.OnRegex(`^(设置复读禁言次数\s*)([0-9]+)`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			_, bantime := readdb(gid)
			d := &data{
				GrpID:       gid,
				RepeatLimit: int(math.Str2Int64(ctx.State["regex_matched"].([]string)[2])),
				BanTime:     bantime,
			}
			err := db.Insert("data", d)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("当前群聊设置复读禁言次数为", d.RepeatLimit))
		})
	en.OnRegex(`^(设置复读禁言时间\s*)([0-9]+)分钟`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			repeatlimit, _ := readdb(gid)
			d := &data{
				GrpID:       gid,
				RepeatLimit: repeatlimit,
				BanTime:     math.Str2Int64(ctx.State["regex_matched"].([]string)[2]),
			}
			err := db.Insert("data", d)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("当前群聊设置复读禁言时间为", d.BanTime, "分钟"))
		})
}

func readdb(gid int64) (int, int64) {
	var d data
	err := db.Find("data", &d, "where gid = "+strconv.FormatInt(gid, 10))
	if err != nil {
		return 3, 60
	}
	return d.RepeatLimit, d.BanTime
}
