// Package rsshub rss_hub订阅插件
package rsshub

import (
	"context"
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/rsshub/domain"
	"github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// 初始化 repo
var (
	rssRepo domain.RssDomain
	initErr error
	// getRssRepo repo 初始化方法，单例
	getRssRepo = ctxext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		logrus.WithContext(context.Background()).Infoln("RssHub订阅姬：初始化")
		rssRepo, initErr = domain.NewRssDomain(engine.DataFolder() + "rsshub.db")
		if initErr != nil {
			ctx.SendChain(message.Text("RssHub订阅姬：初始化失败", initErr.Error()))
			return false
		}
		return true
	})
)

var (
	// 注册插件
	engine = control.Register("rsshub", &ctrl.Options[*zero.Ctx]{
		// 默认不启动
		DisableOnDefault: false,
		Brief:            "RssHub订阅姬",
		// 详细帮助
		Help: "RssHub订阅姬desu~ \n" +
			"支持的详细订阅列表可见 https://rsshub.netlify.app/ \n" +
			"- 添加rsshub订阅-rsshub路由 \n" +
			"- 删除rsshub订阅-rsshub路由 \n" +
			"例：添加rsshub订阅-/bangumi/tv/calendar/today\n" +
			"- 查看rsshub订阅列表 \n" +
			"- rsshub同步 \n" +
			"Tips: 需要配合job一起使用, 全局只需要设置一个, 无视响应状态推送, 下为例子\n" +
			"记录在\"@every 10m\"触发的指令)\n" +
			"rsshub同步",
		// 插件数据存储路径
		PrivateDataFolder: "rsshub",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("RSS订阅姬现在启动了哦"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("RSS订阅姬现在关闭了哦"))
		},
	}).ApplySingle(zbpCtxExt.DefaultSingle)
)

// init 命令路由
func init() {
	engine.OnFullMatch("rsshub同步", zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 群组-频道推送视图  map[群组]推送内容数组
		groupToFeedsMap, err := rssRepo.Sync(context.Background())
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：同步任务失败 ", err))
			return
		}
		// 没有更新的[群组-频道推送视图]则不推送
		if len(groupToFeedsMap) == 0 {
			logrus.Info("rsshub未发现更新")
			return
		}
		sendRssUpdateMsg(ctx, groupToFeedsMap)
	})
	// 添加订阅
	engine.OnRegex(`^添加rsshub订阅-(.+)$`, zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		routeStr := ctx.State["regex_matched"].([]string)[1]
		rv, _, isSubExisted, err := rssRepo.Subscribe(context.Background(), ctx.Event.GroupID, routeStr)
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：添加失败", err.Error()))
			return
		}
		if isSubExisted {
			ctx.SendChain(message.Text("RSS订阅姬：已存在，更新成功"))
		} else {
			ctx.SendChain(message.Text("RSS订阅姬：添加成功"))
		}
		// 添加成功，发送订阅源快照
		msg, err := createRssUpdateMsg(ctx, rv)
		if len(msg) == 0 || err != nil {
			ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("RssHub推送错误", err))
			return
		}
		if id := ctx.Send(msg).ID(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 发送失败订阅源快照，可能被风控了"))
		}
	})
	engine.OnRegex(`^删除rsshub订阅-(.+)$`, zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		routeStr := ctx.State["regex_matched"].([]string)[1]
		err := rssRepo.Unsubscribe(context.Background(), ctx.Event.GroupID, routeStr)
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：删除失败 ", err.Error()))
			return
		}
		ctx.SendChain(message.Text(fmt.Sprintf("RSS订阅姬：删除%s成功", routeStr)))
	})
	engine.OnFullMatch("查看rsshub订阅列表", zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		rv, err := rssRepo.GetSubscribedChannelsByGroupID(context.Background(), ctx.Event.GroupID)
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：查询失败 ", err.Error()))
			return
		}
		// 添加成功，发送订阅源信息
		var msg message.Message
		msg = append(msg, message.Text("RSS订阅姬：当前订阅列表"))
		for _, v := range rv {
			msg = append(msg, message.Text(formatRssToTextMsg(v)))
		}
		ctx.SendChain(msg...)
	})
}

// sendRssUpdateMsg 发送Rss更新消息
func sendRssUpdateMsg(ctx *zero.Ctx, groupToFeedsMap map[int64][]*domain.RssClientView) {
	for groupID, views := range groupToFeedsMap {
		logrus.Infof("RssHub插件在群 %d 触发推送检查", groupID)
		for _, view := range views {
			if view == nil || len(view.Contents) == 0 {
				continue
			}
			msg, err := createRssUpdateMsg(ctx, view)
			if len(msg) == 0 || err != nil {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(rssHubPushErrMsg, err))
				continue
			}
			logrus.Infof("RssHub插件在群 %d 开始推送 %s", groupID, view.Source.Title)
			ctx.SendGroupMessage(groupID, message.Text(view.Source.Title+"\n[RSS订阅姬定时推送]"))
			if res := ctx.SendGroupForwardMessage(groupID, msg); !res.Exists() {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(rssHubPushErrMsg))
			}
		}
	}
}

// createRssUpdateMsg 创建Rss更新消息
func createRssUpdateMsg(ctx *zero.Ctx, view *domain.RssClientView) (message.Message, error) {
	msgSlice, err := formatRssToMsg(view)
	if err != nil {
		return nil, err
	}
	msg := make(message.Message, len(msgSlice))
	for i, item := range msgSlice {
		msg[i] = fakeSenderForwardNode(ctx.Event.SelfID, item...)
	}
	// 发送文字版
	//msg:= formatRssToTextMsg(view)
	//if err != nil {
	//	return nil, err
	//}
	//msg := make(message.Message, 2)
	//msg[0] = message.Image("base64://" + binary.BytesToString(pic))
	//msg[1] = message.Text(ctx.Event.SelfID, message.Text(view.Source.Link))
	return msg, nil
}
