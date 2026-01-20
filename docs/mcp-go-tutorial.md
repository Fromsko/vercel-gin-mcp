# MCP-Go v0.42.0 全方位使用教程

## 概述

MCP-Go 是 Model Context Protocol (MCP) 的 Go 语言实现，用于在 LLM 应用程序和外部数据源/工具之间建立标准化连接。

## 安装

```bash
go get github.com/mark3labs/mcp-go@v0.42.0
```

## 快速开始

### 基础服务器示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    // 创建 MCP 服务器
    s := server.NewMCPServer(
        "Calculator Demo",
        "1.0.0",
        server.WithToolCapabilities(false),
        server.WithRecovery(),
    )

    // 添加工具
    calculatorTool := mcp.NewTool("calculate",
        mcp.WithDescription("执行基本算术运算"),
        mcp.WithString("operation",
            mcp.Required(),
            mcp.Description("运算操作 (add, subtract, multiply, divide)"),
            mcp.Enum("add", "subtract", "multiply", "divide"),
        ),
        mcp.WithNumber("x",
            mcp.Required(),
            mcp.Description("第一个数字"),
        ),
        mcp.WithNumber("y",
            mcp.Required(),
            mcp.Description("第二个数字"),
        ),
    )

    // 添加工具处理器
    s.AddTool(calculatorTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        op, err := request.RequireString("operation")
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }

        x, err := request.RequireFloat("x")
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }

        y, err := request.RequireFloat("y")
        if err != nil {
            return mcp.NewToolResultError(err.Error()), nil
        }

        var result float64
        switch op {
        case "add":
            result = x + y
        case "subtract":
            result = x - y
        case "multiply":
            result = x * y
        case "divide":
            if y == 0 {
                return mcp.NewToolResultError("不能除以零"), nil
            }
            result = x / y
        }

        return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
    })

    // 启动服务器
    if err := server.ServeStdio(s); err != nil {
        fmt.Printf("服务器错误: %v\n", err)
    }
}
```

## 核心概念

### 1. 服务器 (Server)

MCP 服务器是核心组件，提供工具、资源和提示。

```go
s := server.NewMCPServer(
    "服务器名称",
    "版本号",
    server.WithToolCapabilities(true),  // 启用工具功能
    server.WithResourceCapabilities(true),  // 启用资源功能
    server.WithRecovery(),  // 启用恢复机制
)
```

### 2. 工具 (Tools)

工具提供可执行的功能。

```go
tool := mcp.NewTool("tool_name",
    mcp.WithDescription("工具描述"),
    mcp.WithString("param1",
        mcp.Required(),
        mcp.Description("参数描述"),
    ),
    mcp.WithNumber("param2",
        mcp.Description("数字参数"),
    ),
)

s.AddTool(tool, handlerFunc)
```

### 3. 资源 (Resources)

资源提供数据访问。

```go
resource := mcp.NewResource(
    "resource://data",
    "数据资源",
    "application/json",
)

s.AddResource(resource, handlerFunc)
```

### 4. 提示 (Prompts)

提示定义交互模板。

```go
prompt := mcp.NewPrompt("template_name",
    mcp.WithDescription("提示描述"),
    mcp.WithString("variable",
        mcp.Required(),
        mcp.Description("变量描述"),
    ),
)

s.AddPrompt(prompt, handlerFunc)
```

## 传输层 (Transports)

MCP-Go 支持三种传输层：

### 1. STDIO (标准输入输出)

```go
// 用于命令行工具
server.ServeStdio(s)
```

### 2. SSE (Server-Sent Events)

```go
// SSE 服务器示例
sseServer := server.NewMCPServer(
    "SSE Server",
    "1.0.0",
    server.WithToolCapabilities(true),
)

// 设置连接丢失处理器
sseServer.SetConnectionLostHandler(func(sessionID string) {
    log.Printf("客户端连接丢失: %s", sessionID)
    // 实现重连逻辑
})

// 启动 SSE 服务器
http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
    server.ServeSSE(w, r, sseServer)
})

http.ListenAndServe(":8080", nil)
```

### 3. HTTP (流式 HTTP)

```go
// HTTP 服务器示例
http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
    server.ServeHTTP(w, r, s)
})
```

## SSE 详细使用

### 基本 SSE 服务器

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

func main() {
    s := server.NewMCPServer(
        "SSE Demo",
        "1.0.0",
        server.WithToolCapabilities(true),
    )

    // 添加工具
    tool := mcp.NewTool("echo",
        mcp.WithDescription("回显输入文本"),
        mcp.WithString("text",
            mcp.Required(),
            mcp.Description("要回显的文本"),
        ),
    )

    s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        text, _ := request.RequireString("text")
        return mcp.NewToolResultText("回显: " + text), nil
    })

    // 设置连接丢失处理器
    s.SetConnectionLostHandler(func(sessionID string) {
        log.Printf("连接丢失: %s", sessionID)
    })

    // 启动 HTTP 服务器
    http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
        server.ServeSSE(w, r, s)
    })

    log.Println("SSE 服务器启动在 :8080/mcp")
    http.ListenAndServe(":8080", nil)
}
```

### SSE 客户端连接

客户端可以通过以下方式连接：

```javascript
// JavaScript 客户端示例
const eventSource = new EventSource('/mcp');

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('收到消息:', data);
};

eventSource.onerror = function(event) {
    console.error('连接错误:', event);
    // 实现重连逻辑
    setTimeout(() => {
        eventSource = new EventSource('/mcp');
    }, 1000);
};
```

## 会话管理

### 基础会话处理

```go
// 实现自定义会话
type MySession struct {
    id           string
    notifChannel chan mcp.JSONRPCNotification
    isInitialized bool
    // 添加自定义字段
    userData     map[string]interface{}
}

// 实现 ClientSession 接口
func (s *MySession) SessionID() string {
    return s.id
}

func (s *MySession) NotificationChannel() chan<- mcp.JSONRPCNotification {
    return s.notifChannel
}

func (s *MySession) Initialize() {
    s.isInitialized = true
}

func (s *MySession) Initialized() bool {
    return s.isInitialized
}

// 注册会话
session := &MySession{
    id:           "user-123",
    notifChannel: make(chan mcp.JSONRPCNotification, 10),
    userData:     make(map[string]interface{}),
}

if err := s.RegisterSession(context.Background(), session); err != nil {
    log.Printf("注册会话失败: %v", err)
}
```

### 会话特定工具

```go
// 实现 SessionWithTools 接口
type MyAdvancedSession struct {
    MySession
    sessionTools map[string]server.ServerTool
}

func (s *MyAdvancedSession) GetSessionTools() map[string]server.ServerTool {
    return s.sessionTools
}

func (s *MyAdvancedSession) SetSessionTools(tools map[string]server.ServerTool) {
    s.sessionTools = tools
}

// 添加会话特定工具
userTool := mcp.NewTool("user_data",
    mcp.WithDescription("访问用户特定数据"),
)

err := s.AddSessionTool(
    session.SessionID(),
    userTool,
    func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        return mcp.NewToolResultText("用户数据: " + session.SessionID()), nil
    },
)
```

### 发送通知

```go
// 向特定客户端发送通知
err := s.SendNotificationToSpecificClient(
    session.SessionID(),
    "notification/update",
    map[string]interface{}{
        "message": "新数据可用!",
        "data":    session.userData,
    },
)
```

## 请求钩子 (Request Hooks)

```go
// 添加请求前钩子
s.AddBeforeHook(func(ctx context.Context, request interface{}) error {
    log.Printf("收到请求: %+v", request)
    return nil
})

// 添加请求后钩子
s.AddAfterHook(func(ctx context.Context, request interface{}, response interface{}) error {
    log.Printf("发送响应: %+v", response)
    return nil
})
```

## 工具处理器中间件

```go
// 创建中间件
middleware := func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // 前处理
        log.Printf("执行工具: %s", request.Params.Name)

        // 执行下一个处理器
        result, err := next(ctx, request)

        // 后处理
        if err != nil {
            log.Printf("工具执行错误: %v", err)
        }

        return result, err
    }
}

// 应用中间件
s.WithToolMiddleware(middleware)
```

## 完整示例：带 SSE 和会话管理

```go
package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

type SessionManager struct {
    sessions map[string]*UserSession
    mu       sync.RWMutex
}

type UserSession struct {
    id           string
    notifChannel chan mcp.JSONRPCNotification
    isInitialized bool
    data         map[string]interface{}
}

func (s *UserSession) SessionID() string { return s.id }
func (s *UserSession) NotificationChannel() chan<- mcp.JSONRPCNotification { return s.notifChannel }
func (s *UserSession) Initialize() { s.isInitialized = true }
func (s *UserSession) Initialized() bool { return s.isInitialized }

func main() {
    sessionMgr := &SessionManager{
        sessions: make(map[string]*UserSession),
    }

    s := server.NewMCPServer(
        "Advanced MCP Server",
        "1.0.0",
        server.WithToolCapabilities(true),
    )

    // 添加会话管理工具
    sessionTool := mcp.NewTool("manage_session",
        mcp.WithDescription("管理用户会话"),
        mcp.WithString("action",
            mcp.Required(),
            mcp.Enum("create", "get", "delete"),
        ),
        mcp.WithString("session_id",
            mcp.Description("会话 ID"),
        ),
    )

    s.AddTool(sessionTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        action, _ := request.RequireString("action")

        switch action {
        case "create":
            sessionID := generateSessionID()
            session := &UserSession{
                id:           sessionID,
                notifChannel: make(chan mcp.JSONRPCNotification, 10),
                data:         make(map[string]interface{}),
            }

            sessionMgr.mu.Lock()
            sessionMgr.sessions[sessionID] = session
            sessionMgr.mu.Unlock()

            if err := s.RegisterSession(ctx, session); err != nil {
                return mcp.NewToolResultError(err.Error()), nil
            }

            return mcp.NewToolResultJSON(map[string]string{
                "session_id": sessionID,
            }), nil

        case "get":
            sessionID, _ := request.RequireString("session_id")
            sessionMgr.mu.RLock()
            session, exists := sessionMgr.sessions[sessionID]
            sessionMgr.mu.RUnlock()

            if !exists {
                return mcp.NewToolResultError("会话不存在"), nil
            }

            data, _ := json.Marshal(session.data)
            return mcp.NewToolResultText(string(data)), nil

        case "delete":
            sessionID, _ := request.RequireString("session_id")
            sessionMgr.mu.Lock()
            delete(sessionMgr.sessions, sessionID)
            sessionMgr.mu.Unlock()

            s.UnregisterSession(ctx, sessionID)
            return mcp.NewToolResultText("会话已删除"), nil
        }

        return mcp.NewToolResultError("未知操作"), nil
    })

    // 设置连接丢失处理
    s.SetConnectionLostHandler(func(sessionID string) {
        log.Printf("客户端断开连接: %s", sessionID)
        sessionMgr.mu.Lock()
        delete(sessionMgr.sessions, sessionID)
        sessionMgr.mu.Unlock()
    })

    // 启动服务器
    http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
        server.ServeSSE(w, r, s)
    })

    log.Println("服务器启动在 :8080/mcp")
    http.ListenAndServe(":8080", nil)
}

func generateSessionID() string {
    // 生成唯一会话 ID
    return "session-" + randomString(8)
}

func randomString(n int) string {
    // 实现随机字符串生成
    const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}
```

## 最佳实践

1. **错误处理**: 始终返回适当的错误信息
2. **资源管理**: 使用 context 管理请求生命周期
3. **并发安全**: 在处理共享数据时使用适当的同步机制
4. **日志记录**: 添加详细的日志以便调试
5. **连接管理**: 实现适当的重连和超时处理
6. **性能优化**: 使用缓存和批处理减少延迟

## 注意事项

- MCP-Go 仍在积极开发中
- 某些高级功能可能尚未完全实现
- SSE 传输需要处理 HTTP/2 超时断开
- 会话管理需要适当的清理机制
