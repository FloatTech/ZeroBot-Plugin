package ymgal

import (
	"encoding/json"
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
	"time"
)

// gdb 得分数据库
var gdb *ymgaldb

// ymgaldb galgame图片数据库
type ymgaldb gorm.DB

// Ymgal gal图片储存结构体
type Ymgal struct {
	ID                 int64  `gorm:"column:id" json:"id,omitempty"`
	Title              string `gorm:"column:title" json:"title"`
	PictureType        string `gorm:"column:picture_type" json:"picture_type"`
	PictureDescription string `gorm:"column:picture_description;type:varchar(1024)" json:"picture_description"`
	PictureList        string `gorm:"column:picture_list;type:varchar(20000)" json:"picture_list"`
}

// TableName ...
func (Ymgal) TableName() string {
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
	gdb.AutoMigrate(&Ymgal{})
	return (*ymgaldb)(gdb)
}

func (gdb *ymgaldb) insertOrUpdateYmgalByID(id int64, ymgalMap map[string]interface{}) (err error) {
	db := (*gorm.DB)(gdb)
	y := Ymgal{}
	ymgalMapJson, err := json.Marshal(ymgalMap)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	err = json.Unmarshal(ymgalMapJson, &y)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	y.ID = id
	if err = db.Debug().Model(&Ymgal{}).First(&y, "id = ? ", id).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Debug().Model(&Ymgal{}).Create(&y).Error // newUser not user
		}
	} else {
		err = db.Debug().Model(&Ymgal{}).Where("id = ? ", id).Update(ymgalMap).Error
	}
	return
}

func (gdb *ymgaldb) randomYmgal(pictureType string) (y Ymgal) {
	db := (*gorm.DB)(gdb)
	var count int
	db.Debug().Model(&Ymgal{}).Where("picture_type = ?", pictureType).Count(&count).Offset(rand.Intn(count)).Take(&y)
	return
}

func (gdb *ymgaldb) getYmgalByKey(pictureType, key string) (y Ymgal) {
	db := (*gorm.DB)(gdb)
	var count int
	db.Debug().Model(&Ymgal{}).Where("picture_type = ? and (picture_description like ? or title like ?) ", pictureType, "%"+key+"%", "%"+key+"%").Count(&count).Offset(rand.Intn(count)).Take(&y)
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
	maxCgPageNumber       int
	maxEmoticonPageNumber int
	cgURL                 = webURL + "/search?type=picset&sort=default&category=" + url.QueryEscape(cgType) + "&page="
	emoticonURL           = webURL + "/search?type=picset&sort=default&category=" + url.QueryEscape(emoticonType) + "&page="
	commonPageNumberExpr  = "//*[@id='pager-box']/div/a[@class='icon item pager-next']/preceding-sibling::a[1]/text()"
	cgIDList              []string
	emoticonIDList        []string
)

func initPageNumber() {
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
}

func getPicId(pageNumber int, pictureType string) {
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
	initPageNumber()
	for i := 1; i <= maxCgPageNumber; i++ {
		getPicId(i, cgType)
		time.Sleep(time.Millisecond * 500)
	}
	for i := 1; i <= maxEmoticonPageNumber; i++ {
		getPicId(i, emoticonType)
		time.Sleep(time.Millisecond * 500)
	}
	for _, v := range cgIDList {
		storeCgPic(v)
		time.Sleep(time.Millisecond * 500)
	}
	for _, v := range emoticonIDList {
		storeEmoticonPic(v)
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
	y := Ymgal{
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	ymgalJson, err := json.Marshal(&y)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	var ymgalMap map[string]interface{}
	err = json.Unmarshal(ymgalJson, &ymgalMap)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	err = gdb.insertOrUpdateYmgalByID(picID, ymgalMap)
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
	y := Ymgal{
		Title:              title,
		PictureType:        pictureType,
		PictureDescription: pictureDescription,
		PictureList:        pictureList,
	}
	ymgalJson, err := json.Marshal(&y)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	var ymgalMap map[string]interface{}
	err = json.Unmarshal(ymgalJson, &ymgalMap)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
	err = gdb.insertOrUpdateYmgalByID(picID, ymgalMap)
	if err != nil {
		log.Errorln("[ymgal]:", err)
	}
}
