// Package imagefinder 关键字搜图
package imagefinder

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/url"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/pixiv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/pool"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/web"
)

type resultjson struct {
	Data struct {
		Illusts []struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			AltTitle    string `json:"altTitle"`
			Description string `json:"description"`
			Sanity      int    `json:"sanity"`
		} `json:"illusts"`
	} `json:"data"`
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func init() {
	control.Register("imgfinder", &control.Options{
		DisableOnDefault: false,
		Help: "关键字搜图\n" +
			"- 来张 [xxx]",
	}).OnRegex(`^来张\s?(.*)$`, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			soutujson, err := soutuapi(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			rannum := rand.Intn(len(soutujson.Data.Illusts))
			illust, err := pixiv.Works(soutujson.Data.Illusts[rannum].ID)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			u := illust.ImageUrls[0]
			n := u[strings.LastIndex(u, "/")+1 : len(u)-4]
			f := illust.Path(0)

			err = pool.SendImageFromPool(n, f, func() error {
				// 下载图片
				return illust.DownloadToCache(0)
			}, ctxext.SendFakeForwardToGroup(ctx), ctxext.GetFirstMessageInForward(ctx))
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
		})
}

// soutuapi 请求api
func soutuapi(keyword string) (r resultjson, err error) {
	var data []byte
	for i := 0; i < 3; i++ {
		data, err = web.GetData("https://copymanga.azurewebsites.net/api/pixivel?" + url.QueryEscape(keyword) + "?page=0")
		if err != nil {
			process.SleepAbout1sTo2s()
			continue
		}
		err = json.Unmarshal(data, &r)
		if err == nil && r.Error {
			err = errors.New(r.Message)
		}
		return
	}
	return
}
