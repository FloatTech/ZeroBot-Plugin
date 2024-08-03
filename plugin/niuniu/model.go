package niuniu

import (
	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strconv"
	"sync"
	"time"
)

type Model struct {
	sql *sql.Sqlite
	sync.RWMutex
}

type UserInfo struct {
	Uid  int64
	Long float64
	Id   int
}

var (
	db    = &Model{sql: &sql.Sqlite{}}
	getdb = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		db.sql.DBPath = en.DataFolder() + "niuniu.db"
		err := db.sql.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		return true
	})
)

func (db *Model) CreateGidTable(gid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Create(strconv.FormatInt(gid, 10), &UserInfo{})
}

func (db *Model) Findniuniu(gid, uid int64) (float64, error) {
	db.RLock()
	defer db.RUnlock()
	u := UserInfo{}
	err := db.sql.Find(strconv.FormatInt(gid, 10), &u, "where Uid = "+strconv.FormatInt(uid, 10))
	return u.Long, err
}

func (db *Model) Insertniuniu(u UserInfo, gid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Insert(strconv.FormatInt(gid, 10), &u)
}

func (db *Model) Deleteniuniu(gid, uid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Del(strconv.FormatInt(gid, 10), "where Uid = "+strconv.FormatInt(uid, 10))
}

func (db *Model) readAllTable(gid int64) ([]UserInfo, error) {
	db.Lock()
	defer db.Unlock()
	a, err := sql.FindAll[UserInfo](db.sql, strconv.FormatInt(gid, 10), "where Id  = 1")
	slice := convertSocialHostInfoPointersToSlice(a)
	return slice, err
}

// 返回一个不是指针类型的切片
func convertSocialHostInfoPointersToSlice(pointers []*UserInfo) []UserInfo {
	var slice []UserInfo
	for _, ptr := range pointers {
		if ptr != nil {
			slice = append(slice, *ptr)
		}
	}
	return slice
}
