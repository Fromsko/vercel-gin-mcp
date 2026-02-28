package tools

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	if err != nil || bodyContent.Len() == 0 {
		// 回退到纯 HTTP + goquery（带 UA），适配反爬/JS 渲染站点
		return httpFallbackFetch(url)
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
			fmt.Fprintf(&md, "%s %s\n\n", prefix, text)
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
			fmt.Fprintf(&md, "[%s](%s)\n", text, href)
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
		fmt.Fprintf(&md, "# %s\n\n", title)
	}

	md.WriteString(content)

	return md.String()
}

// QuickFetch 快速抓取（便捷函数）
func QuickFetch(url string) (*ScrapeResult, error) {
	return NewScraper().FetchToMarkdown(url)
}

// httpFallbackFetch 使用原生 HTTP + goquery 回退抓取
func httpFallbackFetch(url string) (*ScrapeResult, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request failed: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse html failed: %w", err)
	}

	title := strings.TrimSpace(doc.Find("title").First().Text())
	var md strings.Builder

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, sel *goquery.Selection) {
		tag := goquery.NodeName(sel)
		if len(tag) == 2 && tag[0] == 'h' {
			level := tag[1] - '0'
			prefix := strings.Repeat("#", int(level))
			text := cleanText(sel.Text())
			if text != "" {
				fmt.Fprintf(&md, "%s %s\n\n", prefix, text)
			}
		}
	})

	doc.Find("p").Each(func(_ int, sel *goquery.Selection) {
		text := cleanText(sel.Text())
		if text != "" {
			md.WriteString(text + "\n\n")
		}
	})

	doc.Find("ul li, ol li").Each(func(_ int, sel *goquery.Selection) {
		text := cleanText(sel.Text())
		if text != "" {
			md.WriteString("- " + text + "\n")
		}
	})

	doc.Find("pre, code").Each(func(_ int, sel *goquery.Selection) {
		text := sel.Text()
		if text != "" {
			md.WriteString("```\n" + text + "\n```\n\n")
		}
	})

	doc.Find("a[href]").Each(func(_ int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		text := cleanText(sel.Text())
		if text != "" && href != "" && !strings.HasPrefix(href, "#") {
			fmt.Fprintf(&md, "[%s](%s)\n", text, href)
		}
	})

	return &ScrapeResult{
		URL:      url,
		Title:    title,
		Markdown: formatMarkdown(title, md.String()),
	}, nil
}
