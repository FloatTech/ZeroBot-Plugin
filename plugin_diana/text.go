package diana

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
	ARRAY []string
)

func init() {
	go func() {
		time.Sleep(time.Second)
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		loadText()
		ARRAY = compo.Array
	}()
}

func loadText() {
	if _, err := os.Stat(pbfile); err == nil || os.IsExist(err) {
		f, err := os.Open(pbfile)
		if err == nil {
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					compo.Unmarshal(data)
				}
			}
		}
	}
}

func addText(txt string) error {
	if txt != "" {
		ARRAY = append(ARRAY, txt)
		data, err := compo.Marshal()
		if err == nil {
			if _, err := os.Stat(datapath); err == nil || os.IsExist(err) {
				f, err1 := os.OpenFile(pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
				if err1 != nil {
					return err1
				} else {
					defer f.Close()
					_, err2 := f.Write(data)
					return err2
				}
			}
		}
		return err
	}
	return nil
}
