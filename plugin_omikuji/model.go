package omikuji

import "strconv"

type kuji struct {
	ID   uint8  `db:"id"`
	Text string `db:"text"`
}

// 返回一个解签
func getKujiByBango(id uint8) string {
	var s kuji
	err := db.Find("kuji", &s, "where id = "+strconv.Itoa(int(id)))
	if err != nil {
		return err.Error()
	}
	return s.Text
}
