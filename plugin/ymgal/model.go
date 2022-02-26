package ymgal

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	_ "github.com/fumiama/sqlite3" // import sql
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// gdb 得分数据库
var gdb *ymgaldb

// ymgaldb galgame图片数据库
type ymgaldb gorm.DB

var mu sync.RWMutex

// ymgal gal图片储存结构体
type ymgal struct {
	ID                 int64  `gorm:"column:id" `
	Title              string `gorm:"column:title" `
	PictureType        string `gorm:"column:picture_type" `
	PictureDescription string `gorm:"column:picture_description;type:varchar(1024)" `
	PictureList        string `gorm:"column:picture_list;type:varchar(20000)" `
}

// TableName ...
func (ymgal) TableName() string {
	return "ymgal"
}

// initialize 初始化ymgaldb数据库
func initialize(dbpath string) *ymgaldb {
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
	gdb.AutoMigrate(&ymgal{})
	return (*ymgaldb)(gdb)
}

func (gdb *ymgaldb) insertOrUpdateYmgalByID(id int64, title, pictureType, pictureDescription, pictureList string) (err error) {
	db := (*gorm.DB)(gdb)
	y := ymgal{
		ID:                 id,
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	if err = db.Debug().Model(&ymgal{}).First(&y, "id = ? ", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Debug().Model(&ymgal{}).Create(&y).Error // newUser not user
		}
	} else {
		err = db.Debug().Model(&ymgal{}).Where("id = ? ", id).Update(map[string]interface{}{
			"title":               title,
			"picture_type":        pictureType,
			"picture_description": pictureDescription,
			"picture_list":        pictureList,
		}).Error
	}
	return
}

func (gdb *ymgaldb) getYmgalByID(id string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	db.Debug().Model(&ymgal{}).Where("id = ?", id).Take(&y)
	return
}

func (gdb *ymgaldb) randomYmgal(pictureType string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	var count int
	s := db.Debug().Model(&ymgal{}).Where("picture_type = ?", pictureType).Count(&count)
	if count == 0 {
		return
	}
	s.Offset(rand.Intn(count)).Take(&y)
	return
}

func (gdb *ymgaldb) getYmgalByKey(pictureType, key string) (y ymgal) {
	db := (*gorm.DB)(gdb)
	var count int
	s := db.Debug().Model(&ymgal{}).Where("picture_type = ? and (picture_description like ? or title like ?) ", pictureType, "%"+key+"%", "%"+key+"%").Count(&count)
	if count == 0 {
		return
	}
	s.Offset(rand.Intn(count)).Take(&y)
	return
}

const (
	webURL       = "https://www.ymgal.com"
	cgType       = "Gal CG"
	emoticonType = "其他"
	webPicURL    = webURL + "/co/picset/"
	reNumber     = `\d+`
)

var (
	cgURL                = webURL + "/search?type=picset&sort=default&category=" + url.QueryEscape(cgType) + "&page="
	emoticonURL          = webURL + "/search?type=picset&sort=default&category=" + url.QueryEscape(emoticonType) + "&page="
	commonPageNumberExpr = "//*[@id='pager-box']/div/a[@class='icon item pager-next']/preceding-sibling::a[1]/text()"
	cgIDList             []string
	emoticonIDList       []string
)

func initPageNumber() (maxCgPageNumber, maxEmoticonPageNumber int) {
	doc, err := htmlquery.LoadURL(cgURL + "1")
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	maxCgPageNumber, err = strconv.Atoi(htmlquery.FindOne(doc, commonPageNumberExpr).Data)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	doc, err = htmlquery.LoadURL(emoticonURL + "1")
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	maxEmoticonPageNumber, err = strconv.Atoi(htmlquery.FindOne(doc, commonPageNumberExpr).Data)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	return
}

func getPicID(pageNumber int, pictureType string) {
	var picURL string
	if pictureType == cgType {
		picURL = cgURL + strconv.Itoa(pageNumber)
	} else if pictureType == emoticonType {
		picURL = emoticonURL + strconv.Itoa(pageNumber)
	}
	doc, err := htmlquery.LoadURL(picURL)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	list := htmlquery.Find(doc, "//*[@id='picset-result-list']/ul/div/div[1]/a")
	for i := 0; i < len(list); i++ {
		re := regexp.MustCompile(reNumber)
		picID := re.FindString(list[i].Attr[0].Val)
		if pictureType == cgType {
			cgIDList = append(cgIDList, picID)
		} else if pictureType == emoticonType {
			emoticonIDList = append(emoticonIDList, picID)
		}
	}

}

func updatePic() {
	maxCgPageNumber, maxEmoticonPageNumber := initPageNumber()
	for i := 1; i <= maxCgPageNumber; i++ {
		getPicID(i, cgType)
		time.Sleep(time.Millisecond * 500)
	}
	for i := 1; i <= maxEmoticonPageNumber; i++ {
		getPicID(i, emoticonType)
		time.Sleep(time.Millisecond * 500)
	}
CGLOOP:
	for i := len(cgIDList) - 1; i >= 0; i-- {
		mu.RLock()
		y := gdb.getYmgalByID(cgIDList[i])
		mu.RUnlock()
		if y.PictureList == "" {
			mu.Lock()
			storeCgPic(cgIDList[i])
			mu.Unlock()
		} else {
			break CGLOOP
		}
		time.Sleep(time.Millisecond * 500)
	}
EMOTICONLOOP:
	for i := len(emoticonIDList) - 1; i >= 0; i-- {
		mu.RLock()
		y := gdb.getYmgalByID(emoticonIDList[i])
		mu.RUnlock()
		if y.PictureList == "" {
			mu.Lock()
			storeEmoticonPic(emoticonIDList[i])
			mu.Unlock()
		} else {
			break EMOTICONLOOP
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func storeCgPic(picIDStr string) {
	picID, err := strconv.ParseInt(picIDStr, 10, 64)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	pictureType := cgType
	doc, err := htmlquery.LoadURL(webPicURL + picIDStr)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	title := htmlquery.FindOne(doc, "//meta[@name='name']").Attr[1].Val
	pictureDescription := htmlquery.FindOne(doc, "//meta[@name='description']").Attr[1].Val
	pictureNumberStr := htmlquery.FindOne(doc, "//div[@class='meta-info']/div[@class='meta-right']/span[2]/text()").Data
	re := regexp.MustCompile(reNumber)
	pictureNumber, err := strconv.Atoi(re.FindString(pictureNumberStr))
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	pictureList := ""
	for i := 1; i <= pictureNumber; i++ {
		picURL := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='main-picset-warp']/div/div[2]/div/div[@class='swiper-wrapper']/div[%d]", i)).Attr[1].Val
		if i == 1 {
			pictureList += picURL
		} else {
			pictureList += "," + picURL
		}
	}
	err = gdb.insertOrUpdateYmgalByID(picID, title, pictureType, pictureDescription, pictureList)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}

}

func storeEmoticonPic(picIDStr string) {
	picID, err := strconv.ParseInt(picIDStr, 10, 64)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	pictureType := emoticonType
	doc, err := htmlquery.LoadURL(webPicURL + picIDStr)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	title := htmlquery.FindOne(doc, "//meta[@name='name']").Attr[1].Val
	pictureDescription := htmlquery.FindOne(doc, "//meta[@name='description']").Attr[1].Val
	pictureNumberStr := htmlquery.FindOne(doc, "//div[@class='meta-info']/div[@class='meta-right']/span[2]/text()").Data
	re := regexp.MustCompile(reNumber)
	pictureNumber, err := strconv.Atoi(re.FindString(pictureNumberStr))
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	pictureList := ""
	for i := 1; i <= pictureNumber; i++ {
		picURL := htmlquery.FindOne(doc, fmt.Sprintf("//*[@id='main-picset-warp']/div/div[@class='stream-list']/div[%d]/img", i)).Attr[1].Val
		if i == 1 {
			pictureList += picURL
		} else {
			pictureList += "," + picURL
		}
	}
	err = gdb.insertOrUpdateYmgalByID(picID, title, pictureType, pictureDescription, pictureList)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
}
