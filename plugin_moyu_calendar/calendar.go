// Package moyucalendar 摸鱼人日历
package moyucalendar

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	control "github.com/FloatTech/zbputils/control"
	"github.com/fumiama/cron"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/order"
)

func init() {
	control.Register("moyucalendar", order.AcquirePrio(), &control.Options{
		DisableOnDefault: true,
		Help: "摸鱼人日历\n" +
			"- /启用 moyucalendar\n" +
			"- /禁用 moyucalendar",
	})

	// 定时任务每天8点执行一次
	c := cron.New()
	_, err := c.AddFunc("* 8 * * *", func() { calendar() })
	if err == nil {
		c.Start()
	}
}

func calendar() {
	m, ok := control.Lookup("moyucalendar")
	if !ok {
		return
	}
	image, _ := crew()
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		for _, g := range ctx.GetGroupList().Array() {
			grp := g.Get("group_id").Int()
			if m.IsEnabledIn(grp) {
				ctx.SendGroupMessage(grp, []message.MessageSegment{message.Image(image)})
			}
		}
		return true
	})

}

var newest, _ = regexp.Compile(`uigs="account_article_0" href="(/link.+?)">`)
var weixin, _ = regexp.Compile(`url \+= '(.+)';`)
var rili, _ = regexp.Compile(`data-src="(.{0,300})" data-type="png" data-w="540"`)

func crew() (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://weixin.sogou.com/weixin?type=1&s_from=input&query=%E6%91%B8%E9%B1%BC%E4%BA%BA%E6%97%A5%E5%8E%86", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("status not ok")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	match := newest.FindStringSubmatch(string(b))
	if len(match) < 2 {
		return "", errors.New("newest not found")
	}
	var link = "https://weixin.sogou.com" + match[1]
	reqa, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}
	reqa.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	var cookies = make([]string, 0, 4)
	for _, cookie := range resp.Cookies() {
		if cookie.Name != "ABTEST" && cookie.Name != "SNUID" && cookie.Name != "IPLOC" && cookie.Name != "SUID" {
			continue
		}
		cookies = append(cookies, cookie.Name+"="+cookie.Value)
	}
	reqa.Header.Set("Cookie", strings.Join(cookies, "; "))
	respa, err := client.Do(reqa)
	if err != nil {
		return "", err
	}
	defer respa.Body.Close()
	if respa.StatusCode != http.StatusOK {
		return "", errors.New("status not ok")
	}
	br := bufio.NewReader(respa.Body)
	var weixinurl = make([]string, 0)
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
		weixinurl = append(weixinurl, strings.ReplaceAll(matcha[1], "@", ""))
	}
	if len(weixinurl) == 0 {
		return "", errors.New("weixin url not found")
	}
	reqw, err := http.NewRequest("GET", strings.Join(weixinurl, ""), nil)
	reqa.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.182 Safari/537.36")
	if err != nil {
		return "", err
	}
	respw, err := client.Do(reqw)
	if err != nil {
		return "", err
	}
	defer respw.Body.Close()
	if respw.StatusCode != http.StatusOK {
		return "", errors.New("status not ok")
	}
	bw, _ := ioutil.ReadAll(respw.Body)
	matchw := rili.FindStringSubmatch(string(bw))
	if len(matchw) < 2 {
		return "", errors.New("calendar not found")
	}
	return matchw[1], nil
}
