package nativesetu

import (
	"image"
	"io/fs"
	"os"
	"sync"

	"github.com/corona10/goimagehash"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
	"github.com/FloatTech/ZeroBot-Plugin/utils/process"
	"github.com/FloatTech/ZeroBot-Plugin/utils/sql"
)

// setuclass holds setus in a folder, which is the class name.
type setuclass struct {
	ImgID uint64 `db:"imgid"` // ImgID 图片唯一 id (dhash)
	Name  string `db:"name"`  // Name 图片名
}

var (
	setuclasses []string
	db          = &sql.Sqlite{DBPath: dbfile}
	mu          sync.RWMutex
)

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		if file.IsExist(cfgfile) {
			b, err := os.ReadFile(cfgfile)
			if err == nil {
				setupath = helper.BytesToString(b)
				logrus.Println("[nsetu] set setu dir to", setupath)
			}
		}
	}()
}

func scanall(path string) error {
	setuclasses = setuclasses[:0]
	model := &setuclass{}
	root := os.DirFS(path)
	return fs.WalkDir(root, "./", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			clsn := d.Name()
			mu.Lock()
			err = db.Create(clsn, model)
			setuclasses = append(setuclasses, clsn)
			mu.Unlock()
			if err == nil {
				err = scanclass(root, clsn)
				if err != nil {
					return err
				}
			}
		}
		return err
	})
}

func scanclass(root fs.FS, clsn string) error {
	return fs.WalkDir(root, clsn, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			f, e := os.Open(path)
			if e != nil {
				return e
			}
			img, _, e := image.Decode(f)
			if e != nil {
				return e
			}
			dh, e := goimagehash.DifferenceHash(img)
			if e != nil {
				return e
			}
			mu.Lock()
			err = db.Insert(clsn, &setuclass{ImgID: dh.GetHash(), Name: d.Name()})
			mu.Unlock()
		}
		return err
	})
}
