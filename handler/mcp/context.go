package mcp

import "encoding/json"

// H 类似 gin.H 的灵活 map 类型
type H map[string]any

// Context 工具执行上下文，用于传递状态
type Context struct {
	Name      string
	Arguments map[string]any
	server    *Server
}

// String 获取字符串参数
func (c *Context) String(key string) string {
	if v, ok := c.Arguments[key].(string); ok {
		return v
	}
	return ""
}

// Int 获取整数参数
func (c *Context) Int(key string) int {
	if v, ok := c.Arguments[key].(float64); ok {
		return int(v)
	}
	return 0
}

// Float 获取浮点数参数
func (c *Context) Float(key string) float64 {
	if v, ok := c.Arguments[key].(float64); ok {
		return v
	}
	return 0
}

// Bool 获取布尔参数
func (c *Context) Bool(key string) bool {
	if v, ok := c.Arguments[key].(bool); ok {
		return v
	}
	return false
}

// Has 检查参数是否存在
func (c *Context) Has(key string) bool {
	_, ok := c.Arguments[key]
	return ok
}

// Text 返回文本结果
func (c *Context) Text(text string) *ToolResult {
	return &ToolResult{
		Content: []Content{{Type: "text", Text: text}},
	}
}

// JSON 返回 JSON 格式结果
func (c *Context) JSON(data any) *ToolResult {
	b, err := json.Marshal(data)
	if err != nil {
		return c.Error("JSON marshal error: " + err.Error())
	}
	return &ToolResult{
		Content: []Content{{Type: "text", Text: string(b)}},
	}
}

// Markdown 返回 Markdown 格式结果
func (c *Context) Markdown(md string) *ToolResult {
	return &ToolResult{
		Content: []Content{{Type: "text", Text: md}},
	}
}

// Error 返回错误结果
func (c *Context) Error(msg string) *ToolResult {
	return &ToolResult{
		Content: []Content{{Type: "text", Text: msg}},
		IsError: true,
	}
}
