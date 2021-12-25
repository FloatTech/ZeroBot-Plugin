package timer

import (
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

// Timer 计时器
type Timer struct {
	Id                          uint32 `db:"id"`
	En1Month4Day5Week3Hour5Min6 int32  `db:"emdwhm"`
	Selfid                      int64  `db:"sid"`
	GrpId                       int64  `db:"gid"`
	Alert                       string `db:"alert"`
	Cron                        string `db:"cron"`
	Url                         string `db:"url"`
}

func (t *Timer) InsertInto(db *sql.Sqlite) error {
	return db.Insert("timer", t)
}

/*
func getTimerFrom(db *sql.Sqlite, id uint32) (t Timer, err error) {
	err = db.Find("timer", &t, "where id = "+strconv.Itoa(int(id)))
	return
}
*/
