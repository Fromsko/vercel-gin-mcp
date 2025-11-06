package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tools 定义所有可用的工具
var Tools = []mcp.Tool{
	mcp.NewTool("hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	),
}

// RegisterTools 注册所有工具到 MCP 服务器
func RegisterTools(mcpServer *server.MCPServer, handlers map[string]server.ToolHandlerFunc) {
	for _, tool := range Tools {
		if handler, ok := handlers[tool.Name]; ok {
			mcpServer.AddTool(tool, handler)
		}
	}
}
