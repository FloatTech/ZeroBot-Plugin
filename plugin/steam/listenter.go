package steam

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// ----------------------- 远程调用 ----------------------
const (
	apiurl    = "https://api.steampowered.com/"                          // steam API 调用地址
	statusurl = "ISteamUser/GetPlayerSummaries/v2/?key=%+v&steamids=%+v" // 根据用户steamID获取用户状态
)

var (
	apiKey   string
	apiKeyMu sync.Mutex
)

func init() {
	engine.OnRegex(`^steam绑定\s*api\s*key\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		apiKeyMu.Lock()
		defer apiKeyMu.Unlock()
		apiKey = ctx.State["regex_matched"].([]string)[1]
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		err := m.SetExtra(apiKey)
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: 保存apikey失败！"))
			return
		}
		ctx.SendChain(message.Text("保存apikey成功！"))
	})
	engine.OnFullMatch("查看apikey", zero.OnlyPrivate, zero.SuperUserPermission, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		apiKeyMu.Lock()
		defer apiKeyMu.Unlock()
		ctx.SendChain(message.Text("apikey为: ", apiKey))
	})
	engine.OnFullMatch("拉取steam订阅", getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		su := zero.BotConfig.SuperUsers[0]
		// 获取所有处于监听状态的用户信息
		infos, err := database.findAll()
		if err != nil {
			// 挂了就给管理员发消息
			ctx.SendPrivateMessage(su, message.Text("[steam] ERROR: ", err))
			return
		}
		if len(infos) == 0 {
			return
		}
		// 收集这波用户的streamId，然后查当前的状态，并建立信息映射表
		streamIDs := make([]string, len(infos))
		localPlayerMap := make(map[int64]*player)
		for i := 0; i < len(infos); i++ {
			streamIDs[i] = strconv.FormatInt(infos[i].SteamID, 10)
			localPlayerMap[infos[i].SteamID] = infos[i]
		}
		// 将所有用户状态查一遍
		playerStatus, err := getPlayerStatus(streamIDs...)
		if err != nil {
			// 出错就发消息
			ctx.SendPrivateMessage(su, message.Text("[steam] ERROR: ", err))
			return
		}
		// 遍历返回的信息做对比，假如信息有变化则发消息
		now := time.Now()
		msg := make(message.Message, 0, len(playerStatus))
		for _, playerInfo := range playerStatus {
			msg = msg[:0]
			localInfo := localPlayerMap[playerInfo.SteamID]
			// 排除不需要处理的情况
			if localInfo.GameID == 0 && playerInfo.GameID == 0 {
				continue
			}
			// 打开游戏
			if localInfo.GameID == 0 && playerInfo.GameID != 0 {
				msg = append(msg, message.Text(playerInfo.PersonaName, "正在玩", playerInfo.GameExtraInfo))
				localInfo.LastUpdate = now.Unix()
			}
			// 更换游戏
			if localInfo.GameID != 0 && playerInfo.GameID != localInfo.GameID && playerInfo.GameID != 0 {
				msg = append(msg, message.Text(playerInfo.PersonaName, "玩了", (now.Unix()-localInfo.LastUpdate)/60, "分钟后, 丢下了", localInfo.GameExtraInfo, ", 转头去玩", playerInfo.GameExtraInfo))
				localInfo.LastUpdate = now.Unix()
			}
			// 关闭游戏
			if playerInfo.GameID != localInfo.GameID && playerInfo.GameID == 0 {
				msg = append(msg, message.Text(playerInfo.PersonaName, "玩了", (now.Unix()-localInfo.LastUpdate)/60, "分钟后, 关掉了", localInfo.GameExtraInfo))
				localInfo.LastUpdate = 0
			}
			if len(msg) != 0 {
				groups := strings.Split(localInfo.Target, ",")
				for _, groupString := range groups {
					group, err := strconv.ParseInt(groupString, 10, 64)
					if err != nil {
						ctx.SendPrivateMessage(su, message.Text("[steam] ERROR: ", err, "\nOTHER: SteamID ", localInfo.SteamID))
						continue
					}
					ctx.SendGroupMessage(group, msg)
				}
			}
			// 更新数据
			localInfo.GameID = playerInfo.GameID
			localInfo.GameExtraInfo = playerInfo.GameExtraInfo
			if err = database.update(localInfo); err != nil {
				ctx.SendPrivateMessage(su, message.Text("[steam] ERROR: ", err, "\nEXP: 更新数据失败\nOTHER: SteamID ", localInfo.SteamID))
			}
		}
	})
}

// getPlayerStatus 获取用户状态
func getPlayerStatus(streamIDs ...string) ([]*player, error) {
	players := make([]*player, 0)
	// 拼接请求地址
	apiKeyMu.Lock()
	url := fmt.Sprintf(apiurl+statusurl, apiKey, strings.Join(streamIDs, ","))
	apiKeyMu.Unlock()
	// 拉取并解析数据
	data, err := web.GetData(url)
	if err != nil {
		return players, err
	}
	dataStr := binary.BytesToString(data)
	index := gjson.Get(dataStr, "response.players.#").Uint()
	for i := uint64(0); i < index; i++ {
		players = append(players, &player{
			SteamID:       gjson.Get(dataStr, fmt.Sprintf("response.players.%d.steamid", i)).Int(),
			PersonaName:   gjson.Get(dataStr, fmt.Sprintf("response.players.%d.personaname", i)).String(),
			GameID:        gjson.Get(dataStr, fmt.Sprintf("response.players.%d.gameid", i)).Int(),
			GameExtraInfo: gjson.Get(dataStr, fmt.Sprintf("response.players.%d.gameextrainfo", i)).String(),
		})
	}
	return players, nil
}
