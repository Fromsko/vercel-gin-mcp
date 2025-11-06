package main

import (
	"log"
	"mcp-server/config"
	"mcp-server/handlers"
	"mcp-server/server"
	"mcp-server/server/mcp"

	mcpgo "github.com/mark3labs/mcp-go/server"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 创建 MCP 服务器
	mcpServer := server.NewMCPServer(cfg)

	// 定义工具处理器
	handlers := map[string]mcpgo.ToolHandlerFunc{
		"hello_world": handlers.HelloWorldHandler,
	}

	// 注册工具
	mcp.RegisterTools(mcpServer.MCPServer, handlers)

	// 启动服务器
	if err := server.Start(mcpServer, cfg); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
