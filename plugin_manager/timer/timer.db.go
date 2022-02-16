package timer

import (
	sql "github.com/FloatTech/sqlite"
)

// Timer 计时器
type Timer struct {
	ID                          uint32 `db:"id"`
	En1Month4Day5Week3Hour5Min6 int32  `db:"emdwhm"`
	SelfID                      int64  `db:"sid"`
	GrpID                       int64  `db:"gid"`
	Alert                       string `db:"alert"`
	Cron                        string `db:"cron"`
	URL                         string `db:"url"`
}

// InsertInto 插入自身
func (t *Timer) InsertInto(db *sql.Sqlite) error {
	return db.Insert("timer", t)
}

/*
func getTimerFrom(db *sql.Sqlite, id uint32) (t Timer, err error) {
	err = db.Find("timer", &t, "where id = "+strconv.Itoa(int(id)))
	return
}
*/
