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

	"github.com/FloatTech/zbputils/binary"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/process"
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
	}).OnFullMatch("摸鱼人日历").SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			image, err := crew()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
			}
			ctx.SendChain(message.Image(image))
		})

	// 定时任务每天8点执行一次
	_, err := process.CronTab.AddFunc("30 8 * * *", func() {
		m, ok := control.Lookup("moyucalendar")
		if !ok {
			return
		}
		image, err := crew()
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

var newest = regexp.MustCompile(`href="(/link.+?)" id="sogou_vr_11002601_title_0" uigs="article_title_0"`)
var weixin = regexp.MustCompile(`url \+= '(.+)';`)
var calendar = regexp.MustCompile(`data-src="(.{0,300})" data-type="png" data-w="540"`)

func crew() (string, error) {
	client := &http.Client{}
	u, _ := url.Parse("https://weixin.sogou.com/weixin")
	u.RawQuery = url.Values{
		"type":   []string{"2"},
		"s_from": []string{"input"},
		"query":  []string{fmt.Sprintf("摸鱼人日历 %d月%d日", time.Now().Month(), time.Now().Day())},
	}.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
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
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	match := newest.FindStringSubmatch(binary.BytesToString(b))
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
		matcha := weixin.FindStringSubmatch(binary.BytesToString(b))
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
	bw, _ := io.ReadAll(respw.Body)
	today, err := regexp.Compile(time.Now().Format("2006-01-02"))
	if err != nil {
		return "", err
	}
	if !today.Match(bw) {
		return "", errors.New("today not found")
	}
	matchw := calendar.FindStringSubmatch(binary.BytesToString(bw))
	if len(matchw) < 2 {
		return "", errors.New("calendar not found")
	}
	return matchw[1], nil
}
