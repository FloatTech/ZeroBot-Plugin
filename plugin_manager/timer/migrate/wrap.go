package main

import "time"

func (m *Timer) En() (en bool) {
	return m.En1Month4Day5Week3Hour5Min6&0x800000 != 0
}

func (m *Timer) Month() (mon time.Month) {
	mon = time.Month((m.En1Month4Day5Week3Hour5Min6 & 0x780000) >> 19)
	if mon == 0b1111 {
		mon = -1
	}
	return
}

func (m *Timer) Day() (d int) {
	d = int((m.En1Month4Day5Week3Hour5Min6 & 0x07c000) >> 14)
	if d == 0b11111 {
		d = -1
	}
	return
}

func (m *Timer) Week() (w time.Weekday) {
	w = time.Weekday((m.En1Month4Day5Week3Hour5Min6 & 0x003800) >> 11)
	if w == 0b111 {
		w = -1
	}
	return
}

func (m *Timer) Hour() (h int) {
	h = int((m.En1Month4Day5Week3Hour5Min6 & 0x0007c0) >> 6)
	if h == 0b11111 {
		h = -1
	}
	return
}

func (m *Timer) Minute() (min int) {
	min = int(m.En1Month4Day5Week3Hour5Min6 & 0x00003f)
	if min == 0b111111 {
		min = -1
	}
	return
}

func (m *Timer) SetEn(en bool) {
	if en {
		m.En1Month4Day5Week3Hour5Min6 |= 0x800000
	} else {
		m.En1Month4Day5Week3Hour5Min6 &= 0x7fffff
	}
}

func (m *Timer) SetMonth(mon time.Month) {
	m.En1Month4Day5Week3Hour5Min6 = ((int32(mon) << 19) & 0x780000) | (m.En1Month4Day5Week3Hour5Min6 & 0x87ffff)
}

func (m *Timer) SetDay(d int) {
	m.En1Month4Day5Week3Hour5Min6 = ((int32(d) << 14) & 0x07c000) | (m.En1Month4Day5Week3Hour5Min6 & 0xf83fff)
}

func (m *Timer) SetWeek(w time.Weekday) {
	m.En1Month4Day5Week3Hour5Min6 = ((int32(w) << 11) & 0x003800) | (m.En1Month4Day5Week3Hour5Min6 & 0xffc7ff)
}

func (m *Timer) SetHour(h int) {
	m.En1Month4Day5Week3Hour5Min6 = ((int32(h) << 6) & 0x0007c0) | (m.En1Month4Day5Week3Hour5Min6 & 0xfff83f)
}

func (m *Timer) SetMinute(min int) {
	m.En1Month4Day5Week3Hour5Min6 = (int32(min) & 0x00003f) | (m.En1Month4Day5Week3Hour5Min6 & 0xffffc0)
}
