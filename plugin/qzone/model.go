package qzone

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
)

// qdb qq空间数据库全局变量
var qdb *qzonedb

// qzonedb qq空间数据库结构体
type qzonedb gorm.DB

// initialize 初始化
func initialize(dbpath string) *qzonedb {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return nil
		}
		defer f.Close()
	}
	qdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		panic(err)
	}
	qdb.AutoMigrate(&qzoneConfig{}).AutoMigrate(&emotion{})
	return (*qzonedb)(qdb)
}

// qzoneConfig qq空间初始化信息
type qzoneConfig struct {
	ID     uint   `gorm:"primary_key;AUTO_INCREMENT"`
	QQ     int64  `gorm:"column:qq;unique;not null"`
	Cookie string `gorm:"column:cookie;type:varchar(1024)"`
}

// TableName 表名
func (qzoneConfig) TableName() string {
	return "qzone_config"
}

func (qdb *qzonedb) insertOrUpdate(qq int64, cookie string) (err error) {
	db := (*gorm.DB)(qdb)
	qc := qzoneConfig{
		QQ:     qq,
		Cookie: cookie,
	}
	var oqc qzoneConfig
	err = db.Take(&oqc, "qq = ?", qc.QQ).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Create(&qc).Error
		}
		return
	}
	err = db.Model(&oqc).Updates(qc).Error
	return
}

func (qdb *qzonedb) getByUin(qq int64) (qc qzoneConfig, err error) {
	db := (*gorm.DB)(qdb)
	err = db.Take(&qc, "qq = ?", qq).Error
	return
}

// emotion 说说信息
type emotion struct {
	gorm.Model
	Anonymous bool   `gorm:"column:anonymous"`
	QQ        int64  `gorm:"column:qq"`
	Msg       string `gorm:"column:msg"`
	Status    int    `gorm:"column:status"` // 1-审核中,2-同意,3-拒绝
	Tag       string `gorm:"column:tag"`
}

func (e emotion) textBrief() (t string) {
	t = fmt.Sprintf("序号: %v\nQQ: %v\n创建时间: %v\n", e.ID, e.QQ, e.CreatedAt.Format("2006-01-02 15:04:05"))
	switch e.Status {
	case 1:
		t += "状态: 审核中\n"
	case 2:
		t += "状态: 同意\n"
	case 3:
		t += "状态: 拒绝\n"
	}
	if e.Anonymous {
		t += "匿名: 是"
	} else {
		t += "匿名: 否"
	}
	return
}

// TableName 表名
func (emotion) TableName() string {
	return "emotion"
}

func (qdb *qzonedb) saveEmotion(e emotion) (id int64, err error) {
	db := (*gorm.DB)(qdb)
	err = db.Create(&e).Error
	id = int64(e.ID)
	return
}

func (qdb *qzonedb) getEmotionByIDList(idList []int64) (el []emotion, err error) {
	db := (*gorm.DB)(qdb)
	err = db.Find(&el, "id in (?)", idList).Error
	return
}

func (qdb *qzonedb) getLoveEmotionByStatus(status int, pageNum int) (el []emotion, err error) {
	db := (*gorm.DB)(qdb)
	if status == 0 {
		err = db.Order("created_at desc").Limit(5).Offset(pageNum*5).Find(&el, "tag like ?", "%"+loveTag+"%").Error
		return
	}
	err = db.Order("created_at desc").Limit(5).Offset(pageNum*5).Find(&el, "status = ? and tag like ?", status, "%"+loveTag+"%").Error
	return
}

func (qdb *qzonedb) updateEmotionStatusByIDList(idList []int64, status int) (err error) {
	db := (*gorm.DB)(qdb)
	err = db.Model(&emotion{}).Where("id in (?)", idList).Update("status", status).Error
	return
}
