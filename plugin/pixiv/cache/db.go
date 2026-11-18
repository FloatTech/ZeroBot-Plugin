// Package cache ...
package cache

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/pixiv/model"
	"github.com/jinzhu/gorm"
)

// DB ...
type DB struct {
	*gorm.DB
	mu sync.Mutex
}

// NewDB ...
func NewDB(path string) *DB {
	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&model.IllustCache{}, &model.SentImage{}, &model.RefreshToken{}, &model.GroupR18Permission{}, &model.PixivProxyConfig{}).Error; err != nil {
		panic(err)
	}
	sqlDB := db.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	db.LogMode(false)
	return &DB{db, sync.Mutex{}}
}

// CheckGroupR18Permission ...
func (db *DB) CheckGroupR18Permission(gid int64) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	var count int64
	db.Model(&model.GroupR18Permission{}).Where("group_id = ?", gid).Count(&count)
	return count > 0
}

func (db *DB) findByKeyword(gid int64, keyword string, limit int, r18Req bool) ([]model.IllustCache, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var results []model.IllustCache
	query := db.Model(&model.IllustCache{}).
		Where("keyword = ?", keyword).
		Where("pid NOT IN (?)", db.Model(&model.SentImage{}).Where("group_id = ?", gid).Select("pid").SubQuery()).
		Order("bookmarks DESC").
		Limit(limit)

	query = query.Where("r18 = ?", r18Req)

	err := query.Find(&results).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return results, nil
}

func (db *DB) findByTag(gid int64, tag string, needed int, r18Req bool) ([]model.IllustCache, error) {
	if needed <= 0 {
		return nil, nil
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	var results []model.IllustCache
	escapedTag := strings.ReplaceAll(tag, "\"", "\\\"")
	likePattern := "%\"" + escapedTag + "\"%"

	query := db.Model(&model.IllustCache{}).
		Where("tags LIKE ?", likePattern).
		Where("pid NOT IN (?)", db.Model(&model.SentImage{}).Where("group_id = ?", gid).Select("pid").SubQuery()).
		Order("bookmarks DESC").
		Limit(needed)

	if !r18Req {
		query = query.Where("r18 = ?", false)
	}

	err := query.Find(&results).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return results, nil
}

// CountIllustsSmart ...
func (db *DB) CountIllustsSmart(gid int64, keyword string, r18Req bool) (int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var count int64

	query := db.Model(&model.IllustCache{}).
		Where("keyword = ?", keyword).
		Where("pid NOT IN (?)", db.Model(&model.SentImage{}).Where("group_id = ?", gid).Select("pid").SubQuery())

	if !r18Req {
		query = query.Where("r18 = ?", false)
	}
	err := query.Count(&count).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	return count, nil
}

// FindIllustsSmart ...
func (db *DB) FindIllustsSmart(gid int64, keyword string, limit int, r18Req bool) ([]model.IllustCache, error) {
	seen := make(map[int64]struct{})
	results := make([]model.IllustCache, 0, limit)

	// 1. keyword 严格查询
	kwRes, err := db.findByKeyword(gid, keyword, limit, r18Req)
	if err != nil {
		return nil, err
	}
	for _, ill := range kwRes {
		results = append(results, ill)
		seen[ill.PID] = struct{}{}
	}

	if len(results) >= limit {
		return results[:limit], nil
	}

	// 2. tag 查询补齐
	need := limit - len(results)
	for _, tagKeyword := range strings.Fields(keyword) {
		tagRes, err := db.findByTag(gid, tagKeyword, need, r18Req)
		if err != nil {
			return nil, err
		}

		for _, ill := range tagRes {
			if _, ok := seen[ill.PID]; ok {
				continue
			}
			results = append(results, ill)
			seen[ill.PID] = struct{}{}
			if len(results) >= limit {
				break
			}
		}

		if len(results) >= limit {
			break
		}

		need = limit - len(results)
	}

	return results, nil
}

// GetIllustIDsByKeyword ...
func (db *DB) GetIllustIDsByKeyword(keyword string) ([]int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var illustIDs []int64
	err := db.Model(&model.IllustCache{}).Where("keyword = ?", keyword).Pluck("pid", &illustIDs).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return illustIDs, nil
}

// GetSentPictureIDs ...
func (db *DB) GetSentPictureIDs(gid int64) ([]int64, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	var pictureIDs []int64
	err := db.Model(&model.SentImage{}).Where("group_id = ?", gid).Pluck("pid", &pictureIDs).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return pictureIDs, nil
}

// DeleteIllustByPID 删除指定 PID 的插画缓存记录
func (db *DB) DeleteIllustByPID(pid int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.Where("pid = ?", pid).Delete(&model.IllustCache{}).Error
}

// Create wraps gorm.DB.Create with a mutex to avoid SQLITE_BUSY when requests are concurrent.
func (db *DB) Create(value interface{}) *gorm.DB {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.DB.Create(value)
}

// Save wraps gorm.DB.Save with a mutex to avoid SQLITE_BUSY when requests are concurrent.
func (db *DB) Save(value interface{}) *gorm.DB {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.DB.Save(value)
}

// Delete wraps gorm.DB.Delete with a mutex to avoid SQLITE_BUSY when requests are concurrent.
func (db *DB) Delete(value interface{}, where ...interface{}) *gorm.DB {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.DB.Delete(value, where...)
}
