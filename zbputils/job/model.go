package job

import sql "github.com/FloatTech/sqlite"

type cmd struct {
	ID   int64  `db:"id"`
	Cron string `db:"cron"`
	Cmd  string `db:"cmd"`
}

var db = &sql.Sqlite{}
