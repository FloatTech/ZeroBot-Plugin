// Package moyu 摸鱼
package moyu

import (
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/fumiama/cron"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

func init() { // 插件主体
	// 定时任务每天10点执行一次
	c := cron.New()
	_, err := c.AddFunc("0 10 * * *", func() { sendNotice() })
	if err == nil {
		c.Start()
	}

	control.Register("moyu", order.PrioMoyu, &control.Options{
		DisableOnDefault: true,
		Help: "moyu\n" +
			"- 添加摸鱼提醒\n" +
			"- 删除摸鱼提醒",
	}).OnFullMatch("删除摸鱼提醒", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m, ok := control.Lookup("moyu")
			if ok {
				if m.IsEnabledIn(ctx.Event.GroupID) {
					m.Disable(ctx.Event.GroupID)
					ctx.Send(message.Text("删除成功！"))
				} else {
					ctx.Send(message.Text("未启用！"))
				}
			} else {
				ctx.Send(message.Text("找不到该服务！"))
			}
		})

	zero.OnFullMatch("添加摸鱼提醒", zero.OnlyGroup, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			m, ok := control.Lookup("moyu")
			if ok {
				if m.IsEnabledIn(ctx.Event.GroupID) {
					ctx.Send(message.Text("已启用！"))
				} else {
					m.Enable(ctx.Event.GroupID)
					ctx.Send(message.Text("添加成功！"))
				}
			} else {
				ctx.Send(message.Text("找不到该服务！"))
			}
		})
}

// 获取数据拼接消息链并发送
func sendNotice() {
	m, ok := control.Lookup("moyu")
	if ok {
		if holidaymap == nil {
			ok = false
			if registry.Connect() != nil {
				return
			}
			holidaymap = make(map[string]*Holiday, 32)
		}
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				grp := g.Get("group_id").Int()
				if m.IsEnabledIn(grp) {
					ctx.SendGroupMessage(grp,
						[]message.MessageSegment{
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
						},
					)
				}
			}
			return true
		})
		if !ok {
			_ = registry.Close()
		}
	}
}
