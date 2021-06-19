package picsearcher

import (
	"fmt"
	"os"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/context"
	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/pixiv"
	apiutils "github.com/Yiwen-Chan/ZeroBot-Plugin/api/utils"
	utils "github.com/Yiwen-Chan/ZeroBot-Plugin/picsearcher/utils"
)

var (
	CACHEPATH = os.TempDir() // 缓冲图片路径
)

func init() { // 插件主体
	if strings.Contains(CACHEPATH, "\\") {
		CACHEPATH += "\\picsch\\"
	} else {
		CACHEPATH += "/picsch/"
	}
	apiutils.CreatePath(CACHEPATH)
	// 根据PID搜图
	zero.OnRegex(`^搜图(\d+)$`).SetBlock(true).SetPriority(30).
		Handle(func(ctx *zero.Ctx) {
			id := apiutils.Str2Int(ctx.State["regex_matched"].([]string)[1])
			ctx.Send("少女祈祷中......")
			// 获取P站插图信息
			illust := &pixiv.Illust{}
			if err := illust.IllustInfo(id); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 下载P站插图
			savePath, err := illust.PixivPicDown(CACHEPATH)
			if err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 发送搜索结果
			ctx.Send(illust.DetailPic(savePath))
			illust.RmPic(CACHEPATH)
		})
	// 以图搜图
	zero.OnMessage(FullMatchText("以图搜图", "搜索图片", "以图识图"), context.MustHasPicture()).SetBlock(true).SetPriority(999).
		Handle(func(ctx *zero.Ctx) {
			// 开始搜索图片
			ctx.Send("少女祈祷中......")
			for _, pic := range ctx.State["image_url"].([]string) {
				fmt.Println(pic)
				if m, err := utils.SauceNaoSearch(pic); err == nil {
					ctx.SendChain(m...) // 返回SauceNAO的结果
					continue
				} else {
					ctx.SendChain(message.Text("ERROR: ", err))
				}
				if m, err := utils.Ascii2dSearch(pic); err == nil {
					ctx.SendChain(m...) // 返回Ascii2d的结果
					continue
				} else {
					ctx.SendChain(message.Text("ERROR: ", err))
				}
			}
		})
}

// FullMatchText 如果信息中文本完全匹配则返回 true
func FullMatchText(src ...string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		msg := ctx.Event.Message
		for _, elem := range msg {
			if elem.Type == "text" {
				text := elem.Data["text"]
				text = strings.ReplaceAll(text, " ", "")
				text = strings.ReplaceAll(text, "\r", "")
				text = strings.ReplaceAll(text, "\n", "")
				for _, s := range src {
					if text == s {
						return true
					}
				}
			}
		}
		return false
	}
}
