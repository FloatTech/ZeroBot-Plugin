// Package maofly 漫画猫漫画
package maofly

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/zbputils/binary"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/web"
	LZString "github.com/Lazarus/lz-string-go"
	"github.com/antchfx/htmlquery"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"golang.org/x/net/html"
)

var (
	dbpath        string
	imgPre        = "https://mao.mhtupian.com/uploads/"
	searchURL     = "https://www.maofly.com/search.html?q="
	re            = regexp.MustCompile(`let img_data = "(.*?)"`)
	chanTask      chan string
	chanImageUrls chan string
	waitGroup     sync.WaitGroup
	files         []string
	ua            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36"
)

func init() {
	engine := control.Register("maofly", &control.Options{
		DisableOnDefault:  false,
		Help:              "漫画猫\n- 漫画猫[xxx]",
		PrivateDataFolder: "maofly",
	}).ApplySingle(ctxext.DefaultSingle)
	dbpath = engine.DataFolder()
	engine.OnRegex(`^漫画猫\s?(.{1,25})$`, zero.OnlyGroup, getPara).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("少女祈祷中..."))
			indexURL := ctx.State["index_url"].(string)
			title, c, err := getChapter(indexURL)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			if len(c) == 0 || err != nil {
				ctx.SendChain(message.Text(title, "已下架"))
				return
			}
			files = make([]string, 0)
			zipName := dbpath + title + ".zip"
			if file.IsExist(zipName) {
				err := unzip(zipName, ".")
				if err != nil {
					_ = os.RemoveAll(title)
					ctx.SendChain(message.Text("ERROR:", err))
					return
				}
			} else {
				_ = os.MkdirAll(title, 0755)
			}
			chanTask = make(chan string, len(c))
			chanImageUrls = make(chan string, 1000000)
			for i := 0; i < len(c); i++ {
				waitGroup.Add(1)
				key := fmt.Sprintf("%s|%d|%d|%s|%s", title, i, c[i].dataSort, c[i].href, c[i].title)
				go getImgs(key)
			}
			waitGroup.Add(1)
			go checkOK(len(c))

			for i := 0; i < 5; i++ {
				waitGroup.Add(1)
				go downloadFile()
			}
			waitGroup.Wait()

			if err := zipFiles(zipName, files); err != nil {
				_ = os.RemoveAll(title)
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			fmt.Println("Zipped File:", zipName)
			_ = os.RemoveAll(title)
			_ = ctx.CallAction("upload_group_file", zero.Params{"group_id": ctx.Event.GroupID, "file": file.BOTPATH + "/" + zipName, "name": title})
		})
}

func getPara(ctx *zero.Ctx) bool {
	next := zero.NewFutureEvent("message", 999, false, ctx.CheckSession())
	recv, cancel := next.Repeat()
	keyword := ctx.State["regex_matched"].([]string)[1]
	text, a := search(keyword)
	if len(a) == 0 {
		ctx.SendChain(message.Text("没有找到与", keyword, "有关的漫画"))
		return false
	}
	text += ",请输入下载漫画序号:\n"
	for i := 0; i < len(a); i++ {
		text += fmt.Sprintf("%d. [%s]%s\n", i, a[i].title, a[i].href)
	}
	ctx.SendChain(message.Text(text))
	for {
		select {
		case <-time.After(time.Second * 120):
			ctx.SendChain(message.Text("漫画猫指令过期"))
			cancel()
			return false
		case c := <-recv:
			msg := c.Event.Message.ExtractPlainText()
			num, err := strconv.Atoi(msg)
			if err != nil || num < 0 || num >= len(a) {
				ctx.SendChain(message.Text("请输入有效的数字!"))
				continue
			}
			cancel()
			ctx.State["index_url"] = a[num].href
			return true
		}
	}
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
	data, err := web.RequestDataWith(web.NewDefaultClient(), requestURL, "GET", "", ua)
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
	data, err := web.RequestDataWith(web.NewDefaultClient(), indexURL, "GET", "", ua)
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

func checkOK(total int) {
	count := 0
	for {
		_ = <-chanTask
		count++
		if count == total {
			close(chanImageUrls)
			break
		}
	}
	waitGroup.Done()
}

func getImgs(key string) {
	keys := strings.Split(key, "|")
	var data []byte
	data, err := web.RequestDataWith(web.NewDefaultClient(), keys[3], "GET", "", ua)
	for i := 1; err != nil && i <= 10; i++ {
		log.Errorln("[maofly]", err, ",", i, "s后重试")
		time.Sleep(time.Duration(i) * time.Second)
		data, err = web.RequestDataWith(web.NewDefaultClient(), keys[3], "GET", "", ua)
	}
	s := re.FindStringSubmatch(binary.BytesToString(data))[1]
	d, _ := LZString.Decompress(s, "")
	imgs := strings.Split(d, ",")
	for i := 0; i < len(imgs); i++ {
		dir := fmt.Sprintf("%s/%04s %s", keys[0], keys[1], keys[4])
		_ = os.MkdirAll(dir, 0755)
		fileURL := imgPre + path.Dir(imgs[i]) + "/" + strings.ReplaceAll(url.QueryEscape(path.Base(imgs[i])), "+", "%20")
		filePath := fmt.Sprintf("%s/%s-%s-%03d%s", dir, keys[0], keys[1], i+1, path.Ext(imgs[i]))
		dkey := filePath + "|" + fileURL
		chanImageUrls <- dkey
	}
	chanTask <- key
	waitGroup.Done()
}

func downloadFile() {
	for dkey := range chanImageUrls {
		if dkey == "" {
			continue
		}
		filePath := strings.Split(dkey, "|")[0]
		fileURL := strings.Split(dkey, "|")[1]
		if file.IsExist(filePath) {
			files = append(files, filePath)
			continue
		}
		data, err := web.RequestDataWith(web.NewDefaultClient(), fileURL, "GET", "", ua)
		if err != nil {
			data = binary.StringToBytes(fileURL)
			filePath = strings.ReplaceAll(filePath, path.Ext(filePath), ".txt")
			log.Errorln("[maofly]", err)
		}
		err = os.WriteFile(filePath, data, 0666)
		if err != nil {
			log.Errorln("[maofly]", err)
		}
		files = append(files, filePath)
	}
	waitGroup.Done()
}

func zipFiles(filename string, files []string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filename

	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
