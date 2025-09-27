package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// repoStorage db struct for rss
type repoStorage struct {
	orm *gorm.DB
}

// initDB ...
func (s *repoStorage) initDB() (err error) {
	err = s.orm.AutoMigrate(&RssSource{}, &RssContent{}, &RssSubscribe{}).Error
	if err != nil {
		logrus.Errorf("[rsshub initDB] error: %v", err)
		return err
	}
	return nil
	// s.orm.LogMode(true)
}

// GetSubscribesBySource Impl
func (s *repoStorage) GetSubscribesBySource(ctx context.Context, feedPath string) ([]*RssSubscribe, error) {
	logrus.WithContext(ctx).Infof("[rsshub GetSubscribesBySource] feedPath: %s", feedPath)
	rs := make([]*RssSubscribe, 0)
	err := s.orm.Model(&RssSubscribe{}).Joins(fmt.Sprintf("%s left join %s on %s.rss_source_id=%s.id", tableNameRssSubscribe, tableNameRssSource, tableNameRssSubscribe, tableNameRssSource)).
		Where("rss_source.rss_hub_feed_path = ?", feedPath).Select("rss_subscribe.*").Find(&rs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.WithContext(ctx).Errorf("[rsshub GetSubscribesBySource] error: %v", err)
		return nil, err
	}
	return rs, nil
}

// GetIfExistedSubscribe Impl
func (s *repoStorage) GetIfExistedSubscribe(ctx context.Context, gid int64, feedPath string) (*RssSubscribe, bool, error) {
	rs := RssSubscribe{}

	err := s.orm.Table(tableNameRssSubscribe).
		Select("rss_subscribe.id, rss_subscribe.group_id, rss_subscribe.rss_source_id, rss_subscribe.mtime").
		Joins(fmt.Sprintf("INNER JOIN %s ON %s.rss_source_id=%s.id",
			tableNameRssSource, tableNameRssSubscribe, tableNameRssSource)).
		Where("rss_source.rss_hub_feed_path = ? AND rss_subscribe.group_id = ?", feedPath, gid).Scan(&rs).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		logrus.WithContext(ctx).Errorf("[rsshub GetIfExistedSubscribe] error: %v", err)
		return nil, false, err
	}
	if rs.ID == 0 {
		return nil, false, nil
	}
	return &rs, true, nil
}

// ==================== RepoSource ==================== [Start]

// UpsertSource Impl
func (s *repoStorage) UpsertSource(ctx context.Context, source *RssSource) (err error) {
	// Update columns to default value on `id` conflict
	querySource := &RssSource{RssHubFeedPath: source.RssHubFeedPath}
	err = s.orm.First(querySource, "rss_hub_feed_path = ?", querySource.RssHubFeedPath).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = s.orm.Create(source).Omit("id").Error
			if err != nil {
				logrus.WithContext(ctx).Errorf("[rsshub] add source error: %v", err)
				return
			}
		}
		return
	}
	source.ID = querySource.ID
	logrus.WithContext(ctx).Infof("[rsshub] update source: %+v", source.UpdatedParsed)
	err = s.orm.Model(&source).Where(&RssSource{ID: source.ID}).
		Updates(&RssSource{
			Title:         source.Title,
			ChannelDesc:   source.ChannelDesc,
			ImageURL:      source.ImageURL,
			Link:          source.Link,
			UpdatedParsed: source.UpdatedParsed,
			Mtime:         time.Now(),
		}).Error
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub] update source error: %v", err)
		return
	}
	logrus.Println("[rsshub] add source success: ", source.ID)
	return nil
}

// GetSources Impl
func (s *repoStorage) GetSources(ctx context.Context) (sources []RssSource, err error) {
	sources = []RssSource{}
	err = s.orm.Find(&sources, "id > 0").Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("source not found")
		}
		logrus.WithContext(ctx).Errorf("[rsshub] get sources error: %v", err)
		return
	}
	logrus.WithContext(ctx).Infof("[rsshub] get sources success: %d", len(sources))
	return
}

// GetSourceByRssHubFeedLink Impl
func (s *repoStorage) GetSourceByRssHubFeedLink(ctx context.Context, rssHubFeedLink string) (source *RssSource, err error) {
	source = &RssSource{RssHubFeedPath: rssHubFeedLink}
	err = s.orm.Take(source, source).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.WithContext(ctx).Errorf("[rsshub] get source error: %v", err)
		return
	}
	return
}

// DeleteSource Impl
func (s *repoStorage) DeleteSource(ctx context.Context, fID int64) (err error) {
	err = s.orm.Delete(&RssSource{}, "id = ?", fID).Error
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub] storage.DeleteSource: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("source not found")
		}
		return
	}
	return nil
}

// ==================== RepoSource ==================== [End]

// ==================== RepoContent ==================== [Start]

// UpsertContent Impl
func (s *repoStorage) UpsertContent(ctx context.Context, content *RssContent) (err error) {
	// check params
	if content == nil {
		err = errors.New("content is nil")
		return
	}
	// check params.RssHubFeedPath and params.HashID
	if content.RssSourceID < 0 || content.HashID == "" || content.Title == "" {
		err = errors.New("content.RssSourceID or content.HashID or content.Title is empty")
		return
	}
	err = s.orm.Create(content).Omit("id").Error
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub] storage.UpsertContent: %v", err)
		return
	}
	return
}

// DeleteSourceContents Impl
func (s *repoStorage) DeleteSourceContents(ctx context.Context, channelID int64) (rows int64, err error) {
	err = s.orm.Delete(&RssSubscribe{}).Where(&RssSubscribe{RssSourceID: channelID}).Error
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub] storage.DeleteSourceContents: %v", err)
		return
	}
	return
}

// IsContentHashIDExist Impl
func (s *repoStorage) IsContentHashIDExist(ctx context.Context, hashID string) (bool, error) {
	wanted := &RssContent{HashID: hashID}
	err := s.orm.Take(wanted, wanted).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logrus.WithContext(ctx).Errorf("[rsshub] storage.IsContentHashIDExist: %v", err)
		return false, err
	}
	return true, nil
}

// ==================== RepoContent ==================== [End]

// ==================== RepoSubscribe ==================== [Start]

// CreateSubscribe Impl
func (s *repoStorage) CreateSubscribe(ctx context.Context, gid, rssSourceID int64) (err error) {
	// check subscribe
	if rssSourceID < 0 || gid == 0 {
		err = errors.New("gid or rssSourceID is empty")
		return
	}
	err = s.orm.Create(&RssSubscribe{GroupID: gid, RssSourceID: rssSourceID}).Omit("id").Error
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub] storage.CreateSubscribe: %v", err)
		return
	}
	return
}

// DeleteSubscribe Impl
func (s *repoStorage) DeleteSubscribe(ctx context.Context, subscribeID int64) (err error) {
	err = s.orm.Delete(&RssSubscribe{}, "id = ?", subscribeID).Error
	if err != nil {
		logrus.WithContext(ctx).Errorf("[rsshub] storage.DeleteSubscribe error: %v", err)
		return
	}
	return
}

// GetSubscribeByID Impl
func (s *repoStorage) GetSubscribeByID(ctx context.Context, gid int64, subscribeID int64) (res *RssSubscribe, err error) {
	res = &RssSubscribe{}
	err = s.orm.First(res, &RssSubscribe{GroupID: gid, RssSourceID: subscribeID}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		logrus.WithContext(ctx).Errorf("[rsshub] storage.GetSubscribeByID: %v", err)
		return nil, err
	}
	return
}

// GetSubscribedChannelsByGroupID Impl
func (s *repoStorage) GetSubscribedChannelsByGroupID(ctx context.Context, gid int64) (res []*RssSource, err error) {
	res = make([]*RssSource, 0)
	err = s.orm.Model(&RssSource{}).
		Joins(fmt.Sprintf("join %s on rss_source_id=%s.id", tableNameRssSubscribe, tableNameRssSource)).Where("rss_subscribe.group_id = ?", gid).
		Select("rss_source.*").
		Find(&res).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		logrus.WithContext(ctx).Errorf("[rsshub] storage.GetSubscribedChannelsByGroupID: %v", err)
		return
	}
	return
}

// GetSubscribes Impl
func (s *repoStorage) GetSubscribes(ctx context.Context) (res []*RssSubscribe, err error) {
	res = make([]*RssSubscribe, 0)
	err = s.orm.Find(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
			return
		}
		logrus.WithContext(ctx).Errorf("[rsshub] storage.GetSubscribes: %v", err)
		return
	}
	return
}

// ==================== RepoSubscribe ==================== [End]
