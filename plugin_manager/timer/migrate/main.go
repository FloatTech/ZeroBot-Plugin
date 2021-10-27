package main

import (
	"fmt"
	io "io"
	"os"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
)

var timersmap TimersMapOld
var timersmapnew TimersMap

func loadTimers(pbfile string) bool {
	if file.IsExist(pbfile) {
		f, err := os.Open(pbfile)
		if err == nil {
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					err1 = timersmap.Unmarshal(data)
					if err1 == nil {
						return true
					}
				}
			}
		}
	}
	return false
}

// saveTimers 保存当前计时器
func saveTimers(pbfile string) error {
	data, err := timersmapnew.Marshal()
	if err == nil {
		f, err1 := os.OpenFile(pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err1 != nil {
			return err1
		} else {
			_, err2 := f.Write(data)
			f.Close()
			return err2
		}
	}
	return err
}

func main() {
	if len(os.Args) == 3 {
		if loadTimers(os.Args[1]) {
			timersmapnew.Timers = make(map[string]*Timer)
			for s, t := range timersmap.Timers {
				timersmapnew.Timers[s] = &Timer{
					Enable: t.Enable,
					Alert:  t.Alert,
					Url:    t.Url,
					Month:  t.Month,
					Day:    t.Day,
					Hour:   t.Hour,
					Minute: t.Minute,
				}
			}
			saveTimers(os.Args[2])
		}
	} else {
		fmt.Println("用法：旧文件路径 新文件路径")
	}
}
