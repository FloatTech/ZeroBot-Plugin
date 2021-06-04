package utils

import (
	"fmt"
	"sync"

	"github.com/Yiwen-Chan/ZeroBot-Plugin/api/pixiv"
)

// PoolsCache 图片缓冲池
type PoolsCache struct {
	Lock  sync.Mutex
	Max   int
	Path  string
	Group int64
	Pool  map[string][]*pixiv.Illust
}

// NewPoolsCache 返回一个缓冲池对象
func NewPoolsCache() *PoolsCache {
	return &PoolsCache{
		Max:   10,
		Path:  "./data/SetuTime/cache/",
		Group: 1048452984,
		Pool:  map[string][]*pixiv.Illust{},
	}
}

// Size 返回缓冲池指定类型的现有大小
func (p *PoolsCache) Size(type_ string) int {
	return len(p.Pool[type_])
}

// IsFull 返回缓冲池指定类型是否已满
func (p *PoolsCache) IsFull(type_ string) bool {
	return len(p.Pool[type_]) >= p.Max
}

// Push 向缓冲池插入一张图片，返回错误
func (p *PoolsCache) Push(type_ string, illust *pixiv.Illust) (err error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Pool[type_] = append(p.Pool[type_], illust)
	return nil
}

// Push 在缓冲池拿出一张图片，返回错误
func (p *PoolsCache) Pop(type_ string) (illust *pixiv.Illust) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Size(type_) == 0 {
		return
	}
	illust = p.Pool[type_][0]
	p.Pool[type_] = p.Pool[type_][1:]
	return
}

// Push 在缓冲池拿出一张图片，返回指定格式CQ码
func (p *PoolsCache) GetOnePic(type_ string, form string) string {
	var (
		illust = p.Pop(type_)
		file   = fmt.Sprintf("%s%d.jpg", p.Path, illust.Pid)
	)
	switch form {
	case "XML":
		return illust.BigPic(file)
	case "DETAIL":
		return illust.DetailPic(file)
	default:
		return illust.NormalPic(file)
	}
}
