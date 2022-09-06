package antiabuse

import (
	"errors"
	"strconv"
	"strings"
	"sync"

	sqlite "github.com/FloatTech/sqlite"
)

type antidb struct {
	sync.RWMutex
	sqlite.Sqlite
}

type banWord struct {
	Word string `db:"word"`
}

var nilban = &banWord{}

func (db *antidb) isInAntiList(uid, gid int64, msg string) bool {
	grp := strconv.FormatInt(gid, 36)
	db.RLock()
	defer db.RUnlock()
	return db.CanFind(grp, "WHERE instr('"+msg+"', word)>0")
}

func (db *antidb) insertWord(gid int64, word string) error {
	grp := strconv.FormatInt(gid, 36)
	db.Lock()
	defer db.Unlock()
	err := db.Create(grp, nilban)
	if err != nil {
		return err
	}
	return db.Insert(grp, &banWord{Word: word})
}

func (db *antidb) deleteWord(gid int64, word string) error {
	grp := strconv.FormatInt(gid, 36)
	db.Lock()
	defer db.Unlock()
	if n, _ := db.Count(grp); n == 0 {
		return errors.New("本群还没有违禁词~")
	}
	return db.Del(grp, "WHERE word='"+word+"'")
}

func (db *antidb) listWords(gid int64) string {
	grp := strconv.FormatInt(gid, 36)
	word := &banWord{}
	sb := strings.Builder{}
	db.Lock()
	defer db.Unlock()
	_ = db.FindFor(grp, word, "", func() error {
		sb.WriteString(word.Word)
		sb.WriteString(" | ")
		return nil
	})
	if sb.Len() <= 3 {
		return ""
	}
	return sb.String()[:sb.Len()-3]
}
