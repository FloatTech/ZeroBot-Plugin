package jandan

import (
	"github.com/FloatTech/ZeroBot-Plugin/order"
	"github.com/FloatTech/zbputils/process"
	"github.com/antchfx/htmlquery"
	"github.com/fumiama/cron"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

var (
	chanPicture chan string
	pageTotal   int
)

func init() {
	go func() {
		defer order.DoneOnExit()()
		process.SleepAbout1sTo2s()
		log.Println("[jandan/cron] 开启jandan数据库日常更新")
		jandanDaily()
	}()
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
	err := db.Del("picture", "where 1 = 1")
	log.Errorln("[jandan]:", err)
	chanPicture = make(chan string, 100000)
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
	go scorePicture()
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
	id := 1
	for pictureURL := range chanPicture {
		p := picture{
			ID:         uint64(id),
			PictureURL: pictureURL,
		}
		err := db.Insert("picture", &p)
		if err != nil {
			log.Errorln("[jandan]:", err)
		}
		id++
	}
}
