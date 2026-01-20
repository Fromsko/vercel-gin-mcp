package mcp

// ToolHandler 工具处理函数
type ToolHandler func(ctx *Context) *ToolResult

// Tool 工具构建器 - 链式调用风格
type Tool struct {
	name        string
	description string
	properties  map[string]Property
	required    []string
	handler     ToolHandler
}

// NewTool 创建新工具
func NewTool(name string) *Tool {
	return &Tool{
		name:       name,
		properties: make(map[string]Property),
		required:   []string{},
	}
}

// Desc 设置工具描述
func (t *Tool) Desc(desc string) *Tool {
	t.description = desc
	return t
}

// String 添加字符串参数
func (t *Tool) String(name, desc string, required bool) *Tool {
	t.properties[name] = Property{Type: "string", Description: desc}
	if required {
		t.required = append(t.required, name)
	}
	return t
}

// Number 添加数字参数
func (t *Tool) Number(name, desc string, required bool) *Tool {
	t.properties[name] = Property{Type: "number", Description: desc}
	if required {
		t.required = append(t.required, name)
	}
	return t
}

// Bool 添加布尔参数
func (t *Tool) Bool(name, desc string, required bool) *Tool {
	t.properties[name] = Property{Type: "boolean", Description: desc}
	if required {
		t.required = append(t.required, name)
	}
	return t
}

// Handle 设置处理函数
func (t *Tool) Handle(h ToolHandler) *Tool {
	t.handler = h
	return t
}

// toSchema 转换为 ToolSchema
func (t *Tool) toSchema() ToolSchema {
	return ToolSchema{
		Name:        t.name,
		Description: t.description,
		InputSchema: InputSchema{
			Type:       "object",
			Properties: t.properties,
			Required:   t.required,
		},
	}
}
