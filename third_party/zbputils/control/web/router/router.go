// Package router 路由
package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/FloatTech/zbputils/control/web/controller"
	_ "github.com/FloatTech/zbputils/control/web/docs" // swagger数据
	"github.com/FloatTech/zbputils/control/web/middleware"
)

// SetRouters 创建路由
func SetRouters(engine *gin.Engine) {
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 支持跨域
	engine.Use(middleware.Cors(), gin.Logger())

	// 通用接口
	apiRoute := engine.Group("/api")
	apiRoute.Use(middleware.TokenMiddle())
	apiRoute.GET("/getFriendList", controller.GetFriendList)
	apiRoute.GET("/getGroupList", controller.GetGroupList)
	apiRoute.GET("/getGroupMemberList", controller.GetGroupMemberList)
	apiRoute.GET("/getRequestList", controller.GetRequestList)
	apiRoute.POST("/handleRequest", controller.HandleRequest)
	apiRoute.POST("/sendMsg", controller.SendMsg)
	apiRoute.GET("/getUserInfo", controller.GetUserInfo)

	// 任务相关接口
	jobRoute := apiRoute.Group("/job")
	jobRoute.GET("/list", controller.JobList)
	jobRoute.POST("/add", controller.JobAdd)
	jobRoute.POST("/delete", controller.JobDelete)

	// 管理相关接口
	manageRoute := apiRoute.Group("/manage")
	manageRoute.GET("/getPlugin", controller.GetPlugin)
	manageRoute.GET("/getAllPlugin", controller.GetAllPlugin)
	manageRoute.POST("/updatePluginStatus", controller.UpdatePluginStatus)
	manageRoute.POST("/updateResponseStatus", controller.UpdateResponseStatus)
	manageRoute.POST("/updateAllPluginStatus", controller.UpdateAllPluginStatus)

	noVerifyRoute := engine.Group("/api")
	noVerifyRoute.POST("/login", controller.Login)
	noVerifyRoute.GET("/logout", controller.Logout)
	noVerifyRoute.GET("/getPermCode", controller.GetPermCode)
	noVerifyRoute.GET("/getBotList", controller.GetBotList)
	noVerifyRoute.GET("/getLog", controller.GetLog)
	noVerifyRoute.GET("/data", controller.Upgrade)
}
