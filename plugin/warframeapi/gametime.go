package warframeapi

import (
	"fmt"
	"time"

	"github.com/davidscholberg/go-durationfmt"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	GameTimes []gameTime
)

func (t gameTime) getStatus() string {
	if t.Status {
		return t.StatusTrueDes
	} else {
		return t.StatusFalseDes
	}
}
func (t gameTime) getTime() string {
	d := time.Until(t.NextTime)
	durStr, _ := durationfmt.Format(d, "%m分%s秒后")
	return durStr
}

// 游戏时间模拟初始化
func gameTimeInit() {
	//updateWM()
	LoadTime()
	go gameRuntime()
}

func gameRuntime() {
	for {
		time.Sleep(10 * time.Second)
		timeDet()
	}

}

func LoadTime() {
	//updateWM()
	var isfass bool
	if wfapi.CambionCycle.Active == "fass" {
		isfass = false
	}
	GameTimes = []gameTime{
		{"地球平原", wfapi.CetusCycle.Expiry.Local(), wfapi.CetusCycle.IsDay, "白天", "夜晚", 100 * 60, 50 * 60},
		{"金星平原", wfapi.VallisCycle.Expiry.Local(), wfapi.VallisCycle.IsWarm, "温暖", "寒冷", 400, 20 * 60},
		{"火卫二平原", wfapi.CambionCycle.Expiry.Local(), isfass, "fass", "vome", 100 * 60, 50 * 60},
	}

}

func timeDet() {
	for i, v := range GameTimes {
		nt := time.Until(v.NextTime).Seconds()
		switch {
		case nt < 0:
			if v.Status {
				GameTimes[i].NextTime = v.NextTime.Add(time.Duration(v.NightTime) * time.Second)
			} else {
				GameTimes[i].NextTime = v.NextTime.Add(time.Duration(v.DayTime) * time.Second)
			}
			GameTimes[i].Status = !GameTimes[i].Status
			callUser(i, GameTimes[i].Status, 0)
		case nt < float64(5)*60:
			callUser(i, !GameTimes[i].Status, 5)
		case nt < float64(15)*60:
			if i == 2 && !v.Status {
				return
			}
			callUser(i, !GameTimes[i].Status, 15)
		}
	}
}

func callUser(i int, s bool, time int) {
	for group, sl := range sublist {
		msg := []message.MessageSegment{}

		switch {
		case !sl.Min15Tips && !sl.Min5Tips && time == 15:
			sublist[group].Min15Tips = true
		case sl.Min15Tips && !sl.Min5Tips && time == 5:
			sublist[group].Min5Tips = true
		case sl.Min15Tips && sl.Min5Tips && time == 0:
			sublist[group].Min15Tips = false
			sublist[group].Min5Tips = false
		default:
			return
		}

		//if !sl.Min15Tips && !sl.Min5Tips && time == 15 {
		//	sublist[group].Min15Tips = true
		//} else if sl.Min15Tips && !sl.Min5Tips && time == 5 {
		//	sublist[group].Min5Tips = true
		//} else if sl.Min15Tips && sl.Min5Tips && time == 0 {
		//	sublist[group].Min15Tips = false
		//	sublist[group].Min5Tips = false
		//} else {
		//	return
		//}

		for qq, st := range sl.SubUser {
			if st.SubType[i] != nil {
				if *st.SubType[i] == s {
					msg = append(msg, message.At(qq))
				}
			}
		}
		if len(msg) == 0 {
			continue
		}
		if time <= 0 {
			if s {
				msg = append(msg, message.Text(fmt.Sprintf("\n%s白天(%s)到了", GameTimes[i].Name, GameTimes[i].StatusTrueDes)))
			} else {
				msg = append(msg, message.Text(fmt.Sprintf("\n%s夜晚(%s)到了", GameTimes[i].Name, GameTimes[i].StatusFalseDes)))
			}
		} else {
			if s {
				msg = append(msg, message.Text(fmt.Sprintf("\n%s距离白天(%s)还剩下%d分钟", GameTimes[i].Name, GameTimes[i].StatusTrueDes, time)))
			} else {
				msg = append(msg, message.Text(fmt.Sprintf("\n%s距离夜晚(%s)还剩下%d分钟", GameTimes[i].Name, GameTimes[i].StatusFalseDes, time)))
			}
		}

		zero.GetBot(2429160662).SendGroupMessage(group, msg)
	}

}

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
