package timer

import "time"

// En isEnabled 1bit
func (t *Timer) En() (en bool) {
	return t.En1Month4Day5Week3Hour5Min6&0x800000 != 0
}

// Month 4bits
func (t *Timer) Month() (mon time.Month) {
	mon = time.Month((t.En1Month4Day5Week3Hour5Min6 & 0x780000) >> 19)
	if mon == 0b1111 {
		mon = -1
	}
	return
}

// Day 5bits
func (t *Timer) Day() (d int) {
	d = int((t.En1Month4Day5Week3Hour5Min6 & 0x07c000) >> 14)
	if d == 0b11111 {
		d = -1
	}
	return
}

// Week 3bits
func (t *Timer) Week() (w time.Weekday) {
	w = time.Weekday((t.En1Month4Day5Week3Hour5Min6 & 0x003800) >> 11)
	if w == 0b111 {
		w = -1
	}
	return
}

// Hour 5bits
func (t *Timer) Hour() (h int) {
	h = int((t.En1Month4Day5Week3Hour5Min6 & 0x0007c0) >> 6)
	if h == 0b11111 {
		h = -1
	}
	return
}

// Minute 6bits
func (t *Timer) Minute() (min int) {
	min = int(t.En1Month4Day5Week3Hour5Min6 & 0x00003f)
	if min == 0b111111 {
		min = -1
	}
	return
}

// SetEn ...
func (t *Timer) SetEn(en bool) {
	if en {
		t.En1Month4Day5Week3Hour5Min6 |= 0x800000
	} else {
		t.En1Month4Day5Week3Hour5Min6 &= 0x7fffff
	}
}

// SetMonth ...
func (t *Timer) SetMonth(mon time.Month) {
	t.En1Month4Day5Week3Hour5Min6 = ((int32(mon) << 19) & 0x780000) | (t.En1Month4Day5Week3Hour5Min6 & 0x87ffff)
}

// SetDay ...
func (t *Timer) SetDay(d int) {
	t.En1Month4Day5Week3Hour5Min6 = ((int32(d) << 14) & 0x07c000) | (t.En1Month4Day5Week3Hour5Min6 & 0xf83fff)
}

// SetWeek ...
func (t *Timer) SetWeek(w time.Weekday) {
	t.En1Month4Day5Week3Hour5Min6 = ((int32(w) << 11) & 0x003800) | (t.En1Month4Day5Week3Hour5Min6 & 0xffc7ff)
}

// SetHour ...
func (t *Timer) SetHour(h int) {
	t.En1Month4Day5Week3Hour5Min6 = ((int32(h) << 6) & 0x0007c0) | (t.En1Month4Day5Week3Hour5Min6 & 0xfff83f)
}

// SetMinute ...
func (t *Timer) SetMinute(min int) {
	t.En1Month4Day5Week3Hour5Min6 = (int32(min) & 0x00003f) | (t.En1Month4Day5Week3Hour5Min6 & 0xffffc0)
}
