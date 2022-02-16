package jandan

import (
	"sync"

	sql "github.com/FloatTech/sqlite"
)

var db = &sql.Sqlite{}
var mu sync.RWMutex

type picture struct {
	ID  uint64 `db:"id"`
	URL string `db:"url"`
}

func getRandomPicture() (u string, err error) {
	var p picture
	mu.RLock()
	err = db.Pick("picture", &p)
	mu.RUnlock()
	u = p.URL
	return
}
