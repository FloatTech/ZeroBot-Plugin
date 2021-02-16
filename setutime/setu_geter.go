package setutime

import (
	"errors"
	"fmt"
	"time"

	utils "bot/setutime/utils"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/extension/rate"
)

var limit = rate.NewManager(time.Minute*1, 5)

type setuGet struct{} // setuGet 来份色图

func (_ setuGet) GetPluginInfo() zero.PluginInfo { // 返回插件信息
	return zero.PluginInfo{
		Author:     "kanri",
		PluginName: "SetuGet",
		Version:    "0.0.1",
		Details:    "来份色图",
	}
}

var (
	BOTPATH   = utils.PathExecute()        // 当前bot运行目录
	DATAPATH  = BOTPATH + "data/SetuTime/" // 数据目录
	DBPATH    = DATAPATH + "SetuTime.db"   // 数据库路径
	CACHEPATH = DATAPATH + "cache/"        // 缓冲图片路径

	DB = utils.Sqlite{DBPath: DBPATH} // 涩图数据库

	pool = utils.PicsCache{Max: 10} // 图片缓冲池子
)

func init() {
	zero.RegisterPlugin(setuGet{}) // 注册插件

	utils.CACHE_GROUP = 868047498 // 图片缓冲群
	utils.CACHEPATH = CACHEPATH   // 缓冲图片路径

	utils.CreatePath(DBPATH)
	utils.CreatePath(CACHEPATH)

	ecy := &ecy{}
	setu := &setu{}
	scenery := &scenery{}
	DB.DBCreate(ecy)
	DB.DBCreate(setu)
	DB.DBCreate(scenery)
}

// ecy 二次元
type ecy struct {
	utils.Illust
}

// setu 涩图
type setu struct {
	utils.Illust
}

// scenery 风景
type scenery struct {
	utils.Illust
}

func (_ setuGet) Start() { // 插件主体
	zero.OnFullMatchGroup([]string{"来份涩图", "setu", "来份色图"}).SetBlock(true).SetPriority(20).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			if limit.Load(event.UserID).Acquire() == false {
				zero.Send(event, "请稍后重试0x0...")
				return zero.FinishResponse
			}
			var (
				type_  = "setu"
				illust = &setu{}
			)
			// TODO 池子无图片则立刻下载
			length := illust.Len(type_, &pool)
			if length == 0 {
				zero.Send(event, "[SetuTime] 正在填充弹药......")
				if err := DB.DBSelect(illust, "ORDER BY RANDOM() limit 1"); err != nil {
					utils.SendError(event, err) // 查询出一张图片
					return zero.FinishResponse
				}
				if err := illust.Add(type_, &pool); err != nil {
					utils.SendError(event, err) // 向缓冲池添加一张图片
					return zero.FinishResponse
				}
			}
			// TODO 补充池子
			go func() {
				times := utils.Min(pool.Max-length, 2)
				for i := 0; i < times; i++ {
					if err := DB.DBSelect(illust, "ORDER BY RANDOM() limit 1"); err != nil {
						utils.SendError(event, err) // 查询出一张图片
					}
					if err := illust.Add(type_, &pool); err != nil {
						utils.SendError(event, err) // 向缓冲池添加一张图片
					}
				}
			}()
			// TODO 从缓冲池里抽一张
			hash := illust.Get(type_, &pool)
			if utils.XML {
				if id := zero.Send(event, illust.BigPic(hash)); id == 0 {
					utils.SendError(event, errors.New("可能被风控了"))
				}
			} else {
				if id := zero.Send(event, illust.NormalPic()); id == 0 {
					utils.SendError(event, errors.New("可能被风控了"))
				}
			}
			return zero.FinishResponse
		})
	zero.OnFullMatchGroup([]string{"二次元", "ecy", "来份二次元"}).SetBlock(true).SetPriority(20).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			if limit.Load(event.UserID).Acquire() == false {
				zero.Send(event, "请稍后重试0x0...")
				return zero.FinishResponse
			}
			var (
				type_  = "ecy"
				illust = &ecy{}
			)
			// TODO 池子无图片则立刻下载
			length := illust.Len(type_, &pool)
			if length == 0 {
				zero.Send(event, "[SetuTime] 正在填充弹药......")
				if err := DB.DBSelect(illust, "ORDER BY RANDOM() limit 1"); err != nil {
					utils.SendError(event, err) // 查询出一张图片
					return zero.FinishResponse
				}
				if err := illust.Add(type_, &pool); err != nil {
					utils.SendError(event, err) // 向缓冲池添加一张图片
					return zero.FinishResponse
				}
			}
			// TODO 补充池子
			go func() {
				times := utils.Min(pool.Max-length, 2)
				for i := 0; i < times; i++ {
					if err := DB.DBSelect(illust, "ORDER BY RANDOM() limit 1"); err != nil {
						utils.SendError(event, err) // 查询出一张图片
					}
					if err := illust.Add(type_, &pool); err != nil {
						utils.SendError(event, err) // 向缓冲池添加一张图片
					}
				}
			}()
			// TODO 从缓冲池里抽一张
			hash := illust.Get(type_, &pool)
			if utils.XML {
				if id := zero.Send(event, illust.BigPic(hash)); id == 0 {
					utils.SendError(event, errors.New("可能被风控了"))
				}
			} else {
				if id := zero.Send(event, illust.NormalPic()); id == 0 {
					utils.SendError(event, errors.New("可能被风控了"))
				}
			}
			return zero.FinishResponse
		})
	zero.OnFullMatchGroup([]string{"风景", "来份风景"}).SetBlock(true).SetPriority(20).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			if limit.Load(event.UserID).Acquire() == false {
				zero.Send(event, "请稍后重试0x0...")
				return zero.FinishResponse
			}
			var (
				type_  = "scenery"
				illust = &scenery{}
			)
			// TODO 池子无图片则立刻下载
			length := illust.Len(type_, &pool)
			if length == 0 {
				zero.Send(event, "[SetuTime] 正在填充弹药......")
				if err := DB.DBSelect(illust, "ORDER BY RANDOM() limit 1"); err != nil {
					utils.SendError(event, err) // 查询出一张图片
					return zero.FinishResponse
				}
				if err := illust.Add(type_, &pool); err != nil {
					utils.SendError(event, err) // 向缓冲池添加一张图片
					return zero.FinishResponse
				}
			}
			// TODO 补充池子
			go func() {
				times := utils.Min(pool.Max-length, 2)
				for i := 0; i < times; i++ {
					if err := DB.DBSelect(illust, "ORDER BY RANDOM() limit 1"); err != nil {
						utils.SendError(event, err) // 查询出一张图片
					}
					if err := illust.Add(type_, &pool); err != nil {
						utils.SendError(event, err) // 向缓冲池添加一张图片
					}
				}
			}()
			// TODO 从缓冲池里抽一张
			hash := illust.Get(type_, &pool)
			if utils.XML {
				if id := zero.Send(event, illust.BigPic(hash)); id == 0 {
					utils.SendError(event, errors.New("可能被风控了"))
				}
			} else {
				if id := zero.Send(event, illust.NormalPic()); id == 0 {
					utils.SendError(event, errors.New("可能被风控了"))
				}
			}
			return zero.FinishResponse
		})
	zero.OnRegex(`添加涩图(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(21).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var illust = &setu{}
			zero.Send(event, "少女祈祷中......")
			// TODO 查询P站插图信息
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			if err := illust.IllustInfo(id); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			// TODO 下载插画
			if _, err := illust.PixivPicDown(CACHEPATH); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			if id := zero.Send(event, illust.DetailPic()); id == 0 {
				utils.SendError(event, errors.New("可能被风控了"))
				return zero.FinishResponse
			}
			// TODO 添加插画到对应的数据库table
			if err := DB.DBInsert(illust); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			zero.Send(event, "添加成功")
			return zero.FinishResponse
		})
	zero.OnRegex(`添加二次元(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(21).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var illust = &ecy{}
			zero.Send(event, "少女祈祷中......")
			// TODO 查询P站插图信息
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			if err := illust.IllustInfo(id); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			// TODO 下载插画
			if _, err := illust.PixivPicDown(CACHEPATH); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			if id := zero.Send(event, illust.DetailPic()); id == 0 {
				utils.SendError(event, errors.New("可能被风控了"))
				return zero.FinishResponse
			}
			// TODO 添加插画到对应的数据库table
			if err := DB.DBInsert(illust); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			zero.Send(event, "添加成功")
			return zero.FinishResponse
		})
	zero.OnRegex(`添加风景(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(21).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var illust = &scenery{}
			zero.Send(event, "少女祈祷中......")
			// TODO 查询P站插图信息
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			if err := illust.IllustInfo(id); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			// TODO 下载插画
			if _, err := illust.PixivPicDown(CACHEPATH); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			if id := zero.Send(event, illust.DetailPic()); id == 0 {
				utils.SendError(event, errors.New("可能被风控了"))
				return zero.FinishResponse
			}
			// TODO 添加插画到对应的数据库table
			if err := DB.DBInsert(illust); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			zero.Send(event, "添加成功")
			return zero.FinishResponse
		})
	zero.OnRegex(`删除涩图(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(22).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var illust = &setu{}
			// TODO 查询数据库
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			if err := DB.DBDelete(illust, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			zero.Send(event, "删除成功")
			return zero.FinishResponse
		})
	zero.OnRegex(`删除二次元(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(22).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var illust = &ecy{}
			// TODO 查询数据库
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			if err := DB.DBDelete(illust, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			zero.Send(event, "删除成功")
			return zero.FinishResponse
		})
	zero.OnRegex(`删除风景(\d+)`, zero.SuperUserPermission).SetBlock(true).SetPriority(22).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			var illust = &scenery{}
			// TODO 查询数据库
			id := utils.Str2Int(state["regex_matched"].([]string)[1])
			if err := DB.DBDelete(illust, fmt.Sprintf("WHERE pid=%d", id)); err != nil {
				utils.SendError(event, err)
				return zero.FinishResponse
			}
			zero.Send(event, "删除成功")
			return zero.FinishResponse
		})
	// TODO 查询数据库涩图数量
	zero.OnFullMatchGroup([]string{"setu -s", "setu --status"}, zero.SuperUserPermission).SetBlock(true).SetPriority(23).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			setu, _ := DB.DBNum(&setu{})
			ecy, _ := DB.DBNum(&ecy{})
			scenery, _ := DB.DBNum(&scenery{})
			zero.Send(event, fmt.Sprintf("[SetuTime] \n风景：%d \n二次元：%d \n涩图：%d", scenery, ecy, setu))
			return zero.FinishResponse
		})
	// TODO 开xml模式
	zero.OnFullMatchGroup([]string{"setu -x", "setu --xml"}, zero.AdminPermission).SetBlock(true).SetPriority(24).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			utils.XML = true
			zero.Send(event, "[SetuTime] XML->ON")
			return zero.FinishResponse
		})
	// TODO 关xml模式
	zero.OnFullMatchGroup([]string{"setu -p", "setu --pic"}, zero.AdminPermission).SetBlock(true).SetPriority(24).
		Handle(func(matcher *zero.Matcher, event zero.Event, state zero.State) zero.Response {
			utils.XML = false
			zero.Send(event, "[SetuTime] XML->OFF")
			return zero.FinishResponse
		})
}
