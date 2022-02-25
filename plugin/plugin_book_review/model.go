package bookreview

import sql "github.com/FloatTech/sqlite"

type book struct {
	ID         uint64 `db:"id"`
	BookReview string `db:"bookreview"`
}

var db = &sql.Sqlite{}

// 暂时随机选择一个书评
func getBookReviewByKeyword(keyword string) (b book) {
	_ = db.Find("bookreview", &b, "where bookreview LIKE '%"+keyword+"%'")
	return
}

func getRandomBookReview() (b book) {
	_ = db.Pick("bookreview", &b)
	return
}
