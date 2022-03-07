// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"


	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.Register("moyucalendar", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "moyu_calendar\n" +
			"- 日历]",
	})
	engine.OnFullMatch("日历").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			a := rili()
			ctx.SendChain(message.Image("base64://" + helper.BytesToString(a)))
		})
	_, err := process.CronTab.AddFunc("30 8 * * *", func() {
		m, ok := control.Lookup("moyucalendar")
		if !ok {
			return
		}
		a := rili()
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				grp := g.Get("group_id").Int()
				if m.IsEnabledIn(grp) {
					ctx.SendGroupMessage(grp, message.Message{message.Image("base64://" + helper.BytesToString(a))})
					process.SleepAbout1sTo2s()
				}
			}
			return true
		})
	})
	if err != nil {
		panic(err)
	}
}

func rili() []byte {
	url := "https://api.vvhan.com/api/moyu"
	res, _ := http.Get(url)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	c := string(body)
	rese := Base64Encoding(c)
	a := []byte(rese)
	return a
}

func Base64Encoding(str string) string { //Base64编码
	src := []byte(str)
	ress := base64.StdEncoding.EncodeToString(src) //将编码变成字符串
	return ress
}
