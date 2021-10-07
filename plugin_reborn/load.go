package reborn

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	wr "github.com/mroth/weightedrand"
	log "github.com/sirupsen/logrus"
)

const (
	datapath = "data/Reborn"
	jsonfile = datapath + "/rate.json"
	pburl    = "https://codechina.csdn.net/u011570312/ZeroBot-Plugin/-/raw/master/" + jsonfile
)

type rate []struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

var (
	areac *wr.Chooser
)

func init() {
	go func() {
		time.Sleep(time.Second)
		err := os.MkdirAll(datapath, 0755)
		if err != nil {
			panic(err)
		}
		area := make(rate, 226)
		err = load(&area)
		if err != nil {
			panic(err)
		}
		choices := make([]wr.Choice, len(area))
		for i, a := range area {
			choices[i].Item = a.Name
			choices[i].Weight = uint(a.Weight * 1e9)
		}
		areac, err = wr.NewChooser(choices...)
		if err != nil {
			panic(err)
		}
		log.Printf("[Reborn]读取%d个国家/地区", len(area))
	}()
}

// load 加载rate数据
func load(area *rate) error {
	if _, err := os.Stat(jsonfile); err == nil || os.IsExist(err) {
		f, err := os.Open(jsonfile)
		if err == nil {
			defer f.Close()
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					return json.Unmarshal(data, area)
				}
			}
			return err1
		}
	} else { // 如果没有小作文，则从 url 下载
		f, err := os.Create(jsonfile)
		if err != nil {
			return err
		}
		defer f.Close()
		resp, err := http.Get(pburl)
		if err == nil {
			defer resp.Body.Close()
			if resp.ContentLength > 0 {
				log.Printf("[Reborn]从镜像下载国家和地区%d字节...", resp.ContentLength)
				data, err := io.ReadAll(resp.Body)
				if err == nil && len(data) > 0 {
					_, _ = f.Write(data)
					return json.Unmarshal(data, area)
				}
				return err
			}
			return nil
		}
		return err
	}
	return nil
}
