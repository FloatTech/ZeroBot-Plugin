// Package zaobao 易即今日公众号api的今日早报
package zaobao

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	api = "http://api.soyiji.com/news_jpg"
)

func init() { // 插件主体
	engine := control.Register("zaobao", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "zaobao\n" +
			"配合插件job中的记录在'cron'触发的指令使用\n" +
			"------示例------\n" +
			"|每天早上九点定时发送\n" +
			"记录在'00 9 * * *'触发的指令\n" +
			"今日早报",
		PrivateDataFolder: "zaobao",
	})
	os.RemoveAll(engine.DataFolder())
	engine.OnFullMatch("今日早报", zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			if !file.IsExist(file.BOTPATH + engine.DataFolder() + "/zaobao_" + time.Now().Format("2006-01-02") + ".jpg") {
				download(ctx)
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + engine.DataFolder() + "/zaobao_" + time.Now().Format("2006-01-02") + ".jpg"))
			return
		})
}

func download(ctx *zero.Ctx) { // 获取图片链接并且下载
	var engine control.Engine
	client := http.Client{}
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		ctx.SendChain(message.Text("ERROR:", err))
		return
	}
	res, err := client.Do(req)
	if err != nil {
		ctx.SendChain(message.Text("ERROR:", err))
		return
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendChain(message.Text("ERROR:", err))
		return
	}
	defer res.Body.Close()
	url := gjson.ParseBytes(data).Get("url").String()
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		ctx.SendChain(message.Text("ERROR:", err))
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36 Edg/87.0.664.66")
	req.Header.Set("Referer", "safe.soyiji.com")
	res, err = client.Do(req)
	if err != nil {
		ctx.SendChain(message.Text("ERROR:", err))
		return
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		ctx.SendChain(message.Text("ERROR:", err))
		return
	}
	defer res.Body.Close()
	{
		file, err := os.Create(file.BOTPATH + engine.DataFolder() + "/zaobao_" + time.Now().Format("2006-01-02") + ".jpg")
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		_, err = file.Write(data)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		file.Close()
	}
}
