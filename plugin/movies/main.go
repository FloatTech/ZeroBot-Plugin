// Package movies 电影查询
package movies

import (
	"bytes"
	"encoding/json"
	"image"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/imgfactory"
	"github.com/FloatTech/ttl"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/disintegration/imaging"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	apiURL = "https://m.maoyan.com/ajax/"
	ua     = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Mobile Safari/537.36"
)

var todayPic = ttl.NewCache[uint64, []byte](time.Hour * 12)

func init() {
	control.Register("movies", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "电影查询",
		Help: "- 今日电影\n" +
			"- 预售电影",
	}).OnRegex(`^(今日|预售)电影$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		switch ctx.State["regex_matched"].([]string)[1] {
		case "今日":
			todayOnPic := todayPic.Get(0)
			if todayOnPic != nil {
				ctx.SendChain(message.ImageBytes(todayOnPic))
				return
			}
			data, err := web.RequestDataWith(web.NewDefaultClient(), apiURL+"movieOnInfoList", "", "GET", ua, nil)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR1]", err))
				return
			}
			var parsed movieOnList
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text("[EEROR2]:", err))
				return
			}
			if len(parsed.MovieList) == 0 {
				ctx.SendChain(message.Text("今日无电影上映"))
			}
			pic, err := drawOnListPic(parsed)
			if err != nil {
				ctx.SendChain(message.Text("[EEROR3]:", err))
				return
			}
			todayPic.Set(0, pic)
			ctx.SendChain(message.ImageBytes(pic))
		case "预售":
			todayOnPic := todayPic.Get(1)
			if todayOnPic != nil {
				ctx.SendChain(message.ImageBytes(todayOnPic))
				return
			}
			data, err := web.RequestDataWith(web.NewDefaultClient(), apiURL+"comingList?token=", "", "GET", ua, nil)
			if err != nil {
				ctx.SendChain(message.Text("[ERROR4]", err))
				return
			}
			var parsed comingList
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				ctx.SendChain(message.Text("[EEROR5]:", err))
				return
			}
			if len(parsed.Coming) == 0 {
				ctx.SendChain(message.Text("没有预售信息"))
			}
			pic, err := drawComListPic(parsed)
			if err != nil {
				ctx.SendChain(message.Text("[EEROR6]:", err))
				return
			}
			todayPic.Set(1, pic)
			ctx.SendChain(message.ImageBytes(pic))
		}
	})
}

type movieOnList struct {
	MovieList []movieInfo `json:"movieList"`
}
type movieInfo struct {
	ID       int64   `json:"id"`       // 电影ID
	Img      string  `json:"img"`      // 海报
	Version  string  `json:"version"`  // 电影格式
	Nm       string  `json:"nm"`       // 名称
	Sc       float64 `json:"sc"`       // 评分
	Wish     int64   `json:"wish"`     // 观看人数
	Star     string  `json:"star"`     // 演员
	Rt       string  `json:"rt"`       // 上映时间
	ShowInfo string  `json:"showInfo"` // 今日上映信息
}
type comingList struct {
	Coming []comingInfo `json:"coming"`
}
type comingInfo struct {
	ID          int64  `json:"id"`          // 电影ID
	Img         string `json:"img"`         // 海报
	Version     string `json:"version"`     // 电影格式
	ShowInfo    string `json:"showInfo"`    // 今日上映信息
	Nm          string `json:"nm"`          // 名称
	Wish        int64  `json:"wish"`        // 期待人数
	Star        string `json:"star"`        // 演员
	ComingTitle string `json:"comingTitle"` // 上映时间
}

func drawOnListPic(lits movieOnList) (data []byte, err error) {
	index := len(lits.MovieList)
	backgroundURL, err := web.GetData(lits.MovieList[rand.Intn(index)].Img)
	if err != nil {
		return
	}
	back, _, err := image.Decode(bytes.NewReader(backgroundURL))
	if err != nil {
		return
	}
	listPicH := 3000
	listPicW := float64(back.Bounds().Dx()) * float64(listPicH) / float64(back.Bounds().Dy())
	back = imgfactory.Size(back, int(listPicW), listPicH).Image()
	movieCardw := int(listPicW - 100)
	movieCardh := listPicH/index - 20
	wg := &sync.WaitGroup{}
	wg.Add(index)
	movieList := make(map[int]image.Image, index*2)
	for i, movieInfos := range lits.MovieList {
		go func(info movieInfo, index int) {
			defer wg.Done()
			movieCard := gg.NewContext(movieCardw, movieCardh)
			// 	毛玻璃背景
			movieCard.DrawImage(imaging.Blur(back, 8), -50, -((movieCardh+15)*index + 20))
			movieCard.DrawRoundedRectangle(1, 1, float64(movieCardw-1*2), float64(movieCardh-1*2), 16)
			movieCard.SetLineWidth(3)
			movieCard.SetRGBA255(255, 255, 255, 100)
			movieCard.StrokePreserve()
			movieCard.SetRGBA255(255, 255, 255, 140)
			movieCard.Fill()
			// 放置海报
			posterURL, err := web.GetData(info.Img)
			if err != nil {
				return
			}
			poster, _, err := image.Decode(bytes.NewReader(posterURL))
			if err != nil {
				return
			}
			PicH := movieCardh - 20
			picW := int(float64(poster.Bounds().Dx()) * float64(PicH) / float64(poster.Bounds().Dy()))
			movieCard.DrawImage(imgfactory.Size(poster, picW, PicH).Image(), 10, 10)

			err = movieCard.LoadFontFace(text.GlowSansFontFile, 72)
			if err != nil {
				return
			}
			_, nameH := movieCard.MeasureString(info.Nm)
			scale := float64(PicH/4) / nameH // 按比例缩放
			// 写入文字信息
			err = movieCard.LoadFontFace(text.GlowSansFontFile, 72*scale)
			if err != nil {
				return
			}
			nameW, nameH := movieCard.MeasureString(info.Nm)
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(info.Nm, float64(picW)+20*scale, 20*scale+nameH/2, 0, 0.5)
			// 评分
			wish := strconv.FormatInt(info.Wish, 10) + "人已看"
			munW, munH := movieCard.MeasureString(wish)
			movieCard.DrawRoundedRectangle(float64(movieCardw)-munW*0.9-10*scale, float64(movieCardh)-munH*2.4-10*scale, munW*0.9, munH*2.4, 72*0.2)
			switch {
			case info.Sc < 8.4:
				movieCard.SetRGBA255(250, 97, 0, 200)
			case info.Sc > 9:
				movieCard.SetRGBA255(0, 201, 87, 200)
			default:
				movieCard.SetRGBA255(240, 230, 140, 200)
			}
			movieCard.Fill()
			movieCard.DrawRoundedRectangle(float64(movieCardw)-munW*0.9-10*scale, float64(movieCardh)-munH*1.2-10*scale, munW*0.9, munH*1.2, 72*0.2)
			switch {
			case info.Wish < 100000: // 十万以下
				movieCard.SetRGBA255(255, 125, 64, 200)
			case info.Wish > 100000000: // 破亿
				movieCard.SetRGBA255(34, 139, 34, 200)
			default:
				movieCard.SetRGBA255(255, 215, 0, 200)
			}
			movieCard.Fill()
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(strconv.FormatFloat(info.Sc, 'f', 2, 64), float64(movieCardw)-munW*0.9/2-10*scale, float64(movieCardh)-munH*1.2-munH*1.3/2-10*scale, 0.5, 0.5)
			err = movieCard.LoadFontFace(text.GlowSansFontFile, 60*scale)
			if err != nil {
				return
			}
			movieCard.DrawStringAnchored(wish, float64(movieCardw)-10*scale-munW*0.9/2, float64(movieCardh)-munH*1.2/2-11*scale, 0.5, 0.5)
			// 电影ID
			mid := strconv.FormatInt(info.ID, 10)
			midW, _ := movieCard.MeasureString(mid)
			movieCard.DrawRoundedRectangle(float64(picW)+20*scale+nameW+10*scale, 20*scale, midW*1.2, nameH, 72*0.2)
			movieCard.SetRGBA255(221, 221, 221, 200)
			movieCard.Fill()
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(mid, float64(picW)+20*scale+nameW+10*scale+midW*1.2/2, 20*scale+nameH/2, 0.5, 0.5)

			err = movieCard.LoadFontFace(text.GlowSansFontFile, 32*scale)
			if err != nil {
				return
			}
			_, textH := movieCard.MeasureString(info.Star)
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(info.Star, float64(picW)+20*scale, 25*scale+nameH+10*scale+textH/2, 0, 0.5)
			movieType := "2D"
			if info.Version != "" {
				movieType = info.Version
			}
			movieCard.DrawStringAnchored("类型: "+movieType, float64(picW)+20*scale, 25*scale+nameH+10*scale+(textH+10*scale)*1+textH/2, 0, 0.5)
			movieCard.DrawStringAnchored("上映时间: "+info.Rt, float64(picW)+20*scale, 25*scale+nameH+10*scale+(textH+10*scale)*2+textH/2, 0, 0.5)
			movieCard.DrawStringAnchored("今日信息: "+info.ShowInfo, float64(picW)+20*scale, 25*scale+nameH+10*scale+(textH+10*scale)*3+textH/2, 0, 0.5)
			movieList[index] = movieCard.Image()
		}(movieInfos, i)
	}
	wg.Wait()
	canvas := gg.NewContextForImage(back)
	for i, imgs := range movieList {
		canvas.DrawImage(imgs, 50, (movieCardh+15)*i+20)
	}
	data, err = imgfactory.ToBytes(canvas.Image())
	return
}

func drawComListPic(lits comingList) (data []byte, err error) {
	index := len(lits.Coming)
	backgroundURL, err := web.GetData(lits.Coming[rand.Intn(index)].Img)
	if err != nil {
		return
	}
	back, _, err := image.Decode(bytes.NewReader(backgroundURL))
	if err != nil {
		return
	}
	listPicH := 3000
	listPicW := float64(back.Bounds().Dx()) * float64(listPicH) / float64(back.Bounds().Dy())
	back = imgfactory.Size(back, int(listPicW), listPicH).Image()
	movieCardw := int(listPicW - 100)
	movieCardh := listPicH/index - 20
	wg := &sync.WaitGroup{}
	wg.Add(index)
	movieList := make(map[int]image.Image, index*2)
	for i, movieInfos := range lits.Coming {
		go func(info comingInfo, index int) {
			defer wg.Done()
			movieCard := gg.NewContext(movieCardw, movieCardh)
			// 	毛玻璃背景
			movieCard.DrawImage(imaging.Blur(back, 8), -50, -((movieCardh+15)*index + 20))
			movieCard.DrawRoundedRectangle(1, 1, float64(movieCardw-1*2), float64(movieCardh-1*2), 16)
			movieCard.SetLineWidth(3)
			movieCard.SetRGBA255(255, 255, 255, 100)
			movieCard.StrokePreserve()
			movieCard.SetRGBA255(255, 255, 255, 140)
			movieCard.Fill()
			// 放置海报
			posterURL, err := web.GetData(info.Img)
			if err != nil {
				return
			}
			poster, _, err := image.Decode(bytes.NewReader(posterURL))
			if err != nil {
				return
			}
			PicH := movieCardh - 20
			picW := int(float64(poster.Bounds().Dx()) * float64(PicH) / float64(poster.Bounds().Dy()))
			movieCard.DrawImage(imgfactory.Size(poster, picW, PicH).Image(), 10, 10)

			err = movieCard.LoadFontFace(text.GlowSansFontFile, 72)
			if err != nil {
				return
			}
			_, nameH := movieCard.MeasureString(info.Nm)
			scale := float64(PicH/4) / nameH // 按比例缩放
			// 写入文字信息
			err = movieCard.LoadFontFace(text.GlowSansFontFile, 72*scale)
			if err != nil {
				return
			}
			nameW, nameH := movieCard.MeasureString(info.Nm)
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(info.Nm, float64(picW)+20*scale, 20*scale+nameH/2, 0, 0.5)
			// 期待人数
			munW1, _ := movieCard.MeasureString("期待人数")
			wish := strconv.FormatInt(info.Wish, 10)
			munW2, munH := movieCard.MeasureString(wish)
			munW := math.Max(munW1, munW2)
			movieCard.DrawRoundedRectangle(float64(movieCardw)-munW*0.9-10*scale, float64(movieCardh)-munH*2.4-10*scale, munW*0.9, munH*2.4, 72*0.2)
			switch {
			case info.Wish < 1000:
				movieCard.SetRGBA255(250, 97, 0, 200)
			case info.Wish > 100000:
				movieCard.SetRGBA255(0, 201, 87, 200)
			default:
				movieCard.SetRGBA255(240, 230, 140, 200)
			}
			movieCard.Fill()
			movieCard.DrawRoundedRectangle(float64(movieCardw)-munW*0.9-10*scale, float64(movieCardh)-munH*1.2-10*scale, munW*0.9, munH*1.2, 72*0.2)
			movieCard.SetRGBA255(34, 139, 34, 200)
			movieCard.Fill()
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(wish, float64(movieCardw)-munW*0.9/2-10*scale, float64(movieCardh)-munH*1.2-munH*1.3/2-10*scale, 0.5, 0.5)
			err = movieCard.LoadFontFace(text.GlowSansFontFile, 60*scale)
			if err != nil {
				return
			}
			movieCard.DrawStringAnchored("期待人数", float64(movieCardw)-10*scale-munW*0.9/2, float64(movieCardh)-munH*1.2/2-11*scale, 0.5, 0.5)
			// 电影ID
			mid := strconv.FormatInt(info.ID, 10)
			midW, _ := movieCard.MeasureString(mid)
			movieCard.DrawRoundedRectangle(float64(picW)+20*scale+nameW+10*scale, 20*scale, midW*1.2, nameH, 72*0.2)
			movieCard.SetRGBA255(221, 221, 221, 200)
			movieCard.Fill()
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(mid, float64(picW)+20*scale+nameW+10*scale+midW*1.2/2, 20*scale+nameH/2, 0.5, 0.5)

			err = movieCard.LoadFontFace(text.GlowSansFontFile, 32*scale)
			if err != nil {
				return
			}
			_, textH := movieCard.MeasureString(info.Star)
			movieCard.SetRGBA255(30, 30, 30, 255)
			movieCard.DrawStringAnchored(info.Star, float64(picW)+20*scale, 25*scale+nameH+10*scale+textH/2, 0, 0.5)
			movieType := "2D"
			if info.Version != "" {
				movieType = info.Version
			}
			movieCard.DrawStringAnchored("类型: "+movieType, float64(picW)+20*scale, 25*scale+nameH+10*scale+(textH+10*scale)*1+textH/2, 0, 0.5)
			movieCard.DrawStringAnchored("上映时间: "+info.ComingTitle, float64(picW)+20*scale, 25*scale+nameH+10*scale+(textH+10*scale)*2+textH/2, 0, 0.5)
			movieCard.DrawStringAnchored("今日信息: "+info.ShowInfo, float64(picW)+20*scale, 25*scale+nameH+10*scale+(textH+10*scale)*3+textH/2, 0, 0.5)
			movieList[index] = movieCard.Image()
		}(movieInfos, i)
	}
	wg.Wait()
	canvas := gg.NewContextForImage(back)
	for i, imgs := range movieList {
		canvas.DrawImage(imgs, 50, (movieCardh+15)*i+20)
	}
	data, err = imgfactory.ToBytes(canvas.Image())
	return
}
