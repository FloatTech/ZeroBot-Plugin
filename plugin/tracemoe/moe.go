// Package tracemoe 搜番
package tracemoe

import (
	"fmt"

	control "github.com/FloatTech/zbputils/control"
	trmoe "github.com/fumiama/gotracemoe"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

var (
	moe = trmoe.NewMoe("")
)

func init() { // 插件主体
	engine := control.Register("tracemoe", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help:             "tracemoe\n- 搜番 | 搜索番剧[图片]",
	})
	// 以图搜图
	engine.OnKeywordGroup([]string{"搜番", "搜索番剧"}, zero.MustProvidePicture).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			// 开始搜索图片
			ctx.SendChain(message.Text("少女祈祷中......"))
			for _, pic := range ctx.State["image_url"].([]string) {
				fmt.Println(pic)
				if result, err := moe.Search(pic, true, true); err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
				} else if len(result.Result) > 0 {
					r := result.Result[0]
					hint := "我有把握是这个！"
					if r.Similarity < 0.8 {
						hint = "大概是这个？"
					}
					mf := int(r.From / 60)
					mt := int(r.To / 60)
					sf := r.From - float32(mf*60)
					st := r.To - float32(mt*60)
					ctx.SendChain(
						message.Text(hint),
						message.Image(r.Image),
						message.Text(
							"\n",
							"番剧名：", r.Anilist.Title.Native, "\n",
							"话数：", r.Episode, "\n",
							"时间：", mf, ":", sf, "-", mt, ":", st,
						),
					)
				}
			}
		})
}
