package domain

import (
	"bytes"
	"encoding/json"
	"github.com/FloatTech/floatbox/web"
	"github.com/mmcdole/gofeed"
	"net/http"
	"time"
)

//const (
//	acceptHeader = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
//	userHeader   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.135 Safari/537.36 Edg/84.0.522.63"
//)

// RssHubClient rss hub client (http)
type RssHubClient struct {
	*http.Client
}

// FetchFeed 获取rss feed信息
func (c *RssHubClient) FetchFeed(domain, path string) (feed *gofeed.Feed, err error) {
	var data []byte
	data, err = web.RequestDataWith(c.Client, domain+path, "GET", "", web.RandUA(), nil)
	if err != nil {
		return nil, err
	}
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
