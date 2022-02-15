package cpstory

import sql "github.com/FloatTech/sqlite"

type cpstory struct {
	ID    int64  `db:"id"`
	Gong  string `db:"gong"`
	Shou  string `db:"shou"`
	Story string `db:"story"`
}

var db = &sql.Sqlite{}

func getRandomCpStory() (cs cpstory) {
	_ = db.Pick("cp_story", &cs)
	return
}
