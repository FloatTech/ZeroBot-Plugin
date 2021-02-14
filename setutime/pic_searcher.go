package setutime

import (
	utils "bot/setutime/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
)

func init() {
	zero.RegisterPlugin(picSearch{}) // 注册插件
}

type picSearch struct{} // pixivSearch 搜索P站插图

func (_ picSearch) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "PicSearch",
		Version:    "0.0.1",
		Details:    "以图搜图",
	}
}

func (_ picSearch) Start() { // 插件主体
	// TODO 根据PID搜图
	zero.OnRegex(`搜图(\d+)`).SetBlock(true).SetPriority(30).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			zero.Send(event, "少女祈祷中......")
			// TODO 获取P站插图信息
			illust := &utils.Illust{}
			if err := illust.IllustInfo(id); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			// TODO 下载P站插图
			if _, err := illust.PixivPicDown(CACHEPATH); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			// TODO 发送搜索结果
			zero.Send(event, illust.DetailPic)
			return zero.FinishResponse
		})
	// TODO 通过回复以图搜图
	zero.OnRegex(`\[CQ:reply,id=(.*?)\](.*)搜索图片`).SetBlock(true).SetPriority(32).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var pics []string // 图片搜索池子
			// TODO 获取回复的上文图片链接
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			for _, elem := range zero.GetMessage(id).Elements {
				if elem.Type == "image" {
					pics = append(pics, elem.Data["url"])
				}
			}
			// TODO 没有收到图片则向用户索取
			if len(pics) == 0 {
				zero.Send(event, "请发送多张图片！")
				next := matcher.FutureEvent("message", zero.CheckUser(event.UserID))
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
				zero.Send(event, "没有收到图片，搜图结束......")
				return zero.FinishResponse
			}
			// TODO 开始搜索图片
			zero.Send(event, "少女祈祷中......")
			for _, pic := range pics {
				if text, err := utils.SauceNaoSearch(pic); err == nil {
					zero.Send(event, text) // 返回SauceNAO的结果
					continue
				} else {
					utils.SendError(event, err)
				}
				if text, err := utils.Ascii2dSearch(pic); err == nil {
					zero.Send(event, text) // 返回Ascii2d的结果
					continue
				} else {
					utils.SendError(event, err)
				}
			}
			return zero.FinishResponse
		})
	// TODO 通过命令以图搜图
	zero.OnKeywordGroup([]string{"以图识图", "以图搜图", "搜索图片"}).SetBlock(true).SetPriority(33).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var pics []string // 图片搜索池子
			// TODO 获取信息中图片链接
			for _, elem := range event.Message {
				if elem.Type == "image" {
					pics = append(pics, elem.Data["url"])
				}
			}
			// TODO 没有收到图片则向用户索取
			if len(pics) == 0 {
				zero.Send(event, "请发送多张图片！")
				next := matcher.FutureEvent("message", zero.CheckUser(event.UserID))
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
				zero.Send(event, "没有收到图片，搜图结束......")
				return zero.FinishResponse
			}
			// TODO 开始搜索图片
			zero.Send(event, "少女祈祷中......")
			for _, pic := range pics {
				if text, err := utils.SauceNaoSearch(pic); err == nil {
					zero.Send(event, text) // 返回SauceNAO的结果
					continue
				} else {
					utils.SendError(event, err)
				}
				if text, err := utils.Ascii2dSearch(pic); err == nil {
					zero.Send(event, text) // 返回Ascii2d的结果
					continue
				} else {
					utils.SendError(event, err)
				}
			}
			return zero.FinishResponse
		})
}
