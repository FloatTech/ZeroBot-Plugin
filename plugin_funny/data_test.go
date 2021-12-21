package funny

import (
	"crypto/md5"
	"encoding/binary"
	"os"
	"strings"
	"testing"

	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func TestFillData(t *testing.T) {
	data, err := os.ReadFile("laugh.txt")
	if err != nil {
		t.Fatal(err)
	}
	db := &sql.Sqlite{DBPath: "jokes.db"}
	err = db.Create("jokes", &joke{})
	if err != nil {
		t.Fatal(err)
	}
	jokes := strings.Split(helper.BytesToString(data), "\n")
	for _, j := range jokes {
		s := md5.Sum(helper.StringToBytes(j))
		db.Insert("jokes", &joke{ID: binary.LittleEndian.Uint32(s[:4]), Text: j})
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
}
