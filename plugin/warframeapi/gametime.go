package warframeapi

import (
	"sync"
	"time"

	"github.com/davidscholberg/go-durationfmt"
)

var (
	gameTimes [3]*gameTime
)

// TimeString 根据传入的世界编号，获取对应的游戏时间文本
func (t *gameTime) String() string {
	return "平原时间:" + t.daynight() + "\n" +
		"下次更新:" + t.remaintime()
}

// daynight 获取当前游戏时间状态（白天/夜晚）
func (t *gameTime) daynight() string {
	t.rwm.RLock()
	defer t.rwm.RUnlock()
	if t.Status {
		return t.StatusTrueDes
	}
	return t.StatusFalseDes
}

// remaintime 获取下一次时间状态更新的剩余游戏时间（x分x秒）
func (t *gameTime) remaintime() string {
	t.rwm.RLock()
	d := time.Until(t.NextTime)
	t.rwm.RUnlock()
	durStr, _ := durationfmt.Format(d, "%m分%s秒后")
	return durStr
}

// 游戏时间模拟初始化
//func gameTimeInit() {
//	//updateWM()
//	loadTime(wfapi)
//	go gameRuntime()
//}

// 游戏时间模拟
//func gameRuntime() {
//	wfapi, err := getWFAPI()
//	if err != nil {
//		println("ERROR:GetWFAPI失败,", err.Error())
//		return
//	}
//	loadTime(wfapi)
//	for range time.NewTicker(10 * time.Second).C {
//		timeDet()
//	}
//}

// loadTime 根据API返回内容修正游戏时间
func loadTime(api wfAPI) {
	//updateWM()
	isfass := api.CambionCycle.Active == "fass"
	gameTimes = [3]*gameTime{
		{Name: "地球平原", NextTime: api.CetusCycle.Expiry.Local(), Status: api.CetusCycle.IsDay, StatusTrueDes: "白天", StatusFalseDes: "夜晚", DayTime: 100 * 60, NightTime: 50 * 60},
		{Name: "金星平原", NextTime: api.VallisCycle.Expiry.Local(), Status: api.VallisCycle.IsWarm, StatusTrueDes: "温暖", StatusFalseDes: "寒冷", DayTime: 400, NightTime: 20 * 60},
		{Name: "火卫二平原", NextTime: api.CambionCycle.Expiry.Local(), Status: isfass, StatusTrueDes: "fass", StatusFalseDes: "vome", DayTime: 100 * 60, NightTime: 50 * 60},
	}
}

// timeDet游戏时间更新
func timeDet() {
	for _, v := range gameTimes {
		//当前时间对比下一次游戏状态更新时间，看看还剩多少秒
		nt := time.Until(v.NextTime).Seconds()
		//已经过了游戏时间状态更新时间
		if nt < 0 {
			v.rwm.Lock()
			//更新游戏状态，如果是白天就切换到晚上，反之亦然
			if v.Status {
				//计算下次的晚上更新时间
				v.NextTime = v.NextTime.Add(time.Duration(v.NightTime) * time.Second)
			} else {
				//计算下次的白天更新时间
				v.NextTime = v.NextTime.Add(time.Duration(v.DayTime) * time.Second)
			}
			v.rwm.Unlock()
		}
		//暂时只保留时间更新功能
		//switch {
		//case nt < 0:
		//	if v.Status {
		//		v.NextTime = v.NextTime.Add(time.Duration(v.NightTime) * time.Second)
		//	} else {
		//		v.NextTime = v.NextTime.Add(time.Duration(v.DayTime) * time.Second)
		//	}
		//	v.Status = !v.Status
		//
		//	callUser(i, v.Status, 0)
		//case nt < float64(5)*60:
		//	callUser(i, !v.Status, 5)
		//case nt < float64(15)*60:
		//	if i == 2 && !v.Status {
		//		return
		//	}
		//	callUser(i, !v.Status, 15)
		//}
	}
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
	rwm            sync.RWMutex
	Name           string    `json:"name"`      //时间名称
	NextTime       time.Time `json:"time"`      //下次更新时间
	Status         bool      `json:"status"`    //状态
	StatusTrueDes  string    `json:"true_des"`  //状态说明
	StatusFalseDes string    `json:"false_des"` //状态说明
	DayTime        int       `json:"day"`       //白天时长
	NightTime      int       `json:"night"`     //夜间时长
}

//type subList struct {
//	SubUser   map[int64]subType `json:"qq_sub"`
//	Min5Tips  bool              `json:"min5_tips"`
//	Min15Tips bool              `json:"min15_tips"`
//}

//type subType struct {
//	SubType map[int]*bool `json:"sub_type"`
//	SubRaid bool          `json:"sub_raid"`
//}
