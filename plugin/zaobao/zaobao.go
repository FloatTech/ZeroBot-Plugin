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
	"github.com/FloatTech/zbputils/process"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	api = "http://api.soyiji.com/news_jpg"
)

func init() { // 插件主体
	// 定时任务每天9点执行一次
	// api早上8点更新，推荐定时在8点30后
	_, err := process.CronTab.AddFunc("00 09 * * *", func() { sendzaobao() })
	if err != nil {
		panic(err)
	}

	engine := control.Register("zaobao", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "zaobao\n" +
			"- /启用 zaobao\n" +
			"- /禁用 zaobao",
		PrivateDataFolder: "zaobao",
	})
	engine.OnFullMatch("今日早报", zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			err := os.MkdirAll("data/zaobao", 0755)
			if err != nil {
				panic(err)
			}
			if !FileExist(file.BOTPATH + "/data/zaobao/zaobao_" + time.Now().Format("2006-01-02") + ".jpg") {
				download(ctx)
			}
			ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/data/zaobao/zaobao_" + time.Now().Format("2006-01-02") + ".jpg"))
		})
	engine.OnFullMatch("群发今日早报", zero.OnlyGroup).SetBlock(false).
		Handle(func(ctx *zero.Ctx) {
			sendzaobao()
		})
}

func download(ctx *zero.Ctx) { // 获取图片链接并且下载
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
		file, err := os.Create(file.BOTPATH + "/data/zaobao/zaobao_" + time.Now().Format("2006-01-02") + ".jpg")
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

func sendzaobao() { // 发送
	m, ok := control.Lookup("zaobao")
	if ok {
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				groupid := g.Get("group_id").Int()
				if m.IsEnabledIn(groupid) {
					if !FileExist(file.BOTPATH + "/data/zaobao/zaobao_" + time.Now().Format("2006-01-02") + ".jpg") {
						download(ctx)
					}
					ctx.SendGroupMessage(groupid, message.Image("file:///"+file.BOTPATH+"/data/zaobao/zaobao_"+time.Now().Format("2006-01-02")+".jpg"))
				}
			}
			return true
		})
	}
}

// FileExist 判断文件是否存在
func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
