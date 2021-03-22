package setutime

import (
	utils "bot/setutime/utils"
	"fmt"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() { // 插件主体
	// 根据PID搜图
	zero.OnRegex(`^搜图(\d+)$`).SetBlock(true).SetPriority(30).
		Handle(func(ctx *zero.Ctx) {
			id := utils.Str2Int(ctx.State["regex_matched"].([]string)[1])
			ctx.Send("少女祈祷中......")
			// 获取P站插图信息
			illust := &utils.Illust{}
			if err := illust.IllustInfo(id); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 下载P站插图
			if _, err := illust.PixivPicDown(CACHEPATH); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 发送搜索结果
			ctx.Send(illust.DetailPic)
			return
		})
	// 通过回复以图搜图
	zero.OnRegex(`\[CQ:reply,id=(.*?)\](.*)搜索图片`).SetBlock(true).SetPriority(32).
		Handle(func(ctx *zero.Ctx) {
			var pics []string // 图片搜索池子
			// 获取回复的上文图片链接
			id := utils.Str2Int(ctx.State["regex_matched"].([]string)[1])
			for _, elem := range ctx.GetMessage(id).Elements {
				if elem.Type == "image" {
					pics = append(pics, elem.Data["url"])
				}
			}
			// 没有收到图片则向用户索取
			if len(pics) == 0 {
				ctx.Send("请发送多张图片！")
				next := ctx.FutureEvent("message", ctx.CheckSession())
				recv, cancel := next.Repeat()
				for e := range recv { // 循环获取channel发来的信息
					if len(e.Message) == 1 && e.Message[0].Type == "text" {
						cancel() // 如果是纯文本则退出索取
						break
					}
					for _, elem := range e.Message {
						if elem.Type == "image" { // 将信息中的图片添加到搜索池子
							pics = append(pics, elem.Data["url"])
						}
					}
					if len(pics) >= 5 {
						cancel() // 如果是图片数量大于等于5则退出索取
						break
					}
				}
			}
			if len(pics) == 0 {
				ctx.Send("没有收到图片，搜图结束......")
				return
			}
			// 开始搜索图片
			ctx.Send("少女祈祷中......")
			for _, pic := range pics {
				if text, err := utils.SauceNaoSearch(pic); err == nil {
					ctx.Send(text) // 返回SauceNAO的结果
					continue
				} else {
					ctx.Send(fmt.Sprintf("ERROR: %v", err))
				}
				if text, err := utils.Ascii2dSearch(pic); err == nil {
					ctx.Send(text) // 返回Ascii2d的结果
					continue
				} else {
					ctx.Send(fmt.Sprintf("ERROR: %v", err))
				}
			}
			return
		})
	// 通过命令以图搜图
	zero.OnKeywordGroup([]string{"以图识图", "以图搜图", "搜索图片"}).SetBlock(true).SetPriority(33).
		Handle(func(ctx *zero.Ctx) {
			var pics []string // 图片搜索池子
			// 获取信息中图片链接
			for _, elem := range ctx.Event.Message {
				if elem.Type == "image" {
					pics = append(pics, elem.Data["url"])
				}
			}
			// 没有收到图片则向用户索取
			if len(pics) == 0 {
				ctx.Send("请发送多张图片！")
				next := ctx.FutureEvent("message", zero.CheckUser(ctx.Event.UserID))
				recv, cancel := next.Repeat()
				for e := range recv { // 循环获取channel发来的信息
					if len(e.Message) == 1 && e.Message[0].Type == "text" {
						cancel() // 如果是纯文本则退出索取
						break
					}
					for _, elem := range e.Message {
						if elem.Type == "image" { // 将信息中的图片添加到搜索池子
							pics = append(pics, elem.Data["url"])
						}
					}
					if len(pics) >= 5 {
						cancel() // 如果是图片数量大于等于5则退出索取
						break
					}
				}
			}
			if len(pics) == 0 {
				ctx.Send("没有收到图片，搜图结束......")
				return
			}
			// 开始搜索图片
			ctx.Send("少女祈祷中......")
			for _, pic := range pics {
				if text, err := utils.SauceNaoSearch(pic); err == nil {
					ctx.Send(text) // 返回SauceNAO的结果
					continue
				} else {
					ctx.Send(fmt.Sprintf("ERROR: %v", err))
				}
				if text, err := utils.Ascii2dSearch(pic); err == nil {
					ctx.Send(text) // 返回Ascii2d的结果
					continue
				} else {
					ctx.Send(fmt.Sprintf("ERROR: %v", err))
				}
			}
			return
		})
}
