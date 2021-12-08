package main

import (
	"os"

	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

type book struct {
	Id         uint64 `db:"id"`
	BookReview string `db:"bookreview"`
}

func main() {
	db, err := Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	newdb := &sql.Sqlite{DBPath: os.Args[2]}
	err = newdb.Create("bookreview", &book{})
	if err != nil {
		panic(err)
	}
	rs, err := db.Table("book_review").Select("book_review", "").Rows()
	if err != nil {
		panic(err)
	}
	var d string
	var i uint64
	for rs.Next() {
		err := rs.Scan(&d)
		if err != nil {
			panic(err)
		}
		i++
		err = newdb.Insert("bookreview", &book{i, d})
		if err != nil {
			panic(err)
		}
	}
	db.Close()
	newdb.Close()
}
