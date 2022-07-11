package dice

type strjrrp struct {
	GrpID   int64  `db:"gid"`
	Strjrrp string `db:"strjrrp"`
}

type rsl struct {
	GrpID int64 `db:"gid"`
	Rule  int64 `db:"rule"`
}

type set struct {
	UserID int64 `db:"uid"`
	D      int64 `db:"d"`
}
