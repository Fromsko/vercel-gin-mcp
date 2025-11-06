# MCP Server

一个基于 Go 的 MCP (Model Context Protocol) 服务器实现，使用 Gin 框架提供 HTTP 服务。

## 项目结构

```
mcp-server/
├── cmd/                    # 应用程序入口点
│   └── main.go            # 主程序文件
├── config/                # 配置管理
│   └── config.go          # 配置加载和管理
├── handlers/              # 请求处理器
│   └── hello.go           # Hello World 工具处理器
├── server/                # 服务器相关代码
│   ├── server.go          # 服务器管理
│   ├── http/              # HTTP 服务器
│   │   └── server.go      # Gin 路由和中间件
│   └── mcp/               # MCP 相关
│       └── tools.go       # MCP 工具管理
├── go.mod                 # Go 模块定义
├── go.sum                 # Go 模块依赖锁定
└── README.md              # 项目说明
```

## 快速开始

### 配置

项目使用 godotenv 管理环境变量。你可以通过以下两种方式配置服务器：

#### 方式一：使用 .env 文件（推荐）

1. 复制 `.env.example` 文件为 `.env`：
   ```bash
   cp .env.example .env
   ```

2. 根据需要修改 `.env` 文件中的配置值

#### 方式二：使用环境变量

可以通过以下环境变量配置服务器：

- `SERVER_NAME`: 服务器名称（默认: "Demo 🚀"）
- `SERVER_VERSION`: 服务器版本（默认: "1.0.0"）
- `PORT`: 服务器端口（默认: "8080"）
- `HOST`: 服务器主机（默认: "0.0.0.0"）
- `GIN_MODE`: Gin 运行模式（默认: "release"）

> 注意：如果同时设置了 `.env` 文件和环境变量，环境变量的值将优先使用。

### 运行服务器

```bash
# 进入项目目录
cd mcp-server

# 运行服务器
go run cmd/main.go
```

服务器将在 `http://localhost:8080` 启动。

### API 端点

- `GET /`: 服务器欢迎信息
- `GET /health`: 健康检查
- `GET /sse`: SSE 连接端点
- `POST /message`: 消息处理端点

## 开发指南

### 添加新的工具

1. 在 `handlers/` 目录下创建新的处理器文件
2. 在 `server/mcp/tools.go` 中添加工具定义
3. 在 `cmd/main.go` 中注册新的工具

### 修改服务器配置

在 `config/config.go` 中添加新的配置项，并使用环境变量进行配置。

## 依赖

- [Gin](https://github.com/gin-gonic/gin): HTTP Web 框架
- [mcp-go](https://github.com/mark3labs/mcp-go): MCP Go 实现
- [godotenv](https://github.com/joho/godotenv): 从 .env 文件加载环境变量