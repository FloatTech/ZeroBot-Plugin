// Package novel 铅笔小说搜索插件
package novel

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/antchfx/htmlquery"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	ub "github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
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
	downloadURL  = websiteURL + "/modules/article/txtarticle.php?id=%v"
	detailURL    = websiteURL + "/book/%v/"
	idReg        = `/(\d+)/`
)

var (
	cachePath string
	// apikey 由账号和密码拼接而成, 例: zerobot,123456
	apikey   string
	apikeymu sync.Mutex
)

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("novel"),
		Brief:            "铅笔小说网搜索",
		Help: "- 小说[xxx]\n" +
			"- 设置小说配置 zerobot 123456\n" +
			"- 下载小说30298\n" +
			"建议去https://www.23qb.com/ 注册一个账号, 小说下载有积分限制",
		PrivateDataFolder: "novel",
	})
	cachePath = engine.DataFolder() + "cache/"
	_ = os.MkdirAll(cachePath, 0755)
	engine.OnRegex("^小说([\u4E00-\u9FA5A-Za-z0-9]{1,25})$").SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中......"))
			key := getAPIKey(ctx)
			u, p, _ := strings.Cut(key, ",")
			if u == "" {
				u = username
			}
			if p == "" {
				p = password
			}
			cookie, err := login(u, p)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			searchKey := ctx.State["regex_matched"].([]string)[1]
			searchHTML, err := search(searchKey, cookie)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			doc, err := htmlquery.Parse(strings.NewReader(searchHTML))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			htmlTitle := htmlquery.InnerText(htmlquery.FindOne(doc, "/html/head/title"))
			switch htmlTitle {
			case websiteTitle:
				list, err := htmlquery.QueryAll(doc, "//dl[@id='nr']")
				if err != nil {
					ctx.SendChain(message.Text("ERROR: ", err))
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
						ctx.SendChain(message.Text("ERROR: ", err))
						return
					}
					if id := ctx.SendChain(message.Image("base64://" + ub.BytesToString(data))); id.ID() == 0 {
						ctx.SendChain(message.Text("ERROR: 可能被风控了"))
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
	engine.OnRegex(`^设置小说配置\s(.*[^\s$])\s(.+)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			err := setAPIKey(ctx.State["manager"].(*ctrl.Control[*zero.Ctx]), regexMatched[1]+","+regexMatched[2])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置小说配置\nusername: ", regexMatched[1], "\npassword: ", regexMatched[2]))
		})
	engine.OnRegex("^下载小说([0-9]{1,25})$").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			id := regexMatched[1]
			ctx.SendChain(message.Text("少女祈祷中......"))
			key := getAPIKey(ctx)
			u, p, _ := strings.Cut(key, ",")
			if u == "" {
				u = username
			}
			if p == "" {
				p = password
			}
			cookie, err := login(u, p)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			detailHTML, err := detail(id, cookie)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			doc, err := htmlquery.Parse(strings.NewReader(detailHTML))
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			title := htmlquery.InnerText(htmlquery.FindOne(doc, "//*[@id='bookinfo']/div[@class='bookright']/div[@class='d_title']/h1"))
			fileName := filepath.Join(cachePath, title+".txt")
			if file.IsExist(fileName) {
				ctx.UploadThisGroupFile(filepath.Join(file.BOTPATH, fileName), filepath.Base(fileName), "")
				return
			}
			data, err := download(id, cookie)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = os.WriteFile(fileName, ub.StringToBytes(data), 0666)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.UploadThisGroupFile(filepath.Join(file.BOTPATH, fileName), filepath.Base(fileName), "")
		})
}

func login(username, password string) (cookie string, err error) {
	client := &http.Client{}
	usernameData, err := ub.UTF82GBK(ub.StringToBytes(username))
	if err != nil {
		return
	}
	usernameGbk := ub.BytesToString(usernameData)
	passwordData, err := ub.UTF82GBK(ub.StringToBytes(password))
	if err != nil {
		return
	}
	passwordGbk := ub.BytesToString(passwordData)
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
	defer loginResp.Body.Close()
	for _, v := range loginResp.Cookies() {
		cookie += v.Name + "=" + v.Value + ";"
	}
	return
}

func search(searchKey string, cookie string) (searchHTML string, err error) {
	searchKeyData, err := ub.UTF82GBK(ub.StringToBytes(searchKey))
	if err != nil {
		return
	}
	searchKeyGbk := ub.BytesToString(searchKeyData)
	data, err := web.RequestDataWithHeaders(web.NewDefaultClient(), searchURL, "POST", func(r *http.Request) error {
		r.Header.Set("Cookie", cookie)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", ua)
		return nil
	}, strings.NewReader(fmt.Sprintf("searchkey=%s&searchtype=all", url.QueryEscape(searchKeyGbk))))
	if err != nil {
		return
	}
	searchData, err := ub.GBK2UTF8(data)
	if err != nil {
		return
	}
	searchHTML = ub.BytesToString(searchData)
	return
}

func detail(id string, cookie string) (detailHTML string, err error) {
	data, err := web.RequestDataWithHeaders(web.NewDefaultClient(), fmt.Sprintf(detailURL, id), "GET", func(r *http.Request) error {
		r.Header.Set("Cookie", cookie)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", ua)
		return nil
	}, nil)
	if err != nil {
		return
	}
	detailData, err := ub.GBK2UTF8(data)
	if err != nil {
		return
	}
	detailHTML = ub.BytesToString(detailData)
	return
}

func download(id string, cookie string) (downloadHTML string, err error) {
	data, err := web.RequestDataWithHeaders(web.NewDefaultClient(), fmt.Sprintf(downloadURL, id), "GET", func(r *http.Request) error {
		r.Header.Set("Cookie", cookie)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("User-Agent", ua)
		return nil
	}, nil)
	if err != nil {
		return
	}
	downloadHTML = ub.BytesToString(data)
	return
}

func getAPIKey(ctx *zero.Ctx) string {
	apikeymu.Lock()
	defer apikeymu.Unlock()
	if apikey == "" {
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		_ = m.GetExtra(&apikey)
		logrus.Debugln("[novel] get api key:", apikey)
	}
	return apikey
}

func setAPIKey(m *ctrl.Control[*zero.Ctx], key string) error {
	apikeymu.Lock()
	defer apikeymu.Unlock()
	apikey = key
	return m.SetExtra(apikey)
}
