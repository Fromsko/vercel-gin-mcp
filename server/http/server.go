package http

import (
	"net/http"
	"mcp-server/config"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
)

var engine *gin.Engine

// NewServer 创建新的 Gin HTTP 服务器
func NewServer() *http.Server {
	// 设置 Gin 模式
	gin.SetMode(config.Load().GinMode)
	
	// 创建 Gin 实例
	engine = gin.Default()
	
	return &http.Server{
		Handler: engine,
	}
}

// InitializeRoutes 初始化路由
func InitializeRoutes(sseServer *server.SSEServer) {
	// 注册路由
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, MCP Server!",
		})
	})

	// 注册 405 路由
	engine.NoMethod(func(c *gin.Context) {
		c.JSON(405, gin.H{
			"message": "Method Not Allowed",
		})
	})

	// 注册健康检查路由
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Healthy",
		})
	})

	// 将 MCP 的路由交给 SSE 服务器处理
	engine.GET("/sse", func(c *gin.Context) {
		sseServer.SSEHandler().ServeHTTP(c.Writer, c.Request)
	})
	engine.POST("/message", func(c *gin.Context) {
		sseServer.MessageHandler().ServeHTTP(c.Writer, c.Request)
	})
}