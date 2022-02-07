package jandan

import (
	_ "github.com/logoove/sqlite" // use sql
)

type picture struct {
	ID         uint64 `db:"id"`
	PictureURL string `db:"picture_url"`
}

func getRandomPicture() (p picture, err error) {
	err = db.Pick("picture", &p)
	return
}
