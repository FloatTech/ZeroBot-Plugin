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
		logrus.WithContext(context.Background()).Infoln("RSS订阅姬：初始化")
		rssRepo, initErr = domain.NewRssDomain(engine.DataFolder() + "rsshub.db")
		if initErr != nil {
			ctx.SendChain(message.Text("RSS订阅姬：初始化失败", initErr.Error()))
			return false
		}
		return true
	})
)

var (
	// 注册插件
	engine = control.Register("RssHub", &ctrl.Options[*zero.Ctx]{
		// 默认不启动
		DisableOnDefault: false,
		Brief:            "RssHub订阅姬",
		// 详细帮助
		Help: "RssHub订阅姬desu~ 支持的详细订阅列表可见 https://rsshub.netlify.app/ \n" +
			"- 添加RssHub订阅-RssHub路由 \n" +
			"- 删除RssHub订阅-RssHub路由 \n" +
			"例：添加RssHub订阅-/bangumi/tv/calendar/today\n" +
			"- 查看RssHub订阅列表 \n" +
			"- RssHub同步 \n" +
			"Tips: 需要配合job一起使用, 全局只需要设置一个, 无视响应状态推送, 下为例子\n" +
			"记录在\"@every 10m\"触发的指令)\n" +
			"RssHub同步",
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
	engine.OnFullMatch("RssHub同步", zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		m, ok := control.Lookup("RssHub")
		if !ok {
			logrus.Warn("RssHub插件未启用")
			return
		}
		// 群组-频道推送视图  map[群组]推送内容数组
		groupToFeedsMap, err := rssRepo.Sync(context.Background())
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：同步任务失败 ", err.Error()))
			return
		}
		// 没有更新的[群组-频道推送视图]则不推送
		if len(groupToFeedsMap) == 0 {
			logrus.Info("RssHub未发现更新")
			return
		}
		sendRssUpdateMsg(ctx, groupToFeedsMap, m)
	})
	// 添加订阅
	engine.OnRegex(`^添加RssHub订阅-(.+)$`, zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
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
		// 添加成功，发送订阅源信息
		msg := make(message.Message, 0)
		rawMsgSlice := formatRssToTextMsg(rv)
		for _, rm := range rawMsgSlice {
			msg = append(msg, fakeSenderForwardNode(ctx.Event.SelfID, message.Text(rm)))
		}
		//m := message.Message{zbpCtxExt.FakeSenderForwardNode(ctx, msg...)}
		if id := ctx.Send(msg).ID(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 可能被风控了"))
		}
		//ctx.SendChain(msg...)
	})
	engine.OnRegex(`^删除RssHub订阅-(.+)$`, zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		routeStr := ctx.State["regex_matched"].([]string)[1]
		err := rssRepo.Unsubscribe(context.Background(), ctx.Event.GroupID, routeStr)
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：删除失败 ", err.Error()))
			return
		}
		// 添加成功，发送订阅源信息
		var msg []message.MessageSegment
		msg = append(msg, message.Text(fmt.Sprintf("RSS订阅姬：删除%s成功", routeStr)))
		ctx.SendChain(msg...)
	})
	engine.OnFullMatch("查看RssHub订阅列表", zero.OnlyGroup, getRssRepo).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		rv, err := rssRepo.GetSubscribedChannelsByGroupID(context.Background(), ctx.Event.GroupID)
		if err != nil {
			ctx.SendChain(message.Text("RSS订阅姬：查询失败 ", err.Error()))
			return
		}
		// 添加成功，发送订阅源信息
		var msg []message.MessageSegment
		msg = append(msg, message.Text("RSS订阅姬：当前订阅列表"))
		for _, v := range rv {
			msg = append(msg, message.Text(formatRssToTextMsg(v)))
		}
		ctx.SendChain(msg...)
	})
}

// sendRssUpdateMsg 发送Rss更新消息
func sendRssUpdateMsg(ctx *zero.Ctx, groupToFeedsMap map[int64][]*domain.RssClientView, m *ctrl.Control[*zero.Ctx]) {
	for groupID, views := range groupToFeedsMap {
		logrus.Infof("RssHub插件在群 %d 触发推送检查", groupID)
		if !m.IsEnabledIn(groupID) {
			continue
		}
		for _, view := range views {
			if len(view.Contents) == 0 {
				continue
			}
			msg := createRssUpdateMsg(ctx, view)
			if len(msg) == 0 {
				continue
			}
			logrus.Infof("RssHub插件在群 %d 开始推送 %s", groupID, view.Source.Title)
			ctx.SendGroupMessage(groupID, message.Text(view.Source.Title+"\n[RSS订阅姬定时推送]\n"))
			if res := ctx.SendGroupForwardMessage(groupID, msg); !res.Exists() {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("RssHub推送错误"))
			}
		}
	}
}

// createRssUpdateMsg 创建Rss更新消息
func createRssUpdateMsg(ctx *zero.Ctx, view *domain.RssClientView) message.Message {
	msgSlice := formatRssToTextMsg(view)
	msg := make(message.Message, len(msgSlice))
	for i, text := range msgSlice {
		msg[i] = fakeSenderForwardNode(ctx.Event.SelfID, message.Text(text))
	}
	return msg
}
