package warframeapi

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davidscholberg/go-durationfmt"
)

// 游戏时间模拟
type timezone struct {
	sync.RWMutex `json:"-"`
	Name         string    `json:"name"`      // 时间名称
	NextTime     time.Time `json:"time"`      // 下次更新时间
	IsDay        bool      `json:"status"`    // 状态
	DayDesc      string    `json:"true_des"`  // 状态说明
	NightDesc    string    `json:"false_des"` // 状态说明
	DayLen       int       `json:"day"`       // 白天时长
	NightLen     int       `json:"night"`     // 夜间时长
}

type world struct {
	w       [3]*timezone
	hassync uintptr
}

var gameWorld = newworld()

// String 根据传入的世界编号，获取对应的游戏时间文本
func (t *timezone) String() string {
	t.RLock()
	defer t.RUnlock()
	sb := strings.Builder{}
	sb.WriteString("平原时间: ")
	if t.IsDay {
		sb.WriteString(t.DayDesc)
	} else {
		sb.WriteString(t.NightDesc)
	}
	sb.WriteString(", ")
	sb.WriteString("下次更新: ")
	d := time.Until(t.NextTime)
	durStr, _ := durationfmt.Format(d, "%m分%s秒后")
	sb.WriteString(durStr)
	return sb.String()
}

func newworld() (w world) {
	w.w = [3]*timezone{
		{Name: "地球平原", DayDesc: "白天", NightDesc: "夜晚", DayLen: 100 * 60, NightLen: 50 * 60},
		{Name: "金星平原", DayDesc: "温暖", NightDesc: "寒冷", DayLen: 400, NightLen: 20 * 60},
		{Name: "火卫二平原", DayDesc: "fass", NightDesc: "vome", DayLen: 100 * 60, NightLen: 50 * 60},
	}
	return
}

func (w *world) hasSync() bool {
	return atomic.LoadUintptr(&w.hassync) != 0
}

func (w *world) setsync() bool {
	return atomic.CompareAndSwapUintptr(&w.hassync, 0, 1)
}

func (w *world) resetsync() bool {
	return atomic.CompareAndSwapUintptr(&w.hassync, 1, 0)
}

// 根据API返回内容修正游戏时间
func (w *world) refresh(api *wfapi) {
	for _, t := range w.w {
		t.Lock()
	}
	w.w[0].NextTime = api.CetusCycle.Expiry.Local()
	w.w[0].IsDay = api.CetusCycle.IsDay

	w.w[1].NextTime = api.VallisCycle.Expiry.Local()
	w.w[1].IsDay = api.VallisCycle.IsWarm

	w.w[2].NextTime = api.CambionCycle.Expiry.Local()
	w.w[2].IsDay = api.CambionCycle.Active == "fass"
	for _, t := range w.w {
		t.Unlock()
	}
}

// 游戏时间更新
func (w *world) update() {
	if !w.hasSync() {
		return
	}
	for _, t := range w.w {
		t.Lock()
		// 当前时间对比下一次游戏状态更新时间，看看还剩多少秒
		nt := time.Until(t.NextTime).Seconds()
		// 已经过了游戏时间状态更新时间
		if nt < 0 {
			// 更新游戏状态，如果是白天就切换到晚上，反之亦然
			if t.IsDay {
				// 计算下次的晚上更新时间
				t.NextTime = t.NextTime.Add(time.Duration(t.NightLen) * time.Second)
			} else {
				// 计算下次的白天更新时间
				t.NextTime = t.NextTime.Add(time.Duration(t.DayLen) * time.Second)
			}
		}
		t.Unlock()
	}
}
