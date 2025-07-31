// Package aiimage 提供AI画图功能配置
package aiimage

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	sql "github.com/FloatTech/sqlite"
)

// Storage 管理画图配置存储
type Storage struct {
	sync.RWMutex
	db sql.Sqlite
}

var (
	sdb = &Storage{
		db: sql.New("data/aiimage/config.db"),
	}
)

func init() {
	if err := os.MkdirAll("data/aiimage", 0755); err != nil {
		panic(err)
	}
	if err := sdb.db.Open(time.Hour * 24); err != nil {
		panic(err)
	}
	if err := sdb.db.Create("config", &ImageConfig{}); err != nil {
		panic(err)
	}
}

// ImageConfig 存储AI画图配置信息
type ImageConfig struct {
	ID        int64  `db:"id"`        // 主键ID
	APIKey    string `db:"apiKey"`    // API密钥
	APIURL    string `db:"apiUrl"`    // API地址
	ModelName string `db:"modelName"` // 画图模型名称
}

// GetConfig 获取当前配置
func GetConfig() ImageConfig {
	sdb.RLock()
	defer sdb.RUnlock()
	cfg := ImageConfig{}
	_ = sdb.db.Find("config", &cfg, "WHERE id = 1")
	return cfg
}

// SetConfig 设置AI画图配置
func SetConfig(apiKey, apiURL, modelName string) error {
	sdb.Lock()
	defer sdb.Unlock()
	return sdb.db.Insert("config", &ImageConfig{
		ID:        1,
		APIKey:    apiKey,
		APIURL:    apiURL,
		ModelName: modelName,
	})
}

// PrintConfig 返回格式化后的配置信息
func PrintConfig() string {
	cfg := GetConfig()
	var builder strings.Builder
	builder.WriteString("当前AI画图配置:\n")
	builder.WriteString(fmt.Sprintf("• 密钥: %s\n", cfg.APIKey))
	builder.WriteString(fmt.Sprintf("• 接口地址: %s\n", cfg.APIURL))
	builder.WriteString(fmt.Sprintf("• 模型名: %s\n", cfg.ModelName))
	return builder.String()
}
