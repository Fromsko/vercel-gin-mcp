package tools

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gocolly/colly/v2"
)

// ScrapeResult 抓取结果
type ScrapeResult struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

// Scraper 网页抓取器
type Scraper struct {
	collector *colly.Collector
}

// NewScraper 创建新的抓取器
func NewScraper() *Scraper {
	c := colly.NewCollector(
		colly.AllowURLRevisit(),
		colly.MaxDepth(1),
	)
	return &Scraper{collector: c}
}

// FetchToMarkdown 抓取网页并转换为 Markdown
func (s *Scraper) FetchToMarkdown(url string) (*ScrapeResult, error) {
	result := &ScrapeResult{URL: url}

	var bodyContent strings.Builder
	var title string

	s.collector.OnHTML("title", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)
	})

	s.collector.OnHTML("body", func(e *colly.HTMLElement) {
		// 提取主要内容区域
		content := extractContent(e)
		bodyContent.WriteString(content)
	})

	err := s.collector.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	result.Title = title
	result.Markdown = formatMarkdown(title, bodyContent.String())

	return result, nil
}

// extractContent 从 HTML 元素提取内容并转换为 Markdown
func extractContent(e *colly.HTMLElement) string {
	var md strings.Builder

	// 提取标题
	e.ForEach("h1, h2, h3, h4, h5, h6", func(_ int, el *colly.HTMLElement) {
		level := el.Name[1] - '0'
		prefix := strings.Repeat("#", int(level))
		text := cleanText(el.Text)
		if text != "" {
			md.WriteString(fmt.Sprintf("%s %s\n\n", prefix, text))
		}
	})

	// 提取段落
	e.ForEach("p", func(_ int, el *colly.HTMLElement) {
		text := cleanText(el.Text)
		if text != "" {
			md.WriteString(text + "\n\n")
		}
	})

	// 提取列表
	e.ForEach("ul li, ol li", func(_ int, el *colly.HTMLElement) {
		text := cleanText(el.Text)
		if text != "" {
			md.WriteString("- " + text + "\n")
		}
	})

	// 提取代码块
	e.ForEach("pre, code", func(_ int, el *colly.HTMLElement) {
		text := el.Text
		if text != "" {
			md.WriteString("```\n" + text + "\n```\n\n")
		}
	})

	// 提取链接
	e.ForEach("a[href]", func(_ int, el *colly.HTMLElement) {
		href := el.Attr("href")
		text := cleanText(el.Text)
		if text != "" && href != "" && !strings.HasPrefix(href, "#") {
			md.WriteString(fmt.Sprintf("[%s](%s)\n", text, href))
		}
	})

	return md.String()
}

// cleanText 清理文本
func cleanText(s string) string {
	// 移除多余空白
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// formatMarkdown 格式化最终的 Markdown
func formatMarkdown(title, content string) string {
	var md strings.Builder

	if title != "" {
		md.WriteString(fmt.Sprintf("# %s\n\n", title))
	}

	md.WriteString(content)

	return md.String()
}

// QuickFetch 快速抓取（便捷函数）
func QuickFetch(url string) (*ScrapeResult, error) {
	return NewScraper().FetchToMarkdown(url)
}
