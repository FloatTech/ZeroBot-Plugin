package managerplugin

import "time"

type blacklist struct {
	GrpID  int64 `db:"gid"`
	UserID int64 `db:"uid"`
}

type groupban struct {
	GrpID  int64     `db:"gid"`
	UserID int64     `db:"uid"`
	Time   time.Time `db:"time"`
}
