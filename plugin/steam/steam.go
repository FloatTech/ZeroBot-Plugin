// Package steam 获取steam用户状态
package steam

import (
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/math"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Extra:            control.ExtraFromString("steam"),
		Brief:            "steam相关插件",
		Help: "- steam添加订阅 xxxxxxx (可输入需要绑定的 steamid)\n" +
			"- steam删除订阅 xxxxxxx (删除你创建的对于 steamid 的绑定)\n" +
			"- steam查询订阅 (查询本群内所有的绑定对象)\n" +
			"-----------------------\n" +
			"- steam绑定 api key xxxxxxx (密钥在steam网站申请, 申请地址: https://steamcommunity.com/dev/apikey)\n" +
			"- 查看apikey (查询已经绑定的密钥)\n" +
			"- 拉取steam订阅 (使用插件定时任务开始)\n" +
			"-----------------------\n" +
			"Tips: steamID在用户资料页的链接上面, 形如7656119820673xxxx\n" +
			"需要先私聊绑定apikey, 订阅用户之后使用job插件设置定时, 例: \n" +
			"记录在\"@every 1m\"触发的指令\n" +
			"拉取steam订阅",
		PrivateDataFolder: "steam",
	}).ApplySingle(ctxext.DefaultSingle)
)

func init() {
	// 创建绑定流程
	engine.OnRegex(`^steam添加订阅\s*(\d+)$`, zero.OnlyGroup, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		steamidstr := ctx.State["regex_matched"].([]string)[1]
		steamID := math.Str2Int64(steamidstr)
		// 获取用户状态
		playerStatus, err := getPlayerStatus(steamidstr)
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 添加失败, 获取用户信息错误"))
			return
		}
		if len(playerStatus) == 0 {
			ctx.SendChain(message.Text("[steam] ERROR: 需要添加的用户不存在, 请检查id或url"))
			return
		}
		playerData := playerStatus[0]
		// 判断用户是否已经初始化：若未初始化，通过用户的steamID获取当前状态并初始化；若已经初始化则更新用户信息
		info, err := database.find(steamID)
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 添加失败，数据库错误"))
			return
		}
		// 处理数据
		groupID := strconv.FormatInt(ctx.Event.GroupID, 10)
		if info.Target == "" {
			info = player{
				SteamID:       steamID,
				PersonaName:   playerData.PersonaName,
				Target:        groupID,
				GameID:        playerData.GameID,
				GameExtraInfo: playerData.GameExtraInfo,
				LastUpdate:    time.Now().Unix(),
			}
		} else if !strings.Contains(info.Target, groupID) {
			info.Target = strings.Join([]string{info.Target, groupID}, ",")
		}
		// 更新数据库
		if err = database.update(&info); err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 更新数据库失败"))
			return
		}
		ctx.SendChain(message.Text("添加成功"))
	})
	// 删除绑定流程
	engine.OnRegex(`^steam删除订阅\s*(\d+)$`, zero.OnlyGroup, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		steamID := math.Str2Int64(ctx.State["regex_matched"].([]string)[1])
		groupID := strconv.FormatInt(ctx.Event.GroupID, 10)
		// 判断是否已经绑定该steamID，若已绑定就将群列表从推送群列表钟去除
		info, err := database.findWithGroupID(steamID, groupID)
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 删除失败，数据库错误"))
			return
		}
		if info.SteamID == 0 {
			ctx.SendChain(message.Text("[steam] ERROR: 所需要删除的用户不存在。"))
			return
		}
		// 从绑定列表中剔除需要删除的对象
		targets := strings.Split(info.Target, ",")
		newTargets := make([]string, 0)
		for _, target := range targets {
			if target == groupID {
				continue
			}
			newTargets = append(newTargets, target)
		}
		if len(newTargets) == 0 {
			if err = database.del(steamID); err != nil {
				ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 删除失败，数据库错误"))
				return
			}
		} else {
			info.Target = strings.Join(newTargets, ",")
			if err = database.update(&info); err != nil {
				ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 删除失败，数据库错误"))
				return
			}
		}
		ctx.SendChain(message.Text("删除成功"))
	})
	// 查询当前群绑定信息
	engine.OnFullMatch("steam查询订阅", zero.OnlyGroup, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 获取群信息
		groupID := strconv.FormatInt(ctx.Event.GroupID, 10)
		// 获取所有绑定信息
		infos, err := database.findAll()
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err, "\nEXP: 查询订阅失败, 数据库错误"))
			return
		}
		if len(infos) == 0 {
			ctx.SendChain(message.Text("[steam] ERROR: 还未订阅过用户关系！"))
			return
		}
		// 遍历所有信息，如果包含该群就收集对应的steamID
		var sb strings.Builder
		head := " 查询steam订阅成功, 该群订阅的用户有: \n"
		sb.WriteString(head)
		for _, info := range infos {
			if strings.Contains(info.Target, groupID) {
				sb.WriteString(" ")
				sb.WriteString(info.PersonaName)
				sb.WriteString(":")
				sb.WriteString(strconv.FormatInt(info.SteamID, 10))
				sb.WriteString("\n")
			}
		}
		if sb.String() == head {
			ctx.SendChain(message.Text("查询成功，该群暂时还没有被绑定的用户！"))
			return
		}
		// 组装并返回结果
		data, err := text.RenderToBase64(sb.String(), text.FontFile, 400, 18)
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err))
			return
		}
		ctx.SendChain(message.Image("base64://" + binary.BytesToString(data)))
	})
}
