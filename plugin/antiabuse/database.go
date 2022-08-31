package antiabuse

import (
	"fmt"
	"time"

	sqlite "github.com/FloatTech/sqlite"
)

var db = &sqlite.Sqlite{}

type banUser struct {
	UUID    string `db:"uuid"`
	DueTime int64  `db:"due_time"`
}

func insertUser(gid, uid int64) error {
	obj := &banUser{fmt.Sprintf("%d-%d", gid, uid), time.Now().Add(4 * time.Hour).UnixNano()}
	return db.Insert("banUser", obj)
}

func deleteUser(gid, uid int64) error {
	sql := fmt.Sprintf("WHERE uuid=%d-%d", gid, uid)
	return db.Del("banUser", sql)
}

func recoverUser() error {
	if !db.CanFind("banUser", "") {
		return nil
	}
	obj := &banUser{}
	var uuids []string
	err := db.FindFor("banUser", obj, "", func() error {
		if time.Now().UnixNano() < obj.DueTime {
			uuids = append(uuids, obj.UUID)
		} else {
			if err := db.Del("banUser", "WHERE uuid="+obj.UUID); err != nil {
				return err
			}
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

type banWord struct {
	GroupWord string `db:"group_word"`
}

func insertWord(gid int64, word string) error {
	obj := &banWord{fmt.Sprintf("%d-%s", gid, word)}
	return db.Insert("banWord", obj)
}

func deleteWord(gid int64, word string) error {
	sql := fmt.Sprintf("WHERE group_word = %d-%s", gid, word)
	return db.Del("banWord", sql)
}

func recoverWord() error {
	if !db.CanFind("banWord", "") {
		return nil
	}
	obj := &banWord{}
	var groupWords []string
	err := db.FindFor("banWord", obj, "", func() error {
		groupWords = append(groupWords, obj.GroupWord)
		return nil
	},
	)
	if err != nil {
		return err
	}
	wordSet.AddMany(groupWords)
	return nil
}
