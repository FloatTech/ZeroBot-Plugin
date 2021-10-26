// Package timer 群管定时器
package timer

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/utils/file"
)

type Clock struct {
	// 记录每个定时器以便取消
	timersmap TimersMap
	// 定时器map
	timers *(map[string]*Timer)
	// 读写锁
	timersmu sync.RWMutex
	// 定时器存储位置
	pbfile *string
}

var (
	// @全体成员
	atall = message.MessageSegment{
		Type: "at",
		Data: map[string]string{
			"qq": "all",
		},
	}
)

func NewClock(pbfile string) (c Clock) {
	c.loadTimers(pbfile)
	c.timers = &c.timersmap.Timers
	c.pbfile = &pbfile
	return
}

func firstWeek(date *time.Time, week time.Weekday) (d time.Time) {
	d = date.AddDate(0, 0, 1-date.Day())
	for d.Weekday() != week {
		d = d.AddDate(0, 0, 1)
	}
	return
}

func (ts *Timer) nextWakeTime() (date time.Time) {
	date = time.Now()
	m := ts.Month
	d := ts.Day
	h := ts.Hour
	mn := ts.Minute
	w := ts.Week
	unit := time.Duration(int(ts.Minute)-date.Minute()) * time.Minute
	logrus.Debugln("[timer] unit init:", unit)
	if mn >= 0 {
		switch {
		case h < 0:
			if unit <= time.Second {
				unit += time.Hour
			}
		case d < 0 || w < 0:
			if unit <= time.Second {
				unit += time.Hour * 24
			}
		case d == 0 && w >= 0:
			delta := time.Hour * 24 * time.Duration(int(w)-int(date.Weekday()))
			if delta < 0 {
				delta += time.Hour * 24 * 7
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
		mn = int32(date.Minute())
	}
	if h < 0 {
		h = int32(date.Hour())
	} else {
		stable |= 0x8
	}
	if d < 0 {
		d = int32(date.Day())
	} else if d > 0 {
		stable |= 0x4
	} else {
		d = int32(date.Day())
		if w >= 0 {
			stable |= 0x2
		}
	}
	if m < 0 {
		m = int32(date.Month())
	} else {
		stable |= 0x1
	}
	switch stable {
	case 0b0101:
		if ts.Day != int32(time.Now().Day()) || ts.Month != int32(time.Now().Month()) {
			h = 0
		}
	case 0b1001:
		if ts.Month != int32(time.Now().Month()) {
			d = 0
		}
	case 0b0001:
		if ts.Month != int32(time.Now().Month()) {
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
		if ts.Month < 0 {
			if ts.Day > 0 || (ts.Day == 0 && ts.Week >= 0) {
				date = date.AddDate(0, 1, 0)
			} else if ts.Day < 0 || ts.Week < 0 {
				if ts.Hour > 0 {
					date = date.AddDate(0, 0, 1)
				} else if ts.Minute > 0 {
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
	if stable&0x2 != 0 && int32(date.Weekday()) != w {
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
	if ts.Hour < 0 || ts.Hour == int32(time.Now().Hour()) {
		if ts.Minute < 0 || ts.Minute == int32(time.Now().Minute()) {
			zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
				ctx.Event = new(zero.Event)
				ctx.Event.GroupID = int64(ts.Grpid)
				if ts.Url == "" {
					ctx.SendChain(atall, message.Text(ts.Alert))
				} else {
					ctx.SendChain(atall, message.Text(ts.Alert), message.Image(ts.Url).Add("cache", "0"))
				}
				return false
			})
		}
	}
}

// RegisterTimer 注册计时器
func (c *Clock) RegisterTimer(ts *Timer, save bool) {
	key := ts.GetTimerInfo()
	if c.timers != nil {
		t, ok := c.GetTimer(key)
		if t != ts && ok { // 避免重复注册定时器
			t.Enable = false
		}
		c.timersmu.Lock()
		(*c.timers)[key] = ts
		c.timersmu.Unlock()
		if save {
			c.SaveTimers()
		}
	}
	logrus.Printf("[群管]注册计时器[%t]%s", ts.Enable, key)
	for ts.Enable {
		nextdate := ts.nextWakeTime()
		sleepsec := time.Until(nextdate)
		logrus.Printf("[群管]计时器%s将睡眠%ds", key, sleepsec/time.Second)
		time.Sleep(sleepsec)
		if ts.Enable {
			if ts.Month < 0 || ts.Month == int32(time.Now().Month()) {
				if ts.Day < 0 || ts.Day == int32(time.Now().Day()) {
					ts.judgeHM()
				} else if ts.Day == 0 {
					if ts.Week < 0 || ts.Week == int32(time.Now().Weekday()) {
						ts.judgeHM()
					}
				}
			}
		}
	}
}

// CancelTimer 取消计时器
func (c *Clock) CancelTimer(key string) bool {
	t, ok := (*c.timers)[key]
	if ok {
		t.Enable = false
		c.timersmu.Lock()
		delete(*c.timers, key) // 避免重复取消
		c.timersmu.Unlock()
		_ = c.SaveTimers()
	}
	return ok
}

// SaveTimers 保存当前计时器
func (c *Clock) SaveTimers() error {
	c.timersmu.RLock()
	data, err := c.timersmap.Marshal()
	c.timersmu.RUnlock()
	if err == nil {
		c.timersmu.Lock()
		defer c.timersmu.Unlock()
		f, err1 := os.OpenFile(*c.pbfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
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

// ListTimers 列出本群所有计时器
func (c *Clock) ListTimers(grpID uint64) []string {
	// 数组默认长度为map长度,后面append时,不需要重新申请内存和拷贝,效率很高
	if c.timers != nil {
		g := strconv.FormatUint(grpID, 10)
		c.timersmu.RLock()
		keys := make([]string, 0, len(*c.timers))
		for k := range *c.timers {
			if strings.Contains(k, g) {
				start := strings.Index(k, "]")
				msg := strings.ReplaceAll(k[start+1:]+"\n", "-1", "每")
				msg = strings.ReplaceAll(msg, "日0周", "日周天")
				keys = append(keys, msg)
			}
		}
		c.timersmu.RUnlock()
		return keys
	} else {
		return nil
	}
}

func (c *Clock) GetTimer(key string) (t *Timer, ok bool) {
	c.timersmu.RLock()
	t, ok = (*c.timers)[key]
	c.timersmu.RUnlock()
	return
}

func (c *Clock) loadTimers(pbfile string) {
	if file.IsExist(pbfile) {
		f, err := os.Open(pbfile)
		if err == nil {
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					c.timersmap.Unmarshal(data)
					for _, t := range c.timersmap.Timers {
						go c.RegisterTimer(t, false)
					}
					return
				}
			}
		}
	}
	c.timersmap.Timers = make(map[string]*Timer)
}

// GetTimerInfo 获得标准化定时字符串
func (ts *Timer) GetTimerInfo() string {
	return fmt.Sprintf("[%d]%d月%d日%d周%d:%d", ts.Grpid, ts.Month, ts.Day, ts.Week, ts.Hour, ts.Minute)
}

// GetFilledTimer 获得填充好的ts
func GetFilledTimer(dateStrs []string, matchDateOnly bool) *Timer {
	monthStr := []rune(dateStrs[1])
	dayWeekStr := []rune(dateStrs[2])
	hourStr := []rune(dateStrs[3])
	minuteStr := []rune(dateStrs[4])

	var ts Timer
	ts.Month = chineseNum2Int(monthStr)
	if (ts.Month != -1 && ts.Month <= 0) || ts.Month > 12 { // 月份非法
		logrus.Println("[群管]月份非法！")
		return &ts
	}
	lenOfDW := len(dayWeekStr)
	if lenOfDW == 4 { // 包括末尾的"日"
		dayWeekStr = []rune{dayWeekStr[0], dayWeekStr[2]} // 去除中间的十
		ts.Day = chineseNum2Int(dayWeekStr)
		if (ts.Day != -1 && ts.Day <= 0) || ts.Day > 31 { // 日期非法
			logrus.Println("[群管]日期非法1！")
			return &ts
		}
	} else if dayWeekStr[lenOfDW-1] == rune('日') { // xx日
		dayWeekStr = dayWeekStr[:lenOfDW-1]
		ts.Day = chineseNum2Int(dayWeekStr)
		if (ts.Day != -1 && ts.Day <= 0) || ts.Day > 31 { // 日期非法
			logrus.Println("[群管]日期非法2！")
			return &ts
		}
	} else if dayWeekStr[0] == rune('每') { // 每周
		ts.Week = -1
	} else { // 周x
		ts.Week = chineseNum2Int(dayWeekStr[1:])
		if ts.Week == 7 { // 周天是0
			ts.Week = 0
		}
		if ts.Week < 0 || ts.Week > 6 { // 星期非法
			ts.Week = -11
			logrus.Println("[群管]星期非法！")
			return &ts
		}
	}
	if len(hourStr) == 3 {
		hourStr = []rune{hourStr[0], hourStr[2]} // 去除中间的十
	}
	ts.Hour = chineseNum2Int(hourStr)
	if ts.Hour < -1 || ts.Hour > 23 { // 小时非法
		logrus.Println("[群管]小时非法！")
		return &ts
	}
	if len(minuteStr) == 3 {
		minuteStr = []rune{minuteStr[0], minuteStr[2]} // 去除中间的十
	}
	ts.Minute = chineseNum2Int(minuteStr)
	if ts.Minute < -1 || ts.Minute > 59 { // 分钟非法
		logrus.Println("[群管]分钟非法！")
		return &ts
	}
	if !matchDateOnly {
		urlStr := dateStrs[5]
		if urlStr != "" { // 是图片url
			ts.Url = urlStr[3:] // utf-8下用为3字节
			logrus.Println("[群管]" + ts.Url)
			if !strings.HasPrefix(ts.Url, "http") {
				ts.Url = "illegal"
				logrus.Println("[群管]url非法！")
				return &ts
			}
		}
		ts.Alert = dateStrs[6]
		ts.Enable = true
	}
	return &ts
}

// chineseNum2Int 汉字数字转int，仅支持-10～99，最多两位数，其中"每"解释为-1，"每二"为-2，以此类推
func chineseNum2Int(rs []rune) int32 {
	r := -1
	l := len(rs)
	mai := rune('每')
	if unicode.IsDigit(rs[0]) { // 默认可能存在的第二位也为int
		r, _ = strconv.Atoi(string(rs))
	} else {
		if rs[0] == mai {
			if l == 2 {
				r = -chineseChar2Int(rs[1])
			}
		} else if l == 1 {
			r = chineseChar2Int(rs[0])
		} else {
			ten := chineseChar2Int(rs[0])
			if ten != 10 {
				ten *= 10
			}
			ge := chineseChar2Int(rs[1])
			if ge == 10 {
				ge = 0
			}
			r = ten + ge
		}
	}
	return int32(r)
}

// chineseChar2Int 处理单个字符的映射0~10
func chineseChar2Int(c rune) int {
	if c == rune('日') || c == rune('天') { // 周日/周天
		return 7
	} else {
		match := []rune("零一二三四五六七八九十")
		for i, m := range match {
			if c == m {
				return i
			}
		}
		return 0
	}
}
