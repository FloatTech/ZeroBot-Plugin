// Package domain rsshub领域逻辑
package domain

import (
	"context"

	"github.com/mmcdole/gofeed"
	"github.com/sirupsen/logrus"
)

// syncRss 同步所有频道
// 返回：更新的频道&订阅信息 map[int64]*RssClientView
// 1. 获取所有频道
// 2. 遍历所有频道，检查频道是否更新
// 3. 如果更新，获取更新的内容，但是返回的数据
func (repo *RssDomain) syncRss(ctx context.Context) (updated map[int64]*RssClientView, err error) {
	updated = make(map[int64]*RssClientView)
	// 获取所有频道
	sources, err := repo.storage.GetSources(ctx)
	if err != nil {
		return
	}
	// 遍历所有源，获取每个channel对应的rss内容
	rssView := make([]*RssClientView, len(sources))
	for i, channel := range sources {
		var feed *gofeed.Feed
		// 从site获取rss内容
		feed, err = repo.rssHubClient.FetchFeed(channel.RssHubFeedPath)
		// 如果获取失败，则跳过
		if err != nil {
			logrus.WithContext(ctx).Errorf("[rsshub syncRss] fetch path(%+v) error: %v", channel.RssHubFeedPath, err)
			continue
		}
		rv := convertFeedToRssView(0, channel.RssHubFeedPath, feed)
		rssView[i] = rv
	}
	// 检查频道是否更新
	for _, cv := range rssView {
		if cv == nil {
			continue
		}
		var needUpdate bool
		needUpdate, err = repo.checkSourceNeedUpdate(ctx, cv.Source)
		if err != nil {
			logrus.WithContext(ctx).Errorf("[rsshub syncRss] checkSourceNeedUpdate error: %v", err)
			err = nil
			continue
		}
		// 保存
		logrus.WithContext(ctx).Infof("[rsshub syncRss] cv %+v, need update(real): %v", cv.Source, needUpdate)
		// 如果需要更新，更新channel 和 content
		if needUpdate {
			err = repo.storage.UpsertSource(ctx, cv.Source)
			if err != nil {
				logrus.WithContext(ctx).Errorf("[rsshub syncRss] upsert source error: %v", err)
			}
		}
		var updateChannelView = &RssClientView{Source: cv.Source, Contents: []*RssContent{}}
		err = repo.processContentsUpdate(ctx, cv, updateChannelView)
		if err != nil {
			logrus.WithContext(ctx).Errorf("[rsshub syncRss] processContentsUpdate error: %v", err)
			continue
		}
		if len(updateChannelView.Contents) == 0 {
			logrus.WithContext(ctx).Infof("[rsshub syncRss] cv %s, no new content", cv.Source.RssHubFeedPath)
			continue
		}
		updateChannelView.Sort()
		updated[updateChannelView.Source.ID] = updateChannelView
		logrus.WithContext(ctx).Debugf("[rsshub syncRss] cv %s, new contents: %v", cv.Source.RssHubFeedPath, len(updateChannelView.Contents))
	}
	return
}

// checkSourceNeedUpdate 检查频道是否需要更新
func (repo *RssDomain) checkSourceNeedUpdate(ctx context.Context, source *RssSource) (needUpdate bool, err error) {
	var sourceInDB *RssSource
	sourceInDB, err = repo.storage.GetSourceByRssHubFeedLink(ctx, source.RssHubFeedPath)
	if err != nil {
		return
	}
	if sourceInDB == nil {
		logrus.WithContext(ctx).Errorf("[rsshub syncRss] source not found: %v", source.RssHubFeedPath)
		return
	}
	source.ID = sourceInDB.ID
	// 检查是否需要更新到db
	if sourceInDB.IfNeedUpdate(source) {
		needUpdate = true
	}
	return
}

// processContentsUpdate 处理内容(s)更新
func (repo *RssDomain) processContentsUpdate(ctx context.Context, cv *RssClientView, updateChannelView *RssClientView) error {
	var err error
	for _, content := range cv.Contents {
		if content == nil {
			continue
		}
		content.RssSourceID = cv.Source.ID
		var existed bool
		existed, err = repo.processContentItemUpdate(ctx, content)
		if err != nil {
			logrus.WithContext(ctx).Errorf("[rsshub syncRss] upsert content error: %v", err)
			err = nil
			continue
		}
		if !existed {
			updateChannelView.Contents = append(updateChannelView.Contents, content)
			logrus.WithContext(ctx).Infof("[rsshub syncRss] cv %s, add new content: %v", cv.Source.RssHubFeedPath, content.Title)
		}
	}
	return err
}

// processContentItemUpdate 处理单个内容更新
func (repo *RssDomain) processContentItemUpdate(ctx context.Context, content *RssContent) (existed bool, err error) {
	existed, err = repo.storage.IsContentHashIDExist(ctx, content.HashID)
	if err != nil {
		return
	}
	// 不需要更新&不需要发送
	if existed {
		return
	}
	// 保存
	err = repo.storage.UpsertContent(ctx, content)
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub syncRss] upsert content error: %v", err)
		return
	}
	return
}
