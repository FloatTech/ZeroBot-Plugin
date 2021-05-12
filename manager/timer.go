package manager

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type TimeStamp struct {
	enable bool
	alert  string
	url    string
	month  int8
	day    int8
	week   int8
	hour   int8
	minute int8
}

//记录每个定时器以便取消
var timers = make(map[string]*TimeStamp)

func timer(ts TimeStamp, onTimeReached func()) {
	key := getTimerInfo(&ts)
	fmt.Printf("[群管]注册计时器: %s\n", key)
	t, ok := timers[key]
	if ok { //避免重复注册定时器
		t.enable = false
	}
	timers[key] = &ts
	judgeHM := func() {
		if ts.hour < 0 || ts.hour == int8(time.Now().Hour()) {
			if ts.minute < 0 || ts.minute == int8(time.Now().Minute()) {
				onTimeReached()
			}
		}
	}
	for ts.enable {
		if ts.month < 0 || ts.month == int8(time.Now().Month()) {
			if ts.day < 0 || ts.day == int8(time.Now().Day()) {
				judgeHM()
			} else if ts.day == 0 {
				if ts.week < 0 || ts.week == int8(time.Now().Weekday()) {
					judgeHM()
				}
			}
		}
		time.Sleep(time.Minute)
	}
}

//获得标准化定时字符串
func getTimerInfo(ts *TimeStamp) string {
	return fmt.Sprintf("%d月%d日%d周%d:%d", ts.month, ts.day, ts.week, ts.hour, ts.minute)
}

//获得填充好的ts
func getFilledTimeStamp(dateStrs []string, matchDateOnly bool) TimeStamp {
	monthStr := []rune(dateStrs[1])
	dayWeekStr := []rune(dateStrs[2])
	hourStr := []rune(dateStrs[3])
	minuteStr := []rune(dateStrs[4])

	var ts TimeStamp
	ts.month = chineseNum2Int(monthStr)
	if (ts.month != -1 && ts.month <= 0) || ts.month > 12 { //月份非法
		fmt.Println("[群管]月份非法！")
		return ts
	}
	lenOfDW := len(dayWeekStr)
	if lenOfDW == 4 { //包括末尾的"日"
		dayWeekStr = []rune{dayWeekStr[0], dayWeekStr[2]} //去除中间的十
		ts.day = chineseNum2Int(dayWeekStr)
		if (ts.day != -1 && ts.day <= 0) || ts.day > 31 { //日期非法
			fmt.Println("[群管]日期非法1！")
			return ts
		}
	} else if dayWeekStr[lenOfDW-1] == rune('日') { //xx日
		dayWeekStr = dayWeekStr[:lenOfDW-1]
		ts.day = chineseNum2Int(dayWeekStr)
		if (ts.day != -1 && ts.day <= 0) || ts.day > 31 { //日期非法
			fmt.Println("[群管]日期非法2！")
			return ts
		}
	} else if dayWeekStr[0] == rune('每') { //每周
		ts.week = -1
	} else { //周x
		ts.week = chineseNum2Int(dayWeekStr[1:])
		if ts.week == 7 { //周天是0
			ts.week = 0
		}
		if ts.week < 0 || ts.week > 6 { //星期非法
			ts.week = -11
			fmt.Println("[群管]星期非法！")
			return ts
		}
	}
	if len(hourStr) == 3 {
		hourStr = []rune{hourStr[0], hourStr[2]} //去除中间的十
	}
	ts.hour = chineseNum2Int(hourStr)
	if ts.hour < -1 || ts.hour > 23 { //小时非法
		fmt.Println("[群管]小时非法！")
		return ts
	}
	if len(minuteStr) == 3 {
		minuteStr = []rune{minuteStr[0], minuteStr[2]} //去除中间的十
	}
	ts.minute = chineseNum2Int(minuteStr)
	if ts.minute < -1 || ts.minute > 59 { //分钟非法
		fmt.Println("[群管]分钟非法！")
		return ts
	}
	if !matchDateOnly {
		urlStr := dateStrs[5]
		if urlStr != "" { //是图片url
			ts.url = urlStr[3:] //utf-8下用为3字节
			fmt.Println("[群管]" + ts.url)
			if !strings.HasPrefix(ts.url, "http") {
				ts.url = "illegal"
				fmt.Println("[群管]url非法！")
				return ts
			}
		}
		ts.alert = dateStrs[6]
		ts.enable = true
	}
	return ts
}

//汉字数字转int，仅支持-10～99，最多两位数，其中"每"解释为-1，"每两"为-2，以此类推
func chineseNum2Int(rs []rune) int8 {
	r := -1
	l := len(rs)
	mai := rune('每')
	if unicode.IsDigit(rs[0]) { //默认可能存在的第二位也为int
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
	return int8(r)
}

//处理单个字符的映射0~10
func chineseChar2Int(c rune) int {
	if c == rune('日') || c == rune('天') { //周日/周天
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

//@全体成员
func AtAll() message.MessageSegment {
	return message.MessageSegment{
		Type: "at",
		Data: map[string]string{
			"qq": "all",
		},
	}
}

//无缓存发送图片
func ImageNoCache(url string) message.MessageSegment {
	return message.MessageSegment{
		Type: "image",
		Data: map[string]string{
			"file":  url,
			"cache": "0",
		},
	}
}
