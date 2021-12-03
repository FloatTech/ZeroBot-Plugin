package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
	log "github.com/sirupsen/logrus"
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
	gdb.AutoMigrate(&SleepManage{})
	return (*SleepDB)(gdb)
}

func Open(dbpath string) (*SleepDB, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	} else {
		return (*SleepDB)(db), nil
	}
}

func (sdb *SleepDB) Close() error {
	db := (*gorm.DB)(sdb)
	return db.Close()
}

type SleepManage struct {
	gorm.Model
	GroupId   int64     `gorm:"column:group_id"`
	UserId    int64     `gorm:"column:user_id"`
	SleepTime time.Time `gorm:"column:sleep_time"`
}

func (SleepManage) TableName() string {
	return "sleep_manage"
}

// 更新睡眠时间
func (sdb *SleepDB) Sleep(groupId, userId int64) (position int, awakeTime time.Duration) {
	db := (*gorm.DB)(sdb)
	now := time.Now()

	today := now.Add(-time.Hour*time.Duration(3+now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second()))
	st := SleepManage{
		GroupId:   groupId,
		UserId:    userId,
		SleepTime: now,
	}
	if err := db.Debug().Model(&SleepManage{}).Where("group_id = ? and user_id = ?", groupId, userId).First(&st).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&SleepManage{}).Create(&st) // newUser not user
		}
	} else {
		log.Println("sleeptime为", st)
		awakeTime = now.Sub(st.SleepTime)
		db.Debug().Model(&SleepManage{}).Where("group_id = ? and user_id = ?", groupId, userId).Update(
			map[string]interface{}{
				"sleep_time": now,
			})
	}
	db.Debug().Model(&SleepManage{}).Where("group_id = ? and sleep_time <= ? and sleep_time >= ?", groupId, now, today).Count(&position)
	return position, awakeTime
}

// 更新起床时间
func (sdb *SleepDB) GetUp(groupId, userId int64) (position int, sleepTime time.Duration) {
	db := (*gorm.DB)(sdb)
	now := time.Now()
	today := now.Add(-time.Hour*time.Duration(-6+now.Hour()) - time.Minute*time.Duration(now.Minute()) - time.Second*time.Duration(now.Second()))
	st := SleepManage{
		GroupId:   groupId,
		UserId:    userId,
		SleepTime: now,
	}
	if err := db.Debug().Model(&SleepManage{}).Where("group_id = ? and user_id = ?", groupId, userId).First(&st).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&SleepManage{}).Create(&st) // newUser not user
		}
	} else {
		log.Println("sleeptime为", st)
		sleepTime = now.Sub(st.SleepTime)
		db.Debug().Model(&SleepManage{}).Where("group_id = ? and user_id = ?", groupId, userId).Update(
			map[string]interface{}{
				"get_up_time": now,
			})
	}
	db.Debug().Model(&SleepManage{}).Where("group_id = ? and sleep_time <= ? and sleep_time >= ?", groupId, now, today).Count(&position)
	return position, sleepTime
}
