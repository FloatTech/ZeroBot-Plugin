package antiabuse

import (
	"os"
	"sort"
	"testing"
)

func TestInsertWord(t *testing.T) {
	wordMap = make(map[int64]*Set[string])
	defer func() {
		wordMap = make(map[int64]*Set[string])
	}()
	path := "test.db"
	defer func() {
		err := os.Remove("test.db")
		if err != nil {
			t.Fatal(err)
		}
	}()
	db.DBPath = path
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
	err = insertWord(123, "one")
	if err != nil {
		t.Fatal(err)
	}
	if ok := wordMap[123].Include("one"); !ok {
		t.Fatal(`wordMap[123] should found "one" but not`)
	}
	if !db.CanFind("banWord", "WHERE group_id=123 AND word= 'one' ") {
		t.Fatal(`db should found 123-one but not`)
	}
}

func TestDeleteWord(t *testing.T) {
	wordMap = make(map[int64]*Set[string])
	defer func() {
		wordMap = make(map[int64]*Set[string])
	}()
	path := "test.db"
	defer func() {
		err := os.Remove("test.db")
		if err != nil {
			t.Fatal(err)
		}
	}()
	db.DBPath = path
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
	err = insertWord(123, "one")
	if err != nil {
		t.Fatal(err)
	}
	err = deleteWord(123, "one")
	if err != nil {
		t.Fatal(err)
	}
}

func TestShowWord(t *testing.T) {
	wordMap = make(map[int64]*Set[string])
	defer func() {
		wordMap = make(map[int64]*Set[string])
	}()
	path := "test.db"
	defer func() {
		err := os.Remove("test.db")
		if err != nil {
			t.Fatal(err)
		}
	}()
	db.DBPath = path
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
	err = insertWord(123, "one")
	if err != nil {
		t.Fatal(err)
	}
	err = insertWord(123, "one")
	if err != nil {
		t.Fatal(err)
	}
	err = insertWord(123, "two")
	if err != nil {
		t.Fatal(err)
	}
	var db123 []string
	var map123 []string
	obj := &banWord{}
	err = db.FindFor("banWord", obj, "WHERE group_id=123", func() error {
		db123 = append(db123, obj.Word)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(db123)
	if len(db123) != 2 || db123[0] != "one" || db123[1] != "two" {
		t.Fatal("db should found 123-one and 123-two but not")
	}
	map123 = wordMap[123].ToSlice()
	sort.Strings(map123)
	if len(map123) != 2 || map123[0] != "one" || map123[1] != "two" {
		t.Fatal("wordMap[123] should found 123-one and 123-two but not")
	}
}

func TestRecoverWord(t *testing.T) {
	wordMap = make(map[int64]*Set[string])
	defer func() {
		wordMap = make(map[int64]*Set[string])
	}()
	path := "test.db"
	defer func() {
		err := os.Remove("test.db")
		if err != nil {
			t.Fatal(err)
		}
	}()
	db.DBPath = path
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
	err = insertWord(123, "one")
	if err != nil {
		t.Fatal(err)
	}
	err = insertWord(123, "two")
	if err != nil {
		t.Fatal(err)
	}
	wordMap = make(map[int64]*Set[string])
	err = recoverWord()
	if err != nil {
		t.Fatal(err)
	}
	map123 := wordMap[123].ToSlice()
	sort.Strings(map123)
	if len(map123) != 2 || map123[0] != "one" || map123[1] != "two" {
		t.Fatal("wordMap[123] should found 123-one and 123-two but not")
	}
}
