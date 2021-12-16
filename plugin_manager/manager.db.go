package manager

type Welcome struct {
	GrpID int64  `db:"gid"`
	Msg   string `db:"msg"`
}

type Member struct {
	QQ int64 `db:"qq"`
	// github username
	Ghun string `db:"ghun"`
}
