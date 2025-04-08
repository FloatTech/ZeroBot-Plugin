// Package minecraftobserver 通过mc服务器地址获取服务器状态信息并绘制图片发送到QQ群
package minecraftobserver

import (
	"fmt"
	"strings"
	"time"

	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	name = "minecraftobserver"
)

var (
	// 注册插件
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		// 默认不启动
		DisableOnDefault: false,
		Brief:            "Minecraft服务器状态查询/订阅",
		// 详细帮助
		Help: "- mc服务器状态 [服务器IP/URI]\n" +
			"- mc服务器添加订阅 [服务器IP/URI]\n" +
			"- mc服务器取消订阅 [服务器IP/URI]\n" +
			"- mc服务器订阅拉取 （需要插件定时任务配合使用，全局只需要设置一个）" +
			"-----------------------\n" +
			"使用job插件设置定时, 例:" +
			"记录在\"@every 1m\"触发的指令\n" +
			"（机器人回答：您的下一条指令将被记录，在@@every 1m时触发）" +
			"mc服务器订阅拉取",
		// 插件数据存储路径
		PrivateDataFolder: name,
	}).ApplySingle(zbpCtxExt.DefaultSingle)
)

func init() {
	// 状态查询
	engine.OnRegex("^[mM][cC]服务器状态 (.+)$").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		addr := ctx.State["regex_matched"].([]string)[1]
		resp, err := getMinecraftServerStatus(addr)
		if err != nil {
			ctx.Send(message.Text("服务器状态获取失败... 错误信息: ", err))
			return
		}
		status := resp.genServerSubscribeSchema(addr, 0)
		textMsg, iconBase64 := status.generateServerStatusMsg()
		var msg message.Message
		if iconBase64 != "" {
			msg = append(msg, message.Image(iconBase64))
		}
		msg = append(msg, message.Text(textMsg))
		if id := ctx.Send(msg); id.ID() == 0 {
			// logrus.Errorln(logPrefix + "Send failed")
			return
		}
	})
	// 添加订阅
	engine.OnRegex(`^[mM][cC]服务器添加订阅\s*(.+)$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		addr := ctx.State["regex_matched"].([]string)[1]
		status, err := getMinecraftServerStatus(addr)
		if err != nil {
			ctx.Send(message.Text("服务器信息初始化失败，请检查服务器是否可用！\n错误信息: ", err))
			return
		}
		targetID, targetType := warpTargetIDAndType(ctx.Event.GroupID, ctx.Event.UserID)
		err = dbInstance.newSubscribe(addr, targetID, targetType)
		if err != nil {
			ctx.Send(message.Text("订阅添加失败... 错误信息: ", err))
			return
		}
		// 插入数据库（首条，需要更新状态）
		err = dbInstance.updateServerStatus(status.genServerSubscribeSchema(addr, 0))
		if err != nil {
			ctx.Send(message.Text("服务器状态更新失败... 错误信息: ", err))
			return
		}
		if sid := ctx.Send(message.Text(fmt.Sprintf("服务器 %s 订阅添加成功", addr))); sid.ID() == 0 {
			// logrus.Errorln(logPrefix + "Send failed")
			return
		}
		// 成功后立即发送一次状态
		textMsg, iconBase64 := status.genServerSubscribeSchema(addr, 0).generateServerStatusMsg()
		var msg message.Message
		if iconBase64 != "" {
			msg = append(msg, message.Image(iconBase64))
		}
		msg = append(msg, message.Text(textMsg))
		if id := ctx.Send(msg); id.ID() == 0 {
			// logrus.Errorln(logPrefix + "Send failed")
			return
		}
	})
	// 删除
	engine.OnRegex(`^[mM][cC]服务器取消订阅\s*(.+)$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		addr := ctx.State["regex_matched"].([]string)[1]
		// 通过群组id和服务器地址获取服务器状态
		targetID, targetType := warpTargetIDAndType(ctx.Event.GroupID, ctx.Event.UserID)
		err := dbInstance.deleteSubscribe(addr, targetID, targetType)
		if err != nil {
			ctx.Send(message.Text("取消订阅失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		ctx.Send(message.Text("取消订阅成功"))
	})
	// 查看当前渠道的所有订阅
	engine.OnRegex(`^[mM][cC]服务器订阅列表$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		subList, err := dbInstance.getSubscribesByTarget(warpTargetIDAndType(ctx.Event.GroupID, ctx.Event.UserID))
		if err != nil {
			ctx.Send(message.Text("获取订阅列表失败... 错误信息: ", err))
			return
		}
		if len(subList) == 0 {
			ctx.Send(message.Text("当前没有订阅哦"))
			return
		}
		stringBuilder := strings.Builder{}
		stringBuilder.WriteString("[订阅列表]\n")
		for _, v := range subList {
			stringBuilder.WriteString(fmt.Sprintf("服务器地址: %s\n", v.ServerAddr))
		}
		if sid := ctx.Send(message.Text(stringBuilder.String())); sid.ID() == 0 {
			// logrus.Errorln(logPrefix + "Send failed")
			return
		}
	})
	// 查看全局订阅情况（仅限管理员私聊可用）
	engine.OnRegex(`^[mM][cC]服务器全局订阅列表$`, zero.OnlyPrivate, zero.SuperUserPermission, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		subList, err := dbInstance.getAllSubscribes()
		if err != nil {
			ctx.Send(message.Text("获取全局订阅列表失败... 错误信息: ", err))
			return
		}
		if len(subList) == 0 {
			ctx.Send(message.Text("当前一个订阅都没有哦"))
			return
		}
		userID := ctx.Event.UserID
		userName := ctx.CardOrNickName(userID)
		msg := make(message.Message, 0)

		// 按照群组or用户分组来定
		groupSubMap := make(map[int64][]serverSubscribe)
		userSubMap := make(map[int64][]serverSubscribe)
		for _, v := range subList {
			switch v.TargetType {
			case targetTypeGroup:
				groupSubMap[v.TargetID] = append(groupSubMap[v.TargetID], v)
			case targetTypeUser:
				userSubMap[v.TargetID] = append(userSubMap[v.TargetID], v)
			default:
			}
		}

		// 群
		for k, v := range groupSubMap {
			stringBuilder := strings.Builder{}
			stringBuilder.WriteString(fmt.Sprintf("[群 %d]存在以下订阅:\n", k))
			for _, sub := range v {
				stringBuilder.WriteString(fmt.Sprintf("服务器地址: %s\n", sub.ServerAddr))
			}
			msg = append(msg, message.CustomNode(userName, userID, stringBuilder.String()))
		}
		// 个人
		for k, v := range userSubMap {
			stringBuilder := strings.Builder{}
			stringBuilder.WriteString(fmt.Sprintf("[用户 %d]存在以下订阅:\n", k))
			for _, sub := range v {
				stringBuilder.WriteString(fmt.Sprintf("服务器地址: %s\n", sub.ServerAddr))
			}
			msg = append(msg, message.CustomNode(userName, userID, stringBuilder.String()))
		}
		// 合并发送
		ctx.SendPrivateForwardMessage(ctx.Event.UserID, msg)
	})
	// 状态变更通知，全局触发，逐个服务器检查，检查到变更则逐个发送通知
	engine.OnRegex(`^[mM][cC]服务器订阅拉取$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		serverList, err := dbInstance.getAllSubscribes()
		if err != nil {
			su := zero.BotConfig.SuperUsers[0]
			// 如果订阅列表获取失败，通知管理员
			ctx.SendPrivateMessage(su, message.Text(logPrefix, "获取订阅列表失败..."))
			return
		}
		// logrus.Debugln(logPrefix+"global get ", len(serverList), " subscribe(s)")
		serverMap := make(map[string][]serverSubscribe)
		for _, v := range serverList {
			serverMap[v.ServerAddr] = append(serverMap[v.ServerAddr], v)
		}
		changedCount := 0
		for subAddr, oneServerSubList := range serverMap {
			// 查询当前存储的状态
			storedStatus, sErr := dbInstance.getServerStatus(subAddr)
			if sErr != nil {
				// logrus.Errorln(logPrefix+fmt.Sprintf("getServerStatus ServerAddr(%s) error: ", subAddr), sErr)
				continue
			}
			isChanged, changedNotifyMsg, sErr := singleServerScan(storedStatus)
			if sErr != nil {
				// logrus.Errorln(logPrefix+"singleServerScan error: ", sErr)
				continue
			}
			if !isChanged {
				continue
			}
			changedCount++
			// 发送变化信息
			for _, subInfo := range oneServerSubList {
				time.Sleep(100 * time.Millisecond)
				if subInfo.TargetType == targetTypeUser {
					ctx.SendPrivateMessage(subInfo.TargetID, changedNotifyMsg)
				} else if subInfo.TargetType == targetTypeGroup {
					m, ok := control.Lookup(name)
					if !ok {
						continue
					}
					if !m.IsEnabledIn(subInfo.TargetID) {
						continue
					}
					ctx.SendGroupMessage(subInfo.TargetID, changedNotifyMsg)
				}
			}
		}
	})
}

// singleServerScan 单个服务器状态扫描
func singleServerScan(oldSubStatus *serverStatus) (changed bool, notifyMsg message.Message, err error) {
	notifyMsg = make(message.Message, 0)
	newSubStatus := &serverStatus{}
	// 获取服务器状态 & 检查是否需要更新
	rawServerStatus, err := getMinecraftServerStatus(oldSubStatus.ServerAddr)
	if err != nil {
		// logrus.Warnln(logPrefix+"getMinecraftServerStatus error: ", err)
		err = nil
		// 计数器没有超限，增加计数器并跳过
		if cnt, ts := addPingServerUnreachableCounter(oldSubStatus.ServerAddr, time.Now()); cnt < pingServerUnreachableCounterThreshold &&
			time.Since(ts) < pingServerUnreachableCounterTimeThreshold {
			// logrus.Warnln(logPrefix+"server ", oldSubStatus.ServerAddr, " unreachable, counter: ", cnt, " ts:", ts)
			return
		}
		// 不可达计数器已经超限，则更新服务器状态
		// 深拷贝，设置PingDelay为不可达
		newSubStatus = oldSubStatus.deepCopy()
		newSubStatus.PingDelay = pingDelayUnreachable
	} else {
		newSubStatus = rawServerStatus.genServerSubscribeSchema(oldSubStatus.ServerAddr, oldSubStatus.ID)
	}
	if newSubStatus == nil {
		// logrus.Errorln(logPrefix + "newSubStatus is nil")
		return
	}
	// 检查是否有订阅信息变化
	if oldSubStatus.isServerStatusSpecChanged(newSubStatus) {
		// logrus.Warnf(logPrefix+"server subscribe spec changed: (%+v) -> (%+v)", oldSubStatus, newSubStatus)
		changed = true
		// 更新数据库
		err = dbInstance.updateServerStatus(newSubStatus)
		if err != nil {
			// logrus.Errorln(logPrefix+"updateServerSubscribeStatus error: ", err)
			return
		}
		// 纯文本信息
		notifyMsg = append(notifyMsg, message.Text(formatSubStatusChangeText(oldSubStatus, newSubStatus)))
		// 如果有图标变更
		if oldSubStatus.FaviconMD5 != newSubStatus.FaviconMD5 {
			// 有图标变更
			notifyMsg = append(notifyMsg, message.Text("\n-----[图标变更]-----\n"))
			// 旧图标
			notifyMsg = append(notifyMsg, message.Text("[旧]\n"))
			if oldSubStatus.FaviconRaw != "" {
				notifyMsg = append(notifyMsg, message.Image(oldSubStatus.FaviconRaw.toBase64String()))
			} else {
				notifyMsg = append(notifyMsg, message.Text("(空)\n"))
			}
			// 新图标
			notifyMsg = append(notifyMsg, message.Text("[新]\n"))
			if newSubStatus.FaviconRaw != "" {
				notifyMsg = append(notifyMsg, message.Image(newSubStatus.FaviconRaw.toBase64String()))
			} else {
				notifyMsg = append(notifyMsg, message.Text("(空)\n"))
			}
		}
		notifyMsg = append(notifyMsg, message.Text("\n-------最新状态-------\n"))
		// 服务状态
		textMsg, iconBase64 := newSubStatus.generateServerStatusMsg()
		if iconBase64 != "" {
			notifyMsg = append(notifyMsg, message.Image(iconBase64))
		}
		notifyMsg = append(notifyMsg, message.Text(textMsg))
	}
	// 逻辑到达这里，说明状态已经变更 or 无变更且服务器可达，重置不可达计数器
	resetPingServerUnreachableCounter(oldSubStatus.ServerAddr)
	return
}
