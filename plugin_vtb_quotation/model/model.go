package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
	log "github.com/sirupsen/logrus"

	"math/rand"
	"os"
	"strconv"
	"time"
)

var (
	Db   *gorm.DB
	Path = "data/VtbQuotation/vtb.db"
)

func Init() {
	var err error
	if _, err = os.Stat(Path); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(Path)
		if err != nil {
			return
		}
		defer f.Close()
	}
	Db, err = gorm.Open("sqlite3", Path)
	if err != nil {
		panic("failed to connect database")
	}
	Db.AutoMigrate(FirstCategory{}).AutoMigrate(SecondCategory{}).AutoMigrate(ThirdCategory{})
}

//第一品类
type FirstCategory struct {
	gorm.Model
	FirstCategoryIndex       int64  `gorm:"column:first_category_index"`
	FirstCategoryName        string `gorm:"column:first_category_name"`
	FirstCategoryUid         string `gorm:"column:first_category_uid"`
	FirstCategoryDescription string `gorm:"column:first_category_description;type:varchar(1024)"`
	FirstCategoryIconPath    string `gorm:"column:first_category_icon_path"`
}

func (FirstCategory) TableName() string {
	return "first_category"
}

//第二品类
type SecondCategory struct {
	gorm.Model
	SecondCategoryIndex       int64  `gorm:"column:second_category_index"`
	FirstCategoryUid          string `gorm:"column:first_category_uid;association_foreignkey:first_category_uid"`
	SecondCategoryName        string `gorm:"column:second_category_name"`
	SecondCategoryAuthor      string `gorm:"column:second_category_author"`
	SecondCategoryDescription string `gorm:"column:second_category_description"`
}

func (SecondCategory) TableName() string {
	return "second_category"
}

//第三品类
type ThirdCategory struct {
	gorm.Model
	ThirdCategoryIndex       int64  `gorm:"column:third_category_index"`
	SecondCategoryIndex      int64  `gorm:"column:second_category_index"`
	FirstCategoryUid         string `gorm:"column:first_category_uid"`
	ThirdCategoryName        string `gorm:"column:third_category_name"`
	ThirdCategoryPath        string `gorm:"column:third_category_path"`
	ThirdCategoryAuthor      string `gorm:"column:third_category_author"`
	ThirdCategoryDescription string `gorm:"column:third_category_description"`
}

func (ThirdCategory) TableName() string {
	return "third_category"
}

//取出所有vtb
func GetAllFirstCategoryMessage(db *gorm.DB) string {
	firstStepMessage := "请选择一个vtb并发送序号:\n"
	var fc FirstCategory
	rows, err := db.Model(&FirstCategory{}).Rows()
	if err != nil {
		log.Println("数据库读取错误", err)
	}
	if rows == nil {
		return ""
	}
	for rows.Next() {
		db.ScanRows(rows, &fc)
		log.Println(fc)
		firstStepMessage = firstStepMessage + strconv.FormatInt(fc.FirstCategoryIndex, 10) + ". " + fc.FirstCategoryName + "\n"
	}
	return firstStepMessage
}

//取得同一个vtb所有语录类别
func GetAllSecondCategoryMessageByFirstIndex(db *gorm.DB, firstIndex int) string {
	SecondStepMessage := "请选择一个语录类别并发送序号:\n"
	var sc SecondCategory
	var count int
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	db.Model(&SecondCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).Count(&count)
	if count == 0 {
		return ""
	}
	rows, err := db.Model(&SecondCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).Rows()
	if err != nil {
		log.Println("数据库读取错误", err)
	}

	for rows.Next() {
		db.ScanRows(rows, &sc)
		log.Println(sc)
		SecondStepMessage = SecondStepMessage + strconv.FormatInt(sc.SecondCategoryIndex, 10) + ". " + sc.SecondCategoryName + "\n"
	}
	return SecondStepMessage
}

//取得同一个vtb同个类别的所有语录
func GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(db *gorm.DB, firstIndex, secondIndex int) string {
	ThirdStepMessage := "请选择一个语录并发送序号:\n"
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	var count int
	db.Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ?", fc.FirstCategoryUid, secondIndex).Count(&count)
	if count == 0 {
		return ""
	}
	var tc ThirdCategory
	rows, err := db.Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ?", fc.FirstCategoryUid, secondIndex).Rows()
	if err != nil {
		log.Println("数据库读取错误", err)
	}
	for rows.Next() {
		db.ScanRows(rows, &tc)
		log.Println(tc)
		ThirdStepMessage = ThirdStepMessage + strconv.FormatInt(tc.ThirdCategoryIndex, 10) + ". " + tc.ThirdCategoryName + "\n"
	}
	return ThirdStepMessage
}
func GetThirdCategory(db *gorm.DB, firstIndex, secondIndex, thirdIndex int) ThirdCategory {
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?", fc.FirstCategoryUid, secondIndex, thirdIndex).Take(&tc)
	return tc
}

func RandomVtb(db *gorm.DB) ThirdCategory {
	rand.Seed(time.Now().UnixNano())
	var count int
	db.Model(&ThirdCategory{}).Count(&count)
	log.Info("一共有", count, "个")
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Offset(rand.Intn(count)).Take(&tc)
	log.Info(tc)
	return tc
}

func GetFirstCategoryByFirstUid(db *gorm.DB, firstUid string) FirstCategory {
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_uid = ?", firstUid).Take(&fc)
	log.Info(fc)
	return fc
}
