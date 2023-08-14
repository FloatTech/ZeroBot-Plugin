// Package coc coc插件
package coc

import (
	fcext "github.com/FloatTech/floatbox/ctxext"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	// DefaultJSONFile COC面板
	DefaultJSONFile = "COC面板.json"
	// SettingJSONFile 设置
	SettingJSONFile = "/setting.json"
)

var (
	settingGoup = make(map[int64]settingInfo, 256)
	getsetting  = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		gid := ctx.Event.GroupID
		_, ok := settingGoup[gid]
		if ok {
			return true
		}
		settingInfo, err := loadSetting(gid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
		settingGoup[gid] = settingInfo
		return true
	})
)

// 默认的json模版
type cocJSON struct {
	ID        int64       `json:"ID"`   // QQ
	BaseInfo  []baseInfo  `json:"基本信息"` // 基本信息
	Attribute []attribute `json:"属性详情"` // 属性
	Other     []string    `json:"其他描述"`
}

type settingInfo struct {
	DefaultDice int   `json:"默认骰子面数"`
	CocPC       int64 `json:"COC PC"`
	DiceRule    int   `json:"骰子规则[0-2]"`
}

type baseInfo struct {
	Name  string `json:"信息"`
	Value any    `json:"内容"`
}

type attribute struct {
	Name     string `json:"属性"`
	MaxValue int    `json:"最大值"`
	MinValue int    `json:"最小值"`
	Value    int    `json:"当前值"`
}
