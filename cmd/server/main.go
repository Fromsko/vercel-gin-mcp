package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"mcp-server/internal/mcp/core"
	"mcp-server/internal/mcp/tools"
	"mcp-server/internal/web"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 创建 Gin 引擎
	engine := gin.New()
	engine.Use(gin.Recovery())

	// 创建 MCP 服务器
	mcpServer := core.New("vercel-gin-mcp").Version("1.0.0")

	// 注册工具
	registerTools(mcpServer)

	// 创建 Web 服务器
	authCode := os.Getenv("AUTH_CODE")
	webServer := web.New(authCode)

	// 注册 MCP 端点
	engine.POST("/mcp", mcpServer.Handler())

	// 注册 Web 路由 - 使用 web 引擎作为处理器
	webEngine := webServer.GetEngine()
	webServer.SetupRoutes(mcpServer)

	// 使用 webEngine 作为所有非 MCP 路由的处理器
	engine.NoRoute(gin.WrapH(webEngine))

	// 启动服务器
	addr := fmt.Sprintf(":%s", port)
	log.Printf("服务器启动在 http://localhost%s", addr)
	log.Printf("MCP 端点: POST /mcp")
	log.Printf("Web 界面: http://localhost%s/", addr)
	log.Printf("工具管理: http://localhost%s/tools", addr)

	if authCode != "" {
		log.Printf("认证已启用，请使用 AUTH_CODE 环境变量设置的授权码")
	}

	if err := engine.Run(addr); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}

// registerTools 注册所有工具
func registerTools(server *core.Server) {
	// 网页抓取工具
	server.Register(
		core.NewTool("fetch").
			Desc("抓取网页内容并转换为 Markdown 格式").
			String("url", "要抓取的网页 URL", true).
			Handle(func(ctx *core.Context) *core.ToolResult {
				url := ctx.String("url")
				result, err := tools.QuickFetch(url)
				if err != nil {
					return ctx.Error("抓取失败: " + err.Error())
				}
				return ctx.JSON(core.H{
					"url":      result.URL,
					"title":    result.Title,
					"markdown": result.Markdown,
				})
			}),
	)

	// 网页抓取工具 - 仅返回 Markdown
	server.Register(
		core.NewTool("fetch_md").
			Desc("抓取网页内容，仅返回 Markdown 文本").
			String("url", "要抓取的网页 URL", true).
			Handle(func(ctx *core.Context) *core.ToolResult {
				url := ctx.String("url")
				result, err := tools.QuickFetch(url)
				if err != nil {
					return ctx.Error("抓取失败: " + err.Error())
				}
				return ctx.Markdown(result.Markdown)
			}),
	)

	// 并行抓取多个 URL
	server.Register(
		core.NewTool("fetch_multi").
			Desc("并行抓取多个 URL，并返回每个页面的标题和 Markdown 内容").
			String("urls", "逗号分隔的 URL 列表", true).
			Handle(func(ctx *core.Context) *core.ToolResult {
				raw := ctx.String("urls")
				parts := strings.Split(raw, ",")
				var urls []string
				for _, p := range parts {
					u := strings.TrimSpace(p)
					if u != "" {
						urls = append(urls, u)
					}
				}
				if len(urls) == 0 {
					return ctx.Error("无有效 URL")
				}

				type item struct {
					URL      string `json:"url"`
					Title    string `json:"title"`
					Markdown string `json:"markdown"`
					Error    string `json:"error,omitempty"`
				}

				results := make([]item, len(urls))
				var wg sync.WaitGroup
				wg.Add(len(urls))

				for i, u := range urls {
					i, u := i, u
					go func() {
						defer wg.Done()
						res, err := tools.QuickFetch(u)
						if err != nil {
							results[i] = item{URL: u, Error: err.Error()}
							return
						}
						results[i] = item{URL: u, Title: res.Title, Markdown: res.Markdown}
					}()
				}

				wg.Wait()
				return ctx.JSON(core.H{"results": results})
			}),
	)

	// GitHub 仓库文档下载工具
	server.Register(
		core.NewTool("download_docs").
			Desc("从 GitHub 仓库下载文档文件（.md, .txt），返回文件内容").
			String("repo", "GitHub 仓库 URL，如 https://github.com/user/repo", true).
			String("path", "文档路径过滤，如 docs（可选）", false).
			Handle(func(ctx *core.Context) *core.ToolResult {
				repoURL := ctx.String("repo")
				docsPath := ctx.String("path")

				var result *tools.DocsResult
				var err error

				if docsPath != "" {
					result, err = tools.QuickDownloadPath(repoURL, docsPath)
				} else {
					result, err = tools.QuickDownload(repoURL)
				}

				if err != nil {
					return ctx.Error("下载失败: " + err.Error())
				}

				return ctx.JSON(core.H{
					"repo_url": result.RepoURL,
					"owner":    result.Owner,
					"repo":     result.Repo,
					"count":    result.Count,
					"files":    result.Files,
				})
			}),
	)

	// GitHub 仓库文档下载工具 - 返回 Markdown 格式
	server.Register(
		core.NewTool("download_docs_md").
			Desc("从 GitHub 仓库下载文档文件，返回合并的 Markdown 文本").
			String("repo", "GitHub 仓库 URL，如 https://github.com/user/repo", true).
			String("path", "文档路径过滤，如 docs（可选）", false).
			Handle(func(ctx *core.Context) *core.ToolResult {
				repoURL := ctx.String("repo")
				docsPath := ctx.String("path")

				var result *tools.DocsResult
				var err error

				if docsPath != "" {
					result, err = tools.QuickDownloadPath(repoURL, docsPath)
				} else {
					result, err = tools.QuickDownload(repoURL)
				}

				if err != nil {
					return ctx.Error("下载失败: " + err.Error())
				}

				return ctx.Markdown(result.ToMarkdown())
			}),
	)
}
