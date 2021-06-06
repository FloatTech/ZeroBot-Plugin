package timer

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/msgext"
	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/utils"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

type (
	TimeStamp = Timer
	Ctx       = zero.Ctx
)

var (
	//记录每个定时器以便取消
	timersmap TimersMap
	Timers    *(map[string]*Timer)
	//定时器存储位置
	BOTPATH  = utils.PathExecute()       // 当前bot运行目录
	DATAPATH = BOTPATH + "data/manager/" // 数据目录
	PBFILE   = DATAPATH + "timers.pb"
)

func init() {
	go func() {
		time.Sleep(time.Second)
		utils.CreatePath(DATAPATH)
		loadTimers()
		Timers = &timersmap.Timers
	}()
}

func judgeHM(ts *TimeStamp) {
	if ts.Hour < 0 || ts.Hour == int32(time.Now().Hour()) {
		if ts.Minute < 0 || ts.Minute == int32(time.Now().Minute()) {
			zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
				ctx.Event = new(zero.Event)
				ctx.Event.GroupID = int64(ts.Grpid)
				if ts.Url == "" {
					ctx.SendChain(msgext.AtAll(), message.Text(ts.Alert))
				} else {
					ctx.SendChain(msgext.AtAll(), message.Text(ts.Alert), msgext.ImageNoCache(ts.Url))
				}
				return false
			})
		}
	}
}

func RegisterTimer(ts *TimeStamp, save bool) {
	key := GetTimerInfo(ts)
	t, ok := (*Timers)[key]
	if t != ts && ok { //避免重复注册定时器
		t.Enable = false
	}
	(*Timers)[key] = ts
	if save {
		SaveTimers()
	}
	fmt.Printf("[群管]注册计时器[%t]%s\n", ts.Enable, key)
	for ts.Enable {
		if ts.Month < 0 || ts.Month == int32(time.Now().Month()) {
			if ts.Day < 0 || ts.Day == int32(time.Now().Day()) {
				judgeHM(ts)
			} else if ts.Day == 0 {
				if ts.Week < 0 || ts.Week == int32(time.Now().Weekday()) {
					judgeHM(ts)
				}
			}
		}
		time.Sleep(time.Minute)
	}
}

func SaveTimers() error {
	data, err := timersmap.Marshal()
	if err != nil {
		return err
	} else if utils.PathExists(DATAPATH) {
		f, err1 := os.OpenFile(PBFILE, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err1 != nil {
			return err1
		} else {
			defer f.Close()
			_, err2 := f.Write(data)
			return err2
		}
	} else {
		return nil
	}
}

func loadTimers() {
	if utils.PathExists(PBFILE) {
		f, err := os.Open(PBFILE)
		if err == nil {
			data, err1 := io.ReadAll(f)
			if err1 == nil {
				if len(data) > 0 {
					timersmap.Unmarshal(data)
					for _, t := range timersmap.Timers {
						go RegisterTimer(t, false)
					}
					return
				}
			}
		}
	}
	timersmap.Timers = make(map[string]*Timer)
}

//获得标准化定时字符串
func GetTimerInfo(ts *TimeStamp) string {
	return fmt.Sprintf("%d月%d日%d周%d:%d", ts.Month, ts.Day, ts.Week, ts.Hour, ts.Minute)
}

//获得填充好的ts
func GetFilledTimeStamp(dateStrs []string, matchDateOnly bool) *TimeStamp {
	monthStr := []rune(dateStrs[1])
	dayWeekStr := []rune(dateStrs[2])
	hourStr := []rune(dateStrs[3])
	minuteStr := []rune(dateStrs[4])

	var ts TimeStamp
	ts.Month = chineseNum2Int(monthStr)
	if (ts.Month != -1 && ts.Month <= 0) || ts.Month > 12 { //月份非法
		fmt.Println("[群管]月份非法！")
		return &ts
	}
	lenOfDW := len(dayWeekStr)
	if lenOfDW == 4 { //包括末尾的"日"
		dayWeekStr = []rune{dayWeekStr[0], dayWeekStr[2]} //去除中间的十
		ts.Day = chineseNum2Int(dayWeekStr)
		if (ts.Day != -1 && ts.Day <= 0) || ts.Day > 31 { //日期非法
			fmt.Println("[群管]日期非法1！")
			return &ts
		}
	} else if dayWeekStr[lenOfDW-1] == rune('日') { //xx日
		dayWeekStr = dayWeekStr[:lenOfDW-1]
		ts.Day = chineseNum2Int(dayWeekStr)
		if (ts.Day != -1 && ts.Day <= 0) || ts.Day > 31 { //日期非法
			fmt.Println("[群管]日期非法2！")
			return &ts
		}
	} else if dayWeekStr[0] == rune('每') { //每周
		ts.Week = -1
	} else { //周x
		ts.Week = chineseNum2Int(dayWeekStr[1:])
		if ts.Week == 7 { //周天是0
			ts.Week = 0
		}
		if ts.Week < 0 || ts.Week > 6 { //星期非法
			ts.Week = -11
			fmt.Println("[群管]星期非法！")
			return &ts
		}
	}
	if len(hourStr) == 3 {
		hourStr = []rune{hourStr[0], hourStr[2]} //去除中间的十
	}
	ts.Hour = chineseNum2Int(hourStr)
	if ts.Hour < -1 || ts.Hour > 23 { //小时非法
		fmt.Println("[群管]小时非法！")
		return &ts
	}
	if len(minuteStr) == 3 {
		minuteStr = []rune{minuteStr[0], minuteStr[2]} //去除中间的十
	}
	ts.Minute = chineseNum2Int(minuteStr)
	if ts.Minute < -1 || ts.Minute > 59 { //分钟非法
		fmt.Println("[群管]分钟非法！")
		return &ts
	}
	if !matchDateOnly {
		urlStr := dateStrs[5]
		if urlStr != "" { //是图片url
			ts.Url = urlStr[3:] //utf-8下用为3字节
			fmt.Println("[群管]" + ts.Url)
			if !strings.HasPrefix(ts.Url, "http") {
				ts.Url = "illegal"
				fmt.Println("[群管]url非法！")
				return &ts
			}
		}
		ts.Alert = dateStrs[6]
		ts.Enable = true
	}
	return &ts
}

//汉字数字转int，仅支持-10～99，最多两位数，其中"每"解释为-1，"每两"为-2，以此类推
func chineseNum2Int(rs []rune) int32 {
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
	return int32(r)
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
