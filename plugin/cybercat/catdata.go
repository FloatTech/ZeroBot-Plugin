// Package cybercat 云养猫
package cybercat

import (
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var catType = [...]string{"橘猫", "咖啡猫", "猫娘"}

type catdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type catinfo struct {
	User     int64  // 主人
	Name     string // 喵喵名称
	Type     string // 品种
	Satiety  int    // 饱食度
	Mood     int    // 心情
	Weight   int    // 体重
	LastTime int64  // 上次喂养时间
	Food     int    // 猫粮
}

var (
	catdata = &catdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("cybercat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "云养猫",
		Help:              "- 买猫\n- 买猫粮\n- 喂猫\n- 喵喵状态\n- 喵喵pk@对方QQ\n- 。。。。",
		PrivateDataFolder: "cybercat",
	})
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		catdata.db.DBPath = engine.DataFolder() + "catdata.db"
		err := catdata.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		return true
	})
)

func (sql *catdb) insert(gid string, dbinfo catinfo) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create(gid, &catinfo{})
	if err != nil {
		return err
	}
	return sql.db.Insert(gid, &dbinfo)
}

func (sql *catdb) find(gid, uid string) (dbinfo catinfo, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create(gid, &catinfo{})
	if err != nil {
		return
	}
	if !sql.db.CanFind(gid, "where user = "+uid) {
		return catinfo{}, nil // 规避没有该用户数据的报错
	}
	err = sql.db.Find(gid, &dbinfo, "where user = "+uid)
	return
}

func (sql *catdb) del(gid, uid string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Del(gid, "where user = "+uid)
}
