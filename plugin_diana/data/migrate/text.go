package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"unsafe"

	"github.com/RomiChan/protobuf/proto"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

type Composition struct {
	Array []string `protobuf:"bytes,1,rep"`
}

var (
	compo Composition
)

type Text struct {
	Id   int64  `db:"id"`
	Data string `db:"data"`
}

func main() {
	err := LoadText(os.Args[1])
	if err == nil {
		arrl := len(compo.Array)
		fmt.Printf("[Diana]读取%d条小作文\n", arrl)
		db := sql.Sqlite{DBPath: os.Args[2]}
		err = db.Create("text", &Text{})
		if err != nil {
			panic(err)
		}
		for _, d := range compo.Array {
			s := md5.Sum(helper.StringToBytes(d))
			i := *(*int64)(unsafe.Pointer(&s))
			fmt.Printf("[Diana]id: %d\n", i)
			err = db.Insert("text", &Text{
				Id:   i,
				Data: d,
			})
			if err != nil {
				panic(err)
			}
			c, _ := db.Count("text")
			fmt.Println("[Diana]转化", c, "条小作文")
		}
		err = db.Close()
		if err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}
}

// LoadText 加载小作文
func LoadText(pbfile string) error {
	data, err := os.ReadFile(pbfile)
	if err != nil {
		return err
	}
	return proto.Unmarshal(data, &compo)
}
