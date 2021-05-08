package manager

import (
	"fmt"
	"strconv"
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

func timer(ts TimeStamp, onTimeReached func()) {
	fmt.Printf("注册计时器: %d月%d日%d周%d:%d触发\n", ts.month, ts.day, ts.week, ts.hour, ts.minute)
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
	match := []rune("零一二三四五六七八九十")
	for i, m := range match {
		if c == m {
			return i
		}
	}
	return 0
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
