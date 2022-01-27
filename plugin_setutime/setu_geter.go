// Package setutime 来份涩图
package setutime

import (
	"errors"
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
	lock sync.Mutex
	db   *sql.Sqlite
	path string
	max  int
	pool map[string][]*pixiv.Illust
}

func (i *imgpool) list() (l []string) {
	var err error
	l, err = i.db.ListTables()
	if err != nil {
		l = []string{"涩图", "二次元", "风景", "车万"}
	}
	return l
}

func init() { // 插件主体
	limit := rate.NewManager(time.Minute*1, 5)
	engine := control.Register("setutime", order.PrioSetuTime, &control.Options{
		DisableOnDefault: false,
		Help: "涩图\n" +
			"- 来份[涩图/二次元/风景/车万]\n" +
			"- 添加[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- 删除[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- >setu status",
	})
	process.SleepAbout1sTo2s()
	pool := func() *imgpool {
		cache := &imgpool{
			db:   &sql.Sqlite{DBPath: "data/SetuTime/SetuTime.db"},
			path: pixiv.CacheDir,
			max:  10,
			pool: map[string][]*pixiv.Illust{},
		}
		// 如果数据库不存在则下载
		_, _ = fileutil.GetLazyData(cache.db.DBPath, false, false)
		err := cache.db.Open()
		if err != nil {
			panic(err)
		}
		for _, imgtype := range cache.list() {
			if err := cache.db.Create(imgtype, &pixiv.Illust{}); err != nil {
				panic(err)
			}
		}
		return cache
	}()
	engine.OnRegex(`^来份(.*)$`, rule.FirstValueInList(pool.list())).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.SendChain(message.Text("请稍后重试0x0..."))
				return
			}
			var imgtype = ctx.State["regex_matched"].([]string)[1]
			// 补充池子
			go pool.fill(ctx, imgtype)
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
			if id := ctx.SendChain(message.Image(pool.popfile(imgtype))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnRegex(`^添加(.*?)(\d+)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			err := pool.add(ctx, imgtype, id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR:", err))
				return
			}
			ctx.SendChain(message.Text("成功向分类", imgtype, "添加图片", id))
		})

	engine.OnRegex(`^删除(.*?)(\d+)$`, rule.FirstValueInList(pool.list()), zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			// 查询数据库
			if err := pool.remove(imgtype, id); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("删除成功"))
		})

	// 查询数据库涩图数量
	engine.OnFullMatchGroup([]string{">setu status"}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			state := []string{"[SetuTime]"}
			for _, imgtype := range pool.list() {
				num, err := pool.db.Count(imgtype)
				if err != nil {
					num = 0
				}
				state = append(state, "\n")
				state = append(state, imgtype)
				state = append(state, ": ")
				state = append(state, fmt.Sprintf("%d", num))
			}
			ctx.SendChain(message.Text(state))
		})
}

// size 返回缓冲池指定类型的现有大小
func (p *imgpool) size(imgtype string) int {
	return len(p.pool[imgtype])
}

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
	p.lock.Lock()
	p.pool[imgtype] = append(p.pool[imgtype], illust)
	p.lock.Unlock()
}

func (p *imgpool) pop(imgtype string) (illust *pixiv.Illust) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.size(imgtype) == 0 {
		return
	}
	illust = p.pool[imgtype][0]
	p.pool[imgtype] = p.pool[imgtype][1:]
	return
}

func (p *imgpool) file(i *pixiv.Illust) string {
	u := i.ImageUrls[0]
	m, err := imagepool.GetImage(u[strings.LastIndex(u, "/")+1 : len(u)-4])
	if err == nil {
		return m.String()
	}
	filename := fmt.Sprint(i.Pid) + "_p0"
	filepath := fileutil.BOTPATH + `/` + p.path + filename
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

func (p *imgpool) popfile(imgtype string) string {
	return p.file(p.pop(imgtype))
}

// fill 补充池子
func (p *imgpool) fill(ctx *zero.Ctx, imgtype string) {
	times := math.Min(p.max-p.size(imgtype), 2)
	for i := 0; i < times; i++ {
		illust := &pixiv.Illust{}
		// 查询出一张图片
		if err := p.db.Pick(imgtype, illust); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			continue
		}
		// 向缓冲池添加一张图片
		p.push(ctx, imgtype, illust)
		process.SleepAbout1sTo2s()
	}
}

func (p *imgpool) add(ctx *zero.Ctx, imgtype string, id int64) error {
	if err := p.db.Create(imgtype, &pixiv.Illust{}); err != nil {
		return err
	}
	ctx.SendChain(message.Text("少女祈祷中......"))
	// 查询P站插图信息
	illust, err := pixiv.Works(id)
	if err != nil {
		return err
	}
	// 下载插画
	if _, err := illust.DownloadToCache(0, strconv.FormatInt(id, 10)+"_p0"); err != nil {
		return err
	}
	// 发送到发送者
	if id := ctx.SendChain(message.Image(p.file(illust))); id.ID() == 0 {
		return errors.New("可能被风控，发送失败")
	}
	// 添加插画到对应的数据库table
	if err := p.db.Insert(imgtype, illust); err != nil {
		return err
	}
	return nil
}

func (p *imgpool) remove(imgtype string, id int64) error {
	return p.db.Del(imgtype, fmt.Sprintf("WHERE pid=%d", id))
}
