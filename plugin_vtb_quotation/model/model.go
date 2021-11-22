package model

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type VtbDB gorm.DB

func Init(dbpath string) *VtbDB {
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
	return (*VtbDB)(unsafe.Pointer(gdb))
}

func Open(dbpath string) (*VtbDB, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	} else {
		return (*VtbDB)(unsafe.Pointer(db)), nil
	}
}

// FirstCategory 第一品类
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

// SecondCategory 第二品类
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

// ThirdCategory 第三品类
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

// GetAllFirstCategoryMessage 取出所有vtb
func (vdb *VtbDB) GetAllFirstCategoryMessage() string {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	firstStepMessage := "请选择一个vtb并发送序号:\n"
	var fc FirstCategory
	rows, err := db.Model(&FirstCategory{}).Rows()
	if err != nil {
		logrus.Errorln("[vtb/model]数据库读取错误", err)
	}
	if rows == nil {
		return ""
	}
	for rows.Next() {
		db.ScanRows(rows, &fc)
		// logrus.Println(fc)
		firstStepMessage = firstStepMessage + strconv.FormatInt(fc.FirstCategoryIndex, 10) + ". " + fc.FirstCategoryName + "\n"
	}
	return firstStepMessage
}

// GetAllSecondCategoryMessageByFirstIndex 取得同一个vtb所有语录类别
func (vdb *VtbDB) GetAllSecondCategoryMessageByFirstIndex(firstIndex int) string {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
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
		logrus.Errorln("[vtb/model]数据库读取错误", err)
	}

	for rows.Next() {
		db.ScanRows(rows, &sc)
		// logrus.Println(sc)
		SecondStepMessage = SecondStepMessage + strconv.FormatInt(sc.SecondCategoryIndex, 10) + ". " + sc.SecondCategoryName + "\n"
	}
	return SecondStepMessage
}

// GetAllThirdCategoryMessageByFirstIndexAndSecondIndex 取得同一个vtb同个类别的所有语录
func (vdb *VtbDB) GetAllThirdCategoryMessageByFirstIndexAndSecondIndex(firstIndex, secondIndex int) string {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
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
		logrus.Errorln("[vtb/model]数据库读取错误", err)
	}
	for rows.Next() {
		db.ScanRows(rows, &tc)
		// logrus.Println(tc)
		ThirdStepMessage = ThirdStepMessage + strconv.FormatInt(tc.ThirdCategoryIndex, 10) + ". " + tc.ThirdCategoryName + "\n"
	}
	return ThirdStepMessage
}

// GetThirdCategory
func (vdb *VtbDB) GetThirdCategory(firstIndex, secondIndex, thirdIndex int) ThirdCategory {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_index = ?", firstIndex).First(&fc)
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?", fc.FirstCategoryUid, secondIndex, thirdIndex).Take(&tc)
	return tc
}

func (vdb *VtbDB) RandomVtb() ThirdCategory {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	rand.Seed(time.Now().UnixNano())
	var count int
	db.Model(&ThirdCategory{}).Count(&count)
	// logrus.Info("一共有", count, "个")
	var tc ThirdCategory
	db.Model(&ThirdCategory{}).Offset(rand.Intn(count)).Take(&tc)
	// logrus.Info(tc)
	return tc
}

func (vdb *VtbDB) GetFirstCategoryByFirstUid(firstUid string) FirstCategory {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	var fc FirstCategory
	db.Model(FirstCategory{}).Where("first_category_uid = ?", firstUid).Take(&fc)
	// logrus.Info(fc)
	return fc
}

func (vdb *VtbDB) Close() error {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	return db.Close()
}

const vtbUrl = "https://vtbkeyboard.moe/api/get_vtb_list"

func (vdb *VtbDB) GetVtbList() []string {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbUrl, nil)
	if err != nil {
		logrus.Errorln(err)
	}
	// 自定义Header
	req.Header.Set("User-Agent", randua())
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorln(err)
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln(err)
	}
	// logrus.Println(string(bytes))
	vtbListStr, err := strconv.Unquote(strings.Replace(strconv.Quote(string(bytes)), `\\u`, `\u`, -1))
	if err != nil {
		logrus.Errorln(err)
	}
	// logrus.Println(vtbListStr)
	uidList := make([]string, 0)
	count := gjson.Get(vtbListStr, "#").Int()
	for i := int64(0); i < count; i++ {
		item := gjson.Get(vtbListStr, strconv.FormatInt(i, 10))
		logrus.Println(item)
		fc := FirstCategory{
			FirstCategoryIndex:       i,
			FirstCategoryName:        item.Get("name").String(),
			FirstCategoryDescription: item.Get("description").String(),
			FirstCategoryIconPath:    item.Get("icon_path").String(),
			FirstCategoryUid:         item.Get("uid").String(),
		}
		logrus.Println(fc)
		//db.Model(FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).FirstOrCreate(&fc)
		if err := db.Debug().Model(&FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).First(&fc).Error; err != nil {
			// error handling...
			if gorm.IsRecordNotFoundError(err) {
				db.Debug().Model(&FirstCategory{}).Create(&fc) // newUser not user
			}
		} else {
			db.Debug().Model(&FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).Update(
				map[string]interface{}{
					"first_category_index":       i,
					"first_category_name":        item.Get("name").String(),
					"first_category_description": item.Get("description").String(),
					"first_category_icon_path":   item.Get("icon_path").String(),
				})
		}
		uidList = append(uidList, fc.FirstCategoryUid)
	}

	// logrus.Println(uidList)
	return uidList
}

func (vdb *VtbDB) StoreVtb(uid string) {
	db := (*gorm.DB)(unsafe.Pointer(vdb))
	vtbUrl := "https://vtbkeyboard.moe/api/get_vtb_page?uid=" + uid
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbUrl, nil)
	if err != nil {
		logrus.Errorln(err)
	}
	// 自定义Header
	req.Header.Set("User-Agent", randua())
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorln(err)
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln(err)
	}
	//logrus.Println(string(bytes))
	vtbStr, err := strconv.Unquote(strings.Replace(strconv.Quote(string(bytes)), `\\u`, `\u`, -1))
	if err != nil {
		logrus.Errorln(err)
	}
	// logrus.Println(vtbListStr)
	secondCount := gjson.Get(vtbStr, "data.voices.#").Int()
	logrus.Println("二级品类一共有", secondCount)
	for secondIndex := int64(0); secondIndex < secondCount; secondIndex++ {
		secondItem := gjson.Get(vtbStr, "data.voices."+strconv.FormatInt(secondIndex, 10))
		logrus.Println(secondItem)
		sc := SecondCategory{
			SecondCategoryName:        secondItem.Get("categoryName").String(),
			SecondCategoryIndex:       secondIndex,
			SecondCategoryAuthor:      secondItem.Get("author").String(),
			SecondCategoryDescription: secondItem.Get("categoryDescription.zh-CN").String(),
			FirstCategoryUid:          uid,
		}
		// logrus.Println(sc)
		// db.Model(SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).FirstOrCreate(&sc)
		if err := db.Debug().Model(&SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).First(&sc).Error; err != nil {
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
		logrus.Println("三级品类一共有", thirdCount)
		for thirdIndex := int64(0); thirdIndex < thirdCount; thirdIndex++ {
			thirdItem := secondItem.Get("voiceList." + strconv.FormatInt(thirdIndex, 10))
			logrus.Println(thirdItem)
			tc := ThirdCategory{
				ThirdCategoryName:        thirdItem.Get("name").String(),
				ThirdCategoryIndex:       thirdIndex,
				ThirdCategoryDescription: thirdItem.Get("description.zh-CN").String(),
				FirstCategoryUid:         uid,
				SecondCategoryIndex:      secondIndex,
				ThirdCategoryPath:        thirdItem.Get("path").String(),
				ThirdCategoryAuthor:      thirdItem.Get("author").String(),
			}
			logrus.Println(tc)
			//db.Model(ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
			//	uid, secondIndex, thirdIndex).FirstOrCreate(&tc)
			if err := db.Debug().Model(&ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
				uid, secondIndex, thirdIndex).First(&tc).Error; err != nil {
				// error handling...
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

var agent = [...]string{
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
	"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
	"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
	"User-Agent,Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"User-Agent, Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
	"User-Agent,Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
}

func randua() string {
	return agent[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(agent))]
}
