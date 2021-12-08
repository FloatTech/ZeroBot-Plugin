package plugin_book_review

type book struct {
	Id         uint64 `db:"id"`
	BookReview string `db:"bookreview"`
}

// 暂时随机选择一个书评
func getBookReviewByKeyword(keyword string) (b book) {
	db.Find("bookreview", &b, "where bookreview LIKE '%"+keyword+"%'")
	return
}

func getRandomBookReview() (b book) {
	db.Pick("bookreview", &b)
	return
}
