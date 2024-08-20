// Package niuniu 牛牛大作战
package niuniu

import (
	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	"github.com/shopspring/decimal"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"
)

type model struct {
	sql sql.Sqlite
	sync.RWMutex
}

type userInfo struct {
	UID       int64
	Length    float64
	UserCount int
}
type users []*userInfo

var (
	db    = &model{}
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

func (m users) newPositive() []userInfo {
	var m1 []userInfo
	for _, i2 := range m {
		if i2.Length > 0 {
			m1 = append(m1, *i2)
		}
	}
	return m1
}

func (m users) newNegative() []userInfo {
	var m1 []userInfo
	for _, i2 := range m {
		if i2.Length <= 0 {
			m1 = append(m1, *i2)
		}
	}
	return m1
}

// 牛子深度
func (m users) sortUsersByNegativeLength() users {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Length > m[j].Length
	})
	var newUsers []*userInfo
	for i := len(m); i >= 0; i-- {
		newUsers = append(newUsers, m[i])
	}
	return newUsers
}

// 牛子长度
func (m users) sortUsersByLength() users {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Length > m[j].Length
	})
	return m
}

func (db *model) randLength() decimal.Decimal {
	return decimal.NewFromFloat(float64(rand.Intn(9)+1) + float64(rand.Intn(100))/100)
}

func (db *model) createGIDTable(gid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Create(strconv.FormatInt(gid, 10), &userInfo{})
}

func (db *model) findniuniu(gid, uid int64) (float64, error) {
	db.RLock()
	defer db.RUnlock()
	u := userInfo{}
	err := db.sql.Find(strconv.FormatInt(gid, 10), &u, "where UID = "+strconv.FormatInt(uid, 10))
	return u.Length, err
}

func (db *model) insertniuniu(u userInfo, gid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Insert(strconv.FormatInt(gid, 10), &u)
}

func (db *model) deleteniuniu(gid, uid int64) error {
	db.Lock()
	defer db.Unlock()
	return db.sql.Del(strconv.FormatInt(gid, 10), "where UID = "+strconv.FormatInt(uid, 10))
}

func (db *model) readAllTable(gid int64) (users, error) {
	db.Lock()
	defer db.Unlock()
	a, err := sql.FindAll[userInfo](&db.sql, strconv.FormatInt(gid, 10), "where UserCount  = 1")
	return a, err
}
