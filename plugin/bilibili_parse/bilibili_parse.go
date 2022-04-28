// Package bilibiliparse b站视频链接解析
package bilibiliparse

import (
	"regexp"
	"strings"

	"github.com/FloatTech/zbputils/control"
	"github.com/antchfx/htmlquery"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	bilibiliRe = "https://www.bilibili.com/video/av[0-9]+|https://www.bilibili.com/video/BV[0-9a-zA-Z]+|https://b23.tv/[0-9a-zA-Z]+|https://www.bilibili.com/video/bv[0-9a-zA-Z]+"
	validRe    = "https://www.bilibili.com/video/(BV[0-9a-zA-Z]+)"
)

func init() {
	engine := control.Register("bilibiliparse", &control.Options{
		DisableOnDefault: false,
		Help: "b站视频链接解析\n" +
			"- https://www.bilibili.com/video/BV1xx411c7BF | https://www.bilibili.com/video/av1605 | https://b23.tv/I8uzWCA | https://www.bilibili.com/video/bv1xx411c7BF",
	})

	engine.OnRegex(bilibiliRe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		bilibiliURL := ctx.State["regex_matched"].([]string)[0]
		m, err := parseURL(bilibiliURL)
		if err != nil {
			ctx.SendChain(message.Text("ERROR:", err))
			return
		}
		if len(m) != 0 {
			ctx.Send(m)
		}
	})
}

func parseURL(bilibiliURL string) (m message.Message, err error) {
	doc, err := htmlquery.LoadURL(bilibiliURL)
	if err != nil {
		return
	}
	videoURL := htmlquery.FindOne(doc, "/html/head/meta[@itemprop='url']").Attr[2].Val
	re := regexp.MustCompile(validRe)
	if !re.MatchString(videoURL) {
		return
	}
	bv := re.FindStringSubmatch(videoURL)[1]
	title := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/h1/span/text()").Data
	m = append(m, message.Text(title+"\n"))
	upName := strings.TrimSpace(htmlquery.FindOne(doc, "//*[@id='v_upinfo']/div[2]/div[1]/a[1]/text()").Data)
	fanNumber := htmlquery.InnerText(htmlquery.FindOne(doc, "//i[@class='van-icon-general_addto_s']").NextSibling.NextSibling)
	m = append(m, message.Text("up："+upName+"，粉丝："+fanNumber+"\n"))
	view := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/div/span[@class='view']/text()").Data
	dm := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/div/span[@class='dm']/text()").Data
	m = append(m, message.Text(view+dm+"\n"))
	t := htmlquery.FindOne(doc, "//*[@id='viewbox_report']/div/span[3]/text()").Data
	m = append(m, message.Text(t))
	image := htmlquery.FindOne(doc, "/html/head/meta[@itemprop='image']").Attr[2].Val
	m = append(m, message.Image(image))
	like := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='like']/text()").Data
	coin := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='coin']/text()").Data
	m = append(m, message.Text("\n点赞：", strings.TrimSpace(like)+"，投币：", strings.TrimSpace(coin)+"\n"))
	collect := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='collect']/text()").Data
	share := htmlquery.FindOne(doc, "//*[@id='arc_toolbar_report']/div[1]/span[@class='share']/text()").Data
	m = append(m, message.Text("收藏：", strings.TrimSpace(collect)+"，分享：", strings.TrimSpace(share)+"\n"))
	m = append(m, message.Text(bv))
	return
}
