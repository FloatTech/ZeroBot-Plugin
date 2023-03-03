package steam

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

// ----------------------- 远程调用 ----------------------
const (
	URL       = "https://api.steampowered.com/"                          // steam API 调用地址
	StatusURL = "ISteamUser/GetPlayerSummaries/v2/?key=%+v&steamids=%+v" // 根据用户steamID获取用户状态
)

var (
	apiKey       string
	steamKeyFile = engine.DataFolder() + "apikey.txt"
)

func init() {
	go func() {
		if file.IsNotExist(steamKeyFile) {
			_, err := os.Create(steamKeyFile)
			if err != nil {
				panic(err)
			}
		}
		apikey, err := os.ReadFile(steamKeyFile)
		if err != nil {
			panic(err)
		}
		apiKey = binary.BytesToString(apikey)
	}()
	engine.OnRegex(`初始化steam链接密钥\s*(.*)$`, zero.OnlyPrivate, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 直接赋值给持久化字短
		apiKey = ctx.State["regex_matched"].([]string)[1]
		// 持久化到本地文件
		if err := os.WriteFile(steamKeyFile, binary.StringToBytes(apiKey), 0777); err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: 持久化密钥失败！"))
			return
		}
		ctx.SendChain(message.Text("设置链接密钥成功！"))
	})
	engine.OnFullMatch("查看当前steam链接密钥", zero.OnlyPrivate, zero.SuperUserPermission, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		ctx.SendChain(message.Text("链接密钥为: ", apiKey))
	})
	engine.OnFullMatch("拉取steam绑定用户状态", getDB).SetBlock(true).Handle(listenUserChange)
}

// listenUserChange 用于监听用户的信息变化
func listenUserChange(ctx *zero.Ctx) {
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
	streamIds := make([]string, len(infos))
	localPlayerMap := make(map[string]player)
	for i, info := range infos {
		streamIds[i] = info.SteamID
		localPlayerMap[info.SteamID] = info
	}
	// 将所有用户状态查一遍
	playerStatus, err := getPlayerStatus(streamIds)
	if err != nil {
		// 出错就发消息
		ctx.SendPrivateMessage(su, message.Text("[steam] ERROR: ", err))
		return
	}
	// 遍历返回的信息做对比，假如信息有变化则发消息
	now := time.Now()
	for _, playerInfo := range playerStatus {
		var msg message.Message
		localInfo := localPlayerMap[playerInfo.SteamID]
		// 排除不需要处理的情况
		if localInfo.GameID == "" && playerInfo.GameID == "" {
			continue
		}
		// 打开游戏
		if localInfo.GameID == "" && playerInfo.GameID != "" {
			msg = append(msg, message.Text(playerInfo.PersonaName, "正在玩", playerInfo.GameExtraInfo))
			localInfo.LastUpdate = now.Unix()
		}
		// 更换游戏
		if localInfo.GameID != "" && playerInfo.GameID != localInfo.GameID && playerInfo.GameID != "" {
			msg = append(msg, message.Text(playerInfo.PersonaName, "玩了", (now.Unix()-localInfo.LastUpdate)/60, "分钟后, 丢下了", localInfo.GameExtraInfo, ", 转头去玩", playerInfo.GameExtraInfo))
			localInfo.LastUpdate = now.Unix()
		}
		// 关闭游戏
		if playerInfo.GameID != localInfo.GameID && playerInfo.GameID == "" {
			msg = append(msg, message.Text(playerInfo.PersonaName, "玩了", (now.Unix()-localInfo.LastUpdate)/60, "分钟后, 关掉了", localInfo.GameExtraInfo))
			localInfo.LastUpdate = 0
		}
		if msg != nil {
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
}

// getPlayerStatus 获取用户状态
func getPlayerStatus(streamIds []string) ([]*player, error) {
	players := make([]*player, 0)
	// 拼接请求地址
	url := fmt.Sprintf(URL+StatusURL, apiKey, strings.Join(streamIds, ","))
	logrus.Debugln("[steamstatus] getPlayerStatus url:", url)
	// 拉取并解析数据
	data, err := web.GetData(url)
	if err != nil {
		return players, err
	}
	dataStr := binary.BytesToString(data)
	logrus.Debugln("[steamstatus] getPlayerStatus data:", dataStr)
	index := gjson.Get(dataStr, "response.players.#").Uint()
	for i := uint64(0); i < index; i++ {
		players = append(players, &player{
			SteamID:       gjson.Get(dataStr, fmt.Sprintf("response.players.%d.steamid", i)).String(),
			PersonaName:   gjson.Get(dataStr, fmt.Sprintf("response.players.%d.personaname", i)).String(),
			GameID:        gjson.Get(dataStr, fmt.Sprintf("response.players.%d.gameid", i)).String(),
			GameExtraInfo: gjson.Get(dataStr, fmt.Sprintf("response.players.%d.gameextrainfo", i)).String(),
		})
	}
	logrus.Debugln("[steamstatus] getPlayerStatus players:", players)
	return players, nil
}
