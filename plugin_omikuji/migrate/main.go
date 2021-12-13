package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

type signature struct {
	Id   uint64 `db:"id"`
	Text string `db:"text"`
}

type kuji struct {
	Id   uint8  `db:"id"`
	Text string `db:"text"`
}

func main() {
	db := &sql.Sqlite{DBPath: os.Args[1]}
	newdb := &sql.Sqlite{DBPath: os.Args[2]}
	err := newdb.Create("kuji", &kuji{})
	if err != nil {
		panic(err)
	}
	err = db.Create("signature", &signature{})
	if err != nil {
		panic(err)
	}

	fmt.Println(db.Count("signature"))
	s := &signature{}
	k := &kuji{}
	for i := 1; i <= 100; i++ {
		db.Find("signature", s, "where id = "+strconv.Itoa(i))
		fmt.Println("insert: ", s.Text[:57])
		k.Id = uint8(i)
		k.Text = s.Text
		newdb.Insert("kuji", k)
	}

	db.Close()
	newdb.Close()
}
