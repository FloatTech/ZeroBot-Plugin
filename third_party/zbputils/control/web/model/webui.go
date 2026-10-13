// Package model 用户模型类
package model

import (
	"os"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// User webui用户数据
type User struct {
	ID       int64  `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
}

const webuiFolder = "data/webui/"

var (
	udb = sql.New(webuiFolder + "user.db")
	mu  sync.RWMutex
)

func init() {
	_ = os.MkdirAll(webuiFolder, 0755)
	err := udb.Open(time.Hour)
	if err != nil {
		panic(err)
	}
	err = udb.Create("user", &User{})
	if err != nil {
		panic(err)
	}
}

// CreateOrUpdateUser 创建或修改用户密码
func CreateOrUpdateUser(u *User) error {
	mu.RLock()
	defer mu.RUnlock()
	var fu User
	err := udb.Find("user", &fu, "WHERE username = '"+u.Username+"' AND password = '"+u.Password+"'")
	canFind := err == nil && fu.Username == u.Username
	if canFind {
		err = udb.Del("user", "WHERE username = '"+u.Username+"'")
		if err != nil {
			return err
		}
	}
	err = udb.Insert("user", u)
	return err
}

// FindUser 查找webui账号
func FindUser(username, password string) (u User, err error) {
	mu.Lock()
	defer mu.Unlock()
	err = udb.Find("user", &u, "WHERE username = '"+username+"' AND password = '"+password+"'")
	return
}
