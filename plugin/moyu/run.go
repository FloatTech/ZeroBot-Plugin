// Package moyu 摸鱼
package moyu

import (
	"sync"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	msg        message.Message
	mu         sync.Mutex
	lastupdate time.Time
)

func init() { // 插件主体
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: true,
		Brief:            "摸鱼提醒",
		Help: "- /启用 moyu\n" +
			"- /禁用 moyu\n" +
			"- 记录在\"0 10 * * *\"触发的指令\n" +
			"   - 摸鱼提醒",
	}).OnFullMatch("摸鱼提醒").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			mu.Lock()
			defer mu.Unlock()
			if msg == nil || time.Since(lastupdate) > time.Hour*20 {
				if registry.Connect() != nil {
					return
				}
				msg = message.Message{
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
				lastupdate = time.Now()
			}
			ctx.Send(msg)
		})
}
