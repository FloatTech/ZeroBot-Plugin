// Package saucenao P站ID/saucenao/ascii2d搜图
package saucenao

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/ascii2d"
	"github.com/FloatTech/AnimeAPI/pixiv"
	"github.com/FloatTech/AnimeAPI/saucenao"
	"github.com/FloatTech/AnimeAPI/yandex"

	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/pool"
)

func init() { // 插件主体
	engine := control.Register("saucenao", order.AcquirePrio(), &control.Options{
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
					f := file.BOTPATH + "/" + illust.Path(i)
					n := name + "_p" + strconv.Itoa(i)
					var m *pool.Image
					if file.IsNotExist(f) {
						m, err = pool.GetImage(n)
						if err == nil {
							err = file.DownloadTo(m.String(), f, true)
							if err != nil {
								ctx.SendChain(message.Text("ERROR: ", err))
								return
							}
							break
						}
						logrus.Debugln("[sausenao]开始下载", n)
						err = illust.DownloadToCache(i)
						if err == nil {
							m.SetFile(f)
							_, _ = m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
						}
					}
					imgs = append(imgs, message.Image("file:///"+f))
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
	engine.OnKeywordGroup([]string{"以图搜图", "搜索图片", "以图识图"}, zero.OnlyGroup, ctxext.MustProvidePicture).SetBlock(true).
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
						message.Image(result[0].Thumbnail),
						message.Text(
							"\n",
							"相似度：", result[0].Similarity, "\n",
							"标题：", result[0].Title, "\n",
							"插画ID：", result[0].PixivID, "\n",
							"画师：", result[0].MemberName, "\n",
							"画师ID：", result[0].MemberID, "\n",
							"直链：", "https://pixivel.moe/detail?id=", result[0].PixivID,
						),
					)
					continue
				}
				if result, err := yandex.Yandex(pic); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
				} else {
					// 返回SauceNAO的结果
					ctx.SendChain(
						message.Text("我有把握是这个！"),
						message.Text(
							"\n",
							"标题：", result.Title, "\n",
							"插画ID：", result.Pid, "\n",
							"画师：", result.UserName, "\n",
							"画师ID：", result.UserId, "\n",
							"直链：", "https://pixivel.moe/detail?id=", result.Pid,
						),
					)
					continue
				}
				if result, err := ascii2d.Ascii2d(pic); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					continue
				} else {
					var msg message.Message = []message.MessageSegment{
						message.CustomNode(
							ctx.Event.Sender.Name(),
							ctx.Event.UserID,
							"ascii2d搜图结果",
						)}
					for i := 0; i < len(result) && i < 5; i++ {
						msg = append(
							msg,
							message.CustomNode(
								ctx.Event.Sender.Name(),
								ctx.Event.UserID,
								[]message.MessageSegment{
									message.Image(result[i].Thumb),
									message.Text(fmt.Sprintf(
										"标题：%s\n图源：%s\n画师：%s\n画师链接：%s\n图片链接：%s",
										result[i].Name,
										result[i].Type,
										result[i].AuthNm,
										result[i].Author,
										result[i].Link,
									)),
								},
							),
						)
					}
					if id := ctx.SendGroupForwardMessage(
						ctx.Event.GroupID,
						msg,
					).Get("message_id").Int(); id == 0 {
						ctx.SendChain(message.Text("ERROR: 可能被风控了"))
					}
				}
			}
		})
}
