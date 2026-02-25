// Package amongus AmongUs战绩查询插件
package amongus

import (
	"net/url"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/floatbox/binary"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/FloatTech/floatbox/web"
	sql "github.com/FloatTech/sqlite"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/tidwall/gjson"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

const (
	profileAPI = "https://api.toue.mxyx.club/api/profile/"
	tableName  = "amongus_user"
)

// amongusUser 用户绑定信息
type amongusUser struct {
	UserID    int64  `json:"user_id"`    // QQ用户ID (主键)
	AmongusID string `json:"amongus_id"` // AmongUs游戏ID
}

var (
	database amongusDB
	// 开启并检查数据库链接
	getDB = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		database.db = sql.New(engine.DataFolder() + "amongus.db")
		err := database.db.Open(time.Hour)
		if err != nil {
			ctx.SendChain(message.Text("[amongus] ERROR: ", err))
			return false
		}
		if err = database.db.Create(tableName, &amongusUser{}); err != nil {
			ctx.SendChain(message.Text("[amongus] ERROR: ", err))
			return false
		}
		return true
	})
	engine = control.AutoRegister(&ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "AmongUs战绩查询",
		Help: "- 录入信息 xxxx (绑定你的AmongUs ID)\n" +
			"- 查询战绩 (查询你的AmongUs战绩)",
		PrivateDataFolder: "amongus",
	})
)

// amongusDB 数据库操作封装
type amongusDB struct {
	sync.RWMutex
	db sql.Sqlite
}

// insert 插入或更新用户绑定信息
func (adb *amongusDB) insert(user *amongusUser) error {
	adb.Lock()
	defer adb.Unlock()
	return adb.db.Insert(tableName, user)
}

// find 根据QQ用户ID查找绑定信息
func (adb *amongusDB) find(userID int64) (user amongusUser, err error) {
	adb.RLock()
	defer adb.RUnlock()
	err = adb.db.Find(tableName, &user, "WHERE user_id = ?", userID)
	return
}

func init() {
	// 录入信息
	engine.OnRegex(`^录入信息\s*(.+)$`, getDB).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			amongusID := strings.TrimSpace(ctx.State["regex_matched"].([]string)[1])
			if amongusID == "" {
				ctx.SendChain(message.Text("请输入有效的AmongUs ID"))
				return
			}
			userID := ctx.Event.UserID
			err := database.insert(&amongusUser{
				UserID:    userID,
				AmongusID: amongusID,
			})
			if err != nil {
				ctx.SendChain(message.Text("[amongus] 录入失败: ", err))
				return
			}
			ctx.SendChain(message.Text("录入成功！你的AmongUs ID: ", amongusID))
		})

	// 查询战绩
	engine.OnFullMatch("查询战绩", getDB).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			userID := ctx.Event.UserID
			// 从数据库查询用户绑定的 AmongUs ID
			user, err := database.find(userID)
			if err != nil || user.AmongusID == "" {
				ctx.SendChain(message.Text("你还没有录入AmongUs ID，请先使用「录入信息 xxxx」绑定"))
				return
			}
			// 请求 API 获取战绩
			encodedID := url.PathEscape(user.AmongusID)

			// 3. 发起请求
			fullURL := profileAPI + encodedID
			data, err := web.GetData(fullURL)
			if err != nil {
				ctx.SendChain(message.Text("[amongus] 请求失败: ", err))
				return
			}
			// 解析 JSON
			result := gjson.ParseBytes(data)
			if !result.Get("success").Bool() {
				ctx.SendChain(message.Text("查询错误"))
				return
			}
			// 获取 totalStats
			totalStats := result.Get("data.totalStats")
			averageKills := totalStats.Get("averageKills").Float()
			averageTasksCompleted := totalStats.Get("averageTasksCompleted").Float()
			completedAllTasksRate := totalStats.Get("completedAllTasksRate").Float()
			totalMatches := totalStats.Get("totalMatches").Int()
			winRate := totalStats.Get("winRate").Float()

			// 构建文本
			var sb strings.Builder
			sb.WriteString("══ Among Us 战绩 ══\n\n")
			sb.WriteString(fmt.Sprintf("  总场次:        %d\n", totalMatches))
			sb.WriteString(fmt.Sprintf("  总胜率:        %.2f%%\n", winRate))
			sb.WriteString(fmt.Sprintf("  平均击杀:      %.2f\n", averageKills))
			sb.WriteString(fmt.Sprintf("  平均任务完成:  %.2f\n", averageTasksCompleted))
			sb.WriteString(fmt.Sprintf("  总任务完成率:  %.2f%%\n", completedAllTasksRate))
			sb.WriteString("\n════════════════════")

			// 渲染为图片并发送
			imgData, err := text.RenderToBase64(sb.String(), text.FontFile, 400, 20)
			if err != nil {
				ctx.SendChain(message.Text("[amongus] 图片渲染失败: ", err))
				return
			}
			ctx.SendChain(message.Image("base64://" + binary.BytesToString(imgData)))
		})
}
