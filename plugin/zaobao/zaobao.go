// Package zaobao 易即今日公众号api的今日早报
package zaobao

import (
	"sync"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/binary"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/web"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	api     = "http://api.soyiji.com/news_jpg"
	referer = "safe.soyiji.com"
	ua      = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66"
)

var (
	picdata []byte
	mu      sync.RWMutex
	pictime time.Time
)

func init() { // 插件主体
	engine := control.Register("zaobao", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "zaobao\n" +
			"api早上8点更新，推荐定时在8点30后\n" +
			"配合插件job中的记录在'cron'触发的指令使用\n" +
			"------示例------\n" +
			"每天早上九点定时发送\n" +
			"记录在'00 9 * * *'触发的指令\n" +
			"今日早报",
	})
	engine.OnFullMatch("今日早报", zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			err := getdata()
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.ImageBytes(picdata))
		})
}

func getdata() error { // 获取图片链接并且下载
	mu.RLock()
	if time.Since(pictime) > time.Hour*20 {
		mu.RUnlock()
		mu.Lock()
		picdata = nil
		pictime = time.Now()
		mu.Unlock()
		mu.RLock()
	}
	if picdata != nil {
		mu.RUnlock()
		return nil
	}
	mu.RUnlock()
	mu.Lock()
	defer mu.Unlock()
	if picdata != nil {
		return nil
	}
	data, err := web.GetDataWith(web.NewDefaultClient(), api, "GET", "", ua)
	if err != nil {
		return err
	}
	picdata, err = web.GetDataWith(web.NewDefaultClient(), gjson.Get(binary.BytesToString(data), "url").String(), "GET", referer, ua)
	if err != nil {
		return err
	}
	return nil
}
