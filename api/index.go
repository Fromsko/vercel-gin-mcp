package api

import (
	"fmt"
	"net/http"

	"mcp-server/handler/mcp"
	"mcp-server/handler/mcp/tools"

	"github.com/gin-gonic/gin"
)

var engine *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode)
	engine = gin.New()

	// 创建 MCP 服务器 - 链式调用风格
	server := mcp.New("vercel-gin-mcp").Version("1.0.0")

	// 注册工具 - 函数式注册
	server.Register(
		mcp.NewTool("echo").
			Desc("回显输入的文本").
			String("text", "要回显的文本", true).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
				return ctx.Text("回显: " + ctx.String("text"))
			}),
	)

	server.Register(
		mcp.NewTool("add").
			Desc("计算两个数字的和").
			Number("a", "第一个数字", true).
			Number("b", "第二个数字", true).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
				a, b := ctx.Float("a"), ctx.Float("b")
				return ctx.Text(fmt.Sprintf("%.2f + %.2f = %.2f", a, b, a+b))
			}),
	)

	// 网页抓取工具 - gocolly 集成
	server.Register(
		mcp.NewTool("fetch").
			Desc("抓取网页内容并转换为 Markdown 格式").
			String("url", "要抓取的网页 URL", true).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
				url := ctx.String("url")
				result, err := tools.QuickFetch(url)
				if err != nil {
					return ctx.Error("抓取失败: " + err.Error())
				}
				return ctx.JSON(mcp.H{
					"url":      result.URL,
					"title":    result.Title,
					"markdown": result.Markdown,
				})
			}),
	)

	// 网页抓取工具 - 仅返回 Markdown
	server.Register(
		mcp.NewTool("fetch_md").
			Desc("抓取网页内容，仅返回 Markdown 文本").
			String("url", "要抓取的网页 URL", true).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
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
		mcp.NewTool("download_docs").
			Desc("从 GitHub 仓库下载文档文件（.md, .txt），返回文件内容").
			String("repo", "GitHub 仓库 URL，如 https://github.com/user/repo", true).
			String("path", "文档路径过滤，如 docs（可选）", false).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
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

				return ctx.JSON(mcp.H{
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
		mcp.NewTool("download_docs_md").
			Desc("从 GitHub 仓库下载文档文件，返回合并的 Markdown 文本").
			String("repo", "GitHub 仓库 URL，如 https://github.com/user/repo", true).
			String("path", "文档路径过滤，如 docs（可选）", false).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
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

	// 注册 MCP 端点
	engine.POST("/mcp", server.Handler())
}

func Handler(w http.ResponseWriter, r *http.Request) {
	engine.ServeHTTP(w, r)
}
