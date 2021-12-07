package model

import (
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
)

type BrDB gorm.DB

func Initialize(dbpath string) *BrDB {
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
	gdb.AutoMigrate(&BookReview{})
	return (*BrDB)(gdb)
}

func Open(dbpath string) (*BrDB, error) {
	db, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	} else {
		return (*BrDB)(db), nil
	}
}

type BookReview struct {
	gorm.Model
	BookReview string `gorm:"column:book_review"`
}

func (BookReview) TableName() string {
	return "book_review"
}

// 暂时随机选择一个书评
func (brdb *BrDB) GetBookReviewByKeyword(keyword string) (BookReviewList BookReview) {
	db := (*gorm.DB)(brdb)
	rand.Seed(time.Now().UnixNano())
	var count int
	db.Debug().Model(&BookReview{}).Where("book_review LIKE ?", "%"+keyword+"%").Count(&count).Offset(rand.Intn(count)).Take(&BookReviewList)
	log.Println(BookReviewList)
	return BookReviewList
}

func (brdb *BrDB) GetRandomBookReview() (bookReview BookReview) {
	db := (*gorm.DB)(brdb)
	rand.Seed(time.Now().UnixNano())
	var count int
	db.Debug().Model(&BookReview{}).Count(&count).Offset(rand.Intn(count)).Take(&bookReview)
	return bookReview
}

func (brdb *BrDB) Close() error {
	db := (*gorm.DB)(brdb)
	return db.Close()
}
