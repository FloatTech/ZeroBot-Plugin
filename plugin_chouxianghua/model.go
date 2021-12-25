package chouxianghua

type Pinyin struct {
	Word          string `db:"word"`
	Pronunciation string `db:"pronunciation"`
}
type Emoji struct {
	Pronunciation string `db:"pronunciation"`
	Emoji         string `db:"emoji"`
}

func getPronunciationByWord(word string) (p Pinyin) {
	db.Find("pinyin", &p, "where word = '"+word+"'")
	return
}

func getEmojiByPronunciation(pronunciation string) (e Emoji) {
	db.Find("emoji", &e, "where pronunciation = '"+pronunciation+"'")
	return
}
