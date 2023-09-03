// Package imagefinder 关键字搜图
package imagefinder

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/quic-go/quic-go/http3"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/pixiv"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/pool"
)

type resultjson struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		Illusts []struct {
			ID          int64  `json:"id"`
			Title       string `json:"title"`
			AltTitle    string `json:"altTitle"`
			Description string `json:"description"`
			Type        int64  `json:"type"`
			CreateDate  string `json:"createDate"`
			UploadDate  string `json:"uploadDate"`
			Sanity      int64  `json:"sanity"`
			Width       int64  `json:"width"`
			Height      int64  `json:"height"`
			PageCount   int64  `json:"pageCount"`
			Tags        []struct {
				Name        string `json:"name"`
				Translation string `json:"translation"`
			} `json:"tags"`
			Statistic struct {
				Bookmarks int64 `json:"bookmarks"`
				Likes     int64 `json:"likes"`
				Comments  int64 `json:"comments"`
				Views     int64 `json:"views"`
			} `json:"statistic"`
			Image string `json:"image"`
		} `json:"illusts"`
		Scores  []float64 `json:"scores"`
		HasNext bool      `json:"has_next"`
	} `json:"data"`
}

var hrefre = regexp.MustCompile(`<a href=".*">`)

func init() {
	control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "关键字搜图",
		Help:             "- 来张 [xxx]",
	}).OnRegex(`^来张\s?(.*)$`, zero.AdminPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			keyword := ctx.State["regex_matched"].([]string)[1]
			soutujson, err := soutuapi(keyword)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			rannum := rand.Intn(len(soutujson.Data.Illusts))
			il := soutujson.Data.Illusts[rannum]
			illust, err := pixiv.Works(il.ID)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if len(illust.ImageUrls) == 0 {
				ctx.SendChain(message.Text("ERROR: nil image url"))
				return
			}
			u := illust.ImageUrls[0]
			n := u[strings.LastIndex(u, "/")+1 : len(u)-4]
			f := illust.Path(0)

			err = pool.SendImageFromPool(n, f, func() error {
				// 下载图片
				return illust.DownloadToCache(0)
			}, ctxext.SendFakeForwardToGroup(ctx,
				message.Text(
					il.Width, "x", il.Height, "\n",
					"标题: ", il.Title, "\n",
					"副标题: ", il.AltTitle, "\n",
					"ID: ", il.ID, "\n",
					"画师: ", illust.UserName, " (", illust.UserID, ")", "\n",
					"分级:", il.Sanity, "\n",
					hrefre.ReplaceAllString(strings.ReplaceAll(strings.ReplaceAll(il.Description, "<br />", "\n"), "</a>", ""), ""),
					printtags(reflect.ValueOf(&il.Tags)),
				),
			), ctxext.GetFirstMessageInForward(ctx))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		})
}

// soutuapi 请求api
func soutuapi(keyword string) (r resultjson, err error) {
	var data []byte
	data, err = web.RequestDataWith(&http.Client{Transport: &http3.RoundTripper{}},
		"https://api.pixivel.moe/v2/pixiv/illust/search/"+url.QueryEscape(keyword)+"?page=0",
		"GET",
		"https://pixivel.moe/",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.61 Safari/537.36",
		nil,
	)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &r)
	if err == nil && r.Error {
		err = errors.New(r.Message)
	}
	return
}

func printtags(r reflect.Value) string {
	tags := r.Elem()
	s := binary.BytesToString(binary.NewWriterF(func(w *binary.Writer) {
		for i := 0; i < tags.Len(); i++ {
			_ = w.WriteByte('\n')
			tag := tags.Index(i)
			_ = w.WriteByte('#')
			w.WriteString(tag.Field(0).String())
			if !tag.Field(1).IsZero() {
				w.WriteString(" (")
				w.WriteString(tag.Field(1).String())
				w.WriteString(")")
			}
		}
	}))
	return s
}
