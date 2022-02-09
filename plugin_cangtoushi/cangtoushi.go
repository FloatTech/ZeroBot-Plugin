// Package cangtoushi 藏头诗
package cangtoushi

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	control "github.com/FloatTech/zbputils/control"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/zbputils/control/order"
)

const (
	loginURL  = "https://www.shicimingju.com/cangtoushi/"
	searchURL = "https://www.shicimingju.com/cangtoushi/index.html"
	ua        = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	referer   = "https://www.shicimingju.com/cangtoushi/index.html"
)

var (
	gCurCookieJar *cookiejar.Jar
	csrf          string
)

func init() {
	engine := control.Register("cangtoushi", order.AcquirePrio(), &control.Options{
		DisableOnDefault: false,
		Help: "藏头诗\n" +
			"- 藏头诗[xxx]\n- 藏尾诗[xxx]",
	})
	engine.OnRegex(`藏头诗\s?([一-龥]{3,10})$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		kw := ctx.State["regex_matched"].([]string)[1]
		login()
		data, err := search(kw, "7", "0")
		if err != nil {
			log.Errorln("[cangtoushi]:", err)
		}
		text := dealHTML(helper.BytesToString(data))
		ctx.SendChain(message.Text(text))
	})

	engine.OnRegex(`藏尾诗\s?([一-龥]{3,10})$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		kw := ctx.State["regex_matched"].([]string)[1]
		login()
		data, err := search(kw, "7", "2")
		if err != nil {
			log.Errorln("[cangtoushi]:", err)
		}
		text := dealHTML(helper.BytesToString(data))
		ctx.SendChain(message.Text(text))
	})
}

func login() {
	gCurCookieJar, _ = cookiejar.New(nil)
	client := &http.Client{
		Jar: gCurCookieJar,
	}
	request, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	request.Header.Add("User-Agent", ua)
	response, err := client.Do(request)
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	response.Body.Close()
	doc, err := htmlquery.Parse(strings.NewReader(helper.BytesToString(data)))
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	csrf = htmlquery.SelectAttr(htmlquery.FindOne(doc, "//input[@name='_csrf']"), "value")
}

func search(kw, zishu, position string) (data []byte, err error) {
	postStr := fmt.Sprintf("_csrf=%s&kw=%s&zishu=%s&position=%s", url.QueryEscape(csrf), url.QueryEscape(kw), zishu, position)
	log.Println("postStr:", postStr)
	client := &http.Client{
		Jar: gCurCookieJar,
	}
	request, err := http.NewRequest("POST", searchURL, strings.NewReader(postStr))
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	request.Header.Add("Referer", referer)
	request.Header.Add("User-Agent", ua)
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	data, err = io.ReadAll(response.Body)
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	response.Body.Close()
	return
}

func dealHTML(data string) (text string) {
	doc, err := htmlquery.Parse(strings.NewReader(data))
	if err != nil {
		log.Errorln("[cangtoushi]:", err)
	}
	text = htmlquery.InnerText(htmlquery.FindOne(doc, "//div[@class='card']/div[@class='card']"))
	text = strings.ReplaceAll(text, " ", "")
	text = strings.Replace(text, "\n", "", 1)
	return text
}
