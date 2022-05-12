// Package deepdanbooru 二次元图片标签识别
package deepdanbooru

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/FloatTech/AnimeAPI/danbooru"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/img/writer"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func init() { // 插件主体
	engine := control.Register("danbooru", &control.Options{
		DisableOnDefault: false,
		Help: "二次元图片标签识别\n" +
			"- 鉴赏图片[图片]",
		PrivateDataFolder: "danbooru",
	})

	cachefolder := engine.DataFolder()

	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"鉴赏图片"}, zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			for _, url := range ctx.State["image_url"].([]string) {
				t, err := danbooru.TagURL("", url)
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				digest := md5.Sum(helper.StringToBytes(url))
				f := cachefolder + hex.EncodeToString(digest[:])
				if file.IsNotExist(f) {
					_ = writer.SavePNG2Path(f, t)
				}
				ctx.SendChain(message.Image("file:///" + file.BOTPATH + "/" + f))
			}
		})
}
