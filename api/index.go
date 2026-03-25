package api

import (
	"fmt"
	"math"
	"net/http"
	"strings"

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

	// 递归爬取 + 内存 RAG 检索工具
	server.Register(
		mcp.NewTool("rag_crawl").
			Desc("递归爬取网站，建立内存索引，支持关键词检索相关内容").
			String("url", "起始 URL，将从此页面开始递归爬取", true).
			String("query", "检索关键词，在爬取的内容中搜索相关片段", true).
			Number("max_depth", "最大递归深度（默认 2）", false).
			Number("max_pages", "最大爬取页面数（默认 20）", false).
			Number("top_k", "返回最相关的 K 个结果（默认 5）", false).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
				startURL := ctx.String("url")
				query := ctx.String("query")
				maxDepth := ctx.Int("max_depth")
				maxPages := ctx.Int("max_pages")
				topK := ctx.Int("top_k")

				if maxDepth <= 0 {
					maxDepth = 2
				}
				if maxPages <= 0 {
					maxPages = 20
				}
				if topK <= 0 {
					topK = 5
				}

				crawlResult, err := tools.QuickCrawl(startURL, maxDepth, maxPages)
				if err != nil {
					return ctx.Error("爬取失败: " + err.Error())
				}

				if crawlResult.Total == 0 {
					return ctx.Error("未爬取到任何内容")
				}

				index := tools.NewRAGIndex()
				index.AddFromCrawl(crawlResult, 500)

				results := index.Search(query, topK)
				if len(results) == 0 {
					return ctx.JSON(mcp.H{
						"crawled_pages": crawlResult.Total,
						"index_stats":   index.Stats(),
						"results":       []any{},
						"message":       "未找到相关内容，请尝试其他关键词",
					})
				}

				type resultItem struct {
					Title    string  `json:"title"`
					URL      string  `json:"url"`
					Score    float64 `json:"score"`
					Content  string  `json:"content"`
				}

				items := make([]resultItem, 0, len(results))
				for _, r := range results {
					items = append(items, resultItem{
						Title:   r.Chunk.Title,
						URL:     r.Chunk.URL,
						Score:   math.Round(r.Score*100) / 100,
						Content: r.Chunk.Text,
					})
				}

				return ctx.JSON(mcp.H{
					"crawled_pages": crawlResult.Total,
					"index_stats":   index.Stats(),
					"results":       items,
				})
			}),
	)

	// 并发多 URL 爬取 + RAG 检索工具
	server.Register(
		mcp.NewTool("rag_crawl_multi").
			Desc("并发爬取多个网站，建立内存索引，支持关键词检索相关内容").
			String("urls", "逗号分隔的 URL 列表，将并发爬取这些网站", true).
			String("query", "检索关键词，在爬取的内容中搜索相关片段", true).
			Number("max_depth", "最大递归深度（默认 2）", false).
			Number("max_pages_per_url", "每个 URL 最大爬取页面数（默认 10）", false).
			Number("top_k", "返回最相关的 K 个结果（默认 5）", false).
			Handle(func(ctx *mcp.Context) *mcp.ToolResult {
				urlsStr := ctx.String("urls")
				query := ctx.String("query")
				maxDepth := ctx.Int("max_depth")
				maxPagesPerURL := ctx.Int("max_pages_per_url")
				topK := ctx.Int("top_k")

				if maxDepth <= 0 {
					maxDepth = 2
				}
				if maxPagesPerURL <= 0 {
					maxPagesPerURL = 10
				}
				if topK <= 0 {
					topK = 5
				}

				// Parse URLs
				urls := strings.Split(urlsStr, ",")
				for i, url := range urls {
					urls[i] = strings.TrimSpace(url)
				}

				crawlResult, err := tools.QuickCrawlMulti(urls, maxDepth, maxPagesPerURL)
				if err != nil {
					return ctx.Error("爬取失败: " + err.Error())
				}

				if crawlResult.Total == 0 {
					return ctx.Error("未从任何网站爬取到内容")
				}

				index := tools.NewRAGIndex()
				index.AddFromCrawl(crawlResult, 500)

				results := index.Search(query, topK)
				if len(results) == 0 {
					return ctx.JSON(mcp.H{
						"crawled_urls_count": len(urls),
						"crawled_pages":      crawlResult.Total,
						"index_stats":        index.Stats(),
						"results":            []any{},
						"message":            "未找到相关内容，请尝试其他关键词",
					})
				}

				type resultItem struct {
					Title    string  `json:"title"`
					URL      string  `json:"url"`
					Score    float64 `json:"score"`
					Content  string  `json:"content"`
				}

				items := make([]resultItem, 0, len(results))
				for _, r := range results {
					items = append(items, resultItem{
						Title:   r.Chunk.Title,
						URL:     r.Chunk.URL,
						Score:   math.Round(r.Score*100) / 100,
						Content: r.Chunk.Text,
					})
				}

				return ctx.JSON(mcp.H{
					"crawled_urls_count": len(urls),
					"crawled_pages":      crawlResult.Total,
					"index_stats":        index.Stats(),
					"results":            items,
				})
			}),
	)

	// 注册 MCP 端点
	engine.POST("/mcp", server.Handler())
}

func Handler(w http.ResponseWriter, r *http.Request) {
	engine.ServeHTTP(w, r)
}
