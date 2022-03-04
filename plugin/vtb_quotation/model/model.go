// Package model vtb数据库操作
package model

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/FloatTech/zbputils/web"
	_ "github.com/fumiama/sqlite3" // import sql
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
func (vdb *VtbDB) GetAllFirstCategoryMessage() string {
	db := (*gorm.DB)(vdb)
	firstStepMessage := "请选择一个vtb并发送序号:\n"
	var fcl []FirstCategory
	err := db.Debug().Model(&FirstCategory{}).Find(&fcl).Error
	if err != nil {
		log.Errorln("[vtb/model]数据库读取错误", err)
		return ""
	}
	for _, v := range fcl {
		firstStepMessage += strconv.FormatInt(v.FirstCategoryIndex, 10) + ". " + v.FirstCategoryName + "\n"
	}
	return firstStepMessage
}

// GetAllSecondCategoryMessageByFirstIndex 取得同一个vtb所有语录类别
func (vdb *VtbDB) GetAllSecondCategoryMessageByFirstIndex(firstIndex int) string {
	db := (*gorm.DB)(vdb)
	secondStepMessage := "请选择一个语录类别并发送序号:\n"
	var scl []SecondCategory
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	err := db.Debug().Model(&SecondCategory{}).Find(&scl, "first_category_uid = ?", fc.FirstCategoryUID).Error
	if err != nil || len(scl) == 0 {
		log.Errorln("[vtb/model]数据库读取错误", err)
		return ""
	}
	for _, v := range scl {
		secondStepMessage += strconv.FormatInt(v.SecondCategoryIndex, 10) + ". " + v.SecondCategoryName + "\n"
	}
	return secondStepMessage
}

// GetAllThirdCategoryMessageByFirstIndexAndSecondIndex 取得同一个vtb同个类别的所有语录
func (vdb *VtbDB) GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(firstIndex, secondIndex int) string {
	db := (*gorm.DB)(vdb)
	thirdStepMessage := "请选择一个语录并发送序号:\n"
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	var tcl []ThirdCategory
	err := db.Debug().Model(&ThirdCategory{}).Find(&tcl, "first_category_uid = ? and second_category_index = ?", fc.FirstCategoryUID, secondIndex).Error
	if err != nil || len(tcl) == 0 {
		log.Errorln("[vtb/model]数据库读取错误", err)
		return ""
	}
	for _, v := range tcl {
		thirdStepMessage = thirdStepMessage + strconv.FormatInt(v.ThirdCategoryIndex, 10) + ". " + v.ThirdCategoryName + "\n"
	}
	return thirdStepMessage
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
func (vdb *VtbDB) GetVtbList() (uidList []string) {
	db := (*gorm.DB)(vdb)
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbURL, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	// 自定义Header
	req.Header.Set("User-Agent", web.RandUA())
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
		return
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln(err)
		return
	}

	vtbListStr, err := strconv.Unquote(strings.ReplaceAll(strconv.Quote(string(bytes)), `\\u`, `\u`))
	if err != nil {
		log.Errorln(err)
		return
	}

	count := gjson.Get(vtbListStr, "#").Int()
	for i := int64(0); i < count; i++ {
		item := gjson.Get(vtbListStr, strconv.FormatInt(i, 10))
		log.Println(item)
		fc := FirstCategory{
			FirstCategoryIndex:       i,
			FirstCategoryName:        item.Get("name").String(),
			FirstCategoryDescription: item.Get("description").String(),
			FirstCategoryIconPath:    item.Get("icon_path").String(),
			FirstCategoryUID:         item.Get("uid").String(),
		}
		log.Println(fc)

		if err := db.Debug().Model(&FirstCategory{}).First(&fc, "first_category_uid = ?", fc.FirstCategoryUID).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				db.Debug().Model(&FirstCategory{}).Create(&fc) // newUser not user
			}
		} else {
			db.Debug().Model(&FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUID).Update(
				map[string]interface{}{
					"first_category_index":       i,
					"first_category_name":        item.Get("name").String(),
					"first_category_description": item.Get("description").String(),
					"first_category_icon_path":   item.Get("icon_path").String(),
				})
		}
		uidList = append(uidList, fc.FirstCategoryUID)
	}

	return uidList
}

// StoreVtb ...
func (vdb *VtbDB) StoreVtb(uid string) {
	db := (*gorm.DB)(vdb)
	vtbURL := "https://vtbkeyboard.moe/api/get_vtb_page?uid=" + uid
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbURL, nil)
	if err != nil {
		log.Errorln(err)
		return
	}
	// 自定义Header
	req.Header.Set("User-Agent", web.RandUA())
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln(err)
		return
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln(err)
		return
	}

	vtbStr, err := strconv.Unquote(strings.ReplaceAll(strconv.Quote(string(bytes)), `\\u`, `\u`))
	if err != nil {
		log.Errorln(err)
		return
	}

	secondCount := gjson.Get(vtbStr, "data.voices.#").Int()
	log.Println("二级品类一共有", secondCount)
	for secondIndex := int64(0); secondIndex < secondCount; secondIndex++ {
		secondItem := gjson.Get(vtbStr, "data.voices."+strconv.FormatInt(secondIndex, 10))
		log.Println(secondItem)
		sc := SecondCategory{
			SecondCategoryName:        secondItem.Get("categoryName").String(),
			SecondCategoryIndex:       secondIndex,
			SecondCategoryAuthor:      secondItem.Get("author").String(),
			SecondCategoryDescription: secondItem.Get("categoryDescription.zh-CN").String(),
			FirstCategoryUID:          uid,
		}

		if err := db.Debug().Model(&SecondCategory{}).First(&sc, "first_category_uid = ? and second_category_index = ?", uid, secondIndex).Error; err != nil {
			// error handling...
			if gorm.IsRecordNotFoundError(err) {
				db.Debug().Model(&SecondCategory{}).Create(&sc) // newUser not user
			}
		} else {
			db.Debug().Model(&SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).Update(
				map[string]interface{}{
					"second_category_name":        secondItem.Get("categoryName").String(),
					"second_category_author":      secondItem.Get("author").String(),
					"second_category_description": secondItem.Get("categoryDescription.zh-CN").String(),
				})
		}
		thirdCount := secondItem.Get("voiceList.#").Int()
		log.Println("三级品类一共有", thirdCount)
		for thirdIndex := int64(0); thirdIndex < thirdCount; thirdIndex++ {
			thirdItem := secondItem.Get("voiceList." + strconv.FormatInt(thirdIndex, 10))
			log.Println(thirdItem)
			tc := ThirdCategory{
				ThirdCategoryName:        thirdItem.Get("name").String(),
				ThirdCategoryIndex:       thirdIndex,
				ThirdCategoryDescription: thirdItem.Get("description.zh-CN").String(),
				FirstCategoryUID:         uid,
				SecondCategoryIndex:      secondIndex,
				ThirdCategoryPath:        thirdItem.Get("path").String(),
				ThirdCategoryAuthor:      thirdItem.Get("author").String(),
			}
			log.Println(tc)

			if err := db.Debug().Model(&ThirdCategory{}).First(&tc, "first_category_uid = ? and second_category_index = ? and third_category_index = ?",
				uid, secondIndex, thirdIndex).Error; err != nil {
				if gorm.IsRecordNotFoundError(err) {
					db.Debug().Model(&ThirdCategory{}).Create(&tc) // newUser not user
				}
			} else {
				db.Debug().Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
					uid, secondIndex, thirdIndex).Update(
					map[string]interface{}{
						"third_category_name":        thirdItem.Get("name").String(),
						"third_category_description": thirdItem.Get("description.zh-CN").String(),
						"third_category_path":        thirdItem.Get("path").String(),
						"third_category_author":      thirdItem.Get("author").String(),
					})
			}
		}
	}
}
