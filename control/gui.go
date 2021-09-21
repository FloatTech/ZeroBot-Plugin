// Package control
// @Description: 该文件提供了对前端的支持
package control

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	// 依赖gin监听server
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	// 前端静态文件
	"github.com/huoxue1/test3"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	engine *zero.Engine
	// 向前端推送消息的ws链接
	conn *websocket.Conn
)

func init() {
	// 监听后端
	go Controller()
	// 注册消息handle
	MessageHandle()
	engine = Register("gui", &Options{
		DisableOnDefault: false,
		Help:             "向webui推送信息",
	})
}

// websocket的协议升级
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Controller() {
	if log.GetLevel() != log.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
	}
	engine := gin.New()
	// 支持跨域
	engine.Use(Cors())
	// 注册静态文件
	engine.StaticFS("/dist", http.FS(test3.Dist))
	engine.POST("/get_bots", GetBots)
	engine.POST("/get_group_list", GetGroupList)
	engine.POST("/get_friend_list", GetFriendList)
	// 注册主路径路由，使其跳转到主页面
	engine.GET("/", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/dist/dist/default.html")
	})
	engine.POST("/update_plugin_states", func(context *gin.Context) {

	})
	// 获取插件列表
	engine.POST("/get_plugins", func(context *gin.Context) {
		var datas []map[string]interface{}
		forEach(func(key string, manager *Control) bool {
			datas = append(datas, map[string]interface{}{"ID": 1, "HandleType": "", "Name": key, "Enable": manager.isEnabledIn(0)})
			return true
		})
		context.JSON(200, datas)
	})

	engine.GET("/get_log", func(context *gin.Context) {

	})
	// 获取前端标签
	engine.GET("/get_label", func(context *gin.Context) {
		context.JSON(200, "ZeroBot-Plugin")
	})

	// 发送信息
	engine.POST("/send_msg", CallApi)
	engine.GET("/data", data)
	log.Infoln("the webui is running http://127.0.0.1:3000")
	if err := engine.Run("127.0.0.1:3000"); err != nil {
		log.Debugln(err.Error())
	}

}

// GetFriendList
/**
 * @Description: 获取好友列表
 * @param context
 * example
 */
func GetFriendList(context *gin.Context) {
	selfID, err := strconv.Atoi(context.PostForm("self_id"))
	if err != nil {
		log.Errorln(err.Error())
		var data map[string]interface{}
		err := context.BindJSON(&data)
		if err != nil {
			log.Errorln(err.Error())
			log.Errorln("绑定错误")
			return
		}
		selfID = int(data["self_id"].(float64))
	}
	bot := zero.GetBot(int64(selfID))
	var resp []interface{}
	list := bot.GetFriendList().String()
	err = json.Unmarshal([]byte(list), &resp)
	if err != nil {
		log.Errorln(err.Error())
		log.Errorln("解析json错误")
	}
	context.JSON(200, resp)
}

// GetGroupList
/**
 * @Description: 获取群列表
 * @param context
 * example
 */
func GetGroupList(context *gin.Context) {
	selfID, err := strconv.Atoi(context.PostForm("self_id"))
	if err != nil {

		var data map[string]interface{}
		err := context.BindJSON(&data)
		if err != nil {
			log.Errorln(err.Error())
			return
		}
		selfID = int(data["self_id"].(float64))
	}

	bot := zero.GetBot(int64(selfID))
	var resp []interface{}
	list := bot.GetGroupList().String()
	err = json.Unmarshal([]byte(list), &resp)
	if err != nil {
		log.Errorln(err.Error())
	}
	context.JSON(200, resp)
}

// GetBots
/**
 * @Description: 获取机器人qq号
 * @param context
 * example
 */
func GetBots(context *gin.Context) {
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
func MessageHandle() {
	matcher := engine.OnMessage().SetBlock(false).SetPriority(1)

	matcher.Handle(func(ctx *zero.Ctx) {

		if conn != nil {
			err := conn.WriteJSON(ctx.Event)
			if err != nil {
				log.Debugln("向发送错误")
				return
			}
		}

	})
}

// data
/**
 * @Description: 连接ws，向前端推送message
 * @param context
 * example
 */
func data(context *gin.Context) {
	con, err := upGrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		return
	}
	conn = con
}

// CallApi
/**
 * @Description: 前端调用发送信息
 * @param context
 * example
 */
func CallApi(context *gin.Context) {
	selfID, err := strconv.ParseInt(context.PostForm("self_id"), 10, 64)
	id, err := strconv.ParseInt(context.PostForm("id"), 10, 64)
	message1 := context.PostForm("message")
	messageType := context.PostForm("message_type")
	if err != nil {
		var data map[string]interface{}
		err := context.BindJSON(&data)
		if err != nil {
			context.JSON(404, nil)
			return
		}
		selfID = int64(data["self_id"].(float64))
		id = int64(data["id"].(float64))
		message1 = data["message"].(string)
		messageType = data["message_type"].(string)

	}
	bot := zero.GetBot(selfID)
	var msgID int64
	if messageType == "group" {
		msgID = bot.SendGroupMessage(id, message.ParseMessageFromString(message1))
	} else {
		msgID = bot.SendPrivateMessage(id, message.ParseMessageFromString(message1))
	}
	context.JSON(200, msgID)
}

// Cors
/**
 * @Description: 支持跨域访问
 * @return gin.HandlerFunc
 * example
 */
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin") //请求头部
		if origin != "" {
			//接收客户端发送的origin （重要！）
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//允许跨域设置可以返回其他子段，可以自定义字段
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session, Content-Type")
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
			//设置缓存时间
			c.Header("Access-Control-Max-Age", "172800")
			//允许客户端传递校验信息比如 cookie (重要)
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		//允许类型校验
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
