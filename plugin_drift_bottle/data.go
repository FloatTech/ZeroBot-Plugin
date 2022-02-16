package driftbottle

import (
	"fmt"
	"hash/crc64"
	"strconv"
	"sync"

	sql "github.com/FloatTech/sqlite"
	"github.com/FloatTech/zbputils/binary"
)

type bottle struct {
	ID   int64  `db:"id"`   // ID qq_grp_name_msg 的 crc64
	QQ   int64  `db:"qq"`   // QQ 发送者 qq
	Grp  int64  `db:"grp"`  // Grp 限制抽出的群 / 人（负数）
	Name string `db:"name"` // Name 发送者 昵称
	Msg  string `db:"msg"`  // Msg 消息，纯文本
}

var sea = &sql.Sqlite{}
var seamu sync.RWMutex

func newBottle(qq, grp int64, name, msg string) *bottle {
	id := int64(crc64.Checksum(binary.StringToBytes(fmt.Sprintf("%d_%d_%s_%s", qq, grp, name, msg)), crc64.MakeTable(crc64.ISO)))
	return &bottle{ID: id, QQ: qq, Grp: grp, Name: name, Msg: msg}
}

func (b *bottle) throw(db *sql.Sqlite, channel string) error {
	seamu.Lock()
	defer seamu.Unlock()
	return db.Insert(channel, b)
}

func (b *bottle) destroy(db *sql.Sqlite, channel string) error {
	seamu.Lock()
	defer seamu.Unlock()
	return db.Del(channel, "WHERE id="+strconv.FormatInt(b.ID, 10))
}

// fetchBottle grp != 0
func fetchBottle(db *sql.Sqlite, channel string, grp int64) (*bottle, error) {
	seamu.RLock()
	defer seamu.RUnlock()
	b := new(bottle)
	return b, db.Find(channel, b, "WHERE grp=0 or grp="+strconv.FormatInt(grp, 10)+" ORDER BY RANDOM() limit 1")
}

func createChannel(db *sql.Sqlite, channel string) error {
	seamu.Lock()
	defer seamu.Unlock()
	return db.Create(channel, &bottle{})
}
