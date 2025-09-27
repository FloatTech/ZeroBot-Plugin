package domain

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// RssDomain RssRepo定义
type RssDomain struct {
	storage      *repoStorage
	rssHubClient *RssHubClient
}

// NewRssDomain 新建RssDomain，调用方保证单例模式
func NewRssDomain(dbPath string) (*RssDomain, error) {
	return newRssDomain(dbPath)
}

func newRssDomain(dbPath string) (*RssDomain, error) {
	if _, err := os.Stat(dbPath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		defer f.Close()
	}
	orm, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		logrus.Errorf("[rsshub NewRssDomain] open db error: %v", err)
		panic(err)
	}
	repo := &RssDomain{
		storage:      &repoStorage{orm: orm},
		rssHubClient: &RssHubClient{Client: http.DefaultClient},
	}
	err = repo.storage.initDB()
	if err != nil {
		logrus.Errorf("[rsshub NewRssDomain] open db error: %v", err)
		panic(err)
	}
	return repo, nil
}

// Subscribe QQ群订阅Rss频道
func (repo *RssDomain) Subscribe(ctx context.Context, gid int64, feedPath string) (
	rv *RssClientView, isChannelExisted, isSubExisted bool, err error) {
	// 验证
	feed, err := repo.rssHubClient.FetchFeed(feedPath)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] add source error: %v", err)
		return
	}
	logrus.WithContext(ctx).Infof("[rsshub Subscribe] try get source success: %v", len(feed.Title))
	// 新建source结构体
	rv = convertFeedToRssView(0, feedPath, feed)
	feedChannel, err := repo.storage.GetSourceByRssHubFeedLink(ctx, feedPath)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] query source by feedPath error: %v", err)
		return
	}
	// 如果已经存在
	if feedChannel != nil {
		logrus.WithContext(ctx).Warningf("[rsshub Subscribe] source existed: %v", feedChannel)
		isChannelExisted = true
	} else {
		// 不存在的情况，要把更新时间置空，保证下一次同步时能够更新
		rv.Source.UpdatedParsed = time.Time{}
	}
	// 保存
	err = repo.storage.UpsertSource(ctx, rv.Source)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] save source error: %v", err)
		return
	}
	logrus.Infof("[rsshub Subscribe] save/update source success %v", rv.Source.ID)
	// 添加群号到订阅
	subscribe, err := repo.storage.GetSubscribeByID(ctx, gid, rv.Source.ID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] query subscribe error: %v", err)
		return
	}
	logrus.WithContext(ctx).Infof("[rsshub Subscribe] query subscribe success: %v", subscribe)
	// 如果已经存在，直接返回
	if subscribe != nil {
		isSubExisted = true
		logrus.WithContext(ctx).Infof("[rsshub Subscribe] subscribe existed: %v", subscribe)
		return
	}
	// 如果不存在，保存
	err = repo.storage.CreateSubscribe(ctx, gid, rv.Source.ID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] save subscribe error: %v", err)
		return
	}
	logrus.WithContext(ctx).Infof("[rsshub Subscribe] success: %v", len(rv.Contents))
	return
}

// Unsubscribe 群组取消订阅
func (repo *RssDomain) Unsubscribe(ctx context.Context, gid int64, feedPath string) (err error) {
	existedSubscribes, ifExisted, err := repo.storage.GetIfExistedSubscribe(ctx, gid, feedPath)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] query sub by route error: %v", err)
		return errors.New("数据库错误")
	}
	logrus.WithContext(ctx).Infof("[rsshub Subscribe] query source by route success: %v", existedSubscribes)
	// 如果不存在订阅关系，直接返回
	if !ifExisted || existedSubscribes == nil {
		logrus.WithContext(ctx).Infof("[rsshub Subscribe] source existed: %v", ifExisted)
		return errors.New("频道不存在")
	}
	err = repo.storage.DeleteSubscribe(ctx, existedSubscribes.ID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] delete source error: %v", err)
		return errors.New("删除失败")
	}
	// 查询是否还有群订阅这个频道
	subscribesNeedsToDel, err := repo.storage.GetSubscribesBySource(ctx, feedPath)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Subscribe] query source by route error: %v", err)
		return
	}
	// 没有群订阅的时候，把频道删除
	if len(subscribesNeedsToDel) == 0 {
		err = repo.storage.DeleteSource(ctx, existedSubscribes.RssSourceID)
		if err != nil {
			logrus.WithContext(ctx).Errorf("[rsshub Subscribe] delete source error: %v", err)
			return errors.New("清除频道信息失败")
		}
	}
	return
}

// GetSubscribedChannelsByGroupID 获取群对应的订阅的频道信息
func (repo *RssDomain) GetSubscribedChannelsByGroupID(ctx context.Context, gid int64) ([]*RssClientView, error) {
	channels, err := repo.storage.GetSubscribedChannelsByGroupID(ctx, gid)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub GetSubscribedChannelsByGroupID] GetSubscribedChannelsByGroupID error: %v", err)
		return nil, err
	}
	rv := make([]*RssClientView, len(channels))
	logrus.WithContext(ctx).Infof("[rsshub GetSubscribedChannelsByGroupID] query subscribe success: %v", len(channels))
	for i, cn := range channels {
		rv[i] = &RssClientView{
			Source: cn,
		}
	}
	return rv, nil
}

// Sync 同步任务，按照群组订阅情况做好map切片
func (repo *RssDomain) Sync(ctx context.Context) (groupView map[int64][]*RssClientView, err error) {
	groupView = make(map[int64][]*RssClientView)
	// 获取所有Rss频道
	// 获取所有频道
	updatedViews, err := repo.syncRss(ctx)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Sync] sync rss feed error: %v", err)
		return
	}
	logrus.WithContext(ctx).Infof("[rsshub Sync] updated channels: %v", len(updatedViews))
	subscribes, err := repo.storage.GetSubscribes(ctx)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub Sync] get subscribes error: %v", err)
		return
	}
	for _, subscribe := range subscribes {
		groupView[subscribe.GroupID] = append(groupView[subscribe.GroupID], updatedViews[subscribe.RssSourceID])
	}
	return
}
