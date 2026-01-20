# Go WebDAV 客户端实现

## 概述

WebDAV (Web-based Distributed Authoring and Versioning) 是 HTTP 协议的扩展，允许用户远程管理和编辑服务器上的文件。Go 提供了多个 WebDAV 客户端库。

## 主要库选择

### 1. gowebdav (推荐)

`github.com/studio-b12/gowebdav` - 纯 Go 实现的 WebDAV 客户端库

```bash
go get github.com/studio-b12/gowebdav
```

### 2. go-webdav

`github.com/emersion/go-webdav` - 支持 WebDAV、CalDAV 和 CardDAV

```bash
go get github.com/emersion/go-webdav
```

## gowebdav 使用示例

### 基础连接

```go
package main

import (
    "fmt"
    "log"

    "github.com/studio-b12/gowebdav"
)

func main() {
    // 创建客户端
    client := gowebdav.NewClient(
        "http://your-webdav-server.com/webdav",
        "username",
        "password",
    )

    // 测试连接
    err := client.Connect()
    if err != nil {
        log.Fatal("连接失败:", err)
    }

    fmt.Println("连接成功!")
}
```

### 文件操作

#### 上传文件

```go
// 上传本地文件到远程服务器
func uploadFile(client *gowebdav.Client, localPath, remotePath string) error {
    // 读取本地文件
    data, err := ioutil.ReadFile(localPath)
    if err != nil {
        return err
    }

    // 上传到远程服务器
    return client.Write(remotePath, data, 0644)
}

// 使用示例
err := uploadFile(client, "./local.txt", "/remote/uploaded.txt")
if err != nil {
    log.Fatal("上传失败:", err)
}
```

#### 下载文件

```go
// 从远程服务器下载文件
func downloadFile(client *gowebdav.Client, remotePath, localPath string) error {
    // 读取远程文件
    data, err := client.Read(remotePath)
    if err != nil {
        return err
    }

    // 写入本地文件
    return ioutil.WriteFile(localPath, data, 0644)
}

// 使用示例
err := downloadFile(client, "/remote/data.txt", "./downloaded.txt")
if err != nil {
    log.Fatal("下载失败:", err)
}
```

#### 删除文件

```go
// 删除远程文件
err := client.Remove("/remote/file.txt")
if err != nil {
    log.Fatal("删除失败:", err)
}
```

### 目录操作

#### 创建目录

```go
// 创建目录（递归创建）
err := client.Mkdir("/remote/newfolder", 0755)
if err != nil {
    log.Fatal("创建目录失败:", err)
}

// 创建多级目录
err = client.MkdirAll("/remote/path/to/nested/folder", 0755)
if err != nil {
    log.Fatal("创建多级目录失败:", err)
}
```

#### 列出目录内容

```go
// 列出目录内容
files, err := client.ReadDir("/remote/folder")
if err != nil {
    log.Fatal("读取目录失败:", err)
}

for _, file := range files {
    if file.IsDir() {
        fmt.Printf("目录: %s\n", file.Name())
    } else {
        fmt.Printf("文件: %s (大小: %d)\n", file.Name(), file.Size())
    }
}
```

#### 删除目录

```go
// 删除空目录
err := client.Remove("/remote/emptyfolder")
if err != nil {
    log.Fatal("删除目录失败:", err)
}

// 递归删除目录及其内容
err = client.RemoveAll("/remote/folderwithcontent")
if err != nil {
    log.Fatal("递归删除失败:", err)
}
```

### 文件信息

#### 获取文件信息

```go
// 获取文件/目录信息
info, err := client.Stat("/remote/file.txt")
if err != nil {
    log.Fatal("获取文件信息失败:", err)
}

fmt.Printf("名称: %s\n", info.Name())
fmt.Printf("大小: %d\n", info.Size())
fmt.Printf("修改时间: %s\n", info.ModTime())
fmt.Printf("是否目录: %t\n", info.IsDir())
```

#### 检查文件是否存在

```go
// 检查路径是否存在
exists, err := client.Exists("/remote/file.txt")
if err != nil {
    log.Fatal("检查失败:", err)
}

if exists {
    fmt.Println("文件存在")
} else {
    fmt.Println("文件不存在")
}
```

### 高级功能

#### 复制和移动

```go
// 复制文件/目录
err := client.Copy("/remote/source.txt", "/remote/destination.txt")
if err != nil {
    log.Fatal("复制失败:", err)
}

// 移动/重命名文件/目录
err = client.Rename("/remote/oldname.txt", "/remote/newname.txt")
if err != nil {
    log.Fatal("移动失败:", err)
}
```

#### 读取大文件（流式）

```go
// 创建远程文件读取器
reader, err := client.ReadStream("/remote/largefile.zip")
if err != nil {
    log.Fatal("创建读取器失败:", err)
}
defer reader.Close()

// 创建本地文件
localFile, err := os.Create("./largefile.zip")
if err != nil {
    log.Fatal("创建本地文件失败:", err)
}
defer localFile.Close()

// 复制数据
_, err = io.Copy(localFile, reader)
if err != nil {
    log.Fatal("下载失败:", err)
}
```

#### 写入大文件（流式）

```go
// 打开本地文件
localFile, err := os.Open("./largefile.zip")
if err != nil {
    log.Fatal("打开本地文件失败:", err)
}
defer localFile.Close()

// 创建远程文件写入器
writer, err := client.WriteStream("/remote/upload.zip", 0644)
if err != nil {
    log.Fatal("创建写入器失败:", err)
}
defer writer.Close()

// 复制数据
_, err = io.Copy(writer, localFile)
if err != nil {
    log.Fatal("上传失败:", err)
}
```

### SSL/TLS 配置

```go
package main

import (
    "crypto/tls"
    "log"

    "github.com/studio-b12/gowebdav"
)

func main() {
    // 自定义 TLS 配置
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true, // 跳过证书验证（仅用于测试）
        // MinVersion:         tls.VersionTLS12,
    }

    // 创建带自定义配置的客户端
    client := gowebdav.NewClient(
        "https://your-webdav-server.com/webdav",
        "username",
        "password",
    )

    // 设置 TLS 配置
    client.SetTransport(&http.Transport{
        TLSClientConfig: tlsConfig,
    })

    // 测试连接
    err := client.Connect()
    if err != nil {
        log.Fatal("连接失败:", err)
    }

    log.Println("HTTPS 连接成功!")
}
```

### 自定义 HTTP 客户端

```go
package main

import (
    "net/http"
    "time"

    "github.com/studio-b12/gowebdav"
)

func createCustomClient() *gowebdav.Client {
    // 创建自定义 HTTP 客户端
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        10,
            IdleConnTimeout:     30 * time.Second,
            DisableCompression:  false,
        },
    }

    // 创建 WebDAV 客户端
    client := gowebdav.NewClient(
        "http://your-webdav-server.com/webdav",
        "username",
        "password",
    )

    // 设置自定义 HTTP 客户端
    client.SetTransport(httpClient.Transport)

    return client
}
```

## 使用 emersion/go-webdav

### 基础示例

```go
package main

import (
    "context"
    "log"

    "github.com/emersion/go-webdav"
    "github.com/emersion/go-webdav/client"
)

func main() {
    // 创建客户端
    c, err := client.NewClient(
        &http.Client{},
        "http://your-webdav-server.com/webdav",
    )
    if err != nil {
        log.Fatal("创建客户端失败:", err)
    }

    // 设置认证
    c.SetBasicAuth("username", "password")

    // 获取文件系统
    fs, err := c.FileSystem()
    if err != nil {
        log.Fatal("获取文件系统失败:", err)
    }

    // 列出文件
    f, err := fs.OpenFile(context.Background(), "/", os.O_RDONLY, 0)
    if err != nil {
        log.Fatal("打开目录失败:", err)
    }
    defer f.Close()

    // 读取目录内容
    entries, err := f.Readdir(-1)
    if err != nil {
        log.Fatal("读取目录失败:", err)
    }

    for _, entry := range entries {
        log.Printf("名称: %s, 大小: %d", entry.Name(), entry.Size())
    }
}
```

## 完整示例：WebDAV 文件同步工具

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/studio-b12/gowebdav"
)

type SyncConfig struct {
    LocalDir  string
    RemoteDir string
    Username  string
    Password  string
    ServerURL string
}

func main() {
    config := parseFlags()

    // 创建 WebDAV 客户端
    client := gowebdav.NewClient(config.ServerURL, config.Username, config.Password)

    // 测试连接
    if err := client.Connect(); err != nil {
        log.Fatal("连接失败:", err)
    }

    // 开始同步
    fmt.Printf("开始同步: %s -> %s\n", config.LocalDir, config.RemoteDir)

    if err := syncToRemote(client, config); err != nil {
        log.Fatal("同步失败:", err)
    }

    fmt.Println("同步完成!")
}

func parseFlags() *SyncConfig {
    config := &SyncConfig{}

    flag.StringVar(&config.LocalDir, "local", "./local", "本地目录")
    flag.StringVar(&config.RemoteDir, "remote", "/remote", "远程目录")
    flag.StringVar(&config.Username, "user", "", "用户名")
    flag.StringVar(&config.Password, "pass", "", "密码")
    flag.StringVar(&config.ServerURL, "url", "", "WebDAV 服务器 URL")

    flag.Parse()

    if config.Username == "" || config.Password == "" || config.ServerURL == "" {
        log.Fatal("必须提供用户名、密码和服务器 URL")
    }

    return config
}

func syncToRemote(client *gowebdav.Client, config *SyncConfig) error {
    // 确保远程目录存在
    if err := client.MkdirAll(config.RemoteDir, 0755); err != nil {
        return fmt.Errorf("创建远程目录失败: %w", err)
    }

    // 遍历本地目录
    return filepath.Walk(config.LocalDir, func(localPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // 计算相对路径
        relPath, err := filepath.Rel(config.LocalDir, localPath)
        if err != nil {
            return err
        }

        // 构建远程路径（使用正斜杠）
        remotePath := config.RemoteDir + "/" + filepath.ToSlash(relPath)

        if info.IsDir() {
            // 创建远程目录
            if err := client.Mkdir(remotePath, info.Mode()); err != nil {
                log.Printf("创建目录失败 %s: %v", remotePath, err)
            }
        } else {
            // 检查文件是否需要更新
            remoteInfo, err := client.Stat(remotePath)
            if err == nil && remoteInfo.ModTime().After(info.ModTime()) {
                fmt.Printf("跳过 %s（远程文件更新）\n", relPath)
                return nil
            }

            // 上传文件
            data, err := os.ReadFile(localPath)
            if err != nil {
                return err
            }

            if err := client.Write(remotePath, data, info.Mode()); err != nil {
                log.Printf("上传文件失败 %s: %v", relPath, err)
            } else {
                fmt.Printf("上传: %s\n", relPath)
            }
        }

        return nil
    })
}
```

## 错误处理

### 常见错误类型

```go
package main

import (
    "errors"
    "fmt"
    "net/http"

    "github.com/studio-b12/gowebdav"
)

func handleWebDAVError(err error) {
    if err == nil {
        return
    }

    // 检查是否是特定状态码错误
    if gowebdav.IsErrCode(err, 404) {
        fmt.Println("文件或目录不存在")
    } else if gowebdav.IsErrCode(err, 401) {
        fmt.Println("认证失败")
    } else if gowebdav.IsErrCode(err, 403) {
        fmt.Println("权限不足")
    } else if gowebdav.IsErrCode(err, 423) {
        fmt.Println("资源被锁定")
    } else {
        fmt.Printf("错误: %v\n", err)
    }
}
```

### 重试机制

```go
func uploadWithRetry(client *gowebdav.Client, localPath, remotePath string, maxRetries int) error {
    var lastErr error

    for i := 0; i < maxRetries; i++ {
        data, err := os.ReadFile(localPath)
        if err != nil {
            return err
        }

        err = client.Write(remotePath, data, 0644)
        if err == nil {
            return nil
        }

        lastErr = err

        // 如果是网络错误，等待后重试
        if isNetworkError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }

        // 其他错误直接返回
        break
    }

    return fmt.Errorf("上传失败，重试 %d 次后仍失败: %w", maxRetries, lastErr)
}

func isNetworkError(err error) bool {
    // 检查是否是网络相关错误
    return strings.Contains(err.Error(), "connection") ||
           strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "network")
}
```

## 最佳实践

1. **连接池**: 复用客户端连接以提高性能
2. **并发控制**: 使用 goroutine 池控制并发上传/下载
3. **错误处理**: 实现适当的重试和错误恢复机制
4. **进度跟踪**: 对于大文件操作，实现进度显示
5. **安全性**: 使用 HTTPS 和安全的认证方式
6. **缓存**: 对于频繁访问的元数据，实现本地缓存

## 注意事项

- WebDAV 服务器实现可能有差异，测试时注意兼容性
- 大文件上传/下载时注意内存使用
- 某些服务器可能不支持所有 WebDAV 功能
- 注意路径分隔符（Windows 使用 `\`，WebDAV 使用 `/`）
- 考虑时区问题，特别是文件时间戳
