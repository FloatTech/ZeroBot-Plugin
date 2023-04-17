// Package vtbwife 抽vtb老婆
package vtbwife

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/PuerkitoBio/goquery"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() { // 插件主体
	engine := control.Register("vtbwife", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "抽vtb老婆",
		Help:             "- 抽vtb(老婆)",
		PublicDataFolder: "VtbWife",
	})
	var keys []string
	engine.OnRegex(`^抽(vtb|VTB)(老婆)?$`, fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			content, err := engine.GetLazyData("wife_list.txt", false)
			if err != nil {
				panic(err)
			}
			// 将文件内容转换为单词
			keys = strings.Split(string(content), "\n")
			log.Println("[vtbwife]加载", len(keys), "位wtb...")
			return true
		})).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var key, u, b string
		var ok bool
		for i := 0; i < 3; i++ {
			key = keys[fcext.RandSenderPerDayN(ctx.Event.UserID, len(keys))+i]
			u, b, ok = geturl(key)
			if !ok {
				continue
			}
			break
		}
		txt := message.Text(
			"\n今天你的VTB老婆是: ", key,
		)
		if id := ctx.SendChain(message.At(ctx.Event.UserID), txt, message.Image(u), message.Text(b)); id.ID() == 0 {
			ctx.SendChain(message.At(ctx.Event.UserID), txt, message.Text("图片发送失败...\n"), message.Text(b))
		}
	})
}

func geturl(kword string) (u, brief string, ok bool) {
	resp, err := http.Get("https://zh.moegirl.org.cn/" + url.QueryEscape(kword))
	if err != nil {
		return "", "", false
	}
	defer resp.Body.Close()
	// 使用goquery解析网页内容
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", false
	}
	u, ok = doc.Find(".infobox-image").Attr("src") // class加.
	doc.Find("style").Remove()
	doc.Find("script").Remove()
	doc.Find(".fans-medal-level").Remove()
	var (
		b   []string
		k   int
		buf strings.Builder
	)
	doc.Find(".moe-infobox").Find("tr").Each(func(i int, s *goquery.Selection) {
		b = append(b, strings.TrimSpace(s.Text()))
	})
	for kk, v := range b {
		v = strings.TrimSpace(v)
		if v == "基本资料" || v == "基本信息" || v == "名字" || v == "名称" {
			k = kk + 1
			break
		}
	}
	if k != 0 {
		buf.WriteString(b[k-1])
		buf.WriteString("\n")
	}
	for ; k < len(b); k++ {
		buf.WriteString(strings.Replace(strings.Replace(b[k], "\n", ": ", 1), "\n", "", 1))
		buf.WriteString("\n")
	}

	brief = strings.TrimSpace(buf.String())
	return
}
