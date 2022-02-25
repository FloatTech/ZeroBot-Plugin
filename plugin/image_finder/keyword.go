// Package imagefinder 关键字搜图
package imagefinder

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/pixiv"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/pool"
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
	control.Register("imgfinder", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "关键字搜图\n" +
			"- 来张 [xxx]",
	}).OnRegex(`^来张\s?(.*)$`, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			soutujson, err := soutuapi(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			rannum := rand.Intn(len(soutujson.Data.Illusts))
			illust, err := pixiv.Works(soutujson.Data.Illusts[rannum].ID)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			u := illust.ImageUrls[0]
			n := u[strings.LastIndex(u, "/")+1 : len(u)-4]
			f := illust.Path(0)

			err = pool.SendImageFromPool(n, f, func() error {
				// 下载图片
				return illust.DownloadToCache(0)
			}, ctxext.Send(ctx), ctxext.GetMessage(ctx))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		})
}

// soutuapi 请求api
func soutuapi(keyword string) (r resultjson, err error) {
	url := "https://api.pixivel.moe/v2/pixiv/illust/search/" + keyword + "?page=0"
	data, err := web.ReqWith(url, "GET", "https://pixivel.moe/", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36")
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &r)
	if err == nil && r.Error {
		err = errors.New(r.Message)
	}
	return
}
