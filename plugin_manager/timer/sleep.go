package timer

import (
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

func firstWeek(date *time.Time, week time.Weekday) (d time.Time) {
	d = date.AddDate(0, 0, 1-date.Day())
	for d.Weekday() != week {
		d = d.AddDate(0, 0, 1)
	}
	return
}

func (ts *Timer) nextWakeTime() (date time.Time) {
	date = time.Now()
	m := ts.Month()
	d := ts.Day()
	h := ts.Hour()
	mn := ts.Minute()
	w := ts.Week()
	var unit time.Duration
	logrus.Debugln("[timer] unit init:", unit)
	if mn >= 0 {
		switch {
		case h < 0:
			if unit <= time.Second {
				unit = time.Hour
			}
		case d < 0 || w < 0:
			if unit <= time.Second {
				unit = time.Hour * 24
			}
		case d == 0 && w >= 0:
			delta := time.Hour * 24 * time.Duration(int(w)-int(date.Weekday()))
			if delta < 0 {
				delta = time.Hour * 24 * 7
			}
			unit += delta
		case m < 0:
			unit = -1
		}
	} else {
		unit = time.Minute
	}
	logrus.Debugln("[timer] unit:", unit)
	stable := 0
	if mn < 0 {
		mn = date.Minute()
	}
	if h < 0 {
		h = date.Hour()
	} else {
		stable |= 0x8
	}
	switch {
	case d < 0:
		d = date.Day()
	case d > 0:
		stable |= 0x4
	default:
		d = date.Day()
		if w >= 0 {
			stable |= 0x2
		}
	}
	if m < 0 {
		m = date.Month()
	} else {
		stable |= 0x1
	}
	switch stable {
	case 0b0101:
		if ts.Day() != time.Now().Day() || ts.Month() != time.Now().Month() {
			h = 0
		}
	case 0b1001:
		if ts.Month() != time.Now().Month() {
			d = 0
		}
	case 0b0001:
		if ts.Month() != time.Now().Month() {
			d = 0
			h = 0
		}
	}
	logrus.Debugln("[timer] stable:", stable)
	logrus.Debugln("[timer] m:", m, "d:", d, "h:", h, "mn:", mn, "w:", w)
	date = time.Date(date.Year(), time.Month(m), int(d), int(h), int(mn), date.Second(), date.Nanosecond(), date.Location())
	logrus.Debugln("[timer] date original:", date)
	if unit > 0 {
		date = date.Add(unit)
	}
	logrus.Debugln("[timer] date after add:", date)
	if time.Until(date) <= 0 {
		if ts.Month() < 0 {
			if ts.Day() > 0 || (ts.Day() == 0 && ts.Week() >= 0) {
				date = date.AddDate(0, 1, 0)
			} else if ts.Day() < 0 || ts.Week() < 0 {
				if ts.Hour() > 0 {
					date = date.AddDate(0, 0, 1)
				} else if ts.Minute() > 0 {
					date = date.Add(time.Hour)
				}
			}
		} else {
			date = date.AddDate(1, 0, 0)
		}
	}
	logrus.Debugln("[timer] date after fix:", date)
	if stable&0x8 != 0 && date.Hour() != int(h) {
		switch {
		case stable&0x4 == 0:
			date = date.AddDate(0, 0, 1).Add(-time.Hour)
		case stable&0x2 == 0:
			date = date.AddDate(0, 0, 7).Add(-time.Hour)
		case stable*0x1 == 0:
			date = date.AddDate(0, 1, 0).Add(-time.Hour)
		default:
			date = date.AddDate(1, 0, 0).Add(-time.Hour)
		}
	}
	logrus.Debugln("[timer] date after s8:", date)
	if stable&0x4 != 0 && date.Day() != int(d) {
		switch {
		case stable*0x1 == 0:
			date = date.AddDate(0, 1, -1)
		default:
			date = date.AddDate(1, 0, -1)
		}
	}
	logrus.Debugln("[timer] date after s4:", date)
	if stable&0x2 != 0 && date.Weekday() != w {
		switch {
		case stable*0x1 == 0:
			date = date.AddDate(0, 1, 0)
		default:
			date = date.AddDate(1, 0, 0)
		}
		date = firstWeek(&date, time.Weekday(w))
	}
	logrus.Debugln("[timer] date after s2:", date)
	if time.Until(date) <= 0 {
		date = time.Now().Add(time.Minute)
	}
	return date
}

func (ts *Timer) judgeHM() {
	if ts.Hour() < 0 || ts.Hour() == time.Now().Hour() {
		if ts.Minute() < 0 || ts.Minute() == time.Now().Minute() {
			if ts.SelfID != 0 {
				ts.sendmsg(ts.GrpID, zero.GetBot(ts.SelfID))
			} else {
				zero.RangeBot(func(id int64, ctx *zero.Ctx) (_ bool) {
					ts.sendmsg(ts.GrpID, ctx)
					return
				})
			}
		}
	}
}
