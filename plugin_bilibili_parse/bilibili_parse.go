// Package bilibiliparse b站视频链接解析
package bilibiliparse

import (
	"regexp"
	"strings"

	"github.com/FloatTech/zbputils/control"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

var (
	bilibiliRe = "https://www.bilibili.com/video/av[0-9]+|https://www.bilibili.com/video/BV[0-9a-zA-Z]+|https://b23.tv/[0-9a-zA-Z]+"
	validRe    = "https://www.bilibili.com/video/(BV[0-9a-zA-Z]+)"
)

func init() {
	engine := control.Register("bilibiliparse", order.PrioBiliBiliParse, &control.Options{
		DisableOnDefault: false,
		Help: "b站视频链接解析\n" +
			"- https://www.bilibili.com/video/BV1xx411c7BF | https://www.bilibili.com/video/av1605 | https://b23.tv/I8uzWCA",
	})

	engine.OnRegex(bilibiliRe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		bilibiliURL := ctx.State["regex_matched"].([]string)[0]
		m := parseURL(bilibiliURL)
		if len(m) != 0 {
			ctx.Send(m)
		}
	})
}

func parseURL(bilibiliURL string) (m message.Message) {
	doc, err := htmlquery.LoadURL(bilibiliURL)
	if err != nil {
		log.Errorln("[bilibiliparse]:访问的链接为", bilibiliURL, ",错误为", err)
	}
	videoURL := htmlquery.FindOne(doc, "/html/head/meta[@itemprop='url']").Attr[2].Val
	re := regexp.MustCompile(validRe)
	if !re.MatchString(videoURL) {
		return
	}
	bv := re.FindStringSubmatch(videoURL)[1]
	title := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/h1/span/text()").Data
	m = append(m, message.Text(title+"\n"))
	view := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/div/span[@class='view']/text()").Data
	dm := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/div/span[@class='dm']/text()").Data
	m = append(m, message.Text(view+dm+"\n"))
	t := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/div/span[3]/text()").Data
	m = append(m, message.Text(t))
	image := htmlquery.FindOne(doc, "/html/head/meta[@itemprop='image']").Attr[2].Val
	m = append(m, message.Image(image))
	like := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='like']/text()").Data
	coin := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='coin']/text()").Data
	m = append(m, message.Text("点赞：", strings.TrimSpace(like)+"投币：", strings.TrimSpace(coin)+"\n"))
	collect := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='collect']/text()").Data
	share := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='share']/text()").Data
	m = append(m, message.Text("收藏：", strings.TrimSpace(collect)+"分享：", strings.TrimSpace(share)+"\n"))
	m = append(m, message.Text(bv))
	return
}
