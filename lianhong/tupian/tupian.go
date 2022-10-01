// Package tupian 图片获取集合
package tupian

import (
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	engine := control.Register("tupian", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Help: "全部图片指令\n" +
			"- 兽耳\n" +
			"- 白毛\n" +
			"- ！原神\n" +
			"- 黑丝\n" +
			"- 白丝\n" +
			"- 随机壁纸\n" +
			"- 星空\n" +
			"- 涩涩哒咩/我要涩涩\n",
	})
	engine.OnFullMatch("兽耳").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://iw233.cn/api.php?sort=cat&type")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatch("白丝").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://api.iw233.cn/api.php?sort=swbs")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatch("黑丝").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://api.iw233.cn/api.php?sort=swhs")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatch("随机壁纸").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://iw233.cn/api.php?sort=iw233&type")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatch("白毛").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://iw233.cn/api.php?sort=yin&type")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatch("星空").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://iw233.cn/api.php?sort=xing&type")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatch("！原神").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("https://sakura.iw233.cn/Tag/API/web/api/op.php")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
	engine.OnFullMatchGroup([]string{"涩涩哒咩", "我要涩涩"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			data, err := web.GetData("http://api.iw233.cn/api.php?sort=st")
			if err != nil {
				ctx.SendChain(message.Text("获取图片失败惹", err))
				return
			}
			ctx.SendChain(message.ImageBytes(data))
		})
}
