package utils

import (
	"fmt"
	"sync"
)

// PoolsCache 图片缓冲池
type PoolsCache struct {
	Lock  sync.Mutex
	Max   int
	Path  string
	Group int64
	Pool  map[string][]*Illust
}

// NewPoolsCache 返回一个缓冲池对象
func NewPoolsCache() *PoolsCache {
	return &PoolsCache{
		Max:   10,
		Path:  "./data/SetuTime/cache/",
		Group: 1048452984,
		Pool:  map[string][]*Illust{},
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
func (p *PoolsCache) Push(type_ string, illust *Illust) (err error) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Pool[type_] = append(p.Pool[type_], illust)
	return nil
}

// Push 在缓冲池拿出一张图片，返回错误
func (p *PoolsCache) Pop(type_ string) (illust *Illust) {
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

// BigPic 返回一张XML大图CQ码
func (i *Illust) BigPic(file string) string {
	var hash = PicHash(file)
	return fmt.Sprintf(`[CQ:xml,data=<?xml version='1.0' 
encoding='UTF-8' standalone='yes' ?><msg serviceID="5" 
templateID="12345" action="" brief="不够涩！" 
sourceMsgId="0" url="" flag="0" adverSign="0" multiMsgFlag="0">
<item layout="0" advertiser_id="0" aid="0"><image uuid="%s.jpg" md5="%s" 
GroupFiledid="2235033681" filesize="81322" local_path="%s.jpg" 
minWidth="200" minHeight="200" maxWidth="500" maxHeight="1000" />
</item><source name="%s⭐(id:%d author:%s)" icon="" 
action="" appid="-1" /></msg>]`,
		hash,
		hash,
		hash,
		i.Title,
		i.Pid,
		i.UserName,
	)
}

// NormalPic 返回一张普通图CQ码
func (i *Illust) NormalPic(file string) string {
	return fmt.Sprintf(`[CQ:image,file=file:///%s]`, file)
}

// DetailPic 返回一张带详细信息的图片CQ码
func (i *Illust) DetailPic(file string) string {
	return fmt.Sprintf(`[SetuTime] %s 
标题：%s 
插画ID：%d 
画师：%s 
画师ID：%d 
直链：https://pixivel.moe/detail?id=%d`,
		i.NormalPic(file),
		i.Title,
		i.Pid,
		i.UserName,
		i.UserId,
		i.Pid,
	)
}
