package omikuji

import "strconv"

type signature struct {
	Id   uint64 `db:"id"`
	Text string `db:"text"`
}

// 返回一个解签
func getSignatureById(id int) (s signature) {
	db.Find("signature", &s, "where id = "+strconv.Itoa(id))
	return
}
