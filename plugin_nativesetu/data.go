package nativesetu

import (
	"bytes"
	"image"
	"io/fs"
	"os"
	"strings"
	"sync"

	"github.com/corona10/goimagehash"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	_ "golang.org/x/image/webp" // import webp decoding

	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/sql"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

// setuclass holds setus in a folder, which is the class name.
type setuclass struct {
	ImgID int64  `db:"imgid"` // ImgID 图片唯一 id (dhash)
	Name  string `db:"name"`  // Name 图片名
	Path  string `db:"path"`  // Path 图片路径
}

var ns = &nsetu{db: &sql.Sqlite{DBPath: dbfile}}

func init() {
	go func() {
		defer order.DoneOnExit()()
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

type nsetu struct {
	db *sql.Sqlite
	mu sync.RWMutex
}

func (n *nsetu) List() (l []string) {
	if file.IsExist(n.db.DBPath) {
		err := n.db.Open()
		if err == nil {
			l, err = n.db.ListTables()
		}
		if err != nil {
			logrus.Errorln("[nsetu]", err)
		}
	}
	return
}

func (n *nsetu) scanall(path string) error {
	model := &setuclass{}
	root := os.DirFS(path)
	_ = n.db.Close()
	_ = os.Remove(n.db.DBPath)
	return fs.WalkDir(root, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			clsn := d.Name()
			if clsn != "." {
				n.mu.Lock()
				err = n.db.Create(clsn, model)
				n.mu.Unlock()
				if err == nil {
					err = n.scanclass(root, path, clsn)
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

func (n *nsetu) scanclass(root fs.FS, path, clsn string) error {
	ds, err := fs.ReadDir(root, path)
	if err != nil {
		return err
	}
	n.mu.Lock()
	_ = n.db.Truncate(clsn)
	n.mu.Unlock()
	for _, d := range ds {
		nm := d.Name()
		ln := strings.ToLower(nm)
		if !d.IsDir() &&
			(strings.HasSuffix(ln, ".jpg") || strings.HasSuffix(ln, ".jpeg") ||
				strings.HasSuffix(ln, ".png") || strings.HasSuffix(ln, ".gif") || strings.HasSuffix(ln, ".webp")) {
			relpath := path + "/" + nm
			logrus.Debugln("[nsetu] read", relpath)
			f, e := fs.ReadFile(root, relpath)
			if e != nil {
				return e
			}
			b := bytes.NewReader(f)
			img, _, e := image.Decode(b)
			if e != nil {
				return e
			}
			dh, e := goimagehash.DifferenceHash(img)
			if e != nil {
				return e
			}
			dhi := int64(dh.GetHash())
			logrus.Debugln("[nsetu] insert", nm, "with id", dhi, "into", clsn)
			n.mu.Lock()
			err = n.db.Insert(clsn, &setuclass{ImgID: dhi, Name: nm, Path: relpath})
			n.mu.Unlock()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
