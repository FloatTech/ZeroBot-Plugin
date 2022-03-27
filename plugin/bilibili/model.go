package bilibili

import (
	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/web"
	_ "github.com/fumiama/sqlite3"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"os"
)

const (
	bilibiliCookie = "bilbili_cookie"
)

var (
	vtbURLs   = [...]string{"https://api.vtbs.moe/v1/short", "https://api.tokyo.vtbs.moe/v1/short", "https://vtbs.musedash.moe/v1/short"}
	dbfile    = engine.DataFolder() + "bilibili.db"
	vdb       *vupdb
	cachePath = engine.DataFolder() + "cache/"
)

// vupdb 分数数据库
type vupdb gorm.DB

type vup struct {
	Mid    int64  `gorm:"column:mid;primary_key"`
	Uname  string `gorm:"column:uname"`
	Roomid int64  `gorm:"column:roomid"`
}

func (vup) TableName() string {
	return "vup"
}

type config struct {
	Key   string `gorm:"column:key;primary_key"`
	Value string `gorm:"column:value"`
}

func (config) TableName() string {
	return "config"
}

// initialize 初始化vtb数据库
func initialize(dbpath string) *vupdb {
	if _, err := os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		panic(err)
	}
	gdb.Debug().AutoMigrate(&vup{}).AutoMigrate(&config{})
	return (*vupdb)(gdb)
}

func (vdb *vupdb) insertVupByMid(mid int64, uname string, roomid int64) (err error) {
	db := (*gorm.DB)(vdb)
	v := vup{
		Mid:    mid,
		Uname:  uname,
		Roomid: roomid,
	}
	if err = db.Debug().Model(&vup{}).First(&v, "mid = ? ", mid).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Debug().Model(&vup{}).Create(&v).Error
		}
	}
	return
}

// filterVup 筛选vup
func (vdb *vupdb) filterVup(ids []int64) (vups []vup, err error) {
	db := (*gorm.DB)(vdb)
	if err = db.Debug().Model(&vup{}).Find(&vups, "mid in (?)", ids).Error; err != nil {
		return vups, err
	}
	return
}

func updateVup() {
	for _, v := range vtbURLs {
		data, err := web.GetData(v)
		if err != nil {
			log.Errorln("[bilibili]:", err)
		}
		gjson.Get(binary.BytesToString(data), "@this").ForEach(func(key, value gjson.Result) bool {
			mid := value.Get("mid").Int()
			uname := value.Get("uname").String()
			roomid := value.Get("roomid").Int()
			vdb.insertVupByMid(mid, uname, roomid)
			return true
		})
	}
}

func (vdb *vupdb) setBilibiliCookie(cookie string) (err error) {
	db := (*gorm.DB)(vdb)
	c := config{
		Key:   bilibiliCookie,
		Value: cookie,
	}
	if err = db.Debug().Model(&config{}).First(&c, "key = ? ", bilibiliCookie).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			err = db.Debug().Model(&config{}).Create(&c).Error
		}
	} else {
		err = db.Debug().Model(&config{}).Where("key = ? ", bilibiliCookie).Update(
			map[string]interface{}{
				"value": cookie,
			}).Error
	}
	return
}

func (vdb *vupdb) getBilibiliCookie() (c config) {
	db := (*gorm.DB)(vdb)
	db.Debug().Model(&config{}).First(&c, "key = ?", bilibiliCookie)
	return
}
