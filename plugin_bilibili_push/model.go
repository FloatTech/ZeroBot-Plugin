package bilibilipush

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite" // import sql
	"os"
	"reflect"
)

// bilibilipushdb bili推送数据库
type bilibilipushdb gorm.DB

type bilibilipush struct {
	ID             int64 `gorm:"column:id;primary_key" json:"id"`
	BilibiliUID    int64 `gorm:"column:bilibili_uid;index:idx_buid_gid" json:"bilibili_uid"`
	GroupID        int64 `gorm:"column:group_id;index:idx_buid_gid" json:"group_id"`
	LiveDisable    int64 `gorm:"column:live_disable;default:0" json:"live_disable"`
	DynamicDisable int64 `gorm:"column:dynamic_disable;default:0" json:"dynamic_disable"`
}

// TableName ...
func (bilibilipush) TableName() string {
	return "bilibili_push"
}

type bilibiliup struct {
	BilibiliUID int64  `gorm:"column:bilibili_uid;primary_key"`
	Name        string `gorm:"column:name"`
}

// TableName ...
func (bilibiliup) TableName() string {
	return "bilibili_up"
}

// Initialize 初始化ScoreDB数据库
func Initialize(dbpath string) *bilibilipushdb {
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
	gdb.AutoMigrate(&bilibilipush{}).AutoMigrate(&bilibiliup{})
	return (*bilibilipushdb)(gdb)
}

// Open ...
func Open(dbpath string) (*bilibilipushdb, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	return (*bilibilipushdb)(db), nil
}

// Close ...
func (bdb *bilibilipushdb) Close() error {
	db := (*gorm.DB)(bdb)
	return db.Close()
}

// insertOrUpdateLiveAndDynamic 插入或更新数据库
func (bdb *bilibilipushdb) insertOrUpdateLiveAndDynamic(bpMap map[string]interface{}) (err error) {
	db := (*gorm.DB)(bdb)
	bp := bilibilipush{}
	data, _ := json.Marshal(&bpMap)
	_ = json.Unmarshal(data, &bp)
	fmt.Println(bp, reflect.TypeOf(bp))
	if err = db.Debug().Model(&bilibilipush{}).First(&bp, "bilibili_uid = ? and group_id = ?", bp.BilibiliUID, bp.GroupID).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = db.Debug().Model(&bilibilipush{}).Create(&bp).Error
		}
	} else {
		err = db.Debug().Model(&bilibilipush{}).Where("bilibili_uid = ? and group_id = ?", bp.BilibiliUID, bp.GroupID).Update(bpMap).Error
	}
	return
}

// getAllGroupByBuid 取得订阅了某up的所有群聊
func (bdb *bilibilipushdb) getAllGroupByBuid(buid int64) (bpl []bilibilipush) {
	db := (*gorm.DB)(bdb)
	db.Debug().Model(&bilibilipush{}).Find(&bpl, "bilibili_uid = ? and (live_disable = 0 or dynamic_disable = 0)", buid)
	return
}

func (bdb *bilibilipushdb) getAllGroupByBuidAndLive(buid int64) (buidList []int64) {
	db := (*gorm.DB)(bdb)
	var bpl []bilibilipush
	db.Debug().Model(&bilibilipush{}).Find(&bpl, "bilibili_uid = ? and live_disable = 0", buid)
	for _, v := range bpl {
		buidList = append(buidList, v.BilibiliUID)
	}
	return
}

func (bdb *bilibilipushdb) getAllGroupByBuidAndDynamic(buid int64) (buidList []int64) {
	db := (*gorm.DB)(bdb)
	var bpl []bilibilipush
	db.Debug().Model(&bilibilipush{}).Find(&bpl, "bilibili_uid = ? and dynamic_disable = 0", buid)
	for _, v := range bpl {
		buidList = append(buidList, v.BilibiliUID)
	}
	return
}

func (bdb *bilibilipushdb) insertBilibiliUp(buid int64, name string) {
	db := (*gorm.DB)(bdb)
	bu := bilibiliup{
		BilibiliUID: buid,
		Name:        name,
	}
	db.Debug().Model(&bilibiliup{}).Create(bu)
}

func (bdb *bilibilipushdb) getBilibiliUpName(buid int64) string {
	db := (*gorm.DB)(bdb)
	bu := bilibiliup{}
	db.Debug().Model(&bilibiliup{}).First(&bu, "bilibili_uid = ?", buid)
	return bu.Name
}

func (bdb *bilibilipushdb) getAllBuid() (buidList []int64) {
	db := (*gorm.DB)(bdb)
	var bul []bilibiliup
	db.Debug().Model(&bilibiliup{}).Find(&bul)
	for _, v := range bul {
		buidList = append(buidList, v.BilibiliUID)
	}
	return
}
