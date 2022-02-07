package jandan

import (
	_ "github.com/logoove/sqlite" // use sql
)

type picture struct {
	ID         uint64 `gorm:"column:id;primary_key"`
	PictureURL string `gorm:"column:picture_url"`
}

func getRandomPicture() (p picture, err error) {
	err = db.Pick("picture", &p)
	return
}
