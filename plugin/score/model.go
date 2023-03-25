package score

import (
	"os"
	"time"

	"github.com/jinzhu/gorm"
)

// sdb 得分数据库
var sdb *scoredb

// scoredb 分数数据库
type scoredb gorm.DB

// scoretable 分数结构体
type scoretable struct {
	UID   int64 `gorm:"column:uid;primary_key"`
	Score int   `gorm:"column:score;default:0"`
}

// TableName ...
func (scoretable) TableName() string {
	return "score"
}

// signintable 签到结构体
type signintable struct {
	UID       int64 `gorm:"column:uid;primary_key"`
	Count     int   `gorm:"column:count;default:0"`
	UpdatedAt time.Time
}

// TableName ...
func (signintable) TableName() string {
	return "sign_in"
}

// initialize 初始化ScoreDB数据库
func initialize(dbpath string) *scoredb {
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
	gdb.AutoMigrate(&scoretable{}).AutoMigrate(&signintable{})
	return (*scoredb)(gdb)
}

// Close ...
func (sdb *scoredb) Close() error {
	db := (*gorm.DB)(sdb)
	return db.Close()
}

// GetScoreByUID 取得分数
func (sdb *scoredb) GetScoreByUID(uid int64) (s scoretable) {
	db := (*gorm.DB)(sdb)
	db.Model(&scoretable{}).FirstOrCreate(&s, "uid = ? ", uid)
	return s
}

// InsertOrUpdateScoreByUID 插入或更新分数
func (sdb *scoredb) InsertOrUpdateScoreByUID(uid int64, score int) (err error) {
	db := (*gorm.DB)(sdb)
	s := scoretable{
		UID:   uid,
		Score: score,
	}
	if err = db.Model(&scoretable{}).First(&s, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			err = db.Model(&scoretable{}).Create(&s).Error // newUser not user
		}
	} else {
		err = db.Model(&scoretable{}).Where("uid = ? ", uid).Update(
			map[string]any{
				"score": score,
			}).Error
	}
	return
}

// GetSignInByUID 取得签到次数
func (sdb *scoredb) GetSignInByUID(uid int64) (si signintable) {
	db := (*gorm.DB)(sdb)
	db.Model(&signintable{}).FirstOrCreate(&si, "uid = ? ", uid)
	return si
}

// InsertOrUpdateSignInCountByUID 插入或更新签到次数
func (sdb *scoredb) InsertOrUpdateSignInCountByUID(uid int64, count int) (err error) {
	db := (*gorm.DB)(sdb)
	si := signintable{
		UID:   uid,
		Count: count,
	}
	if err = db.Model(&signintable{}).First(&si, "uid = ? ", uid).Error; err != nil {
		// error handling...
		if gorm.IsRecordNotFoundError(err) {
			err = db.Model(&signintable{}).Create(&si).Error // newUser not user
		}
	} else {
		err = db.Model(&signintable{}).Where("uid = ? ", uid).Update(
			map[string]any{
				"count": count,
			}).Error
	}
	return
}

func (sdb *scoredb) GetScoreRankByTopN(n int) (st []scoretable, err error) {
	db := (*gorm.DB)(sdb)
	err = db.Model(&scoretable{}).Order("score desc").Limit(n).Find(&st).Error
	return
}

type scdata struct {
	drawedfile string
	picfile    string
	uid        int64
	nickname   string
	inc        int // 增加币
	score      int // 钱包
	level      int
	rank       int
}
