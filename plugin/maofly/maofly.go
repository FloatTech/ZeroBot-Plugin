// Package maofly 漫画猫漫画
package maofly

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/web"
	lzString "github.com/Lazarus/lz-string-go"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/net/html"
)

var (
	imagePre  = "https://mao.mhtupian.com/uploads/"
	searchURL = "https://www.maofly.com/search.html?q="
	re        = regexp.MustCompile(`let img_data = "(.*?)"`)
	ua        = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"
	authority = "mao.mhtupian.com"
	referer   = "https://www.maofly.com/"
)

func init() {
	engine := control.Register("maofly", &control.Options{
		DisableOnDefault:  false,
		Help:              "漫画猫\n- 漫画猫[xxx]",
		PrivateDataFolder: "maofly",
	})
	cachePath := engine.DataFolder() + "cache/"
	go func() {
		_ = os.RemoveAll(cachePath)
		_ = os.MkdirAll(cachePath, 0755)
	}()
	engine.OnRegex(`^漫画猫\s?(.{1,25})$`, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
			recv, cancel := next.Repeat()
			defer cancel()
			keyword := ctx.State["regex_matched"].([]string)[1]
			searchText, a := search(keyword)
			if len(a) == 0 {
				ctx.SendChain(message.Text("没有找到与", keyword, "有关的漫画"))
				return
			}
			imageBytes, err := text.RenderToBase64(getIndexText(searchText, a), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(imageBytes))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR:可能被风控了"))
			}
			step := 0
			errorCount := 0
			var title string
			var cs chapterSlice
			for {
				select {
				case <-time.After(time.Second * 120):
					ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("漫画猫指令过期"))
					return
				case c := <-recv:
					if errorCount >= 3 {
						ctx.SendChain(message.Reply(c.Event.MessageID), message.Text("输入错误太多,请重新发指令"))
						return
					}
					msg := c.Event.Message.ExtractPlainText()
					num, err := strconv.Atoi(msg)
					if err != nil {
						ctx.SendChain(message.Text("请输入数字!"))
						errorCount++
						continue
					}
					switch step {
					case 0:
						if num < 0 || num >= len(a) {
							imageBytes, err := text.RenderToBase64("漫画序号非法!请重新选择!\n"+getIndexText(searchText, a), text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR:", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(imageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR:可能被风控了"))
							}
							errorCount++
							continue
						}
						title, cs, err = getChapter(a[num].href)
						if err != nil {
							ctx.SendChain(message.Text("ERROR:", err))
							return
						}
						if len(cs) == 0 {
							imageBytes, err := text.RenderToBase64(title+"已下架!请重新选择!\n"+getIndexText(searchText, a), text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR:", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(imageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR:可能被风控了"))
							}
							errorCount++
							step = 0
							continue
						}
						imageBytes, err := text.RenderToBase64(getChapterText(cs), text.FontFile, 400, 20)
						if err != nil {
							ctx.SendChain(message.Text("ERROR:", err))
							return
						}
						if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(imageBytes))); id.ID() == 0 {
							ctx.SendChain(message.Text("ERROR:可能被风控了"))
						}
					case 1:
						if num < 0 || num >= len(cs) {
							imageBytes, err := text.RenderToBase64("章节序号非法!请重新选择!\n"+getChapterText(cs), text.FontFile, 400, 20)
							if err != nil {
								ctx.SendChain(message.Text("ERROR:", err))
								return
							}
							if id := ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Image("base64://"+helper.BytesToString(imageBytes))); id.ID() == 0 {
								ctx.SendChain(message.Text("ERROR:可能被风控了"))
							}
							errorCount++
							continue
						}
						data, err := web.RequestDataWith(web.NewDefaultClient(), cs[num].href, "GET", referer, ua)
						for err != nil {
							ctx.SendChain(message.Text("ERROR:", err))
							return
						}
						m := message.Message{ctxext.FakeSenderForwardNode(ctx, message.Text(title, ",", cs[num].title))}
						s := re.FindStringSubmatch(binary.BytesToString(data))[1]
						d, _ := lzString.Decompress(s, "")
						images := strings.Split(d, ",")
						for i := range images {
							imageURL := imagePre + path.Dir(images[i]) + "/" + strings.ReplaceAll(url.QueryEscape(path.Base(images[i])), "+", "%20")
							imagePath := cachePath + fmt.Sprintf("%s-%s-%d%s", title, cs[num].title, i, path.Ext(images[i]))
							err = initImage(imagePath, imageURL)
							if !file.IsExist(imagePath) && err != nil {
								m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image(imageURL)))
							} else {
								m = append(m, ctxext.FakeSenderForwardNode(ctx, message.Image("file:///"+file.BOTPATH+"/"+imagePath)))
							}
						}
						if id := ctx.SendGroupForwardMessage(
							ctx.Event.GroupID,
							m).Get("message_id").Int(); id == 0 {
							ctx.SendChain(message.Text("ERROR:可能被风控或下载图片用时过长，请耐心等待"))
						}
						return
					default:
						return
					}
					step++
				}
			}

		})
}

type chapter struct {
	dataSort int
	href     string
	title    string
}

type chapterSlice []chapter

func (c chapterSlice) Len() int {
	return len(c)
}
func (c chapterSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c chapterSlice) Less(i, j int) bool {
	return c[i].dataSort > c[j].dataSort
}

type a struct {
	href  string
	title string
}

func search(key string) (text string, al []a) {
	requestURL := searchURL + url.QueryEscape(key)
	data, err := web.RequestDataWith(web.NewDefaultClient(), requestURL, "GET", referer, ua)
	if err != nil {
		log.Errorln("[maofly]", err)
		return
	}
	doc, err := htmlquery.Parse(bytes.NewReader(data))
	if err != nil {
		log.Errorln("[maofly]", err)
		return
	}
	text = htmlquery.FindOne(doc, "//div[@class=\"text-muted\"]/text()").Data
	list, err := htmlquery.QueryAll(doc, "//h2[@class=\"mt-0 mb-1 one-line\"]/a")
	if err != nil {
		log.Errorln("[maofly]", err)
		return
	}
	al = make([]a, len(list))
	for i := 0; i < len(list); i++ {
		al[i].href = list[i].Attr[0].Val
		al[i].title = list[i].Attr[1].Val
	}
	return
}

func getChapter(indexURL string) (title string, c chapterSlice, err error) {
	data, err := web.RequestDataWith(web.NewDefaultClient(), indexURL, "GET", referer, ua)
	if err != nil {
		return
	}
	doc, err := htmlquery.Parse(bytes.NewReader(data))
	if err != nil {
		return
	}
	title = htmlquery.FindOne(doc, "//meta[@property=\"og:novel:book_name\"]").Attr[1].Val
	list, err := htmlquery.QueryAll(doc, "//*[@id=\"comic-book-list\"]/div/ol/li")
	if err != nil {
		return
	}
	c = make(chapterSlice, len(list))
	for i := 0; i < len(list); i++ {
		c[i].dataSort, _ = strconv.Atoi(list[i].Attr[1].Val)
		var node *html.Node
		node, err = htmlquery.Query(list[i], "//a")
		if err != nil {
			return
		}
		c[i].href = node.Attr[1].Val
		c[i].title = node.Attr[2].Val
	}
	sort.Sort(c)
	return
}

func getIndexText(pre string, a []a) (text string) {
	text += pre + ",请输入下列漫画序号:\n"
	for i := 0; i < len(a); i++ {
		text += fmt.Sprintf("%d. %s\n", i, a[i].title)
	}
	return
}

func getChapterText(cs chapterSlice) (text string) {
	text = "请输入下列章节序号:\n"
	for i := 0; i < len(cs); i++ {
		text += fmt.Sprintf("%d. %s\n", i, cs[i].title)
	}
	return text
}

func initImage(imagePath, imageURL string) error {
	if file.IsNotExist(imagePath) {
		client := web.NewDefaultClient()
		req, _ := http.NewRequest("GET", imageURL, nil)
		req.Header.Set("User-Agent", ua)
		req.Header.Add("Referer", referer)
		req.Header.Add("Authority", authority)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			s := fmt.Sprintf("status code: %d", resp.StatusCode)
			err = errors.New(s)
			return err
		}
		err = os.WriteFile(imagePath, data, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}
