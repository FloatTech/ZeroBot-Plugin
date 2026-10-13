// Package middleware 中间件
package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

var (
	// LoginCache 登录缓存
	LoginCache = cache.New(24*time.Hour, 12*time.Hour)
)

// Cors 跨域
/**
 * @Description: 支持跨域访问
 * @return gin.HandlerFunc
 * example
 */
func Cors() gin.HandlerFunc {
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
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"result":  nil,
				"message": "",
				"type":    "success",
			})
		}

		c.Next()
	}
}

// TokenMiddle 验证token
func TokenMiddle() gin.HandlerFunc {
	return func(con *gin.Context) {
		// 进行token验证
		token := con.Request.Header.Get("Authorization")
		if token == "" {
			con.JSON(http.StatusUnauthorized, gin.H{
				"code":    2,
				"result":  nil,
				"message": "无权访问, 请登录",
				"type":    "error",
			})
			con.Abort()
			return
		}
		_, found := LoginCache.Get(token)
		if !found {
			con.JSON(http.StatusUnauthorized, gin.H{
				"code":    2,
				"result":  nil,
				"message": "toke无效, 请重新登录",
				"type":    "error",
			})
			con.Abort()
			return
		}
		con.Next()
	}
}
