// Package wallet 公用系统
package wallet

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	sql "github.com/FloatTech/sqlite"
)

// Storage 货币系统
type Storage struct {
	sync.RWMutex
	db *sql.Sqlite
}

// Wallet 钱包
type Wallet struct {
	UID   int64
	Money int
}

// NameData 昵称信息
type NameData struct {
	UID  int64
	Name string
}

var (
	sdb = &Storage{
		db: &sql.Sqlite{
			DBPath: "data/wallet/wallet.db",
		},
	}
)

func init() {
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
	err = sdb.db.Create("storage", &Wallet{})
	if err != nil {
		panic(err)
	}
	err = sdb.db.Create("names", &Wallet{})
	if err != nil {
		panic(err)
	}
}

// GetWalletOf 获取钱包数据
func GetWalletOf(uid int64) (money int) {
	return sdb.getWalletOf(uid).Money
}

// GetWalletInfoGroup 获取多人钱包数据
//
// if sort == true,由高到低排序; if sort == false,由低到高排序
func GetGroupWalletOf(uids []int64, sortable bool) (wallets []Wallet, err error) {
	return sdb.getGroupWalletOf(uids, sortable)
}

// InsertWalletOf 更新钱包(money > 0 增加,money < 0 减少)
func InsertWalletOf(uid int64, money int) error {
	lastMoney := sdb.getWalletOf(uid)
	return sdb.updateWalletOf(uid, lastMoney.Money+money)
}

// 获取钱包数据
func (sql *Storage) getWalletOf(uid int64) (Wallet Wallet) {
	sql.RLock()
	defer sql.RUnlock()
	uidstr := strconv.FormatInt(uid, 10)
	_ = sql.db.Find("storage", &Wallet, "where uid is "+uidstr)
	return
}

// 获取钱包数据组
func (sql *Storage) getGroupWalletOf(uids []int64, issorted bool) (wallets []Wallet, err error) {
	uidstr := make([]string, 0, len(uids))
	for _, uid := range uids {
		uidstr = append(uidstr, strconv.FormatInt(uid, 10))
	}
	sql.RLock()
	defer sql.RUnlock()
	wallets = make([]Wallet, 0, len(uids))
	sort := "ASC"
	if issorted {
		sort = "DESC"
	}
	info := Wallet{}
	err = sql.db.FindFor("storage", &info, "where uid IN ("+strings.Join(uidstr, ", ")+") ORDER BY money "+sort, func() error {
		wallets = append(wallets, info)
		return nil
	})
	return
}

// 更新钱包
func (sql *Storage) updateWalletOf(uid int64, money int) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Insert("storage", &Wallet{
		UID:   uid,
		Money: money,
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
func (sql *Storage) getNameOf(uid int64) (name string) {
	sql.RLock()
	defer sql.RUnlock()
	uidstr := strconv.FormatInt(uid, 10)
	var userdata NameData
	_ = sql.db.Find("names", &userdata, "where uid is "+uidstr)
	return userdata.Name
}

// 记录数据
func (sql *Storage) insertNameOf(uid int64, name string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Insert("names", &NameData{
		UID:  uid,
		Name: name,
	})
}

// 删除数据
func (sql *Storage) del(name string) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Del("names", "where Name is "+name)
}
