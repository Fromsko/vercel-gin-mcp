package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// HelloWorldHandler 处理 hello_world 工具请求
func HelloWorldHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Hello, %s!", name)), nil
}