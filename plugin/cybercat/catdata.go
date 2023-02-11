// Package cybercat 云养猫
package cybercat

import (
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var catType = [...]string{"橘猫", "猫娘", "咖啡猫", "奥西猫", "埃及猫",
	"阿比西尼亚猫", "布偶猫", "波斯猫", "伯曼猫", "巴厘猫",
	"东奇尼猫", "德文卷毛猫", "电烫卷猫", "东方短毛猫", "哈瓦那棕猫",
	"科拉特猫", "柯尼斯卷毛猫", "加拿大无毛猫", "金吉拉", "美国短尾猫",
	"美国卷耳猫", "美国短毛猫", "褴褛猫", "美国硬毛猫", "曼岛无尾猫",
	"缅因猫", "孟买猫", "缅甸猫", "索马里猫", "日本短尾猫",
	"挪威森林猫", "山东狮子猫", "斯可可猫", "喜马拉雅猫", "土耳其梵猫",
	"土耳其安哥拉猫", "西伯利亚猫", "夏特尔猫", "新加坡猫", "英国短毛猫",
	"异国短毛猫", "暹罗猫", "重点色短毛猫", "折耳猫", "未知种"}

type catdb struct {
	db *sql.Sqlite
	sync.RWMutex
}

type catInfo struct {
	User     int64   // 主人
	Name     string  // 喵喵名称
	Type     string  // 品种
	Satiety  float64 // 饱食度
	Mood     int     // 心情
	Weight   float64 // 体重
	LastTime int64   // 上次喂养时间
	Work     int64   // 打工时间
	Food     float64 // 食物数量
}

var (
	catdata = &catdb{
		db: &sql.Sqlite{},
	}
	engine = control.Register("cybercat", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "云养猫",
		Help:              "文字版QQ宠物复刻版(?)- 买猫\n- 买猫粮\n- 买n袋猫粮\n- 喂猫\n- 喂猫n斤猫粮\n- 猫猫打工\n- 猫猫打工[1-9]小时\n- 猫猫状态\n- 喵喵改名叫xxx\n- 喵喵pk@对方QQ\n- 。。。。",
		PrivateDataFolder: "cybercat",
	}).ApplySingle(ctxext.DefaultSingle)
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

func (sql *catdb) insert(gid string, dbInfo catInfo) error {
	sql.Lock()
	defer sql.Unlock()
	err := sql.db.Create(gid, &catInfo{})
	if err != nil {
		return err
	}
	return sql.db.Insert(gid, &dbInfo)
}

func (sql *catdb) find(gid, uid string) (dbInfo catInfo, err error) {
	sql.Lock()
	defer sql.Unlock()
	err = sql.db.Create(gid, &catInfo{})
	if err != nil {
		return
	}
	if !sql.db.CanFind(gid, "where user = "+uid) {
		return catInfo{}, nil // 规避没有该用户数据的报错
	}
	err = sql.db.Find(gid, &dbInfo, "where user = "+uid)
	return
}

func (sql *catdb) del(gid, uid string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Del(gid, "where user = "+uid)
}

func (sql *catdb) delcat(gid, uid string) error {
	sql.Lock()
	defer sql.Unlock()
	dbInfo := catInfo{}
	_ = sql.db.Find(gid, &dbInfo, "where user = "+uid)
	newInfo := catInfo{
		User: dbInfo.User,
		Food: dbInfo.Food,
	}
	return sql.db.Insert(gid, &newInfo)
}
