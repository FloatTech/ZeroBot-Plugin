// Package webctrl 包含 webui 所需的所有内容
package webctrl

import (
	"context"
	"io/fs"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	webui "github.com/FloatTech/ZeroBot-Plugin-Webui"
	"github.com/wdvxdr1123/ZeroBot/message"

	"github.com/FloatTech/zbputils/control/web/controller"
	"github.com/FloatTech/zbputils/control/web/model"
	"github.com/FloatTech/zbputils/control/web/router"
)

var (
	// ListenCtrlChan 启动/停止 webui
	listenCtrlChan = make(chan bool)
)

func init() {
	zero.OnRegex(`^/设置webui用户名\s?(\S+)\s?密码\s?(\S+)$`, zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			err := model.CreateOrUpdateUser(&model.User{Username: regexMatched[1], Password: regexMatched[2]})
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("设置成功"))
			if zero.BotConfig.SuperUsers != nil && len(zero.BotConfig.SuperUsers) > 0 {
				ctx.SendPrivateMessage(zero.BotConfig.SuperUsers[0], message.Text("webui账号\n用户名: ", regexMatched[1], "\n密码: ", regexMatched[2]))
			}
		})
	zero.OnCommand("webui", zero.SuperUserPermission, zero.OnlyToMe).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			args = strings.TrimSpace(args)
			isSuccess := args == "启动"
			listenCtrlChan <- isSuccess
			if isSuccess {
				ctx.SendChain(message.Text("成功, webui启动"))
			} else {
				ctx.SendChain(message.Text("成功, webui停止"))
			}
		})
}

// RunGui 运行webui
//
//	@title			zbp api
//	@version		1.0
//	@description	zbp restful api document
//	@host			127.0.0.1:3000
//	@BasePath		/
func RunGui(addr string) {
	defer func() {
		err := recover()
		if err != nil {
			log.Errorln("[gui] ZeroBot-Plugin-Webui出现不可恢复的错误")
			log.Errorln("[gui] err:", err, ", stack:", debug.Stack())
		}
	}()

	engine := gin.Default()
	router.SetRouters(engine)

	staticEngine := gin.Default()
	df, _ := fs.Sub(webui.Dist, "dist")
	staticEngine.StaticFS("/", http.FS(df))
	server := &http.Server{
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			// 如果 URL 以 /api, /swagger 开头, 走后端路由
			if strings.HasPrefix(request.URL.Path, "/api") || strings.HasPrefix(request.URL.Path, "/swagger") {
				engine.ServeHTTP(writer, request)
				return
			}
			// 否则，走前端路由
			staticEngine.ServeHTTP(writer, request)
		}),
		Addr: addr,
	}
	log.Infoln("[gui] the webui is running on", "http://"+addr)
	log.Infoln("[gui] you can see api by http://" + addr + "/swagger/index.html")
	if err := server.ListenAndServe(); err != nil {
		log.Errorln("[gui] server listen err: ", err.Error())
	}
	for canrun := range listenCtrlChan {
		if canrun {
			if err := server.Shutdown(context.TODO()); err != nil {
				log.Errorln("[gui] server shutdown err: ", err.Error())
			}
			server = &http.Server{
				Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					// 如果 URL 以 /api, /swagger 开头, 走后端路由
					if strings.HasPrefix(request.URL.Path, "/api") || strings.HasPrefix(request.URL.Path, "/swagger") {
						engine.ServeHTTP(writer, request)
						return
					}
					// 否则，走前端路由
					staticEngine.ServeHTTP(writer, request)
				}),
				Addr: addr,
			}
			go func() {
				log.Infoln("[gui] the webui is running on", "http://"+addr)
				log.Infoln("[gui] you can see api by http://" + addr + "/swagger/index.html")
				if err := server.ListenAndServe(); err != nil {
					log.Errorln("[gui] server listen err: ", err.Error())
				}
			}()
		} else {
			if err := server.Shutdown(context.TODO()); err != nil {
				log.Errorln("[gui] server shutdown err: ", err.Error())
			}
			controller.MsgConn = nil
			controller.LogConn = nil
		}
	}
}
