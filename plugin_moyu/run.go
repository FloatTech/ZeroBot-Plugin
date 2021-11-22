// Package acgimage 随机图片与AI点评
package moyu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/FloatTech/ZeroBot-Plugin/control"
	"github.com/robfig/cron"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var Group = []int64{639865824}

//开启的群
type Kq struct {
	GroupID []int64 `json:"群号"` //群号
}

//默认
var MY = Kq{
	GroupID: []int64{111},
}

func init() { // 插件主体
	FansDaily() // 摸鱼提醒
	engine := control.Register("moyu", &control.Options{
		DisableOnDefault: false,
		Help: "moyu摸鱼提醒\n" +
			"- 删除提醒\n" +
			"- 添加提醒\n",
	})

	//登录配置
	os.MkdirAll("config/", 0777)
	_, err := os.Stat(`config/moyu.json`)
	if err == nil {
		f, _ := os.Open("config/moyu.json")
		defer f.Close()
		err := json.NewDecoder(f).Decode(&MY)
		if err != nil {
			fmt.Print("config.json格式错误! 请检查!\n")
			time.Sleep(5 * time.Second)
			return
		}
	} else {
		fp, _ := os.Create("config/moyu.json")
		defer fp.Close()
		data, _ := json.Marshal(MY)
		var out bytes.Buffer
		json.Indent(&out, data, "", "\t")
		out.WriteTo(fp)
		time.Sleep(5 * time.Second)
	}
	engine.OnRegex(`^(添加|删除)提醒$`).
		SetBlock(true).SetPriority(20).Handle(func(ctx *zero.Ctx) {
		fp, _ := os.Create("config/moyu.json")
		defer fp.Close()
		if ctx.State["regex_matched"].([]string)[1] == "添加" {
			MY.GroupID = append(MY.GroupID, ctx.Event.GroupID)
			ctx.Send(message.Text("添加成功！"))
		} else {
			for i, v := range MY.GroupID {
				if v == ctx.Event.GroupID {
					MY.GroupID = append(MY.GroupID[:i], MY.GroupID[i+1:]...)
					break
				}

			}
			ctx.Send(message.Text("删除成功！"))
		}
		data, _ := json.Marshal(MY)
		var out bytes.Buffer
		json.Indent(&out, data, "", "\t")
		out.WriteTo(fp)
	})

}

// 定时任务每天晚上最后2分钟执行一次
func FansDaily() {
	c := cron.New()
	_ = c.AddFunc("0 0 10 * * ?", func() { fansData() })
	c.Start()
}

// 获取数据拼接消息链并发送
func fansData() {
	for _, v := range MY.GroupID {
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			ctx.SendGroupMessage(
				v,
				[]message.MessageSegment{
					message.Text(time.Now().Format("2006-01-02") +
						"上午好，摸鱼人！\n工作再累，一定不要忘记摸鱼哦！有事没事起身去茶水间，去厕所，去廊道走走别老在工位上坐着，钱是老板的,但命是自己的。" +
						Zm() + "\n" +
						Moyu("元旦", 2022, 1, 1) + "\n" +
						Moyu("春节", 2022, 1, 31) + "\n" +
						Moyu("清明节", 2022, 4, 3) + "\n" +
						Moyu("劳动节", 2022, 4, 30) + "\n" +
						Moyu("端午节", 2022, 6, 3) + "\n" +
						Moyu("中秋节", 2022, 9, 10) + "\n" +
						Moyu("国庆节", 2022, 10, 1) + "\n" +
						"\n\n上班是帮老板赚钱，摸鱼是赚老板的钱！最后，祝愿天下所有摸鱼人，都能愉快的渡过每一天…",
					),
				},
			)
			return true
		})

	}
}

// 获取两个时间相差的天数，0表同一天，正数表t1>t2，负数表t1<t2
func Moyu(text string, year int, month time.Month, day int) string {
	currentTime := time.Now()
	t1 := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	t2 := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.Local)
	tt := int(t1.Sub(t2).Hours() / 24)
	if tt >= 0 {
		return "距离" + text + "还有: " + strconv.Itoa(tt) + " 天!"
	} else {
		return "好好享受 " + text + " 假期吧!"
	}
}

func Zm() string {
	t := time.Now().Weekday().String()
	switch {
	case t == "Sunday":
		return "\n好好享受周末吧！"
	case t == "Monday":
		return "\n距离周末还有:4天！"
	case t == "Tuesday":
		return "\n距离周末还有:3天！"
	case t == "Wednesday":
		return "\n距离周末还有:2天！"
	case t == "Thursday":
		return "\n距离周末还有:1天！"
	case t == "Friday":
		return "\n距离周末还有:0天！"
	default:
		return "\n好好享受周末吧！"
	}
}
