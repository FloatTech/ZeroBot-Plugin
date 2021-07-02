package plugin_setutime

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
type Pool struct {
	Lock  sync.Mutex
	DB    *Sqlite
	Path  string
	Group int64
	List  []string
	Max   int
	Pool  map[string][]*pixiv.Illust
	Form  int64
}

// NewPoolsCache 返回一个缓冲池对象
func NewPools() *Pool {
	cache := &Pool{
		DB:    &Sqlite{DBPath: "data/SetuTime/SetuTime.db"},
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
		if err := cache.DB.Create(cache.List[i], &pixiv.Illust{}); err != nil {
			panic(err)
		}
	}
	return cache
}

var (
	POOL  = NewPools()
	limit = rate.NewManager(time.Minute*1, 5)
)

func init() { // 插件主体
	zero.OnRegex(`^来份(.*)$`, FirstValueInList(POOL.List)).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.SendChain(message.Text("少女祈祷中......"))
				return
			}
			var type_ = ctx.State["regex_matched"].([]string)[1]
			// 补充池子
			go func() {
				times := Min(POOL.Max-POOL.Size(type_), 2)
				for i := 0; i < times; i++ {
					illust := &pixiv.Illust{}
					// 查询出一张图片
					if err := POOL.DB.Select(type_, illust, "ORDER BY RANDOM() limit 1"); err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						continue
					}
					// 下载图片
					if _, err := download(illust, POOL.Path); err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						continue
					}
					ctx.SendGroupMessage(POOL.Group, []message.MessageSegment{message.Image(file(illust))})
					// 向缓冲池添加一张图片
					POOL.Push(type_, illust)

					time.Sleep(time.Second * 1)
				}
			}()
			// 如果没有缓存，阻塞5秒
			if POOL.Size(type_) == 0 {
				ctx.SendChain(message.Text("INFO: 正在填充弹药......"))
				<-time.After(time.Second * 5)
				if POOL.Size(type_) == 0 {
					ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
					return
				}
			}
			// 从缓冲池里抽一张
			if id := ctx.SendChain(message.Image(file(POOL.Pop(type_)))); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
			return
		})

	zero.OnRegex(`^添加(.*?)(\d+)$`, FirstValueInList(POOL.List), zero.SuperUserPermission).SetBlock(true).SetPriority(21).
		Handle(func(ctx *zero.Ctx) {
			var (
				type_ = ctx.State["regex_matched"].([]string)[1]
				id, _ = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			ctx.SendChain(message.Text("少女祈祷中......"))
			// 查询P站插图信息
			illust, err := pixiv.Works(id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 下载插画
			if _, err := download(illust, POOL.Path); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 发送到发送者
			if id := ctx.SendChain(message.Image(file(illust))); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控，发送失败"))
				return
			}
			// 添加插画到对应的数据库table
			if err := POOL.DB.Insert(type_, illust); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.Send("添加成功")
			return
		})

	zero.OnRegex(`^删除(.*?)(\d+)$`, FirstValueInList(POOL.List), zero.SuperUserPermission).SetBlock(true).SetPriority(22).
		Handle(func(ctx *zero.Ctx) {
			var (
				type_ = ctx.State["regex_matched"].([]string)[1]
				id, _ = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			// 查询数据库
			if err := POOL.DB.Delete(type_, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
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
			for i := range POOL.List {
				num, err := POOL.DB.Num(POOL.List[i])
				if err != nil {
					num = 0
				}
				state = append(state, "\n")
				state = append(state, POOL.List[i])
				state = append(state, ": ")
				state = append(state, fmt.Sprintf("%d", num))
			}
			ctx.Send(strings.Join(state, ""))
			return
		})
}

// FirstValueInList 判断正则匹配的第一个参数是否在列表中
func FirstValueInList(list []string) zero.Rule {
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

// Min 返回两数最小值
func Min(a, b int) int {
	switch {
	default:
		return a
	case a > b:
		return b
	case a < b:
		return a
	}
}

// Size 返回缓冲池指定类型的现有大小
func (p *Pool) Size(type_ string) int {
	return len(p.Pool[type_])
}

// IsFull 返回缓冲池指定类型是否已满
func (p *Pool) IsFull(type_ string) bool {
	return len(p.Pool[type_]) >= p.Max
}

// Push 向缓冲池插入一张图片
func (p *Pool) Push(type_ string, illust *pixiv.Illust) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	p.Pool[type_] = append(p.Pool[type_], illust)
}

// Push 在缓冲池拿出一张图片
func (p *Pool) Pop(type_ string) (illust *pixiv.Illust) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.Size(type_) == 0 {
		return
	}
	illust = p.Pool[type_][0]
	p.Pool[type_] = p.Pool[type_][1:]
	return
}

func file(i *pixiv.Illust) string {
	filename := fmt.Sprint(i.Pid)
	pwd, _ := os.Getwd()
	filepath := pwd + `/` + POOL.Path + filename
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
