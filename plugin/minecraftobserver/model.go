package minecraftobserver

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
)

// ====================
// DB Schema

// serverStatus 服务器状态
type serverStatus struct {
	// ID 主键
	ID int64 `json:"id" gorm:"column:id;primary_key:pk_id;auto_increment;default:0"`
	// 服务器地址
	ServerAddr string `json:"server_addr" gorm:"column:server_addr;default:'';unique_index:udx_server_addr"`
	// 服务器描述
	Description string `json:"description" gorm:"column:description;default:null;type:CLOB"`
	// 在线玩家
	Players string `json:"players" gorm:"column:players;default:''"`
	// 版本
	Version string `json:"version" gorm:"column:version;default:''"`
	// FaviconMD5 Favicon MD5
	FaviconMD5 string `json:"favicon_md5" gorm:"column:favicon_md5;default:''"`
	// FaviconRaw 原始数据
	FaviconRaw icon `json:"favicon_raw" gorm:"column:favicon_raw;default:null;type:CLOB"`
	// 延迟，不可达时为-1
	PingDelay int64 `json:"ping_delay" gorm:"column:ping_delay;default:-1"`
	// 更新时间
	LastUpdate int64 `json:"last_update" gorm:"column:last_update;default:0"`
}

// serverSubscribe 订阅信息
type serverSubscribe struct {
	// ID 主键
	ID int64 `json:"id" gorm:"column:id;primary_key:pk_id;auto_increment;default:0"`
	// 服务器地址
	ServerAddr string `json:"server_addr" gorm:"column:server_addr;default:'';unique_index:udx_ait"`
	// 推送目标id
	TargetID int64 `json:"target_id" gorm:"column:target_id;default:0;unique_index:udx_ait"`
	// 类型 1：群组 2：个人
	TargetType int64 `json:"target_type" gorm:"column:target_type;default:0;unique_index:udx_ait"`
	// 更新时间
	LastUpdate int64 `json:"last_update" gorm:"column:last_update;default:0"`
}

const (
	// pingDelayUnreachable 不可达
	pingDelayUnreachable = -1
)

// isServerStatusSpecChanged 检查是否有状态变化
func (ss *serverStatus) isServerStatusSpecChanged(newStatus *serverStatus) (res bool) {
	res = false
	if ss == nil || newStatus == nil {
		res = false
		return
	}
	// 描述变化、版本变化、Favicon变化
	if ss.Description != newStatus.Description || ss.Version != newStatus.Version || ss.FaviconMD5 != newStatus.FaviconMD5 {
		res = true
		return
	}
	// 状态由不可达变为可达 or 反之
	if (ss.PingDelay == pingDelayUnreachable && newStatus.PingDelay != pingDelayUnreachable) ||
		(ss.PingDelay != pingDelayUnreachable && newStatus.PingDelay == pingDelayUnreachable) {
		res = true
		return
	}
	return
}

// deepCopy 深拷贝
func (ss *serverStatus) deepCopy() (dst *serverStatus) {
	if ss == nil {
		return
	}
	dst = &serverStatus{}
	*dst = *ss
	return
}

// generateServerStatusMsg 生成服务器状态消息
func (ss *serverStatus) generateServerStatusMsg() (msg string, iconBase64 string) {
	var msgBuilder strings.Builder
	if ss == nil {
		return
	}
	msgBuilder.WriteString(ss.Description)
	msgBuilder.WriteString("\n")
	msgBuilder.WriteString("服务器地址：")
	msgBuilder.WriteString(ss.ServerAddr)
	msgBuilder.WriteString("\n")
	// 版本
	msgBuilder.WriteString("版本：")
	msgBuilder.WriteString(ss.Version)
	msgBuilder.WriteString("\n")
	// Ping
	if ss.PingDelay < 0 {
		msgBuilder.WriteString("Ping延迟：超时\n")
	} else {
		msgBuilder.WriteString("Ping延迟：")
		msgBuilder.WriteString(fmt.Sprintf("%d 毫秒\n", ss.PingDelay))
		msgBuilder.WriteString("在线人数：")
		msgBuilder.WriteString(ss.Players)
	}
	// 图标
	if ss.FaviconRaw != "" && ss.FaviconRaw.checkPNG() {
		iconBase64 = ss.FaviconRaw.toBase64String()
	}
	msg = msgBuilder.String()
	return
}

// DB Schema End

// ====================
// Ping & List Response DTO

// serverPingAndListResp 服务器状态数据传输对象 From mc server response
type serverPingAndListResp struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	Favicon icon
	Delay   time.Duration
}

// icon should be a PNG image that is Base64 encoded
// (without newlines: \n, new lines no longer work since 1.13)
// and prepended with "data:image/png;base64,".
type icon string

// func (i icon) toImage() (icon image.Image, err error) {
//	const prefix = "data:image/png;base64,"
//	if !strings.HasPrefix(string(i), prefix) {
//		return nil, errors.Errorf("server icon should prepended with %s", prefix)
//	}
//	base64png := strings.TrimPrefix(string(i), prefix)
//	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))
//	icon, err = png.Decode(r)
//	return
//}

// checkPNG 检查是否为PNG
func (i icon) checkPNG() bool {
	const prefix = "data:image/png;base64,"
	return strings.HasPrefix(string(i), prefix)
}

// toBase64String 转换为base64字符串
func (i icon) toBase64String() string {
	return "base64://" + strings.TrimPrefix(string(i), "data:image/png;base64,")
}

// genServerSubscribeSchema 将DTO转换为DB Schema
func (dto *serverPingAndListResp) genServerSubscribeSchema(addr string, id int64) *serverStatus {
	if dto == nil {
		return nil
	}
	faviconMD5 := md5.Sum(helper.StringToBytes(string(dto.Favicon)))
	return &serverStatus{
		ID:          id,
		ServerAddr:  addr,
		Description: dto.Description.ClearString(),
		Version:     dto.Version.Name,
		Players:     fmt.Sprintf("%d/%d", dto.Players.Online, dto.Players.Max),
		FaviconMD5:  hex.EncodeToString(faviconMD5[:]),
		FaviconRaw:  dto.Favicon,
		PingDelay:   dto.Delay.Milliseconds(),
		LastUpdate:  time.Now().Unix(),
	}
}

// Ping & List Response DTO End
// ====================

// ====================
// Biz Model
const (
	logPrefix = "[minecraft observer] "
)

// warpTargetIDAndType 转换消息信息到订阅的目标ID和类型
func warpTargetIDAndType(groupID, userID int64) (int64, int64) {
	// 订阅
	var targetID int64
	var targetType int64
	if groupID == 0 {
		targetType = targetTypeUser
		targetID = userID
	} else {
		targetType = targetTypeGroup
		targetID = groupID
	}
	return targetID, targetType
}

// formatSubStatusChangeText 格式化状态变更文本
func formatSubStatusChangeText(oldStatus, newStatus *serverStatus) string {
	var msgBuilder strings.Builder
	if oldStatus == nil || newStatus == nil {
		return ""
	}
	// 变更通知
	msgBuilder.WriteString("[Minecraft服务器状态变更通知]\n")
	// 地址
	msgBuilder.WriteString(fmt.Sprintf("服务器地址: %v\n", oldStatus.ServerAddr))
	// 描述
	if oldStatus.Description != newStatus.Description {
		msgBuilder.WriteString("\n-----[描述变更]-----\n")
		msgBuilder.WriteString(fmt.Sprintf("[旧]\n%v\n", oldStatus.Description))
		msgBuilder.WriteString(fmt.Sprintf("[新]\n%v\n", newStatus.Description))
	}
	// 版本
	if oldStatus.Version != newStatus.Version {
		msgBuilder.WriteString("\n-----[版本变更]-----\n")
		msgBuilder.WriteString(fmt.Sprintf("[旧]\n%v\n", oldStatus.Version))
		msgBuilder.WriteString(fmt.Sprintf("[新]\n%v\n", newStatus.Version))
	}
	// 状态由不可达变为可达，反之
	if oldStatus.PingDelay == pingDelayUnreachable && newStatus.PingDelay != pingDelayUnreachable {
		msgBuilder.WriteString("\n-----[Ping延迟]-----\n")
		msgBuilder.WriteString("[旧]\n超时\n")
		msgBuilder.WriteString(fmt.Sprintf("[新]\n%v毫秒\n", newStatus.PingDelay))
	}
	if oldStatus.PingDelay != pingDelayUnreachable && newStatus.PingDelay == pingDelayUnreachable {
		msgBuilder.WriteString("\n-----[Ping延迟]-----\n")
		msgBuilder.WriteString(fmt.Sprintf("[旧]\n%v毫秒\n", oldStatus.PingDelay))
		msgBuilder.WriteString("[新]\n超时\n")
	}
	return msgBuilder.String()
}

// Biz Model End
// ====================
