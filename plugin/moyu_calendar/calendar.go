// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/control/order"
	"github.com/FloatTech/zbputils/process"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	xpath "github.com/antchfx/htmlquery"
)

var ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36"
var weixin = regexp.MustCompile(`url \+= '(.+)';`)

var client = &http.Client{}

func init() {
	control.Register("moyucalendar", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "摸鱼人日历\n" +
			"- /启用 moyucalendar\n" +
			"- /禁用 moyucalendar",
	}).OnFullMatch("摸鱼人日历").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			title := fmt.Sprintf("摸鱼人日历 %d月%d日", time.Now().Month(), time.Now().Day())
			sg, cookies, err := sougou(title, "摸鱼人日历", ua)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			wx, err := redirect(sg, cookies, ua)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			image, err := calendar(wx, ua)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Image(image))
		})

	// 定时任务每天8点30分执行一次
	_, err := process.CronTab.AddFunc("30 8 * * *", func() {
		m, ok := control.Lookup("moyucalendar")
		if !ok {
			return
		}
		title := fmt.Sprintf("摸鱼人日历 %d月%d日", time.Now().Month(), time.Now().Day())
		sg, cookies, err := sougou(title, "摸鱼人日历", ua)
		if err != nil {
			return
		}
		wx, err := redirect(sg, cookies, ua)
		if err != nil {
			return
		}
		image, err := calendar(wx, ua)
		if err != nil {
			return
		}
		zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
			for _, g := range ctx.GetGroupList().Array() {
				grp := g.Get("group_id").Int()
				if m.IsEnabledIn(grp) {
					ctx.SendGroupMessage(grp, message.Message{message.Image(image)})
					process.SleepAbout1sTo2s()
				}
			}
			return true
		})
	})
	if err != nil {
		panic(err)
	}
}

func sougou(title, publisher, ua string) (string, []*http.Cookie, error) {
	u, _ := url.Parse("https://weixin.sogou.com/weixin")
	u.RawQuery = url.Values{
		"type":   []string{"2"},
		"s_from": []string{"input"},
		"query":  []string{title},
	}.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, errors.New("status not ok")
	}
	// 解析XPATH
	doc, err := xpath.Parse(resp.Body)
	if err != nil {
		return "", nil, err
	}
	// 取出每个返回的结果
	list := xpath.Find(doc, `//*[@class="news-list"]/li/div[@class="txt-box"]`)
	if len(list) == 0 {
		return "", nil, errors.New("sougou result not found")
	}
	var match string
	for i := range list {
		account := xpath.FindOne(list[i], `//div[@class="s-p"]/a[@class="account"]`)
		if account == nil {
			continue
		}
		if xpath.InnerText(account) != publisher {
			continue
		}
		target := xpath.FindOne(list[i], `//h3/a[@target="_blank"]`)
		if target == nil {
			continue
		}
		match = xpath.SelectAttr(target, "href")
		break
	}
	if match == "" {
		return "", nil, errors.New("sougou result not found")
	}
	return "https://weixin.sogou.com" + match, resp.Cookies(), nil
}

func redirect(link string, cookies []*http.Cookie, ua string) (string, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", ua)
	var c = make([]string, 0, 4)
	for _, cookie := range cookies {
		if cookie.Name != "ABTEST" && cookie.Name != "SNUID" &&
			cookie.Name != "IPLOC" && cookie.Name != "SUID" {
			continue
		}
		c = append(c, cookie.Name+"="+cookie.Value)
	}
	req.Header.Set("Cookie", strings.Join(c, "; "))
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("status not ok")
	}
	br := bufio.NewReader(resp.Body)
	var u = make([]string, 0)
	for {
		b, _, err := br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		matcha := weixin.FindStringSubmatch(string(b))
		if len(matcha) < 2 {
			continue
		}
		u = append(u, strings.ReplaceAll(matcha[1], "@", ""))
	}
	if len(u) == 0 {
		return "", errors.New("weixin url not found")
	}
	return strings.Join(u, ""), nil
}

func calendar(link, ua string) (string, error) {
	req, err := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", ua)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("status not ok")
	}
	doc, err := xpath.Parse(resp.Body)
	if err != nil {
		return "", err
	}
	html := xpath.OutputHTML(doc, false)
	if !strings.Contains(html, time.Now().Format("2006-01-02")) {
		return "", errors.New("calendar not today")
	}
	images := xpath.Find(doc, `//*[@id="js_content"]/p/img`)
	if images == nil {
		return "", errors.New("calendar not found")
	}
	var image string
	for i := range images {
		if xpath.SelectAttr(images[i], "data-w") != "540" {
			continue
		}
		image = xpath.SelectAttr(images[i], "data-src")
		break
	}
	if image == "" {
		return "", errors.New("image not found")
	}
	return image, nil
}
