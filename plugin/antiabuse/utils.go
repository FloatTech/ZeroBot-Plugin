package antiabuse

import (
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	_ "unsafe"

	sqlite "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

//go:linkname managers github.com/FloatTech/zbputils/control.managers
var managers ctrl.Manager[*zero.Ctx]
var breakFlag = errors.New("break")
var db = &sqlite.Sqlite{}
var crc32Table = crc32.MakeTable(crc32.IEEE)
var wordMap = make(map[int64]*Set[string])

type banWord struct {
	Crc32ID uint32 `db:"crc32_id"`
	GroupID int64  `db:"group_id"`
	Word    string `db:"word"`
}

func banRule(ctx *zero.Ctx) bool {
	if !ctx.Event.IsToMe {
		return true
	}
	uid := ctx.Event.UserID
	gid := ctx.Event.GroupID
	if managers.IsBlocked(uid) {
		return false
	}
	wordSet := wordMap[gid]
	if wordSet == nil {
		return true
	}
	err := wordSet.Iter(func(word string) error {
		if strings.Contains(ctx.MessageString(), word) {
			if err := managers.DoBlock(uid); err != nil {
				return err
			}
			time.AfterFunc(4*time.Hour, func() {
				if err := managers.DoUnblock(uid); err != nil {
					logrus.Error("do unblock error:", err)
				}
			})
			return breakFlag
		}
		return nil
	})
	if err != nil && err != breakFlag {
		ctx.SendChain(message.Text("block user error:", err))
		return true
	}
	ctx.SetGroupBan(gid, uid, 4*3600)
	ctx.SendChain(message.Text("检测到违禁词,已封禁/屏蔽4小时"))
	return false
}

func insertWord(gid int64, word string) error {
	str := fmt.Sprintf("%d-%s", gid, word)
	checksum := crc32.Checksum([]byte(str), crc32Table)
	obj := &banWord{checksum, gid, word}
	err := db.Insert("banWord", obj)
	if err != nil {
		return err
	}
	if _, ok := wordMap[gid]; !ok {
		wordMap[gid] = NewSet[string]()
	}
	wordMap[gid].Add(word)
	return nil
}

func deleteWord(gid int64, word string) error {
	if _, ok := wordMap[gid]; !ok {
		return errors.New("本群还没有违禁词~")
	}
	if !wordMap[gid].Include(word) {
		return errors.New(word + " 不在本群违禁词集合中")
	}
	str := fmt.Sprintf("%d-%s", gid, word)
	checksum := crc32.Checksum([]byte(str), crc32Table)
	sql := fmt.Sprintf("WHERE crc32_id = %d", checksum)
	err := db.Del("banWord", sql)
	if err != nil {
		return err
	}
	wordMap[gid].Remove(word)
	return nil
}

func recoverWord() error {
	if !db.CanFind("banWord", "") {
		return nil
	}
	obj := &banWord{}
	err := db.FindFor("banWord", obj, "", func() error {
		if _, ok := wordMap[obj.GroupID]; !ok {
			wordMap[obj.GroupID] = NewSet[string]()
		}
		wordMap[obj.GroupID].Add(obj.Word)
		return nil
	},
	)
	return err
}
