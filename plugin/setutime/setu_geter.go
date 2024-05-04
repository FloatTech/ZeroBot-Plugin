// Package setutime 来份涩图
package setutime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/pixiv"
	fcext "github.com/FloatTech/floatbox/ctxext"
	fileutil "github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/process"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	imagepool "github.com/FloatTech/zbputils/img/pool"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// Pools 图片缓冲池
type imgpool struct {
	db     *sql.Sqlite
	dbmu   sync.RWMutex
	path   string
	max    int
	pool   map[string][]*message.MessageSegment
	poolmu sync.Mutex
}

func (p *imgpool) List() (l []string) {
	var err error
	p.dbmu.RLock()
	defer p.dbmu.RUnlock()
	l, err = p.db.ListTables()
	if err != nil {
		l = []string{"涩图", "二次元", "风景", "车万"}
	}
	return l
}

var pool = &imgpool{
	db:   &sql.Sqlite{},
	path: pixiv.CacheDir,
	max:  10,
	pool: make(map[string][]*message.MessageSegment),
}

func init() { // 插件主体
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "涩图",
		Help: "- 来份[涩图/二次元/风景/车万]\n" +
			"- 添加[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- 删除[涩图/二次元/风景/车万][P站图片ID]\n" +
			"- >setu status",
		PublicDataFolder: "SetuTime",
	})

	getdb := fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		// 如果数据库不存在则下载
		pool.db.DBPath = engine.DataFolder() + "SetuTime.db"
		_, _ = engine.GetLazyData("SetuTime.db", false)
		err := pool.db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return false
		}
		for _, imgtype := range pool.List() {
			if err := pool.db.Create(imgtype, &pixiv.Illust{}); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return false
			}
		}
		return true
	})

	engine.OnRegex(`^来份(.+)$`, getdb, fcext.ValueInList(func(ctx *zero.Ctx) string { return ctx.State["regex_matched"].([]string)[1] }, pool)).SetBlock(true).Limit(ctxext.LimitByUser).
		Handle(func(ctx *zero.Ctx) {
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
			m := message.Message{ctxext.FakeSenderForwardNode(ctx, *pool.pop(imgtype))}
			if id := ctx.Send(m).ID(); id == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}
		})

	engine.OnRegex(`^添加\s*([^0-9\s]+)\s*(\d+)$`, zero.SuperUserPermission, getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			var (
				imgtype = ctx.State["regex_matched"].([]string)[1]
				id, _   = strconv.ParseInt(ctx.State["regex_matched"].([]string)[2], 10, 64)
			)
			err := pool.add(ctx, imgtype, id)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功向分类", imgtype, "添加图片", id))
		})

	engine.OnRegex(`^删除\s*([^0-9\s]+)\s*(\d+)$`, getdb, fcext.ValueInList(func(ctx *zero.Ctx) string { return ctx.State["regex_matched"].([]string)[1] }, pool), zero.SuperUserPermission).SetBlock(true).
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
	engine.OnFullMatch(">setu status", getdb).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			state := []string{"[SetuTime]"}
			pool.dbmu.RLock()
			defer pool.dbmu.RUnlock()
			for _, imgtype := range pool.List() {
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
	if len(illust.ImageUrls) == 0 {
		return
	}
	u := illust.ImageUrls[0]
	n := u[strings.LastIndex(u, "/")+1 : len(u)-4]
	m, err := imagepool.GetImage(n)
	var msg message.MessageSegment
	f := fileutil.BOTPATH + "/" + illust.Path(0)
	if err != nil {
		if fileutil.IsNotExist(f) {
			// 下载图片
			if err := illust.DownloadToCache(0); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
		}
		m.SetFile(f)
		_, _ = m.Push(ctxext.SendToSelf(ctx), ctxext.GetMessage(ctx))
		msg = message.Image("file:///" + f)
	} else {
		msg = message.Image(m.String())
		if ctxext.SendToSelf(ctx)(msg) == 0 {
			msg = msg.Add("cache", "0")
		}
	}
	p.poolmu.Lock()
	p.pool[imgtype] = append(p.pool[imgtype], &msg)
	p.poolmu.Unlock()
}

func (p *imgpool) pop(imgtype string) (msg *message.MessageSegment) {
	p.poolmu.Lock()
	defer p.poolmu.Unlock()
	if p.size(imgtype) == 0 {
		return
	}
	msg = p.pool[imgtype][0]
	p.pool[imgtype] = p.pool[imgtype][1:]
	return
}

// fill 补充池子
func (p *imgpool) fill(ctx *zero.Ctx, imgtype string) {
	times := math.Min(p.max-p.size(imgtype), 2)
	p.dbmu.RLock()
	defer p.dbmu.RUnlock()
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
	p.dbmu.Lock()
	defer p.dbmu.Unlock()
	if err := p.db.Create(imgtype, &pixiv.Illust{}); err != nil {
		return err
	}
	ctx.SendChain(message.Text("少女祈祷中......"))
	// 查询P站插图信息
	illust, err := pixiv.Works(id)
	if err != nil {
		return err
	}
	if len(illust.ImageUrls) == 0 {
		return errors.New("nil image url")
	}
	err = imagepool.SendImageFromPool(strconv.FormatInt(illust.Pid, 10)+"_p0", illust.Path(0), func() error {
		return illust.DownloadToCache(0)
	}, ctxext.Send(ctx), ctxext.GetMessage(ctx))
	if err != nil {
		return err
	}
	// 添加插画到对应的数据库table
	return p.db.Insert(imgtype, illust)
}

func (p *imgpool) remove(imgtype string, id int64) error {
	p.dbmu.Lock()
	defer p.dbmu.Unlock()
	return p.db.Del(imgtype, fmt.Sprintf("WHERE pid=%d", id))
}
