// Package saucenao P站ID/saucenao/ascii2d搜图
package saucenao

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/ascii2d"
	"github.com/FloatTech/AnimeAPI/pixiv"
	"github.com/FloatTech/AnimeAPI/saucenao"
)

func init() { // 插件主体
	// 根据 PID 搜图
	zero.OnRegex(`^搜图(\d+)$`).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			id, _ := strconv.ParseInt(ctx.State["regex_matched"].([]string)[1], 10, 64)
			ctx.Send("少女祈祷中......")
			// 获取P站插图信息
			illust, err := pixiv.Works(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 改用 i.pixiv.cat 镜像站
			link := illust.ImageUrls
			link = strings.ReplaceAll(link, "i.pximg.net", "i.pixiv.cat")
			// 发送搜索结果
			ctx.SendChain(
				message.Image(link),
				message.Text(
					"\n",
					"标题：", illust.Title, "\n",
					"插画ID：", illust.Pid, "\n",
					"画师：", illust.UserName, "\n",
					"画师ID：", illust.UserId, "\n",
					"直链：", "https://pixivel.moe/detail?id=", illust.Pid,
				),
			)
		})
	// 以图搜图
	zero.OnKeywordGroup([]string{"以图搜图", "搜索图片", "以图识图"}).SetBlock(true).FirstPriority().
		Handle(func(ctx *zero.Ctx) {
			// 匹配命令
			for _, elem := range ctx.Event.Message {
				if elem.Type == "text" {
					text := strings.ReplaceAll(elem.Data["text"], " ", "")
					if text != ctx.State["keyword"].(string) {
						return
					}
				}
			}
			// 匹配图片
			rule := func() zero.Rule {
				return func(ctx *zero.Ctx) bool {
					var urls = []string{}
					for _, elem := range ctx.Event.Message {
						if elem.Type == "image" {
							urls = append(urls, elem.Data["url"])
						}
					}
					if len(urls) > 0 {
						ctx.State["image_url"] = urls
						return true
					}
					return false
				}
			}
			// 索取图片
			if !rule()(ctx) {
				ctx.SendChain(message.Text("请发送一张图片"))
				next := zero.NewFutureEvent("message", 999, false, zero.CheckUser(ctx.Event.UserID), rule())
				recv, cancel := next.Repeat()
				select {
				case <-time.After(time.Second * 120):
					return
				case e := <-recv:
					cancel()
					newCtx := &zero.Ctx{Event: e, State: zero.State{}}
					if rule()(newCtx) {
						ctx.State["image_url"] = newCtx.State["image_url"]
					}
				}
			}
			// 开始搜索图片
			ctx.Send("少女祈祷中......")
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
