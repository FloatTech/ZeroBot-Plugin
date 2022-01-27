// Package saucenao P站ID/saucenao/ascii2d搜图
package saucenao

import (
	"fmt"
	"strconv"

	"github.com/FloatTech/AnimeAPI/ascii2d"
	"github.com/FloatTech/AnimeAPI/imgpool"
	"github.com/FloatTech/AnimeAPI/pixiv"
	"github.com/FloatTech/AnimeAPI/saucenao"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/process"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

func init() { // 插件主体
	engine := control.Register("saucenao", order.PrioSauceNao, &control.Options{
		DisableOnDefault: false,
		Help: "搜图\n" +
			"- 以图搜图 | 搜索图片 | 以图识图[图片]\n" +
			"- 搜图[P站图片ID]",
	})
	// 根据 PID 搜图
	engine.OnRegex(`^搜图(\d+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			id, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			ctx.SendChain(message.Text("少女祈祷中......"))
			// 获取P站插图信息
			illust, err := pixiv.Works(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if illust.Pid > 0 {
				name := strconv.FormatInt(illust.Pid, 10)
				var imgs message.Message
				for i := range illust.ImageUrls {
					n := name + "_p" + strconv.Itoa(i)
					filepath := file.BOTPATH + "/" + pixiv.CacheDir + n
					f := ""
					m, err := imgpool.GetImage(n)
					if err == nil {
						imgs = append(imgs, message.Image(m.String()))
						continue
					}
					switch {
					case file.IsExist(filepath + ".jpg"):
						f = filepath + ".jpg"
					case file.IsExist(filepath + ".png"):
						f = filepath + ".png"
					case file.IsExist(filepath + ".gif"):
						f = filepath + ".gif"
					default:
						logrus.Debugln("[sausenao]开始下载", n)
						filepath, err = illust.DownloadToCache(i, n)
						if err == nil {
							f = filepath
						}
					}
					if f != "" {
						m.SetFile(f)
						hassent, err := m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
						if err == nil {
							imgs = append(imgs, message.Image(m.String()))
							if hassent {
								process.SleepAbout1sTo2s()
							}
						} else {
							logrus.Debugln("[saucenao]", err)
							imgs = append(imgs, message.Image("file:///"+f))
						}
					}
				}
				txt := message.Text(
					"标题：", illust.Title, "\n",
					"插画ID：", illust.Pid, "\n",
					"画师：", illust.UserName, "\n",
					"画师ID：", illust.UserId, "\n",
					"直链：", "https://pixivel.moe/detail?id=", illust.Pid,
				)
				if imgs != nil {
					// 发送搜索结果
					ctx.Send(append(imgs, message.Text("\n"), txt))
				} else {
					// 图片下载失败，仅发送文字结果
					ctx.SendChain(txt)
				}
			} else {
				ctx.SendChain(message.Text("图片不存在!"))
			}
		})
	// 以图搜图
	engine.OnKeywordGroup([]string{"以图搜图", "搜索图片", "以图识图"}, ctxext.CmdMatch, ctxext.MustGiven).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 开始搜索图片
			ctx.SendChain(message.Text("少女祈祷中......"))
			for _, pic := range ctx.State["image_url"].([]string) {
				fmt.Println(pic)
				if result, err := saucenao.SauceNAO(pic); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
				} else {
					// 返回SauceNAO的结果
					ctx.SendChain(
						message.Text("我有把握是这个！"),
						message.Image(result.Thumbnail),
						message.Text(
							"\n",
							"相似度：", result.Similarity, "\n",
							"标题：", result.Title, "\n",
							"插画ID：", result.PixivID, "\n",
							"画师：", result.MemberName, "\n",
							"画师ID：", result.MemberID, "\n",
							"直链：", "https://pixivel.moe/detail?id=", result.PixivID,
						),
					)
					continue
				}
				if result, err := ascii2d.Ascii2d(pic); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
				} else {
					// 返回Ascii2d的结果
					ctx.SendChain(
						message.Text(
							"大概是这个？", "\n",
							"标题：", result.Title, "\n",
							"插画ID：", result.Pid, "\n",
							"画师：", result.UserName, "\n",
							"画师ID：", result.UserId, "\n",
							"直链：", "https://pixivel.moe/detail?id=", result.Pid,
						),
					)
					continue
				}
			}
		})
}
