// Package deepdanbooru 二次元图片标签识别
package deepdanbooru

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/imgfactory"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault:  false,
		Brief:             "二次元图片标签识别",
		Help:              "- 鉴赏图片[图片]",
		PrivateDataFolder: "danbooru",
	})

	cachefolder := engine.DataFolder()

	// 上传一张图进行评价
	engine.OnKeywordGroup([]string{"鉴赏图片"}, zero.OnlyGroup, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			for _, url := range ctx.State["image_url"].([]string) {
				t, st, err := tagurl("", url)
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
					return
				}
				digest := md5.Sum(helper.StringToBytes(url))
				f := cachefolder + hex.EncodeToString(digest[:])
				if file.IsNotExist(f) {
					_ = imgfactory.SavePNG2Path(f, t)
				}
				m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+file.BOTPATH+"/"+f))}
				m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Text("tags: ", strings.Join(st.tseq, ","))))
				if id := ctx.Send(m).ID(); id == 0 {
					ctx.SendChain(message.Text("ERROR: 可能被风控或下载图片用时过长，请耐心等待"))
				}
			}
		})
}
