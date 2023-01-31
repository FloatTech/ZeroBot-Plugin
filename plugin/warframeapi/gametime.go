package warframeapi

import (
	"sync"
	"time"

	"github.com/davidscholberg/go-durationfmt"
)

// 游戏时间模拟
type gameTime struct {
	rwm            sync.RWMutex
	Name           string    `json:"name"`      // 时间名称
	NextTime       time.Time `json:"time"`      // 下次更新时间
	Status         bool      `json:"status"`    // 状态
	StatusTrueDes  string    `json:"true_des"`  // 状态说明
	StatusFalseDes string    `json:"false_des"` // 状态说明
	DayTime        int       `json:"day"`       // 白天时长
	NightTime      int       `json:"night"`     // 夜间时长
}

var (
	gameTimes [3]*gameTime
)

// TimeString 根据传入的世界编号，获取对应的游戏时间文本
func (t *gameTime) String() string {
	return "平原时间:" + t.daynight() + "\n" +
		"下次更新:" + t.remaintime()
}

// 获取当前游戏时间状态（白天/夜晚）
func (t *gameTime) daynight() string {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	if t.Status {
		return t.StatusTrueDes
	}
	return t.StatusFalseDes
}

// 获取下一次时间状态更新的剩余游戏时间（x分x秒）
func (t *gameTime) remaintime() string {
	t.rwm.RLock()
	d := time.Until(t.NextTime)
	t.rwm.RUnlock()
	durStr, _ := durationfmt.Format(d, "%m分%s秒后")
	return durStr
}

// 根据API返回内容修正游戏时间
func loadTime(api wfAPI) {
	gameTimes = [3]*gameTime{
		{Name: "地球平原", NextTime: api.CetusCycle.Expiry.Local(), Status: api.CetusCycle.IsDay, StatusTrueDes: "白天", StatusFalseDes: "夜晚", DayTime: 100 * 60, NightTime: 50 * 60},
		{Name: "金星平原", NextTime: api.VallisCycle.Expiry.Local(), Status: api.VallisCycle.IsWarm, StatusTrueDes: "温暖", StatusFalseDes: "寒冷", DayTime: 400, NightTime: 20 * 60},
		{Name: "火卫二平原", NextTime: api.CambionCycle.Expiry.Local(), Status: api.CambionCycle.Active == "fass", StatusTrueDes: "fass", StatusFalseDes: "vome", DayTime: 100 * 60, NightTime: 50 * 60},
	}
}

// timeDet游戏时间更新
func timeDet() {
	for _, v := range gameTimes {
		// 当前时间对比下一次游戏状态更新时间，看看还剩多少秒
		nt := time.Until(v.NextTime).Seconds()
		// 已经过了游戏时间状态更新时间
		if nt < 0 {
			v.rwm.Lock()
			// 更新游戏状态，如果是白天就切换到晚上，反之亦然
			if v.Status {
				// 计算下次的晚上更新时间
				v.NextTime = v.NextTime.Add(time.Duration(v.NightTime) * time.Second)
			} else {
				// 计算下次的白天更新时间
				v.NextTime = v.NextTime.Add(time.Duration(v.DayTime) * time.Second)
			}
			v.rwm.Unlock()
		}
	}
}
