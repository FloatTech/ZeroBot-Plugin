package antiabuse

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	sqlite "github.com/FloatTech/sqlite"
	"github.com/FloatTech/ttl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

var managers *ctrl.Manager[*zero.Ctx] // managers lazy load
var db = &sqlite.Sqlite{}
var mu sync.RWMutex

func onDel(uid int64, _ struct{}) {
	if managers == nil {
		return
	}
	if err := managers.DoUnblock(uid); err != nil {
		logrus.Errorln("[antiabuse] do unblock:", err)
	}
}

var cache = ttl.NewCacheOn(4*time.Hour, [4]func(int64, struct{}){nil, nil, onDel, nil})

type banWord struct {
	Word string `db:"word"`
}

var nilban = &banWord{}

func insertWord(gid int64, word string) error {
	grp := strconv.FormatInt(gid, 36)
	mu.Lock()
	defer mu.Unlock()
	err := db.Create(grp, nilban)
	if err != nil {
		return err
	}
	return db.Insert(grp, (*banWord)(unsafe.Pointer(&word)))
}

func deleteWord(gid int64, word string) error {
	grp := strconv.FormatInt(gid, 36)
	mu.Lock()
	defer mu.Unlock()
	if n, _ := db.Count(grp); n == 0 {
		return errors.New("本群还没有违禁词~")
	}
	return db.Del(grp, "WHRER word='"+word+"'")
}

func listWords(gid int64) string {
	grp := strconv.FormatInt(gid, 36)
	word := ""
	ptr := (*banWord)(unsafe.Pointer(&word))
	sb := strings.Builder{}
	mu.Lock()
	defer mu.Unlock()
	_ = db.FindFor(grp, ptr, "", func() error {
		sb.WriteString(word)
		sb.WriteString(" |")
		return nil
	})
	if sb.Len() <= 2 {
		return ""
	}
	return sb.String()[:sb.Len()-2]
}
