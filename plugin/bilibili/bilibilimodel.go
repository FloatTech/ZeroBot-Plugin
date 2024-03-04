package bilibili

import (
	"os"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	"github.com/jinzhu/gorm"
	"github.com/tidwall/gjson"
)

var (
	vtbURLs = [...]string{"https://api.vtbs.moe/v1/short", "https://api.tokyo.vtbs.moe/v1/short", "https://vtbs.musedash.moe/v1/short"}
	vdb     *vupdb
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

// initializeVup 初始化vup数据库
func initializeVup(dbpath string) (*vupdb, error) {
	if _, err := os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	gdb.AutoMigrate(&vup{})
	return (*vupdb)(gdb), nil
}

func (vdb *vupdb) insertVupByMid(mid int64, uname string, roomid int64) (err error) {
	db := (*gorm.DB)(vdb)
	v := vup{
		Mid:    mid,
		Uname:  uname,
		Roomid: roomid,
	}
	if err = db.Model(&vup{}).First(&v, "mid = ? ", mid).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Model(&vup{}).Create(&v).Error
		}
	}
	return
}

// filterVup 筛选vup
func (vdb *vupdb) filterVup(ids []int64) (vups []vup, err error) {
	db := (*gorm.DB)(vdb)
	if err = db.Model(&vup{}).Find(&vups, "mid in (?)", ids).Error; err != nil {
		return vups, err
	}
	return
}

func updateVup() error {
	for _, v := range vtbURLs {
		data, err := web.GetData(v)
		if err != nil {
			return err
		}
		gjson.Get(binary.BytesToString(data), "@this").ForEach(func(_, value gjson.Result) bool {
			mid := value.Get("mid").Int()
			uname := value.Get("uname").String()
			roomid := value.Get("roomid").Int()
			err = vdb.insertVupByMid(mid, uname, roomid)
			return err == nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
