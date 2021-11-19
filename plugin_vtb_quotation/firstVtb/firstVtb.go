package firstVtb

import (
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/utils"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var vtbUrl = "https://vtbkeyboard.moe/api/get_vtb_list"

func GetVtbListStr() string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", vtbUrl, nil)
	if err != nil {
		log.Println(err)
	}
	// 自定义Header
	req.Header.Set("User-Agent", utils.GetAgent())
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	//log.Println(string(bytes))
	vtbListStr, err := strconv.Unquote(strings.Replace(strconv.Quote(string(bytes)), `\\u`, `\u`, -1))
	if err != nil {
		log.Println(err)
	}
	log.Println(vtbListStr)
	return vtbListStr
}
func DealVtbListStr(vtbListStr string) []string {
	uidList := make([]string, 0)
	count := gjson.Get(vtbListStr, "#").Int()
	for i := int64(0); i < count; i++ {
		item := gjson.Get(vtbListStr, strconv.FormatInt(i, 10))
		log.Println(item)
		fc := model.FirstCategory{
			FirstCategoryIndex:       i,
			FirstCategoryName:        item.Get("name").String(),
			FirstCategoryDescription: item.Get("description").String(),
			FirstCategoryIconPath:    item.Get("icon_path").String(),
			FirstCategoryUid:         item.Get("uid").String(),
		}
		log.Println(fc)
		//model.Db.Model(FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).FirstOrCreate(&fc)
		if err := model.Db.Debug().Model(&model.FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).First(&fc).Error; err != nil {
			// error handling...
			if gorm.IsRecordNotFoundError(err) {
				model.Db.Debug().Model(&model.FirstCategory{}).Create(&fc) // newUser not user
			}
		} else {
			model.Db.Debug().Model(&model.FirstCategory{}).Where("first_category_uid = ?", fc.FirstCategoryUid).Update(
				map[string]interface{}{
					"first_category_index":       i,
					"first_category_name":        item.Get("name").String(),
					"first_category_description": item.Get("description").String(),
					"first_category_icon_path":   item.Get("icon_path").String(),
				})
		}
		uidList = append(uidList, fc.FirstCategoryUid)

	}

	log.Println(uidList)
	return uidList
}
