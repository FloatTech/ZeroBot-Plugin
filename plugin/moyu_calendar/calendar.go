// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"sync"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/binary"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"
)

const (
	api     = "https://api.vvhan.com/api/moyu?type=json"
	referer = "api.vvhan.com"
	ua      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.41 Safari/537.36 Edg/101.0.1210.32"
)

var (
	mu      sync.Mutex
	pictime time.Time
	picurl  string
)

func init() { // 插件主体
	engine := control.Register("moyucalendar", &control.Options{
		DisableOnDefault: false,
		Help: "摸鱼人日历\n" +
			"- /启用 moyucalendar\n" +
			"- /禁用 moyucalendar\n" +
			"- 记录在\"30 8 * * *\"触发的指令\n" +
			"   - 摸鱼人日历",
	})
	engine.OnFullMatch("摸鱼人日历", zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			err := getdata()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Image(picurl).Add("cache", 0))
		})
}

func getdata() error { // 获取图片链接并且下载
	mu.Lock()
	defer mu.Unlock()
	if picurl != "" && time.Since(pictime) <= time.Hour*20 {
		return nil
	}
	data, err := web.RequestDataWith(web.NewDefaultClient(), api, "GET", "", ua)
	if err != nil {
		return err
	}
	picurl = gjson.Get(binary.BytesToString(data), "url").String()
	pictime = time.Now()
	return nil
}
