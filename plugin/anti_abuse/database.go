package anti_abuse

import (
	"fmt"
	"time"

	sqlite "github.com/FloatTech/sqlite"
)

var db = &sqlite.Sqlite{}

type BanTime struct {
	UUID    string `db:"uuid"`
	DueTime int64  `db:"due_time"`
}

func insertUser(gid, uid int64) error {
	banTime := &BanTime{fmt.Sprintf("%d-%d", gid, uid), time.Now().Add(4 * time.Hour).UnixNano()}
	return db.Insert("BanTime", banTime)
}

func deleteUser(gid, uid int64) error {
	sql := fmt.Sprintf("WHERE uuid=%d-%d", gid, uid)
	return db.Del("BanTime", sql)
}

func recoverUser() error {
	banTime := &BanTime{}
	var uuids []string
	err := db.FindFor("BanTime", banTime, "", func() error {
		if time.Now().UnixNano() < banTime.DueTime {
			uuids = append(uuids, banTime.UUID)
		}
		return nil
	},
	)
	if err != nil {
		return err
	}
	banSet.AddMany(uuids)
	return nil
}

type BanWord struct {
	GroupWord string `db:"group_word"`
}

func insertWord(gid int64, word string) error {
	banWord := &BanWord{fmt.Sprintf("%d-%s", gid, word)}
	return db.Insert("BanWord", banWord)
}

func deleteWord(gid int64, word string) error {
	sql := fmt.Sprintf("WHERE group_word = %d-%s", gid, word)
	return db.Del("BanWord", sql)
}

func recoverWord() error {
	banWord := &BanWord{}
	var groupWords []string
	err := db.FindFor("BanWord", banWord, "", func() error {
		groupWords = append(groupWords, banWord.GroupWord)
		return nil
	},
	)
	if err != nil {
		return err
	}
	wordSet.AddMany(groupWords)
	return nil
}
