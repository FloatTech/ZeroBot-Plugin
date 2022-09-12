// Package kanban 打印版本信息
package kanban

import (
	"sync"
)

var once sync.Once

func init() {
	once.Do(PrintBanner)
}
