package moyu

import (
	"fmt"
	"strconv"
	"time"
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

// 获取两个时间相差
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
