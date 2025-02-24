// Package minecraftobserver 通过mc服务器地址获取服务器状态信息并绘制图片发送到QQ群
package minecraftobserver

import (
	"errors"
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strings"
	"time"
)

const (
	name = "minecraftobserver"
)

var (
	// 注册插件
	engine = control.Register(name, &ctrl.Options[*zero.Ctx]{
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
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("minecraft observer已启动..."))
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("minecraft observer已关闭..."))
		},
	}).ApplySingle(zbpCtxExt.DefaultSingle)
)

func init() {
	// 状态查询
	engine.OnRegex("^[m|M][c|C]服务器状态 (.+)$").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器状态 ", "")
		resp, err := getMinecraftServerStatus(addr)
		if err != nil || resp == nil {
			logrus.Errorln(logPrefix+"getMinecraftServerStatus error: ", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		status := resp.GenServerSubscribeSchema(addr, 0)
		msg := status.GenerateServerStatusMsg()
		if id := ctx.SendChain(msg...); id.ID() == 0 {
			ctx.SendChain(message.Text("发送失败..."))
			return
		}
	})
	// 添加订阅
	engine.OnRegex(`^[m|M][c|C]服务器添加订阅\s*(.+)$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器添加订阅 ", "")
		ss, err := getMinecraftServerStatus(addr)
		if err != nil || ss == nil {
			logrus.Errorln(logPrefix+"getMinecraftServerStatus error: ", err)
			ctx.SendChain(message.Text("服务器信息初始化失败，请检查服务器是否可用！\n", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		targetID, targetType := warpTargetIDAndType(ctx)
		err = dbInstance.newSubscribe(addr, targetID, targetType)
		if err != nil {
			logrus.Errorln(logPrefix+"newSubscribe error: ", err)
			ctx.SendChain(message.Text("订阅添加失败...", fmt.Sprintf("错误信息: %v", err)))
		}
		// 插入数据库（首条，需要更新状态）
		err = dbInstance.updateServerStatus(ss.GenServerSubscribeSchema(addr, 0))
		if err != nil {
			logrus.Errorln(logPrefix+"updateServerStatus error: ", err)
			ctx.SendChain(message.Text("服务器状态更新失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		ctx.SendChain(message.Text(fmt.Sprintf("服务器 %s 订阅添加成功", addr)))
	})
	// 删除
	engine.OnRegex(`^[m|M][c|C]服务器取消订阅\s*(.+)$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器删除订阅 ", "")
		// 通过群组id和服务器地址获取服务器状态
		targetID, targetType := warpTargetIDAndType(ctx)
		err := dbInstance.deleteSubscribe(addr, targetID, targetType)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.SendChain(message.Text("不存在的订阅！"))
				return
			}
			logrus.Errorln(logPrefix+"deleteSubscribe error: ", err)
			ctx.SendChain(message.Text("订阅删除失败...", fmt.Sprintf("错误信息: %v", err)))
		}
		ctx.SendChain(message.Text("订阅删除成功"))
	})
	// 状态变更通知，全局触发，逐个服务器检查，检查到变更则逐个发送通知
	engine.OnRegex(`^[m|M][c|C]服务器订阅拉取$`, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		serverList, err := dbInstance.getAllSubscribes()
		if err != nil {
			logrus.Errorln(logPrefix+"getAllServer error: ", err)
			return
		}
		logrus.Debugln(logPrefix+"global get ", len(serverList), " subscribe(s)")
		serverMap := make(map[string][]ServerSubscribe)
		for _, v := range serverList {
			serverMap[v.ServerAddr] = append(serverMap[v.ServerAddr], v)
		}
		changedCount := 0
		for subAddr, oneServerSubList := range serverMap {
			// 查询当前存储的状态
			storedStatus, sErr := dbInstance.getServerStatus(subAddr)
			if sErr != nil {
				logrus.Errorln(logPrefix+fmt.Sprintf("getServerStatus ServerAddr(%s) error: ", subAddr), sErr)
				continue
			}
			isChanged, changedNotifyMsg, sErr := singleServerScan(storedStatus)
			if sErr != nil {
				logrus.Errorln(logPrefix+"singleServerScan error: ", sErr)
				continue
			}
			if !isChanged {
				continue
			}
			changedCount++
			logrus.Infoln(logPrefix+"singleServerScan changed in ", subAddr)
			// 发送变化信息
			for _, subInfo := range oneServerSubList {
				logrus.Debugln(logPrefix+" now try to send subInfo to ", subInfo.TargetID, subInfo.TargetType)
				time.Sleep(100 * time.Millisecond)
				if subInfo.TargetType == targetTypeUser {
					if sid := ctx.SendPrivateMessage(subInfo.TargetID, changedNotifyMsg); sid == 0 {
						logrus.Warnln(logPrefix + fmt.Sprintf("SendPrivateMessage to (%d,%d) failed", subInfo.TargetID, subInfo.TargetType))
					}
				} else if subInfo.TargetType == targetTypeGroup {
					m, ok := control.Lookup(name)
					if !ok {
						logrus.Warnln(logPrefix + "control.Lookup empty")
						continue
					}
					if !m.IsEnabledIn(subInfo.TargetID) {
						continue
					}
					if sid := ctx.SendGroupMessage(subInfo.TargetID, changedNotifyMsg); sid == 0 {
						logrus.Warnln(logPrefix + fmt.Sprintf("SendGroupMessage to (%d,%d) failed", subInfo.TargetID, subInfo.TargetType))
					}
				}
			}
		}
		logrus.Debugln(logPrefix + fmt.Sprintf("global scan finished, %d server(s) changed", changedCount))
	})
}

// singleServerScan 单个服务器状态扫描
func singleServerScan(oldSubStatus *ServerStatus) (changed bool, notifyMsg message.Message, err error) {
	notifyMsg = make(message.Message, 0)
	newSubStatus := &ServerStatus{}
	// 获取服务器状态 & 检查是否需要更新
	rawServerStatus, err := getMinecraftServerStatus(oldSubStatus.ServerAddr)
	if err != nil {
		logrus.Warnln(logPrefix+"getMinecraftServerStatus error: ", err)
		err = nil
		// 计数器没有超限，增加计数器并跳过
		if cnt, ts := addPingServerUnreachableCounter(oldSubStatus.ServerAddr, time.Now()); cnt < pingServerUnreachableCounterThreshold &&
			time.Now().Sub(ts) < pingServerUnreachableCounterTimeThreshold {
			logrus.Warnln(logPrefix+"server ", oldSubStatus.ServerAddr, " unreachable, counter: ", cnt, " ts:", ts)
			return
		}
		// 不可达计数器已经超限，则更新服务器状态
		// 深拷贝，设置PingDelay为不可达
		newSubStatus = oldSubStatus.DeepCopy()
		newSubStatus.PingDelay = PingDelayUnreachable
	} else {
		newSubStatus = rawServerStatus.GenServerSubscribeSchema(oldSubStatus.ServerAddr, oldSubStatus.ID)
	}
	if newSubStatus == nil {
		logrus.Errorln(logPrefix + "newSubStatus is nil")
		return
	}
	// 检查是否有订阅信息变化
	if oldSubStatus.IsServerStatusSpecChanged(newSubStatus) {
		logrus.Warnf(logPrefix+"server subscribe spec changed: (%+v) -> (%+v)", oldSubStatus, newSubStatus)
		changed = true
		// 更新数据库
		err = dbInstance.updateServerStatus(newSubStatus)
		if err != nil {
			logrus.Errorln(logPrefix+"updateServerSubscribeStatus error: ", err)
			return
		}
		// 服务状态
		newStatusMsg := newSubStatus.GenerateServerStatusMsg()
		// 变化信息 + 服务状态信息
		notifyMsg = append(notifyMsg, formatSubStatusChange(oldSubStatus, newSubStatus)...)
		notifyMsg = append(notifyMsg, message.Text("\n当前状态:\n"))
		notifyMsg = append(notifyMsg, newStatusMsg...)
	}
	// 逻辑到达这里，说明状态已经变更 or 无变更且服务器可达，重置不可达计数器
	resetPingServerUnreachableCounter(oldSubStatus.ServerAddr)
	return
}
