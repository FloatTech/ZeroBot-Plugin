// Package aiimage 提供AI画图功能配置
package aiimage

import (
	"fmt"
	"strings"
	"sync"

	sql "github.com/FloatTech/sqlite"
)

// storage 管理画图配置存储
type storage struct {
	sync.RWMutex
	db sql.Sqlite
}

// imageConfig 存储AI画图配置信息
type imageConfig struct {
	ID        int64  `db:"id"`        // 主键ID
	APIKey    string `db:"apiKey"`    // API密钥
	APIURL    string `db:"apiUrl"`    // API地址
	ModelName string `db:"modelName"` // 画图模型名称
}

// getConfig 获取当前配置
func (sdb *storage) getConfig() imageConfig {
	sdb.RLock()
	defer sdb.RUnlock()
	cfg := imageConfig{}
	_ = sdb.db.Find("config", &cfg, "WHERE id = 1")
	return cfg
}

// setConfig 设置AI画图配置
func (sdb *storage) setConfig(apiKey, apiURL, modelName string) error {
	sdb.Lock()
	defer sdb.Unlock()
	return sdb.db.Insert("config", &imageConfig{
		ID:        1,
		APIKey:    apiKey,
		APIURL:    apiURL,
		ModelName: modelName,
	})
}

// PrintConfig 返回格式化后的配置信息
func (sdb *storage) PrintConfig() string {
	cfg := sdb.getConfig()
	var builder strings.Builder
	builder.WriteString("当前AI画图配置:\n")
	builder.WriteString(fmt.Sprintf("• 密钥: %s\n", cfg.APIKey))
	builder.WriteString(fmt.Sprintf("• 接口地址: %s\n", cfg.APIURL))
	builder.WriteString(fmt.Sprintf("• 模型名: %s\n", cfg.ModelName))
	return builder.String()
}
