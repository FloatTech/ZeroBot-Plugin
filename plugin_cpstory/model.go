package cpstory

type CpStory struct {
	Id    int64  `db:"id"`
	Gong  string `db:"gong"`
	Shou  string `db:"shou"`
	Story string `db:"story"`
}

func getRandomCpStory() (cs CpStory) {
	_ = db.Pick("cp_story", &cs)
	return
}
