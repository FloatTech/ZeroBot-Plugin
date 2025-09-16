package domain

import "context"

// RepoContent RSS 推送信息存储接口
type RepoContent interface {
	// UpsertContent 添加一条文章
	UpsertContent(ctx context.Context, content *RssContent) error
	// DeleteSourceContents 删除订阅源的所有文章，返回被删除的文章数
	DeleteSourceContents(ctx context.Context, channelID int64) (int64, error)
	// IsContentHashIDExist hash id 对应的文章是否已存在
	IsContentHashIDExist(ctx context.Context, hashID string) (bool, error)
}

// RepoSource RSS 订阅源存储接口
type RepoSource interface {
	// UpsertSource 添加一个订阅源
	UpsertSource(ctx context.Context, rfc *RssSource) error
	// GetSources 获取所有订阅源信息
	GetSources(ctx context.Context) ([]RssSource, error)
	// GetSourceByRssHubFeedLink 通过 rssHub 的 feed 链接获取订阅源信息
	GetSourceByRssHubFeedLink(ctx context.Context, url string) (*RssSource, error)
	// DeleteSource 删除一个订阅源
	DeleteSource(ctx context.Context, fID int64) error
}

// RepoSubscribe RSS 订阅存储接口
type RepoSubscribe interface {
	// CreateSubscribe 添加一个订阅
	CreateSubscribe(ctx context.Context, gid, rssSourceID int64) error
	// DeleteSubscribe 删除一个订阅
	DeleteSubscribe(ctx context.Context, subscribeID int64) error
	// GetSubscribeByID 获取一个订阅
	GetSubscribeByID(ctx context.Context, gid int64, subscribeID int64) (*RssSubscribe, error)
	// GetSubscribes 获取全部订阅
	GetSubscribes(ctx context.Context) ([]*RssSubscribe, error)
}

// RepoMultiQuery 多表查询接口
type RepoMultiQuery interface {
	// GetSubscribesBySource 获取一个源对应的所有订阅群组
	GetSubscribesBySource(ctx context.Context, feedPath string) ([]*RssSubscribe, error)
	// GetIfExistedSubscribe 判断一个群组是否已订阅了一个源
	GetIfExistedSubscribe(ctx context.Context, gid int64, feedPath string) (*RssSubscribe, bool, error)
	// GetSubscribedChannelsByGroupID 获取该群所有的订阅
	GetSubscribedChannelsByGroupID(ctx context.Context, gid int64) ([]*RssSource, error)
}
