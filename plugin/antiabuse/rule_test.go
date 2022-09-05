package antiabuse

import (
	"os"
	"strings"
	"testing"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func mockBanRule(ctx *zero.Ctx) bool {
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
		return true
	}
	if err == errBreak {
		return false
	}
	return true
}

func TestBanRule(t *testing.T) {
	*managers = ctrl.NewManager[*zero.Ctx]("test.db", 0)
	defer func() {
		err := managers.D.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove("test.db")
		if err != nil {
			t.Fatal(err)
		}
	}()
	wordMap = make(map[int64]*Set[string])
	defer func() {
		wordMap = make(map[int64]*Set[string])
	}()
	db.DBPath = "test.db"
	err := db.Open(0)
	defer func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	err = db.Create("banWord", &banWord{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := db.Drop("banWord")
		if err != nil {
			t.Fatal(err)
		}
	}()
	ctx := zero.Ctx{}
	ctx.Event = &zero.Event{}
	ctx.Event.GroupID = 100
	ctx.Event.UserID = 111
	ctx.Event.Message = message.Message{message.Text("one")}
	ctx.Event.IsToMe = true
	ctx1 := zero.Ctx{}
	ctx1.Event = &zero.Event{}
	ctx1.Event.GroupID = 100
	ctx1.Event.UserID = 111
	ctx1.Event.Message = message.Message{message.Text("two")}
	ctx.Event.IsToMe = true
	ctx2 := zero.Ctx{}
	ctx2.Event = &zero.Event{}
	ctx2.Event.GroupID = 100
	ctx2.Event.UserID = 111
	ctx2.Event.Message = message.Message{message.Text("onetwo")}
	ctx2.Event.IsToMe = true
	err = insertWord(100, "one")
	if err != nil {
		t.Fatal(err)
	}
	if mockBanRule(&ctx) {
		t.Fatal("ctx cannot pass by rule")
	}
	if !mockBanRule(&ctx1) {
		t.Fatal("ctx1 should pass by rule")
	}
	if mockBanRule(&ctx2) {
		t.Fatal("ctx2 cannot pass by rule")
	}
	err = deleteWord(100, "one")
	if err != nil {
		t.Fatal(err)
	}
	if !banRule(&ctx) {
		t.Fatal("ctx should pass by rule")
	}
	if !banRule(&ctx1) {
		t.Fatal("ctx1 should pass by rule")
	}
	if !banRule(&ctx2) {
		t.Fatal("ctx2 should pass by rule")
	}
}
