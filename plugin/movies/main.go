// Package movies 电影查询
package movies

import (
	"encoding/json"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/rendercard"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	apiURL = "https://m.maoyan.com/ajax/"
	ua     = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"
)

var (
	mu       sync.RWMutex
	todayPic = make([][]byte, 2)
	lasttime time.Time
	en       = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "电影查询",
		Help: "- 今日电影\n" +
			"- 预售电影",
		PrivateDataFolder: "movies",
	})
)

func init() {
	en.OnFullMatch("今日电影").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if todayPic != nil && time.Since(lasttime) < 12*time.Hour {
			ctx.SendChain(message.ImageBytes(todayPic[0]))
			return
		}
		lasttime = time.Now()
		movieComingList, err := getMovieList("今日电影")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if len(movieComingList) == 0 {
			ctx.SendChain(message.Text("没有今日电影"))
			return
		}
		pic, err := drawOnListPic(movieComingList)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		todayPic[0] = pic
		ctx.SendChain(message.ImageBytes(pic))
	})
	en.OnFullMatch("预售电影").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if todayPic[1] != nil && time.Since(lasttime) < 12*time.Hour {
			ctx.SendChain(message.ImageBytes(todayPic[1]))
			return
		}
		lasttime = time.Now()
		movieComingList, err := getMovieList("预售电影")
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		if len(movieComingList) == 0 {
			ctx.SendChain(message.Text("没有预售电影"))
			return
		}
		pic, err := drawComListPic(movieComingList)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return
		}
		todayPic[1] = pic
		ctx.SendChain(message.ImageBytes(pic))
	})
}

type movieInfo struct {
	ID  int64  `json:"id"`  // 电影ID
	Img string `json:"img"` // 海报

	Nm string `json:"nm"` // 名称

	Dir  string `json:"dir"`  // 导演
	Star string `json:"star"` // 演员

	OriLang string `json:"oriLang"` // 原语言
	Cat     string `json:"cat"`     // 类型

	Version string `json:"version"` // 电影格式
	Rt      string `json:"rt"`      // 上映时间

	ShowInfo    string `json:"showInfo"`    // 今日上映信息
	ComingTitle string `json:"comingTitle"` // 预售信息

	Sc      float64 `json:"sc"`      // 评分
	Wish    int64   `json:"wish"`    // 观看人数
	Watched int64   `json:"watched"` // 观看数
}
type movieOnList struct {
	MovieList []movieInfo `json:"movieList"`
}
type comingList struct {
	MovieList []movieInfo `json:"coming"`
}
type movieShow struct {
	MovieInfo movieInfo `json:"detailMovie"`
}

type cardInfo struct {
	Avatar         image.Image
	TopLeftText    string
	BottomLeftText []string
	RightText      string
	Rank           string
}

func getMovieList(mode string) (movieList []movieInfo, err error) {
	var data []byte
	if mode == "今日电影" {
		data, err = web.RequestDataWith(web.NewDefaultClient(), apiURL+"movieOnInfoList", "", "GET", ua, nil)
		if err != nil {
			return
		}
		var parsed movieOnList
		err = json.Unmarshal(data, &parsed)
		if err != nil {
			return
		}
		movieList = parsed.MovieList
	} else {
		data, err = web.RequestDataWith(web.NewDefaultClient(), apiURL+"comingList?token=", "", "GET", ua, nil)
		if err != nil {
			return
		}
		var parsed comingList
		err = json.Unmarshal(data, &parsed)
		if err != nil {
			return
		}
		movieList = parsed.MovieList
	}
	if len(movieList) == 0 {
		return
	}
	for i, info := range movieList {
		movieID := strconv.FormatInt(info.ID, 10)
		data, err = web.RequestDataWith(web.NewDefaultClient(), apiURL+"detailmovie?movieId="+movieID, "", "GET", ua, nil)
		if err != nil {
			return
		}
		var movieInfo movieShow
		err = json.Unmarshal(data, &movieInfo)
		if err != nil {
			return
		}
		if mode != "今日电影" {
			movieInfo.MovieInfo.ComingTitle = movieList[i].ComingTitle
		}
		movieList[i] = movieInfo.MovieInfo
	}
	// 整理数据，进行排序
	sort.Slice(movieList, func(i, j int) bool {
		if movieList[i].Sc != movieList[j].Sc {
			return movieList[i].Sc > movieList[j].Sc
		}
		if mode == "今日电影" {
			return movieList[i].Watched > movieList[j].Watched
		}
		return movieList[i].Wish > movieList[j].Wish
	})
	return movieList, nil
}
func drawOnListPic(lits []movieInfo) (data []byte, err error) {
	rankinfo := make([]*cardInfo, len(lits))

	wg := &sync.WaitGroup{}
	wg.Add(len(lits))
	for i := 0; i < len(lits); i++ {
		go func(i int) {
			info := lits[i]
			defer wg.Done()
			img, err := avatar(&info)
			if err != nil {
				return
			}
			movieType := "2D"
			if info.Version != "" {
				movieType = info.Version
			}
			watched := ""
			switch {
			case info.Watched > 100000000:
				watched = strconv.FormatFloat(float64(info.Watched)/100000000, 'f', 2, 64) + "亿"
			case info.Watched > 10000:
				watched = strconv.FormatFloat(float64(info.Watched)/10000, 'f', 2, 64) + "万"
			default:
				watched = strconv.FormatInt(info.Watched, 10)
			}
			rankinfo[i] = &cardInfo{
				TopLeftText: info.Nm + " (" + strconv.FormatInt(info.ID, 10) + ")",
				BottomLeftText: []string{
					"导演：" + info.Dir,
					"演员：" + info.Star,
					"标签：" + info.Cat,
					"语言: " + info.OriLang + "    类型: " + movieType,
					"上映时间: " + info.Rt,
				},
				RightText: watched + "人已看",
				Avatar:    img,
				Rank:      strconv.FormatFloat(info.Sc, 'f', 1, 64),
			}
		}(i)
	}
	wg.Wait()
	fontbyte, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	img, err := drawRankingCard(fontbyte, "今日电影", rankinfo)
	if err != nil {
		return
	}
	data, err = imgfactory.ToBytes(img)
	return
}

func drawComListPic(lits []movieInfo) (data []byte, err error) {
	rankinfo := make([]*cardInfo, len(lits))

	wg := &sync.WaitGroup{}
	wg.Add(len(lits))
	for i := 0; i < len(lits); i++ {
		go func(i int) {
			info := lits[i]
			defer wg.Done()
			img, err := avatar(&info)
			if err != nil {
				return
			}
			movieType := "2D"
			if info.Version != "" {
				movieType = info.Version
			}
			wish := ""
			switch {
			case info.Wish > 100000000:
				wish = strconv.FormatFloat(float64(info.Wish)/100000000, 'f', 2, 64) + "亿"
			case info.Wish > 10000:
				wish = strconv.FormatFloat(float64(info.Wish)/10000, 'f', 2, 64) + "万"
			default:
				wish = strconv.FormatInt(info.Wish, 10)
			}
			rankinfo[i] = &cardInfo{
				TopLeftText: info.Nm + " (" + strconv.FormatInt(info.ID, 10) + ")",
				BottomLeftText: []string{
					"导演：" + info.Dir,
					"演员：" + info.Star,
					"标签：" + info.Cat,
					"语言: " + info.OriLang + "    类型: " + movieType,
					"上映时间: " + info.Rt + "    播放时间: " + info.ComingTitle,
				},
				RightText: wish + "人期待",
				Avatar:    img,
				Rank:      strconv.Itoa(i + 1),
			}
		}(i)
	}
	wg.Wait()
	fontbyte, err := file.GetLazyData(text.GlowSansFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	img, err := drawRankingCard(fontbyte, "预售电影", rankinfo)
	if err != nil {
		return
	}
	data, err = imgfactory.ToBytes(img)
	return
}

func drawRankingCard(fontdata []byte, title string, rankinfo []*cardInfo) (img image.Image, err error) {
	line := len(rankinfo)
	const lineh = 130
	const w = 800
	h := 64 + (lineh+14)*line + 20 - 14
	canvas := gg.NewContext(w, h)
	canvas.SetRGBA255(255, 255, 255, 255)
	canvas.Clear()

	cardh, cardw := lineh, 770
	cardspac := 14
	hspac, wspac := 64.0, 16.0
	r := 16.0

	wg := &sync.WaitGroup{}
	wg.Add(line)
	cardimgs := make([]image.Image, line)
	for i := 0; i < line; i++ {
		go func(i int) {
			defer wg.Done()
			card := gg.NewContext(w, cardh)

			card.NewSubPath()

			card.MoveTo(wspac+float64(cardh)/2, 0)

			card.LineTo(wspac+float64(cardw)-r, 0)
			card.DrawArc(wspac+float64(cardw)-r, r, r, gg.Radians(-90), gg.Radians(0))
			card.LineTo(wspac+float64(cardw), float64(cardh)-r)
			card.DrawArc(wspac+float64(cardw)-r, float64(cardh)-r, r, gg.Radians(0), gg.Radians(90))
			card.LineTo(wspac+float64(cardh)/2, float64(cardh))
			card.DrawArc(wspac+r, float64(cardh)-r, r, gg.Radians(90), gg.Radians(180))
			card.LineTo(wspac, r)
			card.DrawArc(wspac+r, r, r, gg.Radians(180), gg.Radians(270))

			card.ClosePath()

			card.ClipPreserve()

			avatar := rankinfo[i].Avatar

			PicH := cardh - 20
			picW := int(float64(avatar.Bounds().Dx()) * float64(PicH) / float64(avatar.Bounds().Dy()))
			card.DrawImageAnchored(imgfactory.Size(avatar, picW, PicH).Image(), int(wspac)+10+picW/2, cardh/2, 0.5, 0.5)

			card.ResetClip()
			card.SetRGBA255(0, 0, 0, 127)
			card.Stroke()

			card.SetRGBA255(240, 210, 140, 200)
			card.DrawRoundedRectangle(wspac+float64(cardw-8-250), (float64(cardh)-50)/2, 250, 50, 25)
			card.Fill()
			card.SetRGB255(rendercard.RandJPColor())
			card.DrawRoundedRectangle(wspac+float64(cardw-8-60), (float64(cardh)-50)/2, 60, 50, 25)
			card.Fill()
			cardimgs[i] = card.Image()
		}(i)
	}

	canvas.SetRGBA255(0, 0, 0, 255)
	err = canvas.ParseFontFace(fontdata, 32)
	if err != nil {
		return
	}
	canvas.DrawStringAnchored(title, w/2, 64/2, 0.5, 0.5)

	err = canvas.ParseFontFace(fontdata, 22)
	if err != nil {
		return
	}
	wg.Wait()
	for i := 0; i < line; i++ {
		canvas.DrawImageAnchored(cardimgs[i], w/2, int(hspac)+((cardh+cardspac)*i), 0.5, 0)
		canvas.DrawStringAnchored(rankinfo[i].TopLeftText, wspac+10+80+10, hspac+float64((cardspac+cardh)*i+cardh*3/16), 0, 0.5)
	}

	// canvas.SetRGBA255(63, 63, 63, 255)
	err = canvas.ParseFontFace(fontdata, 14)
	if err != nil {
		return
	}
	for i := 0; i < line; i++ {
		for j, text := range rankinfo[i].BottomLeftText {
			canvas.DrawStringAnchored(text, wspac+10+80+10, hspac+float64((cardspac+cardh)*i+cardh*6/16)+float64(j*16), 0, 0.5)
		}
	}
	canvas.SetRGBA255(0, 0, 0, 255)
	err = canvas.ParseFontFace(fontdata, 20)
	if err != nil {
		return
	}
	for i := 0; i < line; i++ {
		canvas.DrawStringAnchored(rankinfo[i].RightText, w-wspac-8-60-8, hspac+float64((cardspac+cardh)*i+cardh/2), 1, 0.5)
	}

	canvas.SetRGBA255(255, 255, 255, 255)
	err = canvas.ParseFontFace(fontdata, 28)
	if err != nil {
		return
	}
	for i := 0; i < line; i++ {
		canvas.DrawStringAnchored(rankinfo[i].Rank, w-wspac-8-30, hspac+float64((cardspac+cardh)*i+cardh/2), 0.5, 0.5)
	}

	img = canvas.Image()
	return
}

// avatar 获取电影海报,图片大且多，存本地增加响应速度
func avatar(movieInfo *movieInfo) (pic image.Image, err error) {
	mu.Lock()
	defer mu.Unlock()

	aimgfile := filepath.Join(en.DataFolder(), movieInfo.Nm+"("+strconv.FormatInt(movieInfo.ID, 10)+").jpg")
	if file.IsNotExist(aimgfile) {
		err = file.DownloadTo(movieInfo.Img, aimgfile)
		if err != nil {
			return urlToImg(movieInfo.Img)
		}
	}
	f, err := os.Open(filepath.Join(file.BOTPATH, aimgfile))
	if err != nil {
		return urlToImg(movieInfo.Img)
	}
	defer f.Close()
	pic, _, err = image.Decode(f)
	return
}

func urlToImg(url string) (img image.Image, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	img, _, err = image.Decode(resp.Body)
	return
}
