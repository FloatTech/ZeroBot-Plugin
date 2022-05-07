// Package novel 铅笔小说搜索插件
package novel

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	ub "github.com/FloatTech/zbputils/binary"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
)

const (
	websiteURL   = "https://www.23qb.com"
	websiteTitle = "铅笔小说"
	errorTitle   = "出现错误！"
	username     = "zerobot"
	password     = "123456"
	submit       = "%26%23160%3B%B5%C7%26%23160%3B%26%23160%3B%C2%BC%26%23160%3B"
	ua           = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"
	loginURL     = websiteURL + "/login.php?do=submit&jumpurl=https%3A%2F%2Fwww.23qb.com%2F"
	searchURL    = websiteURL + "/saerch.php"
	idReg        = `/(\d+)/`
)

var gCurCookieJar *cookiejar.Jar

func init() {
	control.Register("novel", &control.Options{
		DisableOnDefault: false,
		Help:             "铅笔小说网搜索\n- 小说[xxx]",
	}).OnRegex("^小说([\u4E00-\u9FA5A-Za-z0-9]{1,25})$").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中......"))
			err := login(username, password)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			searchKey := ctx.State["regex_matched"].([]string)[1]
			searchHTML, err := search(searchKey)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			doc, err := htmlquery.Parse(strings.NewReader(searchHTML))
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			htmlTitle := htmlquery.InnerText(htmlquery.FindOne(doc, "/html/head/title"))
			switch htmlTitle {
			case websiteTitle:
				list, err := htmlquery.QueryAll(doc, "//dl[@id='nr']")
				if err != nil {
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
				if len(list) != 0 {
					txt := ""
					for _, v := range list {
						bookName := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[1]/h3/a[1]"))
						category := htmlquery.InnerText(htmlquery.FindOne(v, "/dt/span[1]"))
						author := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[2]/span[1]"))
						status := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[2]/span[2]"))
						wordNumbers := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[2]/span[3]"))
						description := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[3]"))
						updateTime := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[1]/h3/span[1]"))
						latestChapter := htmlquery.InnerText(htmlquery.FindOne(v, "/dd[4]/a[1]"))

						reg := regexp.MustCompile(idReg)
						id := reg.FindStringSubmatch(htmlquery.SelectAttr(htmlquery.FindOne(v, "/dt/a[1]"), "href"))[1]

						webpageURL := websiteURL + "/book/" + id + "/"
						downloadURL := websiteURL + "/modules/article/txtarticle.php?id=" + id
						txt += fmt.Sprintf("书名:%s\n类型:%s\n作者:%s\n状态:%s\n字数:%s\n简介:%s\n更新时间:%s\n最新章节:%s\n网页链接:%s\n下载地址:%s\n\n", bookName, category, author, status, wordNumbers, description, updateTime, latestChapter, webpageURL, downloadURL)
					}
					data, err := text.RenderToBase64(txt, text.FontFile, 400, 20)
					if err != nil {
						ctx.SendChain(message.Text("ERROR:", err))
						return
					}
					if id := ctx.SendChain(message.Image("base64://" + helper.BytesToString(data))); id.ID() == 0 {
						ctx.SendChain(message.Text("ERROR:可能被风控了"))
					}
				} else {
					text := htmlquery.InnerText(htmlquery.FindOne(doc, "//div[@id='tipss']"))
					text = strings.ReplaceAll(text, " ", "")
					text = strings.ReplaceAll(text, "本站", websiteURL)
					ctx.SendChain(message.Text(text))
				}
			case errorTitle:
				ctx.SendChain(message.Text(errorTitle))
				text := htmlquery.InnerText(htmlquery.FindOne(doc, "//div[@style='text-align: center;padding:10px']"))
				text = strings.ReplaceAll(text, " ", "")
				ctx.SendChain(message.Text(text))
			default:
				bookName := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:book_name']"), "content")
				category := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:category']"), "content")
				author := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:author']"), "content")
				status := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:status']"), "content")
				description := htmlquery.InnerText(htmlquery.FindOne(doc, "//div[@id='bookintro']/p"))
				updateTime := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:update_time']"), "content")
				latestChapter := htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:latest_chapter_name']"), "content")

				reg := regexp.MustCompile(idReg)
				id := reg.FindStringSubmatch(htmlquery.SelectAttr(htmlquery.FindOne(doc, "//meta[@property='og:novel:read_url']"), "content"))[1]
				webpageURL := websiteURL + "/book/" + id + "/"
				downloadURL := websiteURL + "/modules/article/txtarticle.php?id=" + id
				text := fmt.Sprintf("书名:%s\n类型:%s\n作者:%s\n状态:%s\n简介:%s\n更新时间:%s\n最新章节:%s\n网页链接:%s\n下载地址:%s\n", bookName, category, author, status, description, updateTime, latestChapter, webpageURL, downloadURL)
				ctx.SendChain(message.Text(text))
			}
		})
}

func login(username, password string) (err error) {
	gCurCookieJar, _ = cookiejar.New(nil)
	client := &http.Client{
		Jar: gCurCookieJar,
	}
	usernameData, err := ub.UTF82GBK(helper.StringToBytes(username))
	if err != nil {
		return
	}
	usernameGbk := helper.BytesToString(usernameData)
	passwordData, err := ub.UTF82GBK(helper.StringToBytes(password))
	if err != nil {
		return
	}
	passwordGbk := helper.BytesToString(passwordData)
	loginReq, err := http.NewRequest("POST", loginURL, strings.NewReader(fmt.Sprintf("username=%s&password=%s&usecookie=315360000&action=login&submit=%s", url.QueryEscape(usernameGbk), url.QueryEscape(passwordGbk), submit)))
	if err != nil {
		return
	}
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginReq.Header.Set("User-Agent", ua)
	loginResp, err := client.Do(loginReq)
	if err != nil {
		return
	}
	_ = loginResp.Body.Close()
	return
}

func search(searchKey string) (searchHTML string, err error) {
	searchKeyData, err := ub.UTF82GBK(helper.StringToBytes(searchKey))
	if err != nil {
		return
	}
	searchKeyGbk := helper.BytesToString(searchKeyData)
	client := &http.Client{
		Jar: gCurCookieJar,
	}
	searchReq, err := http.NewRequest("POST", searchURL, strings.NewReader(fmt.Sprintf("searchkey=%s&searchtype=all", url.QueryEscape(searchKeyGbk))))
	if err != nil {
		return
	}
	searchReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	searchReq.Header.Set("User-Agent", ua)
	searchResp, err := client.Do(searchReq)
	if err != nil {
		return
	}
	searchData, err := io.ReadAll(searchResp.Body)
	_ = searchResp.Body.Close()
	if err != nil {
		return
	}
	searchData, err = ub.GBK2UTF8(searchData)
	if err != nil {
		return
	}
	searchHTML = helper.BytesToString(searchData)
	return searchHTML, nil
}
