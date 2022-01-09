package score

import (
	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite" // import sql
	"os"
	"time"
)

// ScoreDB 分数数据库
type ScoreDB gorm.DB

type Score struct {
	UID   int64 `gorm:"column:uid;primary_key"`
	Score int   `gorm:"column:score;default:0"`
}

// TableName ...
func (Score) TableName() string {
	return "score"
}

type SignIn struct {
	UID       int64 `gorm:"column:uid;primary_key"`
	Count     int   `gorm:"column:count"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName ...
func (SignIn) TableName() string {
	return "sign_in"
}

func Initialize(dbpath string) *ScoreDB {
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
	gdb.AutoMigrate(&Score{}).AutoMigrate(&SignIn{})
	return (*ScoreDB)(gdb)
}

// Open ...
func Open(dbpath string) (*ScoreDB, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	return (*ScoreDB)(db), nil
}

// Close ...
func (sdb *ScoreDB) Close() error {
	db := (*gorm.DB)(sdb)
	return db.Close()
}

// GetScoreByUID 取得分数
func (sdb *ScoreDB) GetScoreByUID(uid int64) (s Score) {
	db := (*gorm.DB)(sdb)
	db.Debug().Model(&Score{}).FirstOrCreate(&s, "uid = ? ", uid)
	return s
}

// InsertOrUpdateScoreByUID 插入或更新分数
func (sdb *ScoreDB) InsertOrUpdateScoreByUID(uid int64, score int) (err error) {
	db := (*gorm.DB)(sdb)
	s := Score{
		UID:   uid,
		Score: score,
	}
	if err = db.Debug().Model(&Score{}).First(&s, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&Score{}).Create(&s) // newUser not user
		}
	} else {
		err = db.Debug().Model(&Score{}).Where("uid = ? ", uid).Update(
			map[string]interface{}{
				"score": score,
			}).Error
	}
	return
}

// GetSignInByUID 取得签到次数
func (sdb *ScoreDB) GetSignInByUID(uid int64) (si SignIn) {
	db := (*gorm.DB)(sdb)
	db.Debug().Model(&SignIn{}).FirstOrCreate(&si, "uid = ? ", uid)
	return si
}

// InsertOrUpdateSignInCountByUID 插入或更新签到次数
func (sdb *ScoreDB) InsertOrUpdateSignInCountByUID(uid int64, count int) (err error) {
	db := (*gorm.DB)(sdb)
	si := SignIn{
		UID:   uid,
		Count: count,
	}
	if err = db.Debug().Model(&SignIn{}).First(&si, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			db.Debug().Model(&SignIn{}).Create(&si) // newUser not user
		}
	} else {
		err = db.Debug().Model(&SignIn{}).Where("uid = ? ", uid).Update(
			map[string]interface{}{
				"count": count,
			}).Error
	}
	return
}
