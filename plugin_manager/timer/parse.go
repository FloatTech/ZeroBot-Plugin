package timer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
)

// GetTimerInfo 获得标准化定时字符串
func (ts *Timer) GetTimerInfo(grp int64) string {
	if ts.Cron != "" {
		return fmt.Sprintf("[%d]%s", grp, ts.Cron)
	}
	return fmt.Sprintf("[%d]%d月%d日%d周%d:%d", grp, ts.Month, ts.Day, ts.Week, ts.Hour, ts.Minute)
}

// GetFilledCronTimer 获得以cron填充好的ts
func GetFilledCronTimer(croncmd string, alert string, img string, botqq int64) *Timer {
	var ts Timer
	ts.Alert = alert
	ts.Cron = croncmd
	ts.Url = img
	ts.Selfid = botqq
	return &ts
}

// GetFilledTimer 获得填充好的ts
func GetFilledTimer(dateStrs []string, botqq int64, matchDateOnly bool) *Timer {
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
	ts.Selfid = botqq
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
