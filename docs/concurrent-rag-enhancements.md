# Enhanced Go MCP Server with Concurrent Goroutine Pools and Improved RAG

## Overview

This document describes the enhancements made to the Go MCP server to integrate concurrent goroutine pools and improve RAG (Retrieval Augmented Generation) functionality. The changes allow for more efficient concurrent processing of web crawling and search operations.

## Changes Made

### 1. Concurrent Package (`handler/mcp/concurrent/`)

Created a new concurrent package that wraps the `sourcegraph/conc` library to provide:

- **Regular Pool**: For basic concurrent task execution
- **Context Pool**: For context-aware concurrent task execution
- **Simple API**: Easy-to-use functions for common concurrent patterns

```go
// Example usage
pool := concurrent.NewPool(5)  // Max 5 concurrent goroutines
pool.Go(func() error { /* task */ })
pool.Wait()
```

### 2. Enhanced RAG Index (`handler/mcp/tools/rag.go`)

Improved the RAG index with concurrent search functionality:

- **Concurrent Scoring**: Parallel processing of search scoring across documents
- **Optimized Performance**: Uses worker pool to distribute scoring work
- **BM25 Algorithm**: Maintains the same effective ranking algorithm
- **Thread Safety**: Proper synchronization for concurrent access

### 3. Enhanced Crawler (`handler/mcp/tools/crawler.go`)

Added concurrent crawling capabilities:

- **Single URL Crawl**: Original functionality preserved
- **Multi-URL Concurrent Crawl**: New `QuickCrawlMulti` function to crawl multiple URLs concurrently
- **Resource Management**: Proper limits on concurrent crawls to prevent overwhelming target servers

### 4. New Utilities Package (`handler/mcp/utils/`)

Created utility functions to avoid code duplication:

- `MinInt`, `MaxInt` functions for integer comparison

### 5. New MCP Tools in API (`api/index.go`)

Added new MCP tools to leverage concurrent functionality:

- **`rag_crawl`**: Original functionality (single URL crawl + RAG search)
- **`rag_crawl_multi`**: NEW - Concurrently crawl multiple URLs and perform RAG search on combined results

#### New Tool: `rag_crawl_multi`

Parameters:
- `urls`: Comma-separated list of URLs to crawl concurrently
- `query`: Search query to find relevant content across all crawled sites
- `max_depth`: Maximum recursion depth for each site (default: 2)
- `max_pages_per_url`: Max pages to crawl per URL (default: 10)
- `top_k`: Number of top results to return (default: 5)

Returns:
- `crawled_urls_count`: Number of URLs crawled
- `crawled_pages`: Total number of pages crawled across all URLs
- `index_stats`: Statistics about the RAG index
- `results`: Top-k search results with score, title, URL, and content

## Benefits

### Performance Improvements
- **Concurrent Crawling**: Multiple sites can be crawled simultaneously
- **Parallel Search**: Search scoring is distributed across multiple goroutines
- **Resource Efficiency**: Controlled concurrency prevents resource exhaustion

### Scalability
- **Configurable Limits**: Adjustable concurrency limits based on system capacity
- **Context Awareness**: Proper cancellation and timeout handling
- **Memory Management**: Efficient indexing and cleanup

### Developer Experience
- **Simplified API**: Easy-to-use functions for common patterns
- **Consistent Interface**: Similar API patterns across different functionality
- **Error Handling**: Proper error propagation and handling

## Architecture

```
[Client Request] 
       ↓
[MCP Server]
       ↓
[Concurrent Pool] ←→ [Multiple Crawlers]
       ↓
[RAG Index] ←→ [Concurrent Search]
       ↓
[Results Returned]
```

The architecture separates concerns while maintaining efficiency:
- Concurrent pools manage goroutine lifecycle
- Crawlers handle web scraping independently
- RAG index processes and searches content efficiently
- MCP tools provide unified interface to clients

## Usage Examples

### Single Site RAG Search
```
{
  "name": "rag_crawl",
  "arguments": {
    "url": "https://example.com",
    "query": "machine learning",
    "max_depth": 2,
    "max_pages": 10,
    "top_k": 5
  }
}
```

### Multi-Site Concurrent RAG Search
```
{
  "name": "rag_crawl_multi",
  "arguments": {
    "urls": "https://site1.com,https://site2.com,https://site3.com",
    "query": "artificial intelligence",
    "max_depth": 2,
    "max_pages_per_url": 5,
    "top_k": 10
  }
}
```

## Future Enhancements

- **Dynamic Scaling**: Automatically adjust concurrency based on system load
- **Cache Layer**: Cache frequently accessed content for faster retrieval
- **Advanced Search Options**: Support for filtering, faceting, and other search features
- **Monitoring**: Add metrics and observability for concurrent operations
- **Retry Logic**: Implement retry mechanisms for failed operations