// Package lolicon 基于 https://api.lolicon.app 随机图片
package lolicon

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	imagepool "github.com/FloatTech/zbputils/img/pool"
)

const (
	api = "https://api.lolicon.app/setu/v2"
)

type imgpool struct {
	pool map[string][]*message.MessageSegment
	max  int
	pm   sync.Mutex
}

var pool = &imgpool{
	pool: make(map[string][]*message.MessageSegment),
	max:  10,
}

func init() {
	en := control.Register("lolicon", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "lolicon\n" +
			"- 随机图片\n" +
			"- 随机图片 萝莉|少女\n",
	}).ApplySingle(ctxext.DefaultSingle)
	en.OnPrefix("随机图片").Limit(ctxext.LimitByUser).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			rapi := api
			imgtype := strings.TrimSpace(ctx.State["args"].(string))
			if imgtype != "" {
				rapi += "?tag=" + url.QueryEscape(imgtype)
			}
			go pool.fill(ctx, imgtype, rapi)
			if pool.size(imgtype) == 0 {
				ctx.SendChain(message.Text("INFO: 正在填充弹药......"))
				time.Sleep(time.Second * 10)
				if pool.size(imgtype) == 0 {
					ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
					return
				}
			}
			if id := ctx.Send(message.Message{ctxext.FakeSenderForwardNode(ctx, *pool.pick(imgtype))}).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}

func (p *imgpool) size(imgtype string) int {
	return len(p.pool[imgtype])
}

func (p *imgpool) push(ctx *zero.Ctx, imgtype, url string) {
	n := url[strings.LastIndex(url, "/")+1 : len(url)-4]
	m, err := imagepool.GetImage(n)
	var msg message.MessageSegment
	if err != nil {
		m.SetFile(url)
		_, _ = m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
		msg = message.Image(m.String())
	}
	p.pm.Lock()
	p.pool[imgtype] = append(p.pool[imgtype], &msg)
	p.pm.Unlock()
}

func (p *imgpool) fill(ctx *zero.Ctx, imgtype, url string) {
	times := math.Min(p.max-p.size(imgtype), 2)
	for i := 0; i < times; i++ {
		data, err := web.GetData(url)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		json := gjson.ParseBytes(data)
		if e := json.Get("error").Str; e != "" {
			ctx.SendChain(message.Text("ERROR: ", e))
			return
		}
		var imageurl string
		if imageurl = json.Get("data.0.urls.original").Str; imageurl == "" {
			ctx.SendChain(message.Text("未找到相关内容, 换个tag试试吧"))
			return
		}
		imageurl = strings.ReplaceAll(imageurl, "i.pixiv.cat", "i.pixiv.re")
		p.push(ctx, imgtype, imageurl)
	}
}

func (p *imgpool) pick(imgtype string) (msg *message.MessageSegment) {
	p.pm.Lock()
	defer p.pm.Unlock()
	if p.size(imgtype) == 0 {
		return
	}
	msg = p.pool[imgtype][0]
	p.pool[imgtype] = p.pool[imgtype][:1]
	return
}
