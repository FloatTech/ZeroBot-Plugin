// Package rsshub rss_hub订阅插件
package rsshub

import (
	"context"
	"fmt"
	"regexp"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/ZeroBot-Plugin/plugin/rsshub/domain"
)

// 初始化 repo
var (
	rssRepo      *domain.RssDomain
	initErr      error
	regexpForSQL = regexp.MustCompile(`[\^<>\[\]%&\*\(\)\{\}\|\=]|(union\s+select|update\s+|delete\s+|drop\s+|truncate\s+|insert\s+|exec\s+|declare\s+)`)
)

var (
	// 注册插件
	engine = control.Register("rsshub", &ctrl.Options[*zero.Ctx]{
		// 默认不启动
		DisableOnDefault: false,
		Brief:            "rsshub订阅姬",
		// 详细帮助
		Help: "rsshub订阅姬desu~ \n" +
			"支持的详细订阅列表文档可见：\n" +
			"https://rsshub.netlify.app/zh/ \n" +
			"- 添加rsshub订阅-/bookfere/weekly \n" +
			"- 删除rsshub订阅-/bookfere/weekly \n" +
			"- 查看rsshub订阅列表 \n" +
			"- rsshub同步 \n" +
			"Tips: 定时刷新rsshub订阅信息需要配合job一起使用, 全局只需要设置一个, 无视响应状态推送, 下为例子\n" +
			"记录在\"@every 10m\"触发的指令)\n" +
			"rsshub同步",
		// 插件数据存储路径
		PrivateDataFolder: "rsshub",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("rsshub订阅姬现在启动了哦"))
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("rsshub订阅姬现在关闭了哦"))
		},
	}).ApplySingle(zbpCtxExt.DefaultSingle)
)

// init 命令路由
func init() {
	rssRepo, initErr = domain.NewRssDomain(engine.DataFolder() + "rsshub.db")
	if initErr != nil {
		logrus.Errorln("rsshub订阅姬：初始化失败", initErr)
		panic(initErr)
	}
	engine.OnFullMatch("rsshub同步", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 群组-频道推送视图  map[群组]推送内容数组
		groupToFeedsMap, err := rssRepo.Sync(context.Background())
		if err != nil {
			logrus.Errorln("rsshub同步失败", err)
			ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("rsshub同步失败", err))
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
	engine.OnPrefix("添加rsshub订阅-", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		routeStr := ctx.State["args"].(string)
		input := regexpForSQL.ReplaceAllString(routeStr, "")
		logrus.Debugf("添加rsshub订阅：raw(%s), replaced(%s)", routeStr, input)
		rv, _, isSubExisted, err := rssRepo.Subscribe(context.Background(), ctx.Event.GroupID, input)
		if err != nil {
			ctx.SendChain(message.Text("rsshub订阅姬：添加失败", err.Error()))
			return
		}
		if isSubExisted {
			ctx.SendChain(message.Text("rsshub订阅姬：已存在，更新成功"))
		} else {
			ctx.SendChain(message.Text("rsshub订阅姬：添加成功\n", rv.Source.Title))
		}
		// 添加成功，发送订阅源快照
		msg, err := newRssDetailsMsg(ctx, rv)
		if len(msg) == 0 || err != nil {
			ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("rsshub推送错误", err))
			return
		}
		if id := ctx.Send(msg).ID(); id == 0 {
			ctx.SendChain(message.Text("ERROR: 发送订阅源快照失败，可能被风控了"))
		}
	})
	engine.OnPrefix("删除rsshub订阅-", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		routeStr := ctx.State["args"].(string)
		input := regexpForSQL.ReplaceAllString(routeStr, "")
		logrus.Debugf("删除rsshub订阅：raw(%s), replaced(%s)", routeStr, input)
		err := rssRepo.Unsubscribe(context.Background(), ctx.Event.GroupID, input)
		if err != nil {
			ctx.SendChain(message.Text("rsshub订阅姬：删除失败 ", err.Error()))
			return
		}
		ctx.SendChain(message.Text(fmt.Sprintf("rsshub订阅姬：删除%s成功", input)))
	})
	engine.OnFullMatch("查看rsshub订阅列表", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		rv, err := rssRepo.GetSubscribedChannelsByGroupID(context.Background(), ctx.Event.GroupID)
		if err != nil {
			ctx.SendChain(message.Text("rsshub订阅姬：查询失败 ", err.Error()))
			return
		}
		// 添加成功，发送订阅源信息
		msg, err := newRssSourcesMsg(ctx, rv)
		if err != nil {
			ctx.SendChain(message.Text("rsshub订阅姬：查询失败 ", err.Error()))
			return
		}
		if len(msg) == 0 {
			ctx.SendChain(message.Text("ん? 没有订阅的频道哦~"))
			return
		}
		ctx.SendChain(msg...)
	})
}

// sendRssUpdateMsg 发送Rss更新消息
func sendRssUpdateMsg(ctx *zero.Ctx, groupToFeedsMap map[int64][]*domain.RssClientView) {
	for groupID, views := range groupToFeedsMap {
		logrus.Infof("rsshub插件在群 %d 触发推送检查", groupID)
		for _, view := range views {
			if view == nil || len(view.Contents) == 0 {
				continue
			}
			msg, err := newRssDetailsMsg(ctx, view)
			if len(msg) == 0 || err != nil {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(rssHubPushErrMsg, err))
				continue
			}
			logrus.Infof("rsshub插件在群 %d 开始推送 %s", groupID, view.Source.Title)
			ctx.SendGroupMessage(groupID, message.Text(fmt.Sprintf("%s\n该rsshub频道下有更新了哦~", view.Source.Title)))
			if res := ctx.SendGroupForwardMessage(groupID, msg); !res.Exists() {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text(rssHubPushErrMsg))
			}
		}
	}
}
