package warframeapi

import (
	"sync"
	"time"

	"github.com/davidscholberg/go-durationfmt"
)

var (
	gameTimes [3]*gameTime
	rwm       sync.RWMutex
)

func (t *gameTime) getStatus() string {
	rwm.RLock()
	defer rwm.RUnlock()
	if t.Status {
		return t.StatusTrueDes
	}
	return t.StatusFalseDes
}
func (t *gameTime) getTime() string {
	rwm.RLock()
	d := time.Until(t.NextTime)
	rwm.RUnlock()
	durStr, _ := durationfmt.Format(d, "%m分%s秒后")
	return durStr
}

// 游戏时间模拟初始化
//func gameTimeInit() {
//	//updateWM()
//	loadTime(wfapi)
//	go gameRuntime()
//}

func gameRuntime() {
	for range time.NewTicker(10 * time.Second).C {
		timeDet()
	}
}

func loadTime(api wfAPI) {
	//updateWM()
	var isfass bool
	if api.CambionCycle.Active == "fass" {
		isfass = true
	}
	gameTimes = [3]*gameTime{
		{"地球平原", api.CetusCycle.Expiry.Local(), api.CetusCycle.IsDay, "白天", "夜晚", 100 * 60, 50 * 60},
		{"金星平原", api.VallisCycle.Expiry.Local(), api.VallisCycle.IsWarm, "温暖", "寒冷", 400, 20 * 60},
		{"火卫二平原", api.CambionCycle.Expiry.Local(), isfass, "fass", "vome", 100 * 60, 50 * 60},
	}

}

func timeDet() {
	rwm.Lock()

	for _, v := range gameTimes {
		nt := time.Until(v.NextTime).Seconds()
		//暂时只保留时间更新功能
		switch {
		case nt < 0:
			if v.Status {
				v.NextTime = v.NextTime.Add(time.Duration(v.NightTime) * time.Second)
			} else {
				v.NextTime = v.NextTime.Add(time.Duration(v.DayTime) * time.Second)
			}
			v.Status = !v.Status
			//
			//	callUser(i, v.Status, 0)
			//case nt < float64(5)*60:
			//	callUser(i, !v.Status, 5)
			//case nt < float64(15)*60:
			//	if i == 2 && !v.Status {
			//		return
			//	}
			//	callUser(i, !v.Status, 15)
		}
	}
	defer rwm.Unlock()
}

//TODO:订阅功能-待重做
//func callUser(i int, s bool, time int) []message.MessageSegment {
//	msg := []message.MessageSegment{}
//	for group, sl := range sublist {
//
//		switch {
//		case !sl.Min15Tips && !sl.Min5Tips && time == 15: //是否
//			sublist[group].Min15Tips = true
//		case sl.Min15Tips && !sl.Min5Tips && time == 5:
//			sublist[group].Min5Tips = true
//		case sl.Min15Tips && sl.Min5Tips && time == 0:
//			sublist[group].Min15Tips = false
//			sublist[group].Min5Tips = false
//		default:
//			return nil
//		}
//		//if !sl.Min15Tips && !sl.Min5Tips && time == 15 {
//		//	sublist[group].Min15Tips = true
//		//} else if sl.Min15Tips && !sl.Min5Tips && time == 5 {
//		//	sublist[group].Min5Tips = true
//		//} else if sl.Min15Tips && sl.Min5Tips && time == 0 {
//		//	sublist[group].Min15Tips = false
//		//	sublist[group].Min5Tips = false
//		//} else {
//		//	return
//		//}
//		for qq, st := range sl.SubUser {
//			if st.SubType[i] != nil {
//				if *st.SubType[i] == s {
//					msg = append(msg, message.At(qq))
//				}
//			}
//		}
//		if len(msg) == 0 {
//			continue
//		}
//		if time <= 0 {
//			if s {
//				msg = append(msg, message.Text(fmt.Sprintf("\n%s白天(%s)到了", gameTimes[i].Name, gameTimes[i].StatusTrueDes)))
//			} else {
//				msg = append(msg, message.Text(fmt.Sprintf("\n%s夜晚(%s)到了", gameTimes[i].Name, gameTimes[i].StatusFalseDes)))
//			}
//		} else {
//			if s {
//				msg = append(msg, message.Text(fmt.Sprintf("\n%s距离白天(%s)还剩下%d分钟", gameTimes[i].Name, gameTimes[i].StatusTrueDes, time)))
//			} else {
//				msg = append(msg, message.Text(fmt.Sprintf("\n%s距离夜晚(%s)还剩下%d分钟", gameTimes[i].Name, gameTimes[i].StatusFalseDes, time)))
//			}
//		}
//	}
//	return msg
//}

// 游戏时间模拟
type gameTime struct {
	Name           string    `json:"name"`      //时间名称
	NextTime       time.Time `json:"time"`      //下次更新时间
	Status         bool      `json:"status"`    //状态
	StatusTrueDes  string    `json:"true_des"`  //状态说明
	StatusFalseDes string    `json:"false_des"` //状态说明
	DayTime        int       `json:"day"`       //白天时长
	NightTime      int       `json:"night"`     //夜间时长
}

type subList struct {
	SubUser   map[int64]subType `json:"qq_sub"`
	Min5Tips  bool              `json:"min5_tips"`
	Min15Tips bool              `json:"min15_tips"`
}

type subType struct {
	SubType map[int]*bool `json:"sub_type"`
	SubRaid bool          `json:"sub_raid"`
}
