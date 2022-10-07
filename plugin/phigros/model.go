package phigros

type data struct {
	UID      int64  `db:"uid"`
	Gamename string `db:"gamename"`
}

type challen struct {
	UID      int64  `db:"uid"`
	Chall    string `db:"chall"`    // rainbow
	Challnum int64  `db:"challnum"` // 49
}

type result struct {
	ID       int64   `db:"id"`
	Songname string  `db:"songname"` // eg. Shadow
	Diff     string  `db:"diff"`     // AT
	Diffnum  float64 `db:"diffnum"`  // 15.9
	Score    int64   `db:"score"`    // 1000000
	Acc      float64 `db:"acc"`      // 100.00
	Rank     string  `db:"rank"`     // phi
	Rksm     float64 `db:"rksm"`     // 15.90
}

type songdata struct {
	Name   string  `db:"Name"`
	EZ     float64 `db:"Easy"`
	HD     float64 `db:"Hard"`
	IN     float64 `db:"Insane"`
	AT     float64 `db:"Another"`
	Legacy float64 `db:"legacy"`
	ATName string  `db:"ATName"`
}

type max struct {
	ID       int64   `db:"id"`
	Songname string  `db:"songname"` // eg. Shadow
	Diff     string  `db:"diff"`     // AT
	Diffnum  float64 `db:"diffnum"`  // 15.9
	Score    int64   `db:"score"`    // 1000000
	Acc      float64 `db:"acc"`      // 100.00
	Rank     string  `db:"rank"`     // phi
	Rksm     float64 `db:"rksm"`     // 15.90
	Max      float64
}
