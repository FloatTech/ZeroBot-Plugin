package main

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/logoove/sqlite"
)

type BrDB = gorm.DB

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
