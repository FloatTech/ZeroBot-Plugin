package dice

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/manager/timer"
	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

var (
	db    = &sql.Sqlite{}
	clock timer.Clock
)

func init() {
	go func() {
		db.DBPath = engine.DataFolder() + "dice.db"
		clock = timer.NewClock(db)
		err := db.Create("strjrrp", &strjrrp{})
		if err != nil {
			panic(err)
		}
	}()
	engine.OnFullMatchGroup([]string{".jrrp", "。jrrp"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			now := time.Now()
			uid := ctx.Event.UserID
			seed := md5.Sum(helper.StringToBytes(fmt.Sprintf("%d%d%d%d", uid, now.Year(), now.Month(), now.Day())))
			newrand := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
			jrrp := newrand.Intn(100) + 1
			/*var j strjrrp
			err := db.Find("strjrrp", &j, "where gid = "+strconv.FormatInt(ctx.Event.GroupID, 10))
			if err == nil {
				ctx.SendGroupMessage(ctx.Event.GroupID, customjrrp(ctx, j.Strjrrp))
			} else {*/
			ctx.SendChain(message.At(uid), message.Text("阁下今日的人品值为", jrrp, "呢~"))
			//}
		})
	engine.OnRegex(`^设置jrrp([\s\S]*)$`, zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			j := &strjrrp{
				GrpID:   ctx.Event.GroupID,
				Strjrrp: ctx.State["regex_matched"].([]string)[1],
			}
			err := db.Insert("strjrrp", j)
			if err == nil {
				ctx.SendChain(message.Text("记住啦!"))
			} else {
				ctx.SendChain(message.Text("出错啦: ", err))
			}
		})
}

// customjrrp 自定义jrrp
/*func customjrrp(ctx *zero.Ctx, strjrrp string) string {
	now := time.Now()
	uid := ctx.Event.UserID
	seed := md5.Sum(helper.StringToBytes(fmt.Sprintf("%d%d%d%d", uid, now.Year(), now.Month(), now.Day())))
	newrand := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
	jrrp := newrand.Intn(100)+1
	jrrp = strconv.Itoa(jrrp)
	str := strings.ReplaceAll(strjrrp, "{jrrp}",jrrp)
	return str
}*/
