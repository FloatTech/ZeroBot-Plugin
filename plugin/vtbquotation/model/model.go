// Package model vtb数据库操作
package model

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/floatbox/web"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// VtbDB vtb 数据库
type VtbDB gorm.DB

// Initialize ...
func Initialize(dbpath string) *VtbDB {
	var err error
	if _, err = os.Stat(dbpath); err != nil || os.IsNotExist(err) {
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
	gdb.AutoMigrate(FirstCategory{}).AutoMigrate(SecondCategory{}).AutoMigrate(ThirdCategory{})
	return (*VtbDB)(gdb)
}

// Open ...
func Open(dbpath string) (*VtbDB, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	return (*VtbDB)(db), nil
}

// FirstCategory 第一品类
type FirstCategory struct {
	gorm.Model
	FirstCategoryIndex       int64  `gorm:"column:first_category_index"`
	FirstCategoryName        string `gorm:"column:first_category_name"`
	FirstCategoryUID         string `gorm:"column:first_category_uid"`
	FirstCategoryDescription string `gorm:"column:first_category_description;type:varchar(1024)"`
	FirstCategoryIconPath    string `gorm:"column:first_category_icon_path"`
}

// TableName ...
func (FirstCategory) TableName() string {
	return "first_category"
}

// SecondCategory 第二品类
type SecondCategory struct {
	gorm.Model
	SecondCategoryIndex       int64  `gorm:"column:second_category_index"`
	FirstCategoryUID          string `gorm:"column:first_category_uid;association_foreignkey:first_category_uid"`
	SecondCategoryName        string `gorm:"column:second_category_name"`
	SecondCategoryAuthor      string `gorm:"column:second_category_author"`
	SecondCategoryDescription string `gorm:"column:second_category_description"`
}

// TableName ...
func (SecondCategory) TableName() string {
	return "second_category"
}

// ThirdCategory 第三品类
type ThirdCategory struct {
	gorm.Model
	ThirdCategoryIndex       int64  `gorm:"column:third_category_index"`
	SecondCategoryIndex      int64  `gorm:"column:second_category_index"`
	FirstCategoryUID         string `gorm:"column:first_category_uid"`
	ThirdCategoryName        string `gorm:"column:third_category_name"`
	ThirdCategoryPath        string `gorm:"column:third_category_path"`
	ThirdCategoryAuthor      string `gorm:"column:third_category_author"`
	ThirdCategoryDescription string `gorm:"column:third_category_description"`
}

// TableName ...
func (ThirdCategory) TableName() string {
	return "third_category"
}

// GetAllFirstCategoryMessage 取出所有vtb
func (vdb *VtbDB) GetAllFirstCategoryMessage() (string, error) {
	db := (*gorm.DB)(vdb)
	firstStepMessage := "请选择一个vtb并发送序号:\n"
	var fcl []FirstCategory
	err := db.Model(&FirstCategory{}).Find(&fcl).Error
	if err != nil {
		return "", err
	}
	for _, v := range fcl {
		firstStepMessage += strconv.FormatInt(v.FirstCategoryIndex, 10) + ". " + v.FirstCategoryName + "\n"
	}
	return firstStepMessage, nil
}

// GetAllSecondCategoryMessageByFirstIndex 取得同一个vtb所有语录类别
func (vdb *VtbDB) GetAllSecondCategoryMessageByFirstIndex(firstIndex int) (string, error) {
	db := (*gorm.DB)(vdb)
	secondStepMessage := "请选择一个语录类别并发送序号:\n"
	var scl []SecondCategory
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	err := db.Model(&SecondCategory{}).Find(&scl, "first_category_uid = ?", fc.FirstCategoryUID).Error
	if err != nil || len(scl) == 0 {
		return "", err
	}
	for _, v := range scl {
		secondStepMessage += strconv.FormatInt(v.SecondCategoryIndex, 10) + ". " + v.SecondCategoryName + "\n"
	}
	return secondStepMessage, nil
}

// GetAllThirdCategoryMessageByFirstIndexAndSecondIndex 取得同一个vtb同个类别的所有语录
func (vdb *VtbDB) GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(firstIndex, secondIndex int) (string, error) {
	db := (*gorm.DB)(vdb)
	thirdStepMessage := "请选择一个语录并发送序号:\n"
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	var tcl []ThirdCategory
	err := db.Model(&ThirdCategory{}).Find(&tcl, "first_category_uid = ? and second_category_index = ?", fc.FirstCategoryUID, secondIndex).Error
	if err != nil || len(tcl) == 0 {
		return "", err
	}
	for _, v := range tcl {
		thirdStepMessage = thirdStepMessage + strconv.FormatInt(v.ThirdCategoryIndex, 10) + ". " + v.ThirdCategoryName + "\n"
	}
	return thirdStepMessage, nil
}

// GetThirdCategory ...
func (vdb *VtbDB) GetThirdCategory(firstIndex, secondIndex, thirdIndex int) ThirdCategory {
	db := (*gorm.DB)(vdb)
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?", fc.FirstCategoryUID, secondIndex, thirdIndex).Take(&tc)
	return tc
}

// RandomVtb ...
func (vdb *VtbDB) RandomVtb() ThirdCategory {
	db := (*gorm.DB)(vdb)
	var count int
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Count(&count).Offset(rand.Intn(count)).Take(&tc)
	return tc
}

// GetFirstCategoryByFirstUID ...
func (vdb *VtbDB) GetFirstCategoryByFirstUID(firstUID string) FirstCategory {
	db := (*gorm.DB)(vdb)
	var fc FirstCategory
	db.Model(FirstCategory{}).Take(&fc, "first_category_uid = ?", firstUID)
	return fc
}

// Close ...
func (vdb *VtbDB) Close() error {
	db := (*gorm.DB)(vdb)
	return db.Close()
}

const vtbURL = "https://vtbkeyboard.moe/api/get_vtb_list"

// GetVtbList ...
func (vdb *VtbDB) GetVtbList() (uidList []string, err error) {
	db := (*gorm.DB)(vdb)
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbURL, nil)
	if err != nil {
		return
	}
	// 自定义Header
	req.Header.Set("User-Agent", web.RandUA())
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	vtbListStr, err := strconv.Unquote(strings.ReplaceAll(strconv.Quote(string(bytes)), `\\u`, `\u`))
	if err != nil {
		return
	}

	count := gjson.Get(vtbListStr, "#").Int()
	for i := int64(0); i < count; i++ {
		item := gjson.Get(vtbListStr, strconv.FormatInt(i, 10))
		log.Debugln(item)
		fc := FirstCategory{
			FirstCategoryIndex:       i,
			FirstCategoryName:        item.Get("name").String(),
			FirstCategoryDescription: item.Get("description").String(),
			FirstCategoryIconPath:    item.Get("icon_path").String(),
			FirstCategoryUID:         item.Get("uid").String(),
		}
		log.Debugln(fc)

		if err := db.Model(&FirstCategory{}).First(&fc, "first_category_uid = ?", fc.FirstCategoryUID).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				db.Model(&FirstCategory{}).Create(&fc) // newUser not user
			}
		} else {
			db.Model(&FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUID).Update(
				map[string]any{
					"first_category_index":       i,
					"first_category_name":        item.Get("name").String(),
					"first_category_description": item.Get("description").String(),
					"first_category_icon_path":   item.Get("icon_path").String(),
				})
		}
		uidList = append(uidList, fc.FirstCategoryUID)
	}

	return
}

// StoreVtb ...
func (vdb *VtbDB) StoreVtb(uid string) (err error) {
	db := (*gorm.DB)(vdb)
	vtbURL := "https://vtbkeyboard.moe/api/get_vtb_page?uid=" + uid
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbURL, nil)
	if err != nil {
		return
	}
	// 自定义Header
	req.Header.Set("User-Agent", web.RandUA())
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	vtbStr, err := strconv.Unquote(strings.ReplaceAll(strconv.Quote(string(bytes)), `\\u`, `\u`))
	if err != nil {
		return
	}

	secondCount := gjson.Get(vtbStr, "data.voices.#").Int()
	log.Debugln("二级品类一共有", secondCount)
	for secondIndex := int64(0); secondIndex < secondCount; secondIndex++ {
		secondItem := gjson.Get(vtbStr, "data.voices."+strconv.FormatInt(secondIndex, 10))
		log.Debugln(secondItem)
		sc := SecondCategory{
			SecondCategoryName:        secondItem.Get("categoryName").String(),
			SecondCategoryIndex:       secondIndex,
			SecondCategoryAuthor:      secondItem.Get("author").String(),
			SecondCategoryDescription: secondItem.Get("categoryDescription.zh-CN").String(),
			FirstCategoryUID:          uid,
		}

		if err := db.Model(&SecondCategory{}).First(&sc, "first_category_uid = ? and second_category_index = ?", uid, secondIndex).Error; err != nil {
			// error handling...
			if gorm.IsRecordNotFoundError(err) {
				db.Model(&SecondCategory{}).Create(&sc) // newUser not user
			}
		} else {
			db.Model(&SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).Update(
				map[string]any{
					"second_category_name":        secondItem.Get("categoryName").String(),
					"second_category_author":      secondItem.Get("author").String(),
					"second_category_description": secondItem.Get("categoryDescription.zh-CN").String(),
				})
		}
		thirdCount := secondItem.Get("voiceList.#").Int()
		log.Debugln("三级品类一共有", thirdCount)
		for thirdIndex := int64(0); thirdIndex < thirdCount; thirdIndex++ {
			thirdItem := secondItem.Get("voiceList." + strconv.FormatInt(thirdIndex, 10))
			log.Debugln(thirdItem)
			tc := ThirdCategory{
				ThirdCategoryName:        thirdItem.Get("name").String(),
				ThirdCategoryIndex:       thirdIndex,
				ThirdCategoryDescription: thirdItem.Get("description.zh-CN").String(),
				FirstCategoryUID:         uid,
				SecondCategoryIndex:      secondIndex,
				ThirdCategoryPath:        thirdItem.Get("path").String(),
				ThirdCategoryAuthor:      thirdItem.Get("author").String(),
			}
			log.Debugln(tc)

			if err := db.Model(&ThirdCategory{}).First(&tc, "first_category_uid = ? and second_category_index = ? and third_category_index = ?",
				uid, secondIndex, thirdIndex).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					db.Model(&ThirdCategory{}).Create(&tc) // newUser not user
				}
			} else {
				db.Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
					uid, secondIndex, thirdIndex).Update(
					map[string]any{
						"third_category_name":        thirdItem.Get("name").String(),
						"third_category_description": thirdItem.Get("description.zh-CN").String(),
						"third_category_path":        thirdItem.Get("path").String(),
						"third_category_author":      thirdItem.Get("author").String(),
					})
			}
		}
	}
	return
}
