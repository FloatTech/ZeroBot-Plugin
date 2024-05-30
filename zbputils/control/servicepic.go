package control

import (
	"image"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/ttl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/disintegration/imaging"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/FloatTech/rendercard"

	"github.com/FloatTech/zbputils/img/text"
)

const (
	bannerpath = "data/zbpbanner/"
	kanbanpath = "data/Control/"
	bannerurl  = "https://github.com/FloatTech/zbpbanner/raw/main/"
)

type plugininfo struct {
	name   string
	brief  string
	banner string
	status bool
}

var (
	// 标题缓存
	titlecache image.Image
	// 卡片缓存
	cardcache = ttl.NewCache[uint64, image.Image](time.Hour * 12)
	// 阴影缓存
	fullpageshadowcache image.Image
	// lnperpg 每页行数
	lnperpg = 9
	// 字体 GlowSans 数据
	glowsd []byte
	// 字体 Impact 数据
	impactd []byte
)

func init() {
	err := os.MkdirAll(bannerpath, 0755)
	if err != nil {
		panic(err)
	}
	_, err = file.GetLazyData(kanbanpath+"kanban.png", Md5File, false)
	if err != nil {
		panic(err)
	}
	glowsd, err = file.GetLazyData(text.GlowSansFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
	impactd, err = file.GetLazyData(text.ImpactFontFile, Md5File, true)
	if err != nil {
		panic(err)
	}
}

func drawservicesof(gid int64) (imgs []image.Image, err error) {
	pluginlist := make([]plugininfo, len(priomap))
	ForEachByPrio(func(i int, manager *ctrl.Control[*zero.Ctx]) bool {
		pluginlist[i] = plugininfo{
			name:   manager.Service,
			brief:  manager.Options.Brief,
			banner: manager.Options.Banner,
			status: manager.IsEnabledIn(gid),
		}
		return true
	})
	// 分页
	if len(pluginlist) < lnperpg*3 {
		// 如果单页显示数量超出了总数量
		lnperpg = math.Ceil(len(pluginlist), 3)
	}
	if titlecache == nil {
		titlecache, err = (&rendercard.Title{
			Line:          lnperpg,
			LeftTitle:     "服务列表",
			LeftSubtitle:  "service_list",
			RightTitle:    "FloatTech",
			RightSubtitle: "ZeroBot-Plugin",
			TitleFontData: glowsd,
			TextFontData:  impactd,
			ImagePath:     kanbanpath + "kanban.png",
		}).DrawTitle()
		if err != nil {
			return
		}
	}

	cardlist := make([]image.Image, len(pluginlist))
	n := runtime.NumCPU()

	if len(pluginlist) <= n {
		for k, info := range pluginlist {
			banner := ""
			switch {
			case strings.HasPrefix(info.banner, "http"):
				err = file.DownloadTo(info.banner, bannerpath+info.name+".png")
				if err != nil {
					return
				}
				banner = bannerpath + info.name + ".png"
			case info.banner != "":
				banner = info.banner
			default:
				_, err = file.GetCustomLazyData(bannerurl, bannerpath+info.name+".png")
				if err == nil {
					banner = bannerpath + info.name + ".png"
				}
			}
			c := &rendercard.Title{
				IsEnabled:     info.status,
				LeftTitle:     info.name,
				LeftSubtitle:  info.brief,
				ImagePath:     banner,
				TitleFontData: impactd,
				TextFontData:  glowsd,
			}
			h := c.Sum64()
			card := cardcache.Get(h)
			if card == nil {
				card, err = c.DrawCard()
				if err != nil {
					return
				}
				cardcache.Set(h, card)
			}
			cardlist[k] = card
		}
	} else {
		batchsize := len(pluginlist) / n
		wg := sync.WaitGroup{}
		drawcard := func(info []plugininfo, cards []image.Image) {
			defer wg.Done()
			for k, info := range info {
				banner := ""
				switch {
				case strings.HasPrefix(info.banner, "http"):
					err1 := file.DownloadTo(info.banner, bannerpath+info.name+".png")
					if err1 != nil {
						err = err1
						return
					}
					banner = bannerpath + info.name + ".png"
				case info.banner != "":
					banner = info.banner
				default:
					_, err1 := file.GetCustomLazyData(bannerurl, bannerpath+info.name+".png")
					if err1 == nil {
						banner = bannerpath + info.name + ".png"
					}
				}
				c := &rendercard.Title{
					IsEnabled:     info.status,
					LeftTitle:     info.name,
					LeftSubtitle:  info.brief,
					ImagePath:     banner,
					TitleFontData: impactd,
					TextFontData:  glowsd,
				}
				h := c.Sum64()
				card := cardcache.Get(h)
				if card == nil {
					var err1 error
					card, err1 = c.DrawCard()
					if err1 != nil {
						err = err1
						return
					}
					card = rendercard.Fillet(card, 8)
					cardcache.Set(h, card)
				}
				cards[k] = card
			}
		}
		wg.Add(n)
		for i := 0; i < n; i++ {
			a := i * batchsize
			b := (i + 1) * batchsize
			go drawcard(pluginlist[a:b], cardlist[a:b])
		}
		if batchsize*n < len(pluginlist) {
			wg.Add(1)
			d := len(pluginlist) - batchsize*n
			go drawcard(pluginlist[d:], cardlist[d:])
		}
		wg.Wait()
		if err != nil {
			return
		}
	}

	wg := sync.WaitGroup{}
	cardnum := lnperpg * 3
	page := math.Ceil(len(pluginlist), cardnum)
	imgs = make([]image.Image, page)
	x, y := 30+2, 30+300+30+6+4
	if fullpageshadowcache == nil {
		fullpageshadow := gg.NewContextForImage(rendercard.Transparency(titlecache, 0))
		fullpageshadow.SetRGBA255(0, 0, 0, 192)
		for i := 0; i < cardnum; i++ {
			fullpageshadow.DrawRoundedRectangle(float64(x), float64(y), 384-4, 256-4, 0)
			fullpageshadow.Fill()
			x += 384 + 30
			if (i+1)%3 == 0 {
				x = 30 + 2
				y += 256 + 30
			}
		}
		fullpageshadowcache = imaging.Blur(fullpageshadow.Image(), 7)
	}
	wg.Add(page)
	for l := 0; l < page; l++ { // 页数
		go func(l int, islast bool) {
			defer wg.Done()
			x, y := 30+2, 30+300+30+6+4
			var shadowimg image.Image
			if islast && len(pluginlist)-cardnum*l < cardnum {
				shadow := gg.NewContextForImage(rendercard.Transparency(titlecache, 0))
				shadow.SetRGBA255(0, 0, 0, 192)
				for i := 0; i < len(pluginlist)-cardnum*l; i++ {
					shadow.DrawRoundedRectangle(float64(x), float64(y), 384-4, 256-4, 0)
					shadow.Fill()
					x += 384 + 30
					if (i+1)%3 == 0 {
						x = 30 + 2
						y += 256 + 30
					}
				}
				shadowimg = imaging.Blur(shadow.Image(), 7)
			} else {
				shadowimg = fullpageshadowcache
			}
			one := gg.NewContextForImage(titlecache)
			x, y = 30, 30+300+30
			one.DrawImage(shadowimg, 0, 0)
			for i := 0; i < math.Min(cardnum, len(pluginlist)-cardnum*l); i++ {
				one.DrawImage(cardlist[(cardnum*l)+i], x, y)
				x += 384 + 30
				if (i+1)%3 == 0 {
					x = 30
					y += 256 + 30
				}
			}
			imgs[l] = one.Image()
		}(l, l == page-1)
	}
	wg.Wait()
	return
}
