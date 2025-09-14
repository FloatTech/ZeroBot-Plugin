package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/FloatTech/floatbox/web"
	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

// const (
//	acceptHeader = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
//	userHeader   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 Edg/84.0.522.63"
//)

var (
	// RSSHubMirrors RSSHub镜像站地址列表，第一个为默认地址
	rssHubMirrors = []string{
		"https://rsshub.rssforever.com",
		"https://rss.injahow.cn",
	}
)

// RssHubClient rss hub client (http)
type RssHubClient struct {
	*http.Client
}

// FetchFeed 获取rss feed信息
func (c *RssHubClient) FetchFeed(path string) (feed *gofeed.Feed, err error) {
	var data []byte
	// 遍历 rssHubMirrors，直到获取成功
	for _, mirror := range rssHubMirrors {
		data, err = web.RequestDataWith(c.Client, mirror+path, "GET", "", web.RandUA(), nil)
		if err == nil && len(data) > 0 {
			break
		}
	}
	if err != nil {
		logrus.Errorf("[rsshub FetchFeed] fetch feed error: %v", err)
		return nil, err
	}
	if len(data) == 0 {
		logrus.Errorf("[rsshub FetchFeed] fetch feed error: data is empty")
		return nil, errors.New("feed data is empty")
	}
	// data, err = web.RequestDataWith(c.Client, domain+path, "GET", "", web.RandUA(), nil)
	// if err != nil {
	//	return nil, err
	//}
	feed, err = gofeed.NewParser().Parse(bytes.NewBuffer(data))
	if err != nil {
		return
	}
	return
}

func convertFeedToRssView(channelID int64, cPath string, feed *gofeed.Feed) (view *RssClientView) {
	var imgURL string
	if feed.Image != nil {
		imgURL = feed.Image.URL
	}
	view = &RssClientView{
		Source: &RssSource{
			ID:             channelID,
			RssHubFeedPath: cPath,
			Title:          feed.Title,
			ChannelDesc:    feed.Description,
			ImageURL:       imgURL,
			Link:           feed.Link,
			UpdatedParsed:  *(feed.UpdatedParsed),
			Mtime:          time.Now(),
		},
		// 不用定长，后面可能会过滤一些元素再append
		Contents: []*RssContent{},
	}
	// convert feed items to rss content
	for _, item := range feed.Items {
		if item.Link == "" || item.Title == "" {
			continue
		}
		var thumbnail string
		if item.Image != nil {
			thumbnail = item.Image.URL
		}
		var publishedParsed = item.PublishedParsed
		if publishedParsed == nil {
			publishedParsed = &time.Time{}
		}
		aus, _ := json.Marshal(item.Authors)
		view.Contents = append(view.Contents, &RssContent{
			ID:          0,
			HashID:      genHashForFeedItem(item.Link, item.GUID),
			RssSourceID: channelID,
			Title:       item.Title,
			Description: item.Description,
			Link:        item.Link,
			Date:        *publishedParsed,
			Author:      string(aus),
			Thumbnail:   thumbnail,
			Content:     item.Content,
			Mtime:       time.Now(),
		})
	}
	return
}
