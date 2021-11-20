package secondVtb

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/model"
	"github.com/FloatTech/ZeroBot-Plugin/plugin_vtb_quotation/utils"
)

func GetVtbStr(uid string) string {
	vtbUrl := "https://vtbkeyboard.moe/api/get_vtb_page?uid=" + uid
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

func DealVtbStr(vtbStr, uid string) {
	secondCount := gjson.Get(vtbStr, "data.voices.#").Int()
	log.Println("二级品类一共有", secondCount)
	for secondIndex := int64(0); secondIndex < secondCount; secondIndex++ {
		secondItem := gjson.Get(vtbStr, "data.voices."+strconv.FormatInt(secondIndex, 10))
		log.Println(secondItem)
		sc := model.SecondCategory{
			SecondCategoryName:        secondItem.Get("categoryName").String(),
			SecondCategoryIndex:       secondIndex,
			SecondCategoryAuthor:      secondItem.Get("author").String(),
			SecondCategoryDescription: secondItem.Get("categoryDescription.zh-CN").String(),
			FirstCategoryUid:          uid,
		}
		log.Println(sc)
		//model.Db.Model(SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).FirstOrCreate(&sc)
		if err := model.Db.Debug().Model(&model.SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).First(&sc).Error; err != nil {
			// error handling...
			if gorm.IsRecordNotFoundError(err) {
				model.Db.Debug().Model(&model.SecondCategory{}).Create(&sc) // newUser not user
			}
		} else {
			model.Db.Debug().Model(&model.SecondCategory{}).Where("first_category_uid = ? and second_category_index = ?", uid, secondIndex).Update(
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
			tc := model.ThirdCategory{
				ThirdCategoryName:        thirdItem.Get("name").String(),
				ThirdCategoryIndex:       thirdIndex,
				ThirdCategoryDescription: thirdItem.Get("description.zh-CN").String(),
				FirstCategoryUid:         uid,
				SecondCategoryIndex:      secondIndex,
				ThirdCategoryPath:        thirdItem.Get("path").String(),
				ThirdCategoryAuthor:      thirdItem.Get("author").String(),
			}
			log.Println(tc)
			//model.Db.Model(ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
			//	uid, secondIndex, thirdIndex).FirstOrCreate(&tc)
			if err := model.Db.Debug().Model(&model.ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
				uid, secondIndex, thirdIndex).First(&tc).Error; err != nil {
				// error handling...
				if gorm.IsRecordNotFoundError(err) {
					model.Db.Debug().Model(&model.ThirdCategory{}).Create(&tc) // newUser not user
				}
			} else {
				model.Db.Debug().Model(&model.ThirdCategory{}).Where("first_category_uid = ? and second_category_index = ? and third_category_index = ?",
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
