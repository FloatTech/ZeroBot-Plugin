package jandan

import (
	"github.com/FloatTech/zbputils/process"
	"github.com/antchfx/htmlquery"
	"github.com/fumiama/cron"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

var (
	chanPicture = make(chan string, 100000)
	pageTotal   int
)

func init() {
	go func() {
		process.SleepAbout1sTo2s()
		scorePicture()
		log.Println("[jandan/cron] 开启jandan数据库日常更新")
		jandanDaily()
	}()
	travelWebpage()
}

func jandanDaily() {
	c := cron.New()
	_, err := c.AddFunc("10 4 * * *", func() { travelWebpage() })
	if err != nil {
		log.Errorln("定时任务有错误:", err)
	} else {
		log.Println("开启jandan数据库定时任务")
		c.Start()
	}
}

func travelWebpage() {
	pictureList = pictureList[0:0]
	webpageURL := jandanPictureURL
	doc, err := htmlquery.LoadURL(webpageURL)
	if err != nil {
		log.Errorln("[jandan]:", err)
	}
	re := regexp.MustCompile(`\d+`)
	pageTotal, err = strconv.Atoi(re.FindString(htmlquery.FindOne(doc, "//*[@id='comments']/div[2]/div/span[@class='current-comment-page']/text()").Data))
	if err != nil {
		log.Errorln("[jandan]:", err)
	}
	for i := 0; i < pageTotal; i++ {
		doc, err = htmlquery.LoadURL(webpageURL)
		if err != nil {
			log.Errorln("[jandan]:", err)
		}
		picList, err := htmlquery.QueryAll(doc, "//*[@class='view_img_link']")
		if err != nil {
			log.Errorln("[jandan]:", err)
		}
		if len(picList) != 0 {
			for _, v := range picList {
				chanPicture <- "https:" + v.Attr[0].Val
			}
		}
		if i != pageTotal-1 {
			webpageURL = "https:" + htmlquery.FindOne(doc, "//*[@id='comments']/div[@class='comments']/div[@class='cp-pagenavi']/a[@class='previous-comment-page']").Attr[1].Val
		}
	}
}

func scorePicture() {
	for pictureURL := range chanPicture {
		pictureList = append(pictureList, pictureURL)
	}
}
