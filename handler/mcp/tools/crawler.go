package tools

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"mcp-server/handler/mcp/concurrent"
	"mcp-server/handler/mcp/utils"
)

type CrawlPage struct {
	URL      string `json:"url"`
	Title    string `json:"title"`
	Markdown string `json:"markdown"`
}

type CrawlResult struct {
	Pages []CrawlPage `json:"pages"`
	Total int         `json:"total"`
}

type CrawlOptions struct {
	MaxDepth   int
	MaxPages   int
	SameDomain bool
}

func DefaultCrawlOptions() *CrawlOptions {
	return &CrawlOptions{
		MaxDepth:   2,
		MaxPages:   20,
		SameDomain: true,
	}
}

type RecursiveCrawler struct {
	opts    *CrawlOptions
	pages   []CrawlPage
	mu      sync.Mutex
	baseURL *url.URL
}

func NewRecursiveCrawler(opts *CrawlOptions) *RecursiveCrawler {
	if opts == nil {
		opts = DefaultCrawlOptions()
	}
	return &RecursiveCrawler{opts: opts}
}

func (rc *RecursiveCrawler) Crawl(startURL string) (*CrawlResult, error) {
	u, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	rc.baseURL = u

	c := colly.NewCollector(
		colly.MaxDepth(rc.opts.MaxDepth),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainRegexp: rc.baseURL.Host,
		Parallelism:  4,
		Delay:        100,
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if rc.pageCount() >= rc.opts.MaxPages {
			return
		}

		href := e.Request.AbsoluteURL(e.Attr("href"))
		if href == "" {
			return
		}

		link, err := url.Parse(href)
		if err != nil {
			return
		}

		if rc.opts.SameDomain && link.Host != rc.baseURL.Host {
			return
		}

		if link.Scheme != "http" && link.Scheme != "https" {
			return
		}

		e.Request.Visit(href)
	})

	c.OnScraped(func(r *colly.Response) {
		rc.mu.Lock()
		defer rc.mu.Unlock()
		if rc.pageCount() >= rc.opts.MaxPages {
			return
		}

		dom, err := goquery.NewDocumentFromReader(strings.NewReader(string(r.Body)))
		if err != nil {
			return
		}

		title := strings.TrimSpace(dom.Find("title").Text())
		markdown := extractFromDOM(dom)
		if markdown == "" {
			return
		}

		rc.pages = append(rc.pages, CrawlPage{
			URL:      r.Request.URL.String(),
			Title:    title,
			Markdown: formatMarkdown(title, markdown),
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		// skip failed pages
	})

	c.Visit(startURL)
	c.Wait()

	return &CrawlResult{
		Pages: rc.pages,
		Total: len(rc.pages),
	}, nil
}

func (rc *RecursiveCrawler) pageCount() int {
	return len(rc.pages)
}

func extractFromDOM(doc *goquery.Document) string {
	var md strings.Builder

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(_ int, s *goquery.Selection) {
		level := s.Get(0).Data[1] - '0'
		prefix := strings.Repeat("#", int(level))
		text := cleanText(s.Text())
		if text != "" {
			md.WriteString(fmt.Sprintf("%s %s\n\n", prefix, text))
		}
	})

	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if text != "" {
			md.WriteString(text + "\n\n")
		}
	})

	doc.Find("ul li, ol li").Each(func(_ int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if text != "" {
			md.WriteString("- " + text + "\n")
		}
	})

	doc.Find("pre").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		if text != "" {
			md.WriteString("```\n" + text + "\n```\n\n")
		}
	})

	doc.Find("table").Each(func(_ int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if text != "" {
			md.WriteString(text + "\n\n")
		}
	})

	doc.Find("blockquote").Each(func(_ int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if text != "" {
			md.WriteString("> " + text + "\n\n")
		}
	})

	return md.String()
}

func QuickCrawl(startURL string, maxDepth, maxPages int) (*CrawlResult, error) {
	opts := DefaultCrawlOptions()
	if maxDepth > 0 {
		opts.MaxDepth = maxDepth
	}
	if maxPages > 0 {
		opts.MaxPages = maxPages
	}
	return NewRecursiveCrawler(opts).Crawl(startURL)
}

// QuickCrawlMulti concurrently crawls multiple URLs
func QuickCrawlMulti(urls []string, maxDepth, maxPagesPerURL int) (*CrawlResult, error) {
	if len(urls) == 0 {
		return nil, fmt.Errorf("no URLs provided")
	}

	// Limit concurrent crawls to avoid overwhelming
	maxConcurrent := utils.MinInt(3, len(urls)) // cap concurrent crawls
	pool := concurrent.NewPool(maxConcurrent)

	resultsChan := make(chan *CrawlResult, len(urls))
	errorsChan := make(chan error, len(urls))

	for _, url := range urls {
		url := url // capture loop variable
		pool.Go(func() error {
			result, err := QuickCrawl(url, maxDepth, maxPagesPerURL)
			if err != nil {
				errorsChan <- err
				return err
			}
			resultsChan <- result
			return nil
		})
	}

	// Wait for all crawls to complete
	pool.Wait()

	// Close channels after pool completes
	close(resultsChan)
	close(errorsChan)

	// Collect results
	var allResults CrawlResult
	for result := range resultsChan {
		allResults.Pages = append(allResults.Pages, result.Pages...)
		allResults.Total += result.Total
	}

	// Check for any errors
	select {
	case err := <-errorsChan:
		return &allResults, err
	default:
		// No errors
	}

	return &allResults, nil
}
