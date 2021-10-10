package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	manager "github.com/FloatTech/bot-manager"
	// 依赖gin监听server
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	// 前端静态文件
	ctrl "github.com/FloatTech/ZeroBot-Plugin/control"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	// 向前端推送消息的ws链接
	conn *websocket.Conn
	// 向前端推送日志的ws链接
	logConn *websocket.Conn

	l logWriter
)

// logWriter
// @Description:
//
type logWriter struct {
}

// initGui 初始化gui
func initGui() {
	// 将日志重定向到前端hook
	writer := io.MultiWriter(l, os.Stderr)
	log.SetOutput(writer)
	// 监听后端
	go controller()
	// 注册消息handle
	messageHandle()
}

// websocket的协议升级
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func controller() {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorln("[gui]" + "bot-manager出现不可恢复的错误")
			log.Errorln("[gui]", err)
		}
	}()

	engine := gin.New()
	// 支持跨域
	engine.Use(cors())
	// 注册静态文件
	engine.StaticFS("/dist", http.FS(manager.Dist))
	engine.POST("/get_bots", getBots)
	engine.POST("/get_group_list", getGroupList)
	engine.POST("/get_friend_list", getFriendList)
	// 注册主路径路由，使其跳转到主页面
	engine.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/dist/dist/default.html")
	})
	// 更改某个插件状态
	engine.POST("/update_plugin_status", updatePluginStatus)
	// 更改某一个插件在所有群的状态
	engine.POST("/update_plugin_all_group_status", updatePluginAllGroupStatus)
	// 更改所有插件状态
	engine.POST("/update_all_plugin_status", updateAllPluginStatus)
	// 获取所有插件状态
	engine.POST("/get_plugins_status", getPluginsStatus)
	// 获取一个插件状态
	engine.POST("/get_plugin_status", getPluginStatus)
	// 获取插件列表
	engine.POST("/get_plugins", func(context *gin.Context) {
		var datas []map[string]interface{}
		ctrl.ForEach(func(key string, manager *ctrl.Control) bool {
			datas = append(datas, map[string]interface{}{"id": 1, "handle_type": "", "name": key, "enable": manager.IsEnabledIn(0)})
			return true
		})
		context.JSON(200, datas)
	})
	// 链接日志
	engine.GET("/get_log", getLogs)
	// 获取前端标签
	engine.GET("/get_label", func(context *gin.Context) {
		context.JSON(200, "ZeroBot-Plugin")
	})

	// 发送信息
	engine.POST("/send_msg", sendMsg)
	engine.GET("/data", upgrade)
	log.Infoln("[gui] the webui is running http://127.0.0.1:3000")
	log.Infoln("[gui] ", "you input the `ZeroBot-Plugin.exe -g` can disable the gui")
	if err := engine.Run("127.0.0.1:3000"); err != nil {
		log.Debugln("[gui] ", err.Error())
	}
}

// updateAllPluginStatus
/**
 * @Description: 改变所有插件的状态
 * @param context
 * example
 */
func updateAllPluginStatus(context *gin.Context) {
	enable, err := strconv.ParseBool(context.PostForm("enable"))
	if err != nil {
		var parse map[string]interface{}
		err := context.BindJSON(&parse)
		if err != nil {
			log.Errorln("[gui] " + err.Error())
			return
		}
		enable = parse["enable"].(bool)
	}
	var groups []int64
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		for _, group := range ctx.GetGroupList().Array() {
			groups = append(groups, group.Get("group_id").Int())
		}
		return true
	})

	ctrl.ForEach(func(key string, manager *ctrl.Control) bool {
		if enable {
			for _, group := range groups {
				manager.Enable(group)
			}
		} else {
			for _, group := range groups {
				manager.Disable(group)
			}
		}
		return true
	})
	context.JSON(200, nil)
}

// updatePluginAllGroupStatus
/**
 * @Description: 改变插件在所有群的状态
 * @param context
 * example
 */
func updatePluginAllGroupStatus(context *gin.Context) {
	name := context.PostForm("name")
	enable, err := strconv.ParseBool(context.PostForm("enable"))
	if err != nil {
		var parse map[string]interface{}
		err := context.BindJSON(&parse)
		if err != nil {
			log.Errorln("[gui]" + err.Error())
			return
		}
		name = parse["name"].(string)
		enable = parse["enable"].(bool)
	}
	control, b := ctrl.Lookup(name)
	if !b {
		context.JSON(404, nil)
		return
	}
	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		for _, group := range ctx.GetGroupList().Array() {
			if enable {
				control.Enable(group.Get("group_id").Int())
			} else {
				control.Disable(group.Get("group_id").Int())
			}
		}

		return true
	})

	context.JSON(200, nil)
}

// updatePluginStatus
/**
 * @Description: 更改某一个插件状态
 * @param context
 * example
 */
func updatePluginStatus(context *gin.Context) {
	var parse map[string]interface{}
	err := context.BindJSON(&parse)
	if err != nil {
		log.Errorln("[gui] ", err)
		return
	}
	groupID := int64(parse["group_id"].(float64))
	name := parse["name"].(string)
	enable := parse["enable"].(bool)
	fmt.Println(name)
	control, b := ctrl.Lookup(name)
	if !b {
		context.JSON(404, "服务不存在")
		return
	}
	if enable {
		control.Enable(groupID)
	} else {
		control.Disable(groupID)
	}
	context.JSON(200, nil)
}

// getPluginStatus
/**
 * @Description: 获取一个插件的状态
 * @param context
 * example
 */
func getPluginStatus(context *gin.Context) {
	groupID, err := strconv.ParseInt(context.PostForm("group_id"), 10, 64)
	name := context.PostForm("name")
	if err != nil {
		var parse map[string]interface{}
		err := context.BindJSON(&parse)
		if err != nil {
			log.Errorln("[gui]" + err.Error())
			return
		}
		groupID = int64(parse["group_id"].(float64))
		name = parse["name"].(string)
	}
	control, b := ctrl.Lookup(name)
	if !b {
		context.JSON(404, "服务不存在")
		return
	}
	context.JSON(200, gin.H{"enable": control.IsEnabledIn(groupID)})
}

// getPluginsStatus
/**
 * @Description: 获取所有插件的状态
 * @param context
 * example
 */
func getPluginsStatus(context *gin.Context) {
	groupID, err := strconv.ParseInt(context.PostForm("group_id"), 10, 64)
	if err != nil {
		var parse map[string]interface{}
		err := context.BindJSON(&parse)
		if err != nil {
			log.Errorln("[gui]" + err.Error())
			return
		}
		groupID = int64(parse["group_id"].(float64))
	}
	var datas []map[string]interface{}
	ctrl.ForEach(func(key string, manager *ctrl.Control) bool {
		enable := manager.IsEnabledIn(groupID)
		datas = append(datas, map[string]interface{}{"name": key, "enable": enable})
		return true
	})
	context.JSON(200, datas)
}

// getLogs
/**
 * @Description: 连接日志
 * @param context
 * example
 */
func getLogs(context *gin.Context) {
	con1, err := upGrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}
	logConn = con1
}

// getFriendList
/**
 * @Description: 获取好友列表
 * @param context
 * example
 */
func getFriendList(context *gin.Context) {
	selfID, err := strconv.Atoi(context.PostForm("self_id"))
	if err != nil {
		log.Errorln("[gui]" + err.Error())
		var data map[string]interface{}
		err := context.BindJSON(&data)
		if err != nil {
			log.Errorln("[gui]" + err.Error())
			log.Errorln("[gui]" + "绑定错误")
			return
		}
		selfID = int(data["self_id"].(float64))
	}
	bot := zero.GetBot(int64(selfID))
	var resp []interface{}
	list := bot.GetFriendList().String()
	err = json.Unmarshal([]byte(list), &resp)
	if err != nil {
		log.Errorln("[gui]" + err.Error())
		log.Errorln("[gui]" + "解析json错误")
	}
	context.JSON(200, resp)
}

// getGroupList
/**
 * @Description: 获取群列表
 * @param context
 * example
 */
func getGroupList(context *gin.Context) {
	selfID, err := strconv.Atoi(context.PostForm("self_id"))
	if err != nil {
		var data map[string]interface{}
		err := context.BindJSON(&data)
		if err != nil {
			log.Errorln("[gui]" + err.Error())
			return
		}
		selfID = int(data["self_id"].(float64))
	}

	bot := zero.GetBot(int64(selfID))
	var resp []interface{}
	list := bot.GetGroupList().String()
	err = json.Unmarshal([]byte(list), &resp)
	if err != nil {
		log.Errorln("[gui]" + err.Error())
	}
	context.JSON(200, resp)
}

// getBots
/**
 * @Description: 获取机器人qq号
 * @param context
 * example
 */
func getBots(context *gin.Context) {
	var bots []int64

	zero.RangeBot(func(id int64, ctx *zero.Ctx) bool {
		bots = append(bots, id)
		return true
	})
	context.JSON(200, bots)
}

// MessageHandle
/**
 * @Description: 定义一个向前端发送信息的handle
 * example
 */
func messageHandle() {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorln("[gui]" + "bot-manager出现不可恢复的错误")
			log.Errorln("[gui] ", err)
		}
	}()

	matcher := zero.OnMessage().SetBlock(false).SetPriority(1)

	matcher.Handle(func(ctx *zero.Ctx) {
		if conn != nil {
			err := conn.WriteJSON(ctx.Event)
			if err != nil {
				log.Debugln("[gui] " + "向发送错误")
				return
			}
		}
	})
}

// upgrade
/**
 * @Description: 连接ws，向前端推送message
 * @param context
 * example
 */
func upgrade(context *gin.Context) {
	con, err := upGrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}
	conn = con
}

// sendMsg
/**
 * @Description: 前端调用发送信息
 * @param context
 * example
 */
func sendMsg(context *gin.Context) {
	var data map[string]interface{}
	err := context.BindJSON(&data)
	if err != nil {
		context.JSON(404, nil)
		return
	}
	selfID := int64(data["self_id"].(float64))
	id := int64(data["id"].(float64))
	message1 := data["message"].(string)
	messageType := data["message_type"].(string)

	bot := zero.GetBot(selfID)
	var msgID int64
	if messageType == "group" {
		msgID = bot.SendGroupMessage(id, message.ParseMessageFromString(message1))
	} else {
		msgID = bot.SendPrivateMessage(id, message.ParseMessageFromString(message1))
	}
	context.JSON(200, msgID)
}

// cors
/**
 * @Description: 支持跨域访问
 * @return gin.HandlerFunc
 * example
 */
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") // 请求头部
		if origin != "" {
			// 接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			// 服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			// 允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session, Content-Type")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			// 设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			// 允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		// 允许类型校验
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "ok!")
		}

		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic info is: %v", err)
			}
		}()

		c.Next()
	}
}

func (l logWriter) Write(p []byte) (n int, err error) {
	if logConn != nil {
		err := logConn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			return len(p), nil
		}
	}
	return len(p), nil
}
