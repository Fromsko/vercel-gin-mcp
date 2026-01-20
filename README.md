# MCP Server

一个基于 Go 的 MCP (Model Context Protocol) 服务器实现，专为 Vercel 部署优化，提供强大的网页抓取和 GitHub 文档下载功能。

## 特性

- **Vercel 优化**: 专为 Vercel 无服务器函数设计
- **网页抓取**: 集成 gocolly 实现强大的网页内容抓取
- **GitHub 集成**: 支持从 GitHub 仓库批量下载文档
- **MCP 协议**: 完全兼容 Model Context Protocol
- **链式 API**: 优雅的工具注册和配置方式
- **Markdown 输出**: 智能转换网页内容为 Markdown 格式

## 快速开始

### 本地开发

```bash
# 克隆项目
git clone https://github.com/fromsko/vercel-gin-mcp.git
cd vercel-gin-mcp

# 安装依赖
go mod download

# 运行本地服务
go run api/index.go
```

### Vercel 部署

```bash
# 安装 Vercel CLI
npm i -g vercel

# 部署到 Vercel
vercel
```

## 项目结构

```
vercel-gin-mcp/
├── api/                   # Vercel 函数入口
│   └── index.go           # 主入口点和工具注册
├── handler/               # 处理器模块
│   └── mcp/               # MCP 相关实现
│       ├── mcp.go         # MCP 服务器核心
│       └── tools/         # 工具实现
│           ├── fetch.go   # 网页抓取工具
│           └── github.go  # GitHub 工具
├── docs/                  # 文档
│   ├── deployment-guide.md
│   ├── go-webdav-client.md
│   └── mcp-go-tutorial.md
├── go.mod                 # Go 模块定义
├── go.sum                 # Go 模块依赖锁定
├── vercel.json            # Vercel 配置
└── README.md              # 项目说明
```

## 架构设计

### MCP 服务器构建

采用链式调用风格构建 MCP 服务器：

```go
// 创建服务器
server := mcp.New("mcp-server").Version("1.0.0")

// 注册工具
server.Register(
    mcp.NewTool("tool_name").
        Desc("工具描述").
        Param("参数", "参数描述", true).
        Handle(func(ctx *mcp.Context) *mcp.ToolResult {
            // 工具逻辑
        }),
)
```

## API 端点

### Vercel 函数

- `POST https://your-app.vercel.app/mcp` - MCP 协议端点

## 可用工具

- **echo** - 回显输入文本
- **add** - 计算两个数字的和
- **fetch** - 抓取网页内容并转换为 Markdown 格式
- **fetch_md** - 抓取网页内容，仅返回 Markdown 文本
- **download_docs** - 从 GitHub 仓库下载文档文件
- **download_docs_md** - 从 GitHub 仓库下载文档，返回合并的 Markdown

## 工具使用示例

### 网页抓取

```json
{
  "name": "fetch",
  "arguments": {
    "url": "https://example.com"
  }
}
```

### GitHub 文档下载

```json
{
  "name": "download_docs",
  "arguments": {
    "repo": "https://github.com/user/repo",
    "path": "docs"
  }
}
```


## 开发

### 添加新工具

在 `api/index.go` 中注册新工具：

```go
server.Register(
    mcp.NewTool("my_tool").
        Desc("工具描述").
        String("param", "参数描述", true).
        Handle(func(ctx *mcp.Context) *mcp.ToolResult {
            // 实现你的逻辑
            return ctx.Text("结果")
        }),
)
```

### 本地开发

```bash
# 使用 air 实现热重载
go install github.com/air-verse/air@latest
air
```

## 部署

详细的部署指南请参考 [deployment-guide.md](./docs/deployment-guide.md)

## 依赖

- [Gin](https://github.com/gin-gonic/gin) v1.11.0 - HTTP Web 框架
- [go-git](https://github.com/go-git/go-git/v5) v5.16.4 - Git 操作库
- [gocolly](https://github.com/gocolly/colly/v2) v2.3.0 - 网页爬虫框架
- [mergo](https://github.com/dario-cat/mergo) v1.0.0 - 结构体合并工具

## 许可证

[MIT License](./LICENSE)
