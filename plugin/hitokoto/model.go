package hitokoto

import (
	"os"

	"github.com/jinzhu/gorm"
)

// hdb 表情包数据库全局变量
var hdb *hitokotodb

// hitokotodb 表情包数据库
type hitokotodb gorm.DB

// initialize 初始化
func initialize(dbpath string) (db *hitokotodb, err error) {
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil, err
		}
		_ = f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return
	}
	gdb.AutoMigrate(&hitokoto{})
	return (*hitokotodb)(gdb), nil
}

type hitokoto struct {
	ID         int    `json:"id" gorm:"column:id;primary_key"`
	Hitokoto   string `json:"hitokoto" gorm:"column:hitokoto"`
	Type       string `json:"type" gorm:"column:type"`
	From       string `json:"from" gorm:"column:from"`
	FromWho    string `json:"from_who" gorm:"column:from_who"`
	Creator    string `json:"creator" gorm:"column:creator"`
	CreatorUID int    `json:"creator_uid" gorm:"column:creator_uid"`
	Reviewer   int    `json:"reviewer" gorm:"column:reviewer"`
	UUID       string `json:"uuid" gorm:"column:uuid"`
	CreatedAt  string `json:"created_at" gorm:"column:created_at"`
	Category   string `json:"catogory" gorm:"column:category"`
}

// TableName 表名
func (hitokoto) TableName() string {
	return "hitokoto"
}

func (hdb *hitokotodb) getByKey(key string) (b []hitokoto, err error) {
	db := (*gorm.DB)(hdb)
	err = db.Where("hitokoto like ?", "%"+key+"%").Find(&b).Error
	return
}

type result struct {
	Category string
	Count    int
}

func (hdb *hitokotodb) getAllCategory() (results []result, err error) {
	db := (*gorm.DB)(hdb)
	err = db.Table("hitokoto").Select("category, count(1) as count").Group("category").Scan(&results).Error
	return
}

func (hdb *hitokotodb) getByCategory(category string) (h []hitokoto, err error) {
	db := (*gorm.DB)(hdb)
	err = db.Where("category = ?", category).Find(&h).Error
	return
}
