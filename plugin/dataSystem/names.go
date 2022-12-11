package dataSystem

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/wallet"
	"github.com/FloatTech/floatbox/file"
	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// Storage 货币系统
type NameSystem struct {
	sync.RWMutex
	db *sql.Sqlite
}

// NameData 昵称信息
type NameData struct {
	UID  int64
	Name string
}

var (
	sdb = &NameSystem{
		db: &sql.Sqlite{
			DBPath: "data/wallet/wallet.db",
		},
	}
)

func init() {
	helpInfo = append(helpInfo, "----------昵 称 系 统---------"+
		"- @bot 叫我[xxx]\n- 注销昵称 [xxx/qq号/@QQ]")
	if file.IsNotExist("data/wallet") {
		err := os.MkdirAll("data/wallet", 0755)
		if err != nil {
			panic(err)
		}
	}
	err := sdb.db.Open(time.Hour * 24)
	if err != nil {
		panic(err)
	}
	err = sdb.db.Create("names", &NameData{})
	if err != nil {
		panic(err)
	}
	engine.OnRegex(`^叫我\s*([^\s]+(\s+[^\s]+)*)`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		username := ctx.State["regex_matched"].([]string)[1]
		if strings.Contains(username, "[CQ:face,id=") {
			ctx.SendChain(message.Text("昵称不支持表情包哦"))
			return
		}
		if len([]rune(username)) > 10 {
			ctx.SendChain(message.Text("昵称不得长于10个字符"))
			return
		}
		err := SetNameOf(ctx.Event.UserID, username)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("好的,", username))
	})
	engine.OnRegex(`^注销昵称(\s*\[CQ:at,qq=)?(.*[^\]$])`, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		username := ctx.State["regex_matched"].([]string)[1]
		uid, err := strconv.ParseInt(username, 10, 64)
		if err != nil {
			err = CancelNameOf(username)
		} else {
			username = GetNameOf(uid)
			err = CancelNameOf(username)
		}
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		ctx.SendChain(message.Text("注销成功"))
	})
	engine.OnFullMatchGroup([]string{"查看我的钱包", "/钱包"}).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		uid := ctx.Event.UserID
		money := wallet.GetWalletOf(uid)
		ctx.SendChain(message.At(uid), message.Text("你的钱包当前有", money, "ATRI币"))
	})
}

// GetName 获取昵称数据
func GetNameOf(uid int64) (Name string) {
	return sdb.getNameOf(uid)
}

// SetNameOf 记录昵称数据
func SetNameOf(uid int64, name string) error {
	return sdb.insertNameOf(uid, name)
}

// CancelNameOf 注销昵称数据
func CancelNameOf(name string) error {
	return sdb.del(name)
}

// 获取数据
func (sql *NameSystem) getNameOf(uid int64) (name string) {
	sql.RLock()
	defer sql.RUnlock()
	uidstr := strconv.FormatInt(uid, 10)
	var userdata NameData
	_ = sql.db.Find("names", &userdata, "where uid is "+uidstr)
	return userdata.Name
}

// 记录数据
func (sql *NameSystem) insertNameOf(uid int64, name string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Insert("names", &NameData{
		UID:  uid,
		Name: name,
	})
}

// 删除数据
func (sql *NameSystem) del(name string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Del("names", "where Name is "+name)
}
