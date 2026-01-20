package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Server MCP 服务器
type Server struct {
	name    string
	version string
	tools   map[string]*Tool
}

// New 创建新的 MCP 服务器
func New(name string) *Server {
	return &Server{
		name:    name,
		version: "1.0.0",
		tools:   make(map[string]*Tool),
	}
}

// Version 设置版本
func (s *Server) Version(v string) *Server {
	s.version = v
	return s
}

// Register 注册工具
func (s *Server) Register(tool *Tool) *Server {
	s.tools[tool.name] = tool
	return s
}

// Handler 返回 Gin 处理函数
func (s *Server) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, Response{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &Error{Code: -32700, Message: "Parse error"},
			})
			return
		}

		resp := s.handleRequest(&req)
		if resp == nil {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// handleRequest 处理 JSON-RPC 请求
func (s *Server) handleRequest(req *Request) *Response {
	var result any
	var rpcErr *Error

	switch req.Method {
	case "initialize":
		result = InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: Capabilities{
				Tools: &ToolsCapability{ListChanged: false},
			},
			ServerInfo: ServerInfo{
				Name:    s.name,
				Version: s.version,
			},
		}

	case "notifications/initialized":
		return nil

	case "tools/list":
		tools := make([]ToolSchema, 0, len(s.tools))
		for _, t := range s.tools {
			tools = append(tools, t.toSchema())
		}
		result = ToolsListResult{Tools: tools}

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			rpcErr = &Error{Code: -32602, Message: "Invalid params"}
		} else {
			result, rpcErr = s.callTool(&params)
		}

	default:
		rpcErr = &Error{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)}
	}

	resp := &Response{
		JSONRPC: "2.0",
		ID:      req.ID,
	}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}
	return resp
}

// callTool 调用工具
func (s *Server) callTool(params *CallToolParams) (*ToolResult, *Error) {
	tool, ok := s.tools[params.Name]
	if !ok {
		return nil, &Error{Code: -32602, Message: fmt.Sprintf("Unknown tool: %s", params.Name)}
	}

	if tool.handler == nil {
		return nil, &Error{Code: -32603, Message: fmt.Sprintf("Tool has no handler: %s", params.Name)}
	}

	ctx := &Context{
		Name:      params.Name,
		Arguments: params.Arguments,
		server:    s,
	}

	return tool.handler(ctx), nil
}
