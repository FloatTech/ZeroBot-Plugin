// Package gamesystem 基于zbp的猜歌插件
package gamesystem

import (
	"image"

	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/single"
	"github.com/wdvxdr1123/ZeroBot/message"

	// 图片输出
	"github.com/Coloured-glaze/gg"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/img/writer"
	"github.com/FloatTech/rendercard"
	"github.com/FloatTech/zbputils/img/text"
)

const (
	serviceErr = "[gamesystem]error:"
	kanbanpath = "data/Control/icon.jpg"
)

type gameinfo struct {
	Name    string // 游戏名称
	Command string // 游玩指令
	Help    string // 游戏说明
	Rewards string // 奖励说明
}

var (
	// 游戏列表
	gamelist = make([]gameinfo, 0, 100)
	// 插件主体
	engine = control.Register("gamesystem", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "游戏系统",
		Help:             "- 游戏列表",
	}).ApplySingle(single.New(
		single.WithKeyFn(func(ctx *zero.Ctx) int64 { return ctx.Event.GroupID }),
		single.WithPostFn[int64](func(ctx *zero.Ctx) {
			ctx.Break()
			ctx.Send(
				message.ReplyWithMessage(ctx.Event.MessageID,
					message.Text("呜呜呜,我只能当一个游戏的裁判,分身乏术力..."),
				),
			)
		}),
	))
)

type imginfo struct {
	Img  image.Image
	High int
}

func init() {
	engine.OnFullMatch("游戏列表").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			_, err := file.GetLazyData(text.SakuraFontFile, true)
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			var imgs []imginfo
			var yOfLine1 int // 第一列最大高度
			var yOfLine2 int // 第二列最大高度
			for i, info := range gamelist {
				gameCardInfo := rendercard.TextCardInfo{
					FontOfTitle:    text.SakuraFontFile,
					FontOfText:     text.SakuraFontFile,
					Title:          info.Name,
					TitleSetting:   "Center",
					Text:           info,
					DisplaySetting: true,
				}
				img, yOfPic, err := gameCardInfo.DrawTextCard()
				if err != nil {
					ctx.SendChain(message.Text(serviceErr, err))
					return
				}
				if i%2 == 0 { // 第一列
					yOfLine1 += yOfPic + 20
				} else { // 第二列
					yOfLine2 += yOfPic + 20
				}
				var imginfo = imginfo{
					Img:  img,
					High: yOfPic,
				}
				imgs = append(imgs, imginfo)
			}
			lnperpg := math.Ceil(math.Max(yOfLine1, yOfLine2), (256 + 30))
			imgback, err := rendercard.Titleinfo{
				Line:          lnperpg,
				Lefttitle:     "游戏系统",
				Leftsubtitle:  "Game System",
				Righttitle:    "FloatTech",
				Rightsubtitle: "ZeroBot-Plugin",
				Textpath:      text.SakuraFontFile,
				Imgpath:       kanbanpath,
			}.Drawtitle()
			if err != nil {
				ctx.SendChain(message.Text(serviceErr, err))
				return
			}
			yOfLine := []int{0, 0}
			canvas := gg.NewContextForImage(imgback)
			// 插入游戏列表卡片
			for i, imginfo := range imgs {
				canvas.DrawImage(imginfo.Img, 25+620*(i%2), 360+yOfLine[i%2])
				yOfLine[i%2] += imginfo.High + 20
			}
			data, cl := writer.ToBytes(canvas.Image())
			defer cl()
			if id := ctx.SendChain(message.ImageBytes(data)); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})
}
