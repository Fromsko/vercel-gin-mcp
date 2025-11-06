package server

import (
	"mcp-server/config"
	"mcp-server/server/http"

	"github.com/mark3labs/mcp-go/server"
)

// MCPServer MCP 服务器包装器
type MCPServer struct {
	*server.MCPServer
}

// NewMCPServer 创建新的 MCP 服务器
func NewMCPServer(cfg *config.Config) *MCPServer {
	s := server.NewMCPServer(
		cfg.ServerName,
		cfg.ServerVersion,
		server.WithToolCapabilities(false),
	)
	return &MCPServer{MCPServer: s}
}

// Start 启动服务器
func Start(mcpServer *MCPServer, cfg *config.Config) error {
	// 创建 Gin HTTP 服务器
	httpServer := http.NewServer()

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(mcpServer.MCPServer, server.WithHTTPServer(httpServer))

	// 初始化路由
	http.InitializeRoutes(sseServer)

	// 启动服务器
	return sseServer.Start(cfg.Host + ":" + cfg.Port)
}