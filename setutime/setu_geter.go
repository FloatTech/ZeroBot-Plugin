package setutime

import (
	"errors"
	"fmt"
	"strings"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"

	"github.com/Yiwen-Chan/ZeroBot-Plugin/setutime/utils"
)

var limit = rate.NewManager(time.Minute*1, 5)

var (
	BOTPATH  = utils.PathExecute()        // 当前bot运行目录
	DATAPATH = BOTPATH + "data/SetuTime/" // 数据目录
	DBPATH   = DATAPATH + "SetuTime.db"   // 数据库路径

	CACHEPATH        = DATAPATH + "cache/"               // 缓冲图片路径
	CACHEGROUP int64 = 0                                 // 缓冲图片群，为0即可
	PoolList         = []string{"涩图", "二次元", "风景", "车万"} // 可自定义

	DB         = utils.Sqlite{DBPath: DBPATH} // 新建涩图数据库对象
	PoolsCache = utils.NewPoolsCache()        // 新建一个缓冲池对象

	FORM = "PIC" // 默认 PIC 格式
)

func init() {
	PoolsCache.Group = CACHEGROUP // 图片缓冲群
	PoolsCache.Path = CACHEPATH   // 缓冲图片路径

	utils.CreatePath(DBPATH)
	utils.CreatePath(CACHEPATH)

	for i := range PoolList {
		if err := DB.Create(PoolList[i], &utils.Illust{}); err != nil {
			panic(err)
		}
	}
}

func init() { // 插件主体
	zero.OnRegex(`^来份(.*)$`, FirstValueInList(PoolList)).SetBlock(true).SetPriority(20).
		Handle(func(ctx *zero.Ctx) {
			if !limit.Load(ctx.Event.UserID).Acquire() {
				ctx.Send("请稍后重试0x0...")
				return
			}
			var type_ = ctx.State["regex_matched"].([]string)[1]
			// 补充池子
			go func() {
				times := utils.Min(PoolsCache.Max-PoolsCache.Size(type_), 2)
				for i := 0; i < times; i++ {
					illust := &utils.Illust{}
					// 查询出一张图片
					if err := DB.Select(type_, illust, "ORDER BY RANDOM() limit 1"); err != nil {
						ctx.Send(fmt.Sprintf("ERROR: %v", err))
						continue
					}
					ctx.SendGroupMessage(PoolsCache.Group, "正在下载"+illust.ImageUrls)
					file, err := illust.PixivPicDown(PoolsCache.Path)
					if err != nil {
						ctx.Send(fmt.Sprintf("ERROR: %v", err))
						continue
					}
					ctx.SendGroupMessage(PoolsCache.Group, illust.NormalPic(file))
					// 向缓冲池添加一张图片
					if err := PoolsCache.Push(type_, illust); err != nil {
						ctx.Send(fmt.Sprintf("ERROR: %v", err))
						continue
					}
					time.Sleep(time.Second * 1)
				}
			}()
			// 如果没有缓存，阻塞5秒
			if PoolsCache.Size(type_) == 0 {
				ctx.Send("[SetuTime] 正在填充弹药......")
				<-time.After(time.Second * 5)
				if PoolsCache.Size(type_) == 0 {
					ctx.Send("[SetuTime] 等待填充，请稍后再试......")
					return
				}
			}
			// 从缓冲池里抽一张
			if id := ctx.Send(PoolsCache.GetOnePic(type_, FORM)); id == 0 {
				ctx.Send(fmt.Sprintf("ERROR: %v", errors.New("可能被风控了")))
			}
			return
		})

	zero.OnRegex(`^添加(.*?)(\d+)$`, FirstValueInList(PoolList), zero.SuperUserPermission).SetBlock(true).SetPriority(21).
		Handle(func(ctx *zero.Ctx) {
			var (
				type_  = ctx.State["regex_matched"].([]string)[1]
				id     = utils.Str2Int(ctx.State["regex_matched"].([]string)[2])
				illust = &utils.Illust{}
			)
			ctx.Send("少女祈祷中......")
			// 查询P站插图信息

			if err := illust.IllustInfo(id); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			// 下载插画
			if _, err := illust.PixivPicDown(PoolsCache.Path); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			file := fmt.Sprintf("%s%d.jpg", PoolsCache.Path, illust.Pid)
			if id := ctx.Send(illust.DetailPic(file)); id == 0 {
				ctx.Send(fmt.Sprintf("ERROR: %v", "可能被风控，发送失败"))
				return
			}
			// 添加插画到对应的数据库table
			if err := DB.Insert(type_, illust); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			ctx.Send("添加成功")
			return
		})

	zero.OnRegex(`^删除(.*?)(\d+)$`, FirstValueInList(PoolList), zero.SuperUserPermission).SetBlock(true).SetPriority(22).
		Handle(func(ctx *zero.Ctx) {
			var (
				type_ = ctx.State["regex_matched"].([]string)[1]
				id    = utils.Str2Int(ctx.State["regex_matched"].([]string)[2])
			)
			// 查询数据库
			if err := DB.Delete(type_, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
				ctx.Send(fmt.Sprintf("ERROR: %v", err))
				return
			}
			ctx.Send("删除成功")
			return
		})

	// 查询数据库涩图数量
	zero.OnFullMatchGroup([]string{"setu -s", "setu --status", ">setu status"}).SetBlock(true).SetPriority(23).
		Handle(func(ctx *zero.Ctx) {
			state := []string{"[SetuTime]"}
			for i := range PoolList {
				num, err := DB.Num(PoolList[i])
				if err != nil {
					num = 0
				}
				state = append(state, "\n")
				state = append(state, PoolList[i])
				state = append(state, ": ")
				state = append(state, fmt.Sprintf("%d", num))
			}
			ctx.Send(strings.Join(state, ""))
			return
		})
	// 开xml模式
	zero.OnFullMatchGroup([]string{"setu -x", "setu --xml", ">setu xml"}).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			FORM = "XML"
			ctx.Send("[SetuTime] XML->ON")
			return
		})
	// 关xml模式
	zero.OnFullMatchGroup([]string{"setu -p", "setu --pic", ">setu pic"}).SetBlock(true).SetPriority(24).
		Handle(func(ctx *zero.Ctx) {
			FORM = "PIC"
			ctx.Send("[SetuTime] XML->OFF")
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
