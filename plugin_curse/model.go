package curse

type curse struct {
	ID    uint32 `db:"id"`
	Text  string `db:"text"`
	Level string `db:"level"`
}

func getRandomCurseByLevel(level string) (c curse) {
	_ = db.Find("curse", &c, "where level = '"+level+"' ORDER BY RANDOM() limit 1")
	return
}
