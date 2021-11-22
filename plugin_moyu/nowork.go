package moyu

import (
	"fmt"
	"time"
)

type holiday struct {
	name string
	date time.Time
	dur  time.Duration
}

// NewHoliday 节日名 天数 年 月 日
func NewHoliday(name string, dur, year int, month time.Month, day int) *holiday {
	return &holiday{name: name, date: time.Date(year, month, day, 0, 0, 0, 0, time.Local), dur: time.Duration(dur) * time.Hour * 24}
}

// 获取两个时间相差
func (h *holiday) String() string {
	d := time.Until(h.date)
	if d >= 0 {
		return "距离" + h.name + "还有: " + d.String()
	} else if d+h.dur >= 0 {
		return "好好享受 " + h.name + " 假期吧!"
	} else {
		return "今年 " + h.name + " 假期已过"
	}
}

func weekend() string {
	t := time.Now().Weekday()
	switch t {
	case time.Sunday, time.Saturday:
		return "好好享受周末吧！"
	default:
		return fmt.Sprintf("距离周末还有:%d天！", 5-t)
	}
}
