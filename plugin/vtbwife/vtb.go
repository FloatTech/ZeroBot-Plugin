// Package vtbwife 抽vtb老婆
package vtbwife

import (
	"log"
	"net/http"
	"net/url"
	//"os"
	"strings"

	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/web"
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
		Help:             "- 抽vtb",
		PublicDataFolder: "VtbWife",
	})
	var keys []string
	engine.OnRegex(`^抽(vtb|VTB)(老婆)?$`, fcext.DoOnceOnSuccess(
		func(ctx *zero.Ctx) bool {
			content, err := engine.GetLazyData("wife_list.txt", false)
			if err != nil {
				panic(err)
			}
			/*content, err := os.ReadFile(engine.DataFolder() + "wife_list.txt") // 779分界
			if err != nil {
				log.Println("[vtbwife]读取vtbwife数据文件失败: ", err)
				return false
			}*/
			// 将文件内容转换为单词
			keys = strings.Split(string(content), "\n")
			log.Println("[vtbwife]加载", len(keys), "位wife数据...")
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
		if !ok {
			ctx.SendChain(message.Text("-获取图片链接失败"))
			return
		}
		img, err := web.GetData(u)
		if err != nil {
			ctx.SendChain(message.Text("-获取图片失败惹", err))
			return
		}
		txt := message.Text(
			"\n今天你的VTB老婆是: ", key,
		)
		if id := ctx.SendChain(message.At(ctx.Event.UserID), txt, message.ImageBytes(img), message.Text(b)); id.ID() == 0 {
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
	var b []string
	doc.Find(".moe-infobox").Find("tr").Each(func(i int, s *goquery.Selection) {
		b = append(b, strings.TrimSpace(s.Text()))
	})
	var k int
	for kk, vv := range b {
		if strings.TrimSpace(vv) == "基本资料" || strings.TrimSpace(vv) == "基本信息" || strings.TrimSpace(vv) == "名字" || strings.TrimSpace(vv) == "名称" {
			k = kk + 1
			break
		}
	}
	if k != 0 {
		brief = b[k-1] + "\n"
	}
	for ; k < len(b); k++ {
		brief += strings.Replace(strings.Replace(b[k], "\n", ": ", 1), "\n", "", 1) + "\n"
	}
	brief = strings.TrimSpace(brief)
	return
}
