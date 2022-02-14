package driftbottle

import (
	"os"

	sql "github.com/FloatTech/sqlite"
)

const (
	dbpath = "data/driftbottle/"
	dbfile = dbpath + "sea.db"
)

var sea = &sql.Sqlite{DBPath: dbfile}

func init() {
	_ = os.MkdirAll(dbpath, 0755)
	err := sea.Open()
	if err != nil {
		panic(err)
	}
	_ = createChannel(sea, "global")
}
