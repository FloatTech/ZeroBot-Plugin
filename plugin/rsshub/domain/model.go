package domain

import (
	"encoding/hex"
	"hash/fnv"
	"time"
)

// ======== RSS ========[START]

// type SingleFeedItem gofeed.Item

func genHashForFeedItem(link, guid string) string {
	idString := link + "||" + guid
	h := fnv.New32()
	_, _ = h.Write([]byte(idString))
	encoded := hex.EncodeToString(h.Sum(nil))
	return encoded
}

// RssClientView 频道视图
type RssClientView struct {
	Source   *RssSource
	Contents []*RssContent
}

// ======== RSS ========[END]

// ======== DB ========[START]

const (
	tableNameRssSource    = "rss_source"
	tableNameRssContent   = "rss_content"
	tableNameRssSubscribe = "rss_subscribe"
)

// RssSource RSS频道
type RssSource struct {
	// Id 自增id
	ID int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	// RssHubFeedPath 频道路由 用于区分rss_hub 不同的频道 例如: `/bangumi/tv/calendar/today`
	RssHubFeedPath string `gorm:"column:rss_hub_feed_path;not null;unique;" json:"rss_hub_feed_path"`
	// Title 频道标题
	Title string `gorm:"column:title"        json:"title"`
	// ChannelDesc 频道描述
	ChannelDesc string `gorm:"column:channel_desc" json:"channel_desc"`
	// ImageURL 频道图片
	ImageURL string `gorm:"column:image_url"    json:"image_url"`
	// Link 频道链接
	Link string `gorm:"column:link"         json:"link"`
	// UpdatedParsed RSS页面更新时间
	UpdatedParsed time.Time `gorm:"column:updated_parsed" json:"updated_parsed"`
	//// Ctime create time
	// Ctime int64 `gorm:"column:ctime;default:current_timestamp"  json:"ctime"`
	// Mtime update time
	Mtime time.Time `gorm:"column:mtime;default:current_timestamp;" json:"mtime"`
}

// TableName ...
func (RssSource) TableName() string {
	return tableNameRssSource
}

// IfNeedUpdate ...
func (r RssSource) IfNeedUpdate(cmp *RssSource) bool {
	if r.Link != cmp.Link {
		return false
	}
	return r.UpdatedParsed.Unix() < cmp.UpdatedParsed.Unix()
}

// RssContent 订阅的RSS频道的推送信息
type RssContent struct {
	// Id 自增id
	ID          int64     `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	HashID      string    `gorm:"column:hash_id;unique"        json:"hash_id"`
	RssSourceID int64     `gorm:"column:rss_source_id;not null"   json:"rss_source_id"`
	Title       string    `gorm:"column:title"       json:"title"`
	Description string    `gorm:"column:description" json:"description"`
	Link        string    `gorm:"column:link"        json:"link"`
	Date        time.Time `gorm:"column:date"        json:"date"`
	Author      string    `gorm:"column:author"      json:"author"`
	Thumbnail   string    `gorm:"column:thumbnail"   json:"thumbnail"`
	Content     string    `gorm:"column:content"     json:"content"`
	//// Ctime create time
	// Ctime int64 `gorm:"column:ctime;default:current_timestamp"  json:"ctime"`
	// Mtime update time
	Mtime time.Time `gorm:"column:mtime;default:current_timestamp;" json:"mtime"`
}

// TableName ...
func (RssContent) TableName() string {
	return tableNameRssContent
}

// RssSubscribe 订阅关系表：群组-RSS频道
type RssSubscribe struct {
	// Id 自增id
	ID int64 `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	// 订阅群组
	GroupID int64 `gorm:"column:group_id;not null;uniqueIndex:uk_sid_gid"`
	// 订阅频道
	RssSourceID int64 `gorm:"column:rss_source_id;not null;uniqueIndex:uk_sid_gid"`
	//// Ctime create time
	// Ctime int64 `gorm:"column:ctime;default:current_timestamp"  json:"ctime"`
	// Mtime update time
	Mtime time.Time `gorm:"column:mtime;default:current_timestamp;" json:"mtime"`
}

// TableName ...
func (RssSubscribe) TableName() string {
	return tableNameRssSubscribe
}

// ======== DB ========[END]
