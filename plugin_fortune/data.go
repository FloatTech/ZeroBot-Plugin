package fortune

import (
	io "io"
	"os"
	"sync"
)

var (
	conf Conf
	mu   sync.Mutex
)

func loadcfg(name string) error {
	name = base + name
	if _, err := os.Stat(name); err == nil || os.IsExist(err) {
		f, err := os.Open(name)
		if err == nil {
			defer f.Close()
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					return conf.Unmarshal(data)
				}
			}
			return err1
		}
	} else { // 如果没有 cfg，则使用空 map
		conf.Kind = make(map[int64]uint32)
	}
	return nil
}

func savecfg(name string) error {
	name = base + name
	data, err := conf.Marshal()
	if err == nil {
		if _, err := os.Stat(base); err == nil || os.IsExist(err) {
			f, err1 := os.OpenFile(name, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
			if err1 == nil {
				mu.Lock()
				_, err2 := f.Write(data)
				f.Close()
				mu.Unlock()
				return err2
			}
			return err1
		}
	}
	return err
}
