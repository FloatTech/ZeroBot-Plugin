package moyu

import (
	"fmt"
	"strconv"
	"time"

	reg "github.com/fumiama/go-registry"
	"github.com/sirupsen/logrus"
)

// Holiday 节日
type Holiday struct {
	name string
	date time.Time
	dur  time.Duration
}

// NewHoliday 节日名 天数 年 月 日
func NewHoliday(name string, dur, year int, month time.Month, day int) *Holiday {
	return &Holiday{name: name, date: time.Date(year, month, day, 0, 0, 0, 0, time.Local), dur: time.Duration(dur) * time.Hour * 24}
}

var registry = reg.NewRegReader("reilia.westeurope.cloudapp.azure.com:32664", "fumiama")

// GetHoliday 从 reg 服务器获取节日
func GetHoliday(name string) *Holiday {
	var dur, year int
	var month time.Month
	var day int
	ret, err := registry.Get("holiday/" + name)
	if err != nil {
		return NewHoliday(name+err.Error(), 0, 0, 0, 0)
	}
	fmt.Sscanf(ret, "%d_%d_%d_%d", &dur, &year, &month, &day)
	logrus.Debugln("[moyu]获取节日:", name, dur, year, month, day)
	return NewHoliday(name, dur, year, month, day)
}

// String 获取两个时间相差
func (h *Holiday) String() string {
	d := time.Until(h.date)
	switch {
	case d >= 0:
		return "距离" + h.name + "还有: " + strconv.FormatFloat(d.Hours()/24.0, 'f', 2, 64) + "天！"
	case d+h.dur >= 0:
		return "好好享受 " + h.name + " 假期吧!"
	default:
		return "今年 " + h.name + " 假期已过"
	}
}

func weekend() string {
	t := time.Now().Weekday()
	if t == time.Sunday || t == time.Saturday {
		return "好好享受周末吧！"
	}
	return fmt.Sprintf("距离周末还有:%d天！", 5-t)
}
