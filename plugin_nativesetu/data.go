package nativesetu

import (
	"bytes"
	"image"
	"io"
	"io/fs"
	"os"
	"sync"

	"github.com/corona10/goimagehash"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"golang.org/x/image/webp"

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/sql"
)

// setuclass holds setus in a folder, which is the class name.
type setuclass struct {
	ImgID int64  `db:"imgid"` // ImgID 图片唯一 id (dhash)
	Name  string `db:"name"`  // Name 图片名
	Path  string `db:"path"`  // Path 图片路径
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
		if file.IsExist(dbfile) {
			err := db.Open()
			if err == nil {
				setuclasses, err = db.ListTables()
			}
			if err != nil {
				logrus.Errorln("[nsetu]", err)
			}
		}
	}()
}

func scanall(path string) error {
	setuclasses = nil
	model := &setuclass{}
	root := os.DirFS(path)
	_ = db.Close()
	_ = os.Remove(dbfile)
	return fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			clsn := d.Name()
			if clsn != "." {
				mu.Lock()
				err = db.Create(clsn, model)
				setuclasses = append(setuclasses, clsn)
				mu.Unlock()
				if err == nil {
					err = scanclass(root, path, clsn)
					if err != nil {
						logrus.Errorln("[nsetu]", err)
						return err
					}
				}
			}
		}
		return nil
	})
}

func scanclass(root fs.FS, path, clsn string) error {
	ds, err := fs.ReadDir(root, path)
	if err != nil {
		return err
	}
	mu.Lock()
	_ = db.Truncate(clsn)
	mu.Unlock()
	for _, d := range ds {
		if !d.IsDir() {
			relpath := path + "/" + d.Name()
			logrus.Debugln("[nsetu] read", relpath)
			f, e := fs.ReadFile(root, relpath)
			if e != nil {
				return e
			}
			b := bytes.NewReader(f)
			img, _, e := image.Decode(b)
			if e != nil {
				b.Seek(0, io.SeekStart)
				img, e = webp.Decode(b)
			}
			if e != nil {
				return e
			}
			dh, e := goimagehash.DifferenceHash(img)
			if e != nil {
				return e
			}
			dhi := int64(dh.GetHash())
			logrus.Debugln("[nsetu] insert", d.Name(), "with id", dhi, "into", clsn)
			mu.Lock()
			err = db.Insert(clsn, &setuclass{ImgID: dhi, Name: d.Name(), Path: relpath})
			mu.Unlock()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
