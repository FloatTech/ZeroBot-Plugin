package cpstory

type cpstory struct {
	ID    int64  `db:"id"`
	Gong  string `db:"gong"`
	Shou  string `db:"shou"`
	Story string `db:"story"`
}

func getRandomCpStory() (cs cpstory) {
	_ = db.Pick("cp_story", &cs)
	return
}
