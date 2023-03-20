package steam

import (
	"strconv"
	"sync"
	"time"

	fcext "github.com/FloatTech/floatbox/ctxext"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	database streamDB
	// 开启并检查数据库链接
	getDB = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		database.db.DBPath = engine.DataFolder() + "steam.db"
		err := database.db.Open(time.Hour * 24)
		if err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err))
			return false
		}
		if err = database.db.Create(TableListenPlayer, &player{}); err != nil {
			ctx.SendChain(message.Text("[steam] ERROR: ", err))
			return false
		}
		// 校验密钥是否初始化
		m := ctx.State["manager"].(*ctrl.Control[*zero.Ctx])
		_ = m.Manager.Response(steamapikeygid)
		_ = m.Manager.GetExtra(steamapikeygid, &apiKey)
		if apiKey == "" {
			ctx.SendChain(message.Text("ERROR: 未设置steam apikey"))
			return false
		}
		return true
	})
)

// streamDB 继承方法的存储结构
type streamDB struct {
	sync.RWMutex
	db sql.Sqlite
}

const (
	// TableListenPlayer 存储查询用户信息
	TableListenPlayer = "listen_player"
)

// player 用户状态存储结构体
type player struct {
	SteamID       int64  `json:"steam_id"`        // 绑定用户标识ID
	PersonaName   string `json:"persona_name"`    // 用户昵称
	Target        string `json:"target"`          // 信息推送群组
	GameID        int64  `json:"game_id"`         // 游戏ID
	GameExtraInfo string `json:"game_extra_info"` // 游戏信息
	LastUpdate    int64  `json:"last_update"`     // 更新时间
}

// update 如果主键不存在则插入一条新的数据，如果主键存在直接复写
func (sql *streamDB) update(dbInfo *player) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Insert(TableListenPlayer, dbInfo)
}

// find 根据主键查信息
func (sql *streamDB) find(steamID int64) (dbInfo player, err error) {
	sql.Lock()
	defer sql.Unlock()
	condition := "where steam_id = " + strconv.FormatInt(steamID, 10)
	if !sql.db.CanFind(TableListenPlayer, condition) {
		return player{}, nil // 规避没有该用户数据的报错
	}
	err = sql.db.Find(TableListenPlayer, &dbInfo, condition)
	return
}

// findWithGroupID 根据用户steamID和groupID查询信息
func (sql *streamDB) findWithGroupID(steamID int64, groupID string) (dbInfo player, err error) {
	sql.Lock()
	defer sql.Unlock()
	condition := "where steam_id = " + strconv.FormatInt(steamID, 10) + " AND target LIKE '%" + groupID + "%'"
	if !sql.db.CanFind(TableListenPlayer, condition) {
		return player{}, nil // 规避没有该用户数据的报错
	}
	err = sql.db.Find(TableListenPlayer, &dbInfo, condition)
	return
}

// findAll 查询所有库信息
func (sql *streamDB) findAll() (dbInfos []player, err error) {
	sql.Lock()
	defer sql.Unlock()
	var info player
	num, err := sql.db.Count(TableListenPlayer)
	if err != nil || num == 0 {
		return
	}
	dbInfos = make([]player, 0, num)
	err = sql.db.FindFor(TableListenPlayer, &info, "", func() error {
		if info.SteamID != 0 {
			dbInfos = append(dbInfos, info)
		}
		return nil
	})
	return
}

// del 删除指定数据
func (sql *streamDB) del(steamID int64) error {
	sql.Lock()
	defer sql.Unlock()
	return sql.db.Del(TableListenPlayer, "where steam_id = "+strconv.FormatInt(steamID, 10))
}
