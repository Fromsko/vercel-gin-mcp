package api

import (
	"net/http"
	"os"

	"mcp-server/internal/mcp/core"
	"mcp-server/internal/mcp/tools"
	"mcp-server/internal/web"

	"github.com/gin-gonic/gin"
)

var engine *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	engine = gin.New()

	// 创建 MCP 服务器 - 链式调用风格
	server := core.New("vercel-gin-mcp").Version("1.0.0")

	// 注册工具
	registerTools(server)

	// Web 服务器（用于 / 与 /tools）
	authCode := os.Getenv("AUTH_CODE")
	webServer := web.New(authCode)
	webServer.SetupRoutes(server)

	// 注册 MCP 端点
	engine.POST("/mcp", server.Handler())

	// 将其他路由委托给 web 引擎（渲染模板）
	engine.NoRoute(gin.WrapH(webServer.GetEngine()))
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

func Handler(w http.ResponseWriter, r *http.Request) {
	engine.ServeHTTP(w, r)
}
