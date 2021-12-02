package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
	"os"
	"time"
)

type SleepDB gorm.DB

func Initialize(dbpath string) *SleepDB {
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
	gdb.AutoMigrate(&SleepTime{})
	return (*SleepDB)(gdb)
}

type SleepTime struct {
	gorm.Model
	GroupId   int64     `gorm:"column:group_id"`
	UserId    int64     `gorm:"column:user_id"`
	SleepTime time.Time `gorm:"column:sleep_time"`
	GetUpTime time.Time `gorm:"column:get_up_time"`
}

func (SleepTime) TableName() string {
	return "sleep_time"
}

//更新睡眠时间
func (sdb *SleepDB) Sleep(groupId, userId int64) (position int) {
	db := (*gorm.DB)(sdb)
	sleepTime := time.Now()
	st := SleepTime{
		GroupId:   groupId,
		UserId:    userId,
		SleepTime: sleepTime,
	}
	if err := db.Debug().Model(&SleepTime{}).Where("group_id = ? and user_id = ?", groupId, userId).First(&st).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&SleepTime{}).Create(&st) // newUser not user
		}
	} else {
		db.Debug().Model(&SleepTime{}).Where("group_id = ? and user_id = ?", groupId, userId).Update(
			map[string]interface{}{
				"sleep_time": sleepTime,
			})
	}
	db.Debug().Model(&SleepTime{}).Where("group_id = ? and sleep_time <= ?", groupId, sleepTime).Count(&position)
	return position
}

//更新起床时间
func (sdb *SleepDB) GetUp(groupId, userId int64) {
	db := (*gorm.DB)(sdb)
	getUpTime := time.Now()
	st := SleepTime{
		GroupId:   groupId,
		UserId:    userId,
		GetUpTime: getUpTime,
	}
	if err := db.Debug().Model(&SleepTime{}).Where("group_id = ? and user_id = ?", groupId, userId).First(&st).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&SleepTime{}).Create(&st) // newUser not user
		}
	} else {
		db.Debug().Model(&SleepTime{}).Where("group_id = ? and user_id = ?", groupId, userId).Update(
			map[string]interface{}{
				"get_up_time": getUpTime,
			})
	}
}
