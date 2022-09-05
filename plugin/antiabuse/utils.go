package antiabuse

import (
	"errors"
	"fmt"
	"hash/crc32"
	"strings"
	"time"

	sqlite "github.com/FloatTech/sqlite"
	"github.com/FloatTech/ttl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var managers = &ctrl.Manager[*zero.Ctx]{} //managers lazy load
var errBreak = errors.New("break")
var db = &sqlite.Sqlite{}
var crc32Table = crc32.MakeTable(crc32.IEEE)
var wordMap = make(map[int64]*Set[string])

func onDel(uid int64, _ struct{}) {
	if managers == nil {
		return
	}
	if err := managers.DoUnblock(uid); err != nil {
		logrus.Error("do unblock error:", err)
	}
}

var cache = ttl.NewCacheOn[int64, struct{}](4*time.Hour, [4]func(int64, struct{}){
	nil, nil, onDel, nil})

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
	wordSet := wordMap[gid]
	if wordSet == nil {
		return true
	}
	err := wordSet.Iter(func(word string) error {
		if strings.Contains(ctx.MessageString(), word) {
			if err := managers.DoBlock(uid); err != nil {
				return err
			}
			cache.Set(uid, struct{}{})
			return errBreak
		}
		return nil
	})
	if err != nil && err != errBreak {
		ctx.SendChain(message.Text("block user error:", err))
		return true
	}
	if err == errBreak {
		ctx.SetGroupBan(gid, uid, 4*3600)
		ctx.SendChain(message.Text("检测到违禁词,已封禁/屏蔽4小时"))
		return false
	}
	return true
}

func insertWord(gid int64, word string) error {
	str := fmt.Sprintf("%d-%s", gid, word)
	checksum := crc32.Checksum([]byte(str), crc32Table)
	obj := &banWord{checksum, gid, word}
	if _, ok := wordMap[gid]; !ok {
		wordMap[gid] = NewSet[string]()
	}
	wordMap[gid].Add(word)
	return db.Insert("banWord", obj)
}

func deleteWord(gid int64, word string) error {
	if _, ok := wordMap[gid]; !ok {
		return errors.New("本群还没有违禁词~")
	}
	if !wordMap[gid].Include(word) {
		return errors.New(word + " 不在本群违禁词集合中")
	}
	wordMap[gid].Remove(word)
	str := fmt.Sprintf("%d-%s", gid, word)
	checksum := crc32.Checksum([]byte(str), crc32Table)
	sql := fmt.Sprintf("WHERE crc32_id = %d", checksum)
	return db.Del("banWord", sql)
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
