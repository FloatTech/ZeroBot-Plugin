package utils

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	zero "github.com/wdvxdr1123/ZeroBot"
)

var (
	CACHEPATH   string // 图片缓存路径
	CACHE_GROUP int64  // 图片缓存群，用于上传图片到tx服务器
)

// PicsCache 图片缓冲池
type PicsCache struct {
	Lock     sync.Mutex
	Max      int
	ECY      []string
	IECY     []Illust
	SETU     []string
	ISETU    []Illust
	SCENERY  []string
	ISCENERY []Illust
}

// Len 返回当前缓冲池的图片数量
func (this *Illust) Len(type_ string, pool *PicsCache) (length int) {
	switch type_ {
	case "ecy":
		return len(pool.ECY)
	case "setu":
		return len(pool.SETU)
	case "scenery":
		return len(pool.SCENERY)
	}
	return 0
}

// Add 添加图片到缓冲池，返回错误
func (this Illust) Add(type_ string, pool *PicsCache) (err error) {
	// TODO 下载图片
	path, err := this.PixivPicDown(CACHEPATH)
	if err != nil {
		return err
	}
	hash := PicHash(path)
	// TODO 发送到缓存群以上传tx服务器
	if id := zero.SendGroupMessage(CACHE_GROUP, "[CQ:image,file=file:///"+path+"]"); id == 0 {
		return errors.New("send failed")
	}
	// TODO 把hash和插图信息添加到缓冲池
	pool.Lock.Lock()
	defer pool.Lock.Unlock()
	switch type_ {
	case "ecy":
		pool.ECY = append(pool.ECY, hash)
		pool.IECY = append(pool.IECY, this)
	case "setu":
		pool.SETU = append(pool.SETU, hash)
		pool.ISETU = append(pool.ISETU, this)
	case "scenery":
		pool.SCENERY = append(pool.SCENERY, hash)
		pool.ISCENERY = append(pool.ISCENERY, this)
	}
	return nil
}

// Get 从缓冲池里取出一张，返回hash，illust值中Pid和UserName会被改变
func (this *Illust) Get(type_ string, pool *PicsCache) (hash string) {
	pool.Lock.Lock()
	defer pool.Lock.Unlock()
	switch type_ {
	case "ecy":
		if len(pool.ECY) > 0 {
			hash := pool.ECY[0]
			this.Pid = pool.IECY[0].Pid
			this.Title = pool.IECY[0].Title
			this.UserName = pool.IECY[0].UserName
			pool.ECY = pool.ECY[1:]
			pool.IECY = pool.IECY[1:]
			return hash
		}
	case "setu":
		if len(pool.SETU) > 0 {
			hash := pool.SETU[0]
			this.Pid = pool.ISETU[0].Pid
			this.Title = pool.ISETU[0].Title
			this.UserName = pool.ISETU[0].UserName
			pool.SETU = pool.SETU[1:]
			pool.ISETU = pool.ISETU[1:]
			return hash
		}
	case "scenery":
		if len(pool.SCENERY) > 0 {
			hash := pool.SCENERY[0]
			this.Pid = pool.ISCENERY[0].Pid
			this.Title = pool.ISCENERY[0].Title
			this.UserName = pool.ISCENERY[0].UserName
			pool.SCENERY = pool.SCENERY[1:]
			pool.ISCENERY = pool.ISCENERY[1:]
			return hash
		}
	default:
		//
	}
	return ""
}

func GetCQcodePicLink(text string) (url string) {
	text = strings.ReplaceAll(text, "{", "")
	text = strings.ReplaceAll(text, "{", "")
	text = strings.ReplaceAll(text, "-", "")
	if index := strings.Index(text, "."); index != -1 {
		if hash := text[:index]; len(hash) == 32 {
			return fmt.Sprintf("http://gchat.qpic.cn/gchatpic_new//--%s/0", hash)
		}
	}
	return ""
}

// BigPic 返回一张XML大图CQ码
func (this *Illust) BigPic(hash string) string {
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
		this.Title,
		this.Pid,
		this.UserName,
	)
}

// NormalPic 返回一张普通图CQ码
func (this *Illust) NormalPic() string {
	return fmt.Sprintf(`[CQ:image,file=file:///%s%d.jpg]`, CACHEPATH, this.Pid)
}

// DetailPic 返回一张带详细信息的图片CQ码
func (this *Illust) DetailPic() string {
	return fmt.Sprintf(`[SetuTime] %s 标题：%s 
插画ID：%d 
画师：%s 
画师ID：%d 
直链：https://pixivel.moe/detail?id=%d`,
		this.NormalPic(),
		this.Title,
		this.Pid,
		this.UserName,
		this.UserId,
		this.Pid,
	)
}
