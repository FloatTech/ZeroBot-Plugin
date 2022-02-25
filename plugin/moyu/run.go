// Package moyu 摸鱼
package moyu

import (
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

func init() { // 插件主体
	control.Register("moyu", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "moyu\n" +
			"- /启用 moyu\n" +
			"- /禁用 moyu",
	})

	// 定时任务每天10点执行一次
	_, err := process.CronTab.AddFunc("0 10 * * *", func() { sendNotice() })
	if err != nil {
		panic(err)
	}
}

// 获取数据拼接消息链并发送
func sendNotice() {
	m, ok := control.Lookup("moyu")
	if ok {
		if registry.Connect() != nil {
			return
		}
		msg := message.Message{
			message.Text(time.Now().Format("2006-01-02")),
			message.Text("上午好，摸鱼人！\n工作再累，一定不要忘记摸鱼哦！有事没事起身去茶水间，去厕所，去廊道走走别老在工位上坐着，钱是老板的,但命是自己的。\n"),
			message.Text(weekend()),
			message.Text("\n"),
			message.Text(GetHoliday("元旦")),
			message.Text("\n"),
			message.Text(GetHoliday("春节")),
			message.Text("\n"),
			message.Text(GetHoliday("清明节")),
			message.Text("\n"),
			message.Text(GetHoliday("劳动节")),
			message.Text("\n"),
			message.Text(GetHoliday("端午节")),
			message.Text("\n"),
			message.Text(GetHoliday("中秋节")),
			message.Text("\n"),
			message.Text(GetHoliday("国庆节")),
			message.Text("\n"),
			message.Text("上班是帮老板赚钱，摸鱼是赚老板的钱！最后，祝愿天下所有摸鱼人，都能愉快的渡过每一天…"),
		}
		_ = registry.Close()
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				grp := g.Get("group_id").Int()
				if m.IsEnabledIn(grp) {
					ctx.SendGroupMessage(grp, msg)
				}
			}
			return true
		})
	}
}
