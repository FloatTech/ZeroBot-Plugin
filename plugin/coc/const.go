// Package coc coc插件
package coc

import (
	"sync"

	fcext "github.com/FloatTech/floatbox/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
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
	engine = control.Register("coc", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "简易的跑团辅助器",
		Help: "只支持指定的面板格式,需要bot主人后台以群号为文件夹,将COC面板模版复制到文件夹里按对应格式改动后才行。\n" +
			"\n---------------------\n" +
			"coc类指令:\n" +
			"- .loadcoc 以默认的json内容创建coc面板\n" +
			"- .loadcoc[符合本插件的json格式群文件名] 以指定群文件的格式创建coc面板\n" +
			"- .coc 查看信息,如果没有注册就生成随机属性的空白人物面板并绑定\n" +
			"- .coc身份#愚者 生成身份为愚者的面板并绑定\n" +
			"- .coc昵称#张三/身份#愚者 信息填写要与模版相同\n" +
			"\n注:面板主要主要分三个区域“基本信息区”;“属性信息区”和“其他信息区”\n" +
			"coc指令只能注册“基本信息区”和“其他信息区”,\n向“其他信息区”注册时示例为:“.coc昵称#张三/描述#这里是其他信息区”\n" +
			"\n---------------------\n" +
			"r类指令:\n" +
			"- .r 投掷1次默认骰子\n" +
			"- .r5d 投掷5次默认骰子\n" +
			"- .rd12 投掷1次12面骰子\n" +
			"- .r5d12a力量  以力量属性作为权重投掷5次12面骰子\n" +
			"\n---------------------\n" +
			"set类指令\n" +
			"- .set身份#愚者 更改面板基本属性\n" +
			"- .sst [add|del|clr] [段落数] [内容] \n对其他信息进行更改.\n例:.sst clr 2 那之后可以变为1次rd12\n说明:对描述的第2段文字重新编辑为“那之后可以变为1次rd12”\n" +
			"- .sa [骰子表达式:次数d面数a属性] [属性] [数值] [经费] \n花费[经费]ATRI币对[属性]鉴定,成功增加\n例:.sa 1d5a运气 力量 2 100\n说明:花费100ATRI币对力量用运气权重投掷1次5面骰子,成功就+2\n" +
			"- .setpc@玩家 将玩家设为coc管理员\n" +
			"- .setdice[骰子数] 更改默认骰子面数\n" +
			"- .setrule[规则号] 更改默认骰子规则\n" +
			"- .show@玩家 管理员查看指定玩家面板\n" +
			"- .pcset@玩家 身份#愚者/运气#30 管理员更改玩家面板属性\n" +
			"- .kill@玩家 删除角色\n" +
			"\n---------------------\n" +
			"规则列表:\n" +
			"规则0(默认):\n" +
			"大成功:dice=1\n" +
			"大失败:成功率<0.5,dice>95;成功率>=0.5,dice=100;\n" +
			"规则1:\n" +
			"大成功:成功率<0.5,dice=1;成功率>=0.5,dice<6;\n" +
			"大失败:成功率<0.5,dice>95;成功率>=0.5,dice=100;\n" +
			"规则2:\n" +
			"大成功:dice<6\n" +
			"大失败:dice>95",
		PrivateDataFolder: "coc",
	})
	mu          sync.Mutex
	settingGoup = make(map[int64]settingInfo, 256)
	getsetting  = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		gid := ctx.Event.GroupID
		_, ok := settingGoup[gid]
		if ok {
			return true
		}
		_, err := loadSetting(gid)
		if err != nil {
			ctx.SendChain(message.Text("[ERROR]:", err))
			return false
		}
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
