# 部署指南：本地服务与 Vercel 函数

本项目采用抽象设计，支持两种部署模式：本地服务和 Vercel 函数。

## 架构设计

### 抽象层
- `server.Engine` 接口定义了服务器引擎的抽象行为
- `BaseEngine` 提供通用功能实现
- `LocalEngine` 实现本地服务功能
- `VercelEngine` 实现 Vercel 函数功能

### 共享组件
- MCP 服务器逻辑完全共享
- Gin 路由和中间件共享
- 工具处理器在两种模式下都可复用

## 本地服务部署

### 运行方式

```bash
# 方式一：直接运行
go run cmd/local/main.go

# 方式二：指定参数
go run cmd/local/main.go -host 0.0.0.0 -port 8080 -debug

# 方式三：使用配置文件
go run cmd/local/main.go -config .env.local
```

### 参数说明

- `-config`: 配置文件路径（默认：.env）
- `-host`: 服务器主机地址（默认：0.0.0.0）
- `-port`: 服务器端口（默认：8080）
- `-debug`: 启用调试模式

### 端点

- `GET http://localhost:8080/` - 欢迎信息
- `GET http://localhost:8080/health` - 健康检查
- `GET http://localhost:8080/sse` - SSE 连接
- `POST http://localhost:8080/message` - 消息处理

### 开发模式

启用调试模式可以看到详细日志：

```bash
go run cmd/local/main.go -debug
```

## Vercel 函数部署

### 目录结构

```
api/
├── index.go    # 主入口点
├── sse.go      # SSE 端点
└── message.go  # 消息处理端点
```

### 部署步骤

1. **安装 Vercel CLI**

```bash
npm i -g vercel
```

2. **登录 Vercel**

```bash
vercel login
```

3. **部署项目**

```bash
# 在项目根目录执行
vercel
```

4. **配置环境变量（可选）**

在 Vercel Dashboard 中配置：
- `SERVER_NAME`: 服务器名称
- `SERVER_VERSION`: 服务器版本
- `GIN_MODE`: 设置为 "release"

### Vercel 端点

部署后，你的端点将是：
- `https://your-app.vercel.app/` - 欢迎信息
- `https://your-app.vercel.app/api/health` - 健康检查
- `https://your-app.vercel.app/api/sse` - SSE 连接
- `https://your-app.vercel.app/api/message` - 消息处理

### vercel.json 配置

```json
{
  "version": 2,
  "builds": [
    {
      "src": "api/*.go",
      "use": "@vercel/go"
    }
  ],
  "routes": [
    {
      "src": "/api/(.*)",
      "dest": "/api/index.go"
    }
  ],
  "env": {
    "GO_BUILD_FLAGS": "-ldflags '-s -w'"
  }
}
```

## 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `SERVER_NAME` | MCP Server | 服务器名称 |
| `SERVER_VERSION` | 1.0.0 | 服务器版本 |
| `HOST` | 0.0.0.0 | 本地服务主机 |
| `PORT` | 8080/3000 | 端口号（本地/Vercel） |
| `GIN_MODE` | release | Gin 运行模式 |
| `PLATFORM` | auto | 运行平台（自动检测） |

### .env 文件示例

```env
# 本地开发配置
SERVER_NAME=My MCP Server
SERVER_VERSION=1.0.0
HOST=0.0.0.0
PORT=8080
GIN_MODE=debug
PLATFORM=local

# Vercel 配置（通过 Dashboard 设置）
SERVER_NAME=My MCP Server
SERVER_VERSION=1.0.0
GIN_MODE=release
PLATFORM=vercel
```

## 工具注册

两种模式共享相同的工具处理器：

```go
toolHandlers := map[string]mcpgo.ToolHandlerFunc{
    "hello_world": handlers.HelloWorldHandler,
    "echo":        handlers.EchoHandler,
    "get_time":    handlers.GetTimeHandler,
    "calculate":   handlers.CalculateHandler,
}
```

## 最佳实践

### 本地开发

1. 使用调试模式查看详细日志
2. 使用热重载工具（如 air）提高开发效率
3. 配置合适的 `.env.local` 文件

```bash
# 安装 air
go install github.com/air-verse/air@latest

# 运行
air
```

### Vercel 部署

1. 优化二进制大小：使用 `-ldflags '-s -w'`
2. 设置合理的超时时间
3. 使用 Vercel 的环境变量管理敏感信息
4. 启用 Vercel Analytics 监控性能

### 代码组织

1. 保持业务逻辑与部署逻辑分离
2. 使用接口抽象提高可测试性
3. 共享处理器和工具逻辑
4. 使用配置文件管理环境差异

## 故障排除

### 本地服务问题

1. **端口被占用**
   ```bash
   lsof -i :8080  # 查看占用端口的进程
   ```

2. **权限问题**
   ```bash
   sudo go run cmd/local/main.go  # 需要管理员权限时
   ```

### Vercel 部署问题

1. **构建失败**
   - 检查 go.mod 文件
   - 确认所有依赖都是兼容的

2. **运行时错误**
   - 查看 Vercel Functions 日志
   - 检查环境变量配置

3. **性能问题**
   - 使用 Vercel Analytics
   - 优化冷启动时间

## 测试

### 本地测试

```bash
# 测试健康检查
curl http://localhost:8080/health

# 测试 SSE 连接
curl -N http://localhost:8080/sse

# 测试消息处理
curl -X POST http://localhost:8080/message \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"hello_world","arguments":{}}}'
```

### Vercel 测试

```bash
# 替换为你的 Vercel URL
curl https://your-app.vercel.app/api/health

# 测试 SSE
curl -N https://your-app.vercel.app/api/sse
```

## 迁移指南

### 从本地迁移到 Vercel

1. 代码无需修改，使用相同的处理器
2. 配置 Vercel 环境变量
3. 更新客户端的端点 URL
4. 测试所有功能

### 从 Vercel 迁移到本地

1. 创建本地配置文件
2. 运行本地服务器
3. 更新客户端配置
4. 验证功能一致性
