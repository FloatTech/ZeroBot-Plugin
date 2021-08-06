package setutime

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/AnimeAPI/pixiv"
)

// Pools 图片缓冲池
type imgpool struct {
	Lock  sync.Mutex
	DB    *sqlite
	Path  string
	Group int64
	List  []string
	Max   int
	Pool  map[string][]*pixiv.Illust
	Form  int64
}

// NewPoolsCache 返回一个缓冲池对象
func newPools() *imgpool {
	cache := &imgpool{
		DB:    &sqlite{DBPath: "data/SetuTime/SetuTime.db"},
		Path:  "data/SetuTime/cache/",
		Group: 0,
		List:  []string{"涩图", "二次元", "风景", "车万"}, // 可以自己加类别，得自己加图片进数据库
		Max:   10,
		Pool:  map[string][]*pixiv.Illust{},
		Form:  0,
	}
	err := os.MkdirAll(cache.Path, 0755)
	if err != nil {
		panic(err)
	}
	for i := range cache.List {
		if err := cache.DB.create(cache.List[i], &pixiv.Illust{}); err != nil {
			panic(err)
		}
	}
	return cache
}

var (
	pool  = newPools()
	limit = rate.NewManager(time.Minute*1, 5)
)

func init() { // 插件主体
	zero.OnRegex(`^来份(.*)$`, firstValueInList(pool.List)).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.SendChain(message.Text("请稍后重试0x0..."))
				return
			}
			var imgtype = ctx.State["regex_matched"].([]string)[1]
			// 补充池子
			go func() {
				times := min(pool.Max-pool.size(imgtype), 2)
				for i := 0; i < times; i++ {
					illust := &pixiv.Illust{}
					// 查询出一张图片
					if err := pool.DB.find(imgtype, illust, "ORDER BY RANDOM() limit 1"); err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						continue
					}
					// 下载图片
					if _, err := download(illust, pool.Path); err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						continue
					}
					ctx.SendGroupMessage(pool.Group, []message.MessageSegment{message.Image(file(illust))})
					// 向缓冲池添加一张图片
					pool.push(imgtype, illust)

					time.Sleep(time.Second * 1)
				}
			}()
			// 如果没有缓存，阻塞5秒
			if pool.size(imgtype) == 0 {
				ctx.SendChain(message.Text("INFO: 正在填充弹药......"))
				<-time.After(time.Second * 5)
				if pool.size(imgtype) == 0 {
					ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
					return
				}
			}
			// 从缓冲池里抽一张
			if id := ctx.SendChain(message.Image(file(pool.pop(imgtype)))); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			return
		})

	zero.OnRegex(`^添加(.*?)(\d+)$`, firstValueInList(pool.List), zero.SuperUserPermission).SetBlock(true).SetPriority(21).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			ctx.SendChain(message.Text("少女祈祷中......"))
			// 查询P站插图信息
			illust, err := pixiv.Works(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 下载插画
			if _, err := download(illust, pool.Path); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 发送到发送者
			if id := ctx.SendChain(message.Image(file(illust))); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控，发送失败"))
				return
			}
			// 添加插画到对应的数据库table
			if err := pool.DB.insert(imgtype, illust); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send("添加成功")
			return
		})

	zero.OnRegex(`^删除(.*?)(\d+)$`, firstValueInList(pool.List), zero.SuperUserPermission).SetBlock(true).SetPriority(22).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			// 查询数据库
			if err := pool.DB.del(imgtype, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			ctx.Send("删除成功")
			return
		})

	// 查询数据库涩图数量
	zero.OnFullMatchGroup([]string{">setu status"}).SetBlock(true).SetPriority(23).
		Handle(func(ctx *zero.Ctx) {
			state := []string{"[SetuTime]"}
			for i := range pool.List {
				num, err := pool.DB.num(pool.List[i])
				if err != nil {
					num = 0
				}
				state = append(state, "\n")
				state = append(state, pool.List[i])
				state = append(state, ": ")
				state = append(state, fmt.Sprintf("%d", num))
			}
			ctx.Send(strings.Join(state, ""))
			return
		})
}

// firstValueInList 判断正则匹配的第一个参数是否在列表中
func firstValueInList(list []string) zero.Rule {
	return func(ctx *zero.Ctx) bool {
		first := ctx.State["regex_matched"].([]string)[1]
		for i := range list {
			if first == list[i] {
				return true
			}
		}
		return false
	}
}

// min 返回两数最小值
func min(a, b int) int {
	switch {
	default:
		return a
	case a > b:
		return b
	case a < b:
		return a
	}
}

// size 返回缓冲池指定类型的现有大小
func (p *imgpool) size(type_ string) int {
	return len(p.Pool[type_])
}

// isFull 返回缓冲池指定类型是否已满
func (p *imgpool) isFull(type_ string) bool {
	return len(p.Pool[type_]) >= p.Max
}

// push 向缓冲池插入一张图片
func (p *imgpool) push(type_ string, illust *pixiv.Illust) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Pool[type_] = append(p.Pool[type_], illust)
}

// Push 在缓冲池拿出一张图片
func (p *imgpool) pop(type_ string) (illust *pixiv.Illust) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.size(type_) == 0 {
		return
	}
	illust = p.Pool[type_][0]
	p.Pool[type_] = p.Pool[type_][1:]
	return
}

func file(i *pixiv.Illust) string {
	filename := fmt.Sprint(i.Pid)
	pwd, _ := os.Getwd()
	filepath := pwd + `/` + pool.Path + filename
	if _, err := os.Stat(filepath + ".jpg"); err == nil || os.IsExist(err) {
		return `file:///` + filepath + ".jpg"
	}
	if _, err := os.Stat(filepath + ".png"); err == nil || os.IsExist(err) {
		return `file:///` + filepath + ".png"
	}
	if _, err := os.Stat(filepath + ".gif"); err == nil || os.IsExist(err) {
		return `file:///` + filepath + ".gif"
	}
	return ""
}

func download(i *pixiv.Illust, filedir string) (string, error) {
	filename := fmt.Sprint(i.Pid)
	filepath := filedir + filename
	if _, err := os.Stat(filepath + ".jpg"); err == nil || os.IsExist(err) {
		return filepath + ".jpg", nil
	}
	if _, err := os.Stat(filepath + ".png"); err == nil || os.IsExist(err) {
		return filepath + ".png", nil
	}
	if _, err := os.Stat(filepath + ".gif"); err == nil || os.IsExist(err) {
		return filepath + ".gif", nil
	}
	// 下载最大分辨率为 1200 的图片
	link := i.ImageUrls
	link = strings.ReplaceAll(link, "img-original", "img-master")
	link = strings.ReplaceAll(link, "_p0", "_p0_master1200")
	link = strings.ReplaceAll(link, ".png", ".jpg")
	// 下载
	return pixiv.Download(link, filedir, filename)
}
