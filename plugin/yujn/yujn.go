// Package yujn 来源于 https://api.yujn.cn/ 的接口
package yujn

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	yujnURL      = "https://api.yujn.cn"
	zzxjjURL     = yujnURL + "/api/zzxjj.php?type=video"
	baisisURL    = yujnURL + "/api/baisis.php?type=video"
	heisisURL    = yujnURL + "/api/heisis.php?type=video"
	xjjURL       = yujnURL + "/api/xjj.php?type=video"
	tianmeiURL   = yujnURL + "/api/tianmei.php?type=video"
	ndymURL      = yujnURL + "/api/ndym.php?type=video"
	sbklURL      = yujnURL + "/api/sbkl.php?type=video"
	nvgaoURL     = yujnURL + "/api/nvgao.php?type=video"
	luoliURL     = yujnURL + "/api/luoli.php?type=video"
	yuzuURL      = yujnURL + "/api/yuzu.php?type=video"
	xggURL       = yujnURL + "/api/xgg.php?type=video"
	rewuURL      = yujnURL + "/api/rewu.php?type=video"
	diaodaiURL   = yujnURL + "/api/diaodai.php?type=video"
	hanfuURL     = yujnURL + "/api/hanfu.php?type=video"
	jpyzURL      = yujnURL + "/api/jpmt.php?type=video"
	qingchunURL  = yujnURL + "/api/qingchun.php?type=video"
	ksbianzhuang = yujnURL + "/api/ksbianzhuang.php?type=video"
	dybianzhuang = yujnURL + "/api/bianzhuang.php?type=video"
	mengwaURL    = yujnURL + "/api/mengwa.php?type=video"
	chuandaURL   = yujnURL + "/api/chuanda.php?type=video"
	wmscURL      = yujnURL + "/api/wmsc.php?type=video"
	yujieURL     = yujnURL + "/api/yujie.php"
	luchaURL     = yujnURL + "/api/lvcha.php"
	duirenURL    = yujnURL + "/api/duiren.php"
	saohuaURL    = yujnURL + "/api/saohua.php"
	qinghuaURL   = yujnURL + "/api/qinghua.php"
	wuURL        = yujnURL + "/api/text_wu.php"
	wenanURL     = yujnURL + "/api/wenan.php"
	yuyinURL     = yujnURL + "/api/yuyin.php?type=json&from=%v&msg=%v"
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "遇见API",
		Help: "- 小姐姐视频\n- 小姐姐视频2\n- 黑丝视频\n- 白丝视频\n" +
			"- 欲梦视频\n- 甜妹视频\n- 双倍快乐\n- 纯情女高\n" +
			"- 萝莉视频\n- 玉足视频\n- 帅哥视频\n- 热舞视频\n" +
			"- 吊带视频\n- 汉服视频\n- 极品狱卒\n- 清纯视频\n" +
			"- 快手变装\n- 抖音变装\n- 萌娃视频\n- 穿搭视频\n" +
			"- 完美身材\n- 御姐撒娇\n- 绿茶语音\n- 怼人语音\n" +
			"- 随机骚话\n- 随机污句子\n- 随机美句\n- 土味情话\n- 让[丁真|陈泽|梅西|孙笑川|科比|懒羊羊|胡桃|雫るる]说我测尼玛",
	})
)

func init() {
	engine.OnFullMatchGroup([]string{"小姐姐视频"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(zzxjjURL))
	})
	engine.OnFullMatchGroup([]string{"小姐姐视频2"}).SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(xjjURL))
	})
	engine.OnFullMatch("黑丝视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(heisisURL))
	})
	engine.OnFullMatch("白丝视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(baisisURL))
	})
	engine.OnFullMatch("欲梦视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(ndymURL))
	})
	engine.OnFullMatch("甜妹视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(tianmeiURL))
	})
	engine.OnFullMatch("双倍快乐").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(sbklURL))
	})
	engine.OnFullMatch("纯情女高").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(nvgaoURL))
	})
	engine.OnFullMatch("萝莉视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(luoliURL))
	})
	engine.OnFullMatch("玉足视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(yuzuURL))
	})
	engine.OnFullMatch("帅哥视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(xggURL))
	})
	engine.OnFullMatch("热舞视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(rewuURL))
	})
	engine.OnFullMatch("吊带视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(diaodaiURL))
	})
	engine.OnFullMatch("汉服视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(hanfuURL))
	})
	engine.OnFullMatch("极品狱卒").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(jpyzURL))
	})
	engine.OnFullMatch("清纯视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(qingchunURL))
	})
	engine.OnFullMatch("快手变装").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(ksbianzhuang))
	})
	engine.OnFullMatch("抖音变装").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(dybianzhuang))
	})
	engine.OnFullMatch("萌娃视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(mengwaURL))
	})
	engine.OnFullMatch("穿搭视频").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(chuandaURL))
	})
	engine.OnFullMatch("完美身材").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Video(wmscURL))
	})
	engine.OnFullMatch("御姐撒娇").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Record(yujieURL))
	})
	engine.OnFullMatch("绿茶语音").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Record(luchaURL))
	})
	engine.OnFullMatch("怼人语音").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Record(duirenURL))
	})
	engine.OnFullMatch("随机骚话").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(saohuaURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(binary.BytesToString(data)))
	})
	engine.OnFullMatch("土味情话").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(qinghuaURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(binary.BytesToString(data)))
	})
	engine.OnFullMatch("随机污句子").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(wuURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(binary.BytesToString(data)))
	})
	engine.OnFullMatch("随机美句").SetBlock(true).Limit(ctxext.LimitByUser).Handle(func(ctx *zero.Ctx) {
		data, err := web.GetData(wenanURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		text := binary.BytesToString(data)
		text = strings.ReplaceAll(text, "<p>", "")
		text = strings.ReplaceAll(text, "</p>", "")
		ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text(text))
	})
	engine.OnRegex("^让(丁真|陈泽|梅西|孙笑川|科比|懒羊羊|胡桃|雫るる)说([\\s\u4e00-\u9fa5\u3040-\u309F\u30A0-\u30FF\\w\\p{P}\u3000-\u303F\uFF00-\uFFEF]+)$").Limit(ctxext.LimitByGroup).Handle(func(ctx *zero.Ctx) {
		name := ctx.State["regex_matched"].([]string)[1]
		msg := ctx.State["regex_matched"].([]string)[2]
		data, err := web.GetData(fmt.Sprintf(yuyinURL, url.QueryEscape(name), url.QueryEscape(msg)))
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		recordURL := gjson.Get(binary.BytesToString(data), "url").String()
		if recordURL == "" {
			ctx.SendChain(message.Text("ERROR: 语音生成失败"))
			return
		}
		ctx.SendChain(message.Record(recordURL))
	})
}
