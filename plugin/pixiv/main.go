// Package pixiv ...
package pixiv

import (
	"math/rand"
	"os"
	"strconv"

	"github.com/FloatTech/AnimeAPI/pixiv"
	"github.com/FloatTech/AnimeAPI/pixiv/model"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/pixiv/cache"
	"github.com/FloatTech/floatbox/file"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var defaultKeyword = []string{"萝莉", "御姐", "妹妹", "姐姐", "车万"}

var (
	service *Service
)

func init() {
	if file.IsNotExist("data/pixiv") {
		err := os.MkdirAll("data/pixiv", 0775)
		if err != nil {
			panic(err)
		}
	}
	if file.IsNotExist(pixivTempDir) {
		err := os.MkdirAll(pixivTempDir, 0775)
		if err != nil {
			panic(err)
		}
	}

	db := cache.NewDB("data/pixiv/pixiv.db")

	var t1 model.RefreshToken
	if err := db.First(&t1).Error; err != nil {
		log.Warning("Fail fetching token store from database")
	}

	pixivAPI := pixiv.NewPixivAPI(t1.Token)

	var proxyCfg model.PixivProxyConfig
	proxyCfg.Name = "global"
	if err := db.Where("name = ?", proxyCfg.Name).FirstOrCreate(&proxyCfg).Error; err == nil {
		if err := pixivAPI.Client.SetProxy(proxyCfg.Proxy); err != nil {
			log.Warning("设置pixiv代理错误: ", err)
		}
	}

	service = NewService(db, pixivAPI)
}

const help = `
- [x张]涩图 [关键词]
- 每日涩图
- [x张]画师[画师的uid]
- p站搜图[插画pid]
- 设置p站token [token]
- 授权此处使用p站18
- 设置pixiv代理 [http://127.0.0.1:7890]
- 查看pixiv代理
- 清除pixiv代理
[]不用打出来这只是一个占位符
可添加多个关键词每个关键词用空格隔开
默认不发R-18如果要发就加一个R-18关键词
`

func init() {
	engine := control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "Pixiv 图片搜索",
		Help:             help,
	})

	engine.OnFullMatch("授权此处使用p站18", zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		id := ctx.Event.GroupID
		if id == 0 {
			id = -ctx.Event.UserID
		}
		if err := service.DB.Create(&model.GroupR18Permission{GroupID: id}).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("已允许"))
	})

	engine.OnRegex(`^设置p站token\s+(.*)`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		token := ctx.State["regex_matched"].([]string)[1]
		var refreshToken model.RefreshToken
		refreshToken.User = ctx.Event.UserID
		refreshToken.Token = token
		if err := service.DB.Save(&refreshToken).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		service.API.Token.RefreshToken = token

		ctx.SendChain(message.Text("Pixiv Token: ", token))
	})

	engine.OnRegex(`^设置pixiv代理\s+([\s\S]+)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		proxy := ctx.State["regex_matched"].([]string)[1]

		var proxyCfg model.PixivProxyConfig
		proxyCfg.Name = "global"
		if err := service.DB.Where("name = ?", proxyCfg.Name).FirstOrCreate(&proxyCfg).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		proxyCfg.Proxy = proxy
		if err := service.DB.Save(&proxyCfg).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if err := service.API.Client.SetProxy(proxy); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		ctx.SendChain(message.Text("Pixiv client proxy 已设置为: ", proxy))
	})

	engine.OnFullMatch("清除pixiv代理", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var proxyCfg model.PixivProxyConfig
		proxyCfg.Name = "global"
		if err := service.DB.Where("name = ?", proxyCfg.Name).FirstOrCreate(&proxyCfg).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		proxyCfg.Proxy = ""
		if err := service.DB.Save(&proxyCfg).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if err := service.API.Client.SetProxy(""); err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		ctx.SendChain(message.Text("Pixiv client proxy 已清除"))
	})

	engine.OnFullMatch("查看pixiv代理", zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		var proxyCfg model.PixivProxyConfig
		proxyCfg.Name = "global"
		if err := service.DB.Where("name = ?", proxyCfg.Name).FirstOrCreate(&proxyCfg).Error; err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if proxyCfg.Proxy == "" {
			ctx.SendChain(message.Text("Pixiv client proxy: 未设置"))
			return
		}
		ctx.SendChain(message.Text("Pixiv client proxy: ", proxyCfg.Proxy))
	})

	engine.OnRegex(`^p站搜图(\d+)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		rawPID := ctx.State["regex_matched"].([]string)[1]
		pid, err := strconv.ParseInt(rawPID, 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		illust, err := service.API.FetchPixivByPID(pid)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		_ = service.DB.Create(illust)
		service.SendIllusts(ctx, []model.IllustCache{*illust})
	})

	engine.OnRegex(`^(\d+)?张?画师(\d+)`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if !service.Acquire(ctx.Event.UserID) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("上一个任务还没结束，请稍后再试"))
			return
		}
		defer service.Release(ctx.Event.UserID)

		limit := ctx.State["regex_matched"].([]string)[1]
		if limit == "" {
			limit = "1"
		}
		rawUID := ctx.State["regex_matched"].([]string)[2]
		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		uid, err := strconv.ParseInt(rawUID, 10, 64)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		pictureIDs, err := service.DB.GetSentPictureIDs(ctx.Event.GroupID)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		illustInfos, err := service.API.FetchPixivByUser(uid, limitInt, pictureIDs)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}
		if len(illustInfos) == 0 {
			ctx.SendChain(message.Text("没有找到图片"))
			return
		}

		service.SendIllusts(ctx, illustInfos)
	})

	engine.OnRegex(`^每日[色|涩|瑟]图$`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		illusts, err := service.API.FetchPixivRecommend(1)
		if err != nil {
			ctx.SendChain(message.Text("发送涩图失败惹"))
			return
		}
		service.SendIllusts(ctx, []model.IllustCache{illusts[0]})
	})

	engine.OnRegex(`^(\d+)?张?[色|瑟|涩]图\s*(.+)?`).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		limit := ctx.State["regex_matched"].([]string)[1]
		keyword := ctx.State["regex_matched"].([]string)[2]

		if !service.Acquire(ctx.Event.UserID) {
			ctx.SendChain(message.Reply(ctx.Event.MessageID), message.Text("上一个任务还没结束，请稍后再试"))
			return
		}
		defer service.Release(ctx.Event.UserID)

		if limit == "" {
			limit = "1"
		}

		if keyword == "" {
			keyword = defaultKeyword[rand.Intn(len(defaultKeyword))]
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		if limitInt > 10 {
			ctx.SendChain(message.Text("图片太多了"))
			return
		}

		gid := ctx.Event.GroupID
		r18Req := pixiv.IsR18(keyword)
		cleanKeyword := pixiv.RemoveR18Keywords(keyword)

		if gid == 0 {
			gid = -ctx.Event.UserID
		}

		if r18Req && !service.DB.CheckGroupR18Permission(gid) {
			ctx.SendChain(message.Text([]string{
				"笨蛋笨蛋大笨蛋",
				"此处未授权r18",
			}[rand.Intn(2)]))
			return
		}

		cachedIllusts, err := service.DB.FindIllustsSmart(gid, cleanKeyword, limitInt, r18Req)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		cached, _ := service.DB.GetSentPictureIDs(gid)

		for _, ill := range cachedIllusts {
			cached = append(cached, ill.PID)
		}

		illusts, err := service.API.GetIllustsByKeyword(keyword, limitInt, cachedIllusts, cached)
		if err != nil {
			ctx.SendChain(message.Text("ERROR: ", err))
			return
		}

		for _, illust := range illusts {
			log.Print("获取到图片 PID: ", illust.PID)
			_ = service.DB.Create(&illust).Error
		}

		service.SendIllusts(ctx, illusts)
		service.BackgroundCacheFiller(cleanKeyword, 10, r18Req, 5, ctx.Event.GroupID)
	})
}
