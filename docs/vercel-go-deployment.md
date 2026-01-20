# Vercel 部署 Go 程序作为 Handler

## 概述

Vercel 支持部署 Go 程序作为 Serverless Functions。Go 运行时目前处于 Beta 阶段，在所有计划上都可用。

## 基本结构

### 目录结构
```
project-root/
├── api/
│   └── index.go
├── go.mod
└── vercel.json (可选)
```

### 基本示例

创建 `api/index.go` 文件：

```go
package handler

import (
    "fmt"
    "net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<h1>Hello from Go!</h1>")
}
```

## Go 版本管理

### 自动检测
- Vercel 会自动检测项目根目录的 `go.mod` 文件来确定 Go 版本
- 如果没有 `go.mod` 或未定义版本，默认使用 Go 1.20
- 首次检测到版本后会自动下载并缓存，后续部署使用缓存版本

### go.mod 示例
```go
module your-project-name

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
)
```

## 依赖管理

- Go 运行时会自动检测项目根目录的 `go.mod` 文件来安装依赖
- 所有在 `go.mod` 中声明的依赖都会被自动安装

## 构建配置

### 环境变量配置
可以通过 `GO_BUILD_FLAGS` 环境变量提供自定义构建标志：

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "build": {
    "env": {
      "GO_BUILD_FLAGS": "-ldflags '-s -w'"
    }
  }
}
```

### 常用构建标志
- `-ldflags '-s -w'`: 移除调试信息，减小输出文件大小（默认值）
- 可以添加其他链接器标志来优化构建

## 高级用法

### 函数签名要求
- 导出函数必须实现 `http.HandlerFunc` 签名
- 可以使用任何有效的 Go 导出函数名

```go
func Handler(w http.ResponseWriter, r *http.Request) {
    // 你的处理逻辑
}

// 或者使用其他函数名
func APIHandler(w http.ResponseWriter, r *http.Request) {
    // 你的处理逻辑
}
```

### 使用私有包

对于私有 Go 包，需要在 `vercel.json` 中配置：

```json
{
  "build": {
    "env": {
      "GOPRIVATE": "github.com/your-private-org",
      "GOPROXY": "direct"
    }
  }
}
```

### 多个 API 端点

可以在 `api` 目录下创建多个 Go 文件：

```
api/
├── index.go      -> /api/index
├── users.go      -> /api/users
└── posts.go      -> /api/posts
```

每个文件都需要有自己的 `Handler` 函数。

## 框架集成

### 使用 Gin 框架示例

```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func Handler(w http.ResponseWriter, r *http.Request) {
    // 创建 Gin 路由器
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello from Gin on Vercel!",
        })
    })

    r.ServeHTTP(w, r)
}
```

## 部署

1. 将代码推送到 Git 仓库
2. 在 Vercel 中导入项目
3. Vercel 会自动检测 Go 项目并使用 Go 运行时
4. 部署完成后，可以通过 `/api/your-file-name` 访问

## 注意事项

- Go 运行时处于 Beta 阶段，可能有变化
- 冷启动时间可能比 Node.js 长
- 适合处理 CPU 密集型任务
- 确保处理程序是无状态的
- 使用环境变量管理敏感信息

## 调试

### 本地测试
可以使用 Vercel CLI 本地测试：

```bash
npm i -g vercel
vercel dev
```

### 日志查看
部署后可以在 Vercel Dashboard 的 Functions 标签中查看日志。

## 性能优化

1. **减少依赖**: 只导入必要的包
2. **使用构建标志**: `-ldflags '-s -w'` 减小二进制大小
3. **连接池**: 对于数据库连接，使用连接池
4. **缓存**: 合理使用缓存减少冷启动影响
