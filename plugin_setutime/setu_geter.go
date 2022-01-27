// Package setutime 来份涩图
package setutime

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	imagepool "github.com/FloatTech/AnimeAPI/imgpool"
	"github.com/FloatTech/AnimeAPI/pixiv"
	control "github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	fileutil "github.com/FloatTech/zbputils/file"
	"github.com/FloatTech/zbputils/math"
	"github.com/FloatTech/zbputils/process"
	"github.com/FloatTech/zbputils/rule"
	"github.com/FloatTech/zbputils/sql"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/order"
)

// Pools 图片缓冲池
type imgpool struct {
	Lock  sync.Mutex
	DB    *sql.Sqlite
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
		DB:    &sql.Sqlite{DBPath: "data/SetuTime/SetuTime.db"},
		Path:  pixiv.CacheDir,
		Group: 0,
		List:  []string{"涩图", "二次元", "风景", "车万"}, // 可以自己加类别，得自己加图片进数据库
		Max:   10,
		Pool:  map[string][]*pixiv.Illust{},
		Form:  0,
	}
	// 如果数据库不存在则下载
	_, _ = fileutil.GetLazyData(cache.DB.DBPath, false, false)
	for i := range cache.List {
		if err := cache.DB.Create(cache.List[i], &pixiv.Illust{}); err != nil {
			panic(err)
		}
	}
	return cache
}

var (
	pool  *imgpool
	limit = rate.NewManager(time.Minute*1, 5)
)

func init() { // 插件主体
	engine := control.Register("setutime", order.PrioSetuTime, &control.Options{
		DisableOnDefault: false,
		Help: "涩图\n" +
			"- 来份[涩图/二次元/风景/车万]\n" +
			"- 添加[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- 删除[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- >setu status",
	})
	process.SleepAbout1sTo2s()
	pool = newPools()
	engine.OnRegex(`^来份(.*)$`, rule.FirstValueInList(pool.List)).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.SendChain(message.Text("请稍后重试0x0..."))
				return
			}
			var imgtype = ctx.State["regex_matched"].([]string)[1]
			// 补充池子
			go func() {
				times := math.Min(pool.Max-pool.size(imgtype), 2)
				for i := 0; i < times; i++ {
					illust := &pixiv.Illust{}
					// 查询出一张图片
					if err := pool.DB.Pick(imgtype, illust); err != nil {
						ctx.SendChain(message.Text("ERROR: ", err))
						continue
					}
					// 向缓冲池添加一张图片
					pool.push(ctx, imgtype, illust)
					process.SleepAbout1sTo2s()
				}
			}()
			// 如果没有缓存，阻塞10秒
			if pool.size(imgtype) == 0 {
				ctx.SendChain(message.Text("INFO: 正在填充弹药......"))
				time.Sleep(time.Second * 10)
				if pool.size(imgtype) == 0 {
					ctx.SendChain(message.Text("ERROR: 等待填充，请稍后再试......"))
					return
				}
			}
			// 从缓冲池里抽一张
			if id := ctx.SendChain(message.Image(file(pool.pop(imgtype)))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnRegex(`^添加(.*?)(\d+)$`, rule.FirstValueInList(pool.List), zero.SuperUserPermission).SetBlock(true).
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
			if _, err := illust.DownloadToCache(0, strconv.FormatInt(id, 10)+"_p0"); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			// 发送到发送者
			if id := ctx.SendChain(message.Image(file(illust))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控，发送失败"))
				return
			}
			// 添加插画到对应的数据库table
			if err := pool.DB.Insert(imgtype, illust); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("添加成功"))
		})

	engine.OnRegex(`^删除(.*?)(\d+)$`, rule.FirstValueInList(pool.List), zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			// 查询数据库
			if err := pool.DB.Del(imgtype, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("删除成功"))
		})

	// 查询数据库涩图数量
	engine.OnFullMatchGroup([]string{">setu status"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			state := []string{"[SetuTime]"}
			for i := range pool.List {
				num, err := pool.DB.Count(pool.List[i])
				if err != nil {
					num = 0
				}
				state = append(state, "\n")
				state = append(state, pool.List[i])
				state = append(state, ": ")
				state = append(state, fmt.Sprintf("%d", num))
			}
			ctx.SendChain(message.Text(state))
		})
}

// size 返回缓冲池指定类型的现有大小
func (p *imgpool) size(imgtype string) int {
	return len(p.Pool[imgtype])
}

/*
// isFull 返回缓冲池指定类型是否已满
func (p *imgpool) isFull(imgtype string) bool {
	return len(p.Pool[imgtype]) >= p.Max
}*/

// push 向缓冲池插入一张图片
func (p *imgpool) push(ctx *zero.Ctx, imgtype string, illust *pixiv.Illust) {
	u := illust.ImageUrls[0]
	n := u[strings.LastIndex(u, "/")+1 : len(u)-4]
	m, err := imagepool.GetImage(n)
	if err != nil {
		// 下载图片
		f := ""
		if f, err = illust.DownloadToCache(0, n); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		m.SetFile(fileutil.BOTPATH + "/" + f)
		_, _ = m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
	}
	p.Lock.Lock()
	p.Pool[imgtype] = append(p.Pool[imgtype], illust)
	p.Lock.Unlock()
}

// Push 在缓冲池拿出一张图片
func (p *imgpool) pop(imgtype string) (illust *pixiv.Illust) {
	p.Lock.Lock()
	defer p.Lock.Unlock()
	if p.size(imgtype) == 0 {
		return
	}
	illust = p.Pool[imgtype][0]
	p.Pool[imgtype] = p.Pool[imgtype][1:]
	return
}

func file(i *pixiv.Illust) string {
	u := i.ImageUrls[0]
	m, err := imagepool.GetImage(u[strings.LastIndex(u, "/")+1 : len(u)-4])
	if err == nil {
		return m.String()
	}
	filename := fmt.Sprint(i.Pid) + "_p0"
	filepath := fileutil.BOTPATH + `/` + pool.Path + filename
	if fileutil.IsExist(filepath + ".jpg") {
		return `file:///` + filepath + ".jpg"
	}
	if fileutil.IsExist(filepath + ".png") {
		return `file:///` + filepath + ".png"
	}
	if fileutil.IsExist(filepath + ".gif") {
		return `file:///` + filepath + ".gif"
	}
	return ""
}
