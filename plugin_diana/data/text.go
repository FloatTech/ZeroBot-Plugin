// Package data 加载位于 datapath 的小作文
package data

import (
	"io"
	"os"
	"time"
)

const (
	datapath = "data/Diana"
	pbfile   = datapath + "/text.pb"
)

var (
	compo Composition
	// 小作文数组
	Array []string
)

func init() {
	go func() {
		time.Sleep(time.Second)
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		if LoadText() == nil {
			Array = compo.Array
		}
	}()
}

// LoadText 加载小作文
func LoadText() error {
	if _, err := os.Stat(pbfile); err == nil || os.IsExist(err) {
		f, err := os.Open(pbfile)
		if err == nil {
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					return compo.Unmarshal(data)
				}
			}
			return err1
		}
		return err
	}
	return nil
}

// AddText 添加小作文
func AddText(txt string) error {
	if txt != "" {
		compo.Array = append(compo.Array, txt)
		data, err := compo.Marshal()
		if err == nil {
			if _, err := os.Stat(datapath); err == nil || os.IsExist(err) {
				f, err1 := os.OpenFile(pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
				if err1 == nil {
					defer f.Close()
					_, err2 := f.Write(data)
					return err2
				}
				return err1
			}
		}
		return err
	}
	return nil
}
