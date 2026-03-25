package tools

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"mcp-server/handler/mcp/utils"
)

type Chunk struct {
	Text     string `json:"text"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	Position int    `json:"position"`
}

type SearchResult struct {
	Chunk Chunk   `json:"chunk"`
	Score float64 `json:"score"`
}

type RAGIndex struct {
	chunks    []Chunk
	docFreq   map[string]int
	termFreqs []map[string]int
	docLens   []int
	avgDocLen float64
	totalDocs int
}

func NewRAGIndex() *RAGIndex {
	return &RAGIndex{
		docFreq: make(map[string]int),
	}
}

func (r *RAGIndex) AddChunk(chunk Chunk) {
	tokens := tokenize(chunk.Text)
	r.chunks = append(r.chunks, chunk)

	tf := make(map[string]int, len(tokens))
	seen := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		tf[t]++
		seen[t] = true
	}
	r.termFreqs = append(r.termFreqs, tf)

	docLen := len(tokens)
	r.docLens = append(r.docLens, docLen)

	for t := range seen {
		r.docFreq[t]++
	}

	r.totalDocs = len(r.chunks)
	totalLen := 0
	for _, l := range r.docLens {
		totalLen += l
	}
	r.avgDocLen = float64(totalLen) / float64(r.totalDocs)
}

func (r *RAGIndex) AddFromCrawl(result *CrawlResult, chunkSize int) {
	if chunkSize <= 0 {
		chunkSize = 500
	}
	for _, page := range result.Pages {
		chunks := chunkText(page.Markdown, chunkSize)
		for i, text := range chunks {
			if strings.TrimSpace(text) != "" {
				r.AddChunk(Chunk{
					Text:     text,
					URL:      page.URL,
					Title:    page.Title,
					Position: i,
				})
			}
		}
	}
}

func (r *RAGIndex) Search(query string, topK int) []SearchResult {
	if topK <= 0 {
		topK = 5
	}
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 || r.totalDocs == 0 {
		return nil
	}

	k1 := 1.5
	b := 0.75

	// Use concurrent processing for scoring
	numWorkers := utils.MinInt(4, r.totalDocs) // Limit workers to avoid overhead
	if numWorkers <= 0 {
		numWorkers = 1
	}

	// Channel to receive results
	resultsChan := make(chan SearchResult, r.totalDocs)
	var wg sync.WaitGroup

	// Distribute work among workers
	chunkSize := utils.MaxInt(1, len(r.chunks)/numWorkers)
	for i := 0; i < len(r.chunks); i += chunkSize {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for j := start; j < end && j < len(r.chunks); j++ {
				tf := r.termFreqs[j]
				docLen := r.docLens[j]

				score := 0.0
				for _, term := range queryTokens {
					df := r.docFreq[term]
					if df == 0 {
						continue
					}
					freq := tf[term]
					if freq == 0 {
						continue
					}

					idf := math.Log((float64(r.totalDocs-df)+0.5)/(float64(df)+0.5) + 1.0)
					tfNorm := (float64(freq) * (k1 + 1)) / (float64(freq) + k1*(1-b+b*(float64(docLen)/r.avgDocLen)))
					score += idf * tfNorm
				}

				if score > 0 {
					resultsChan <- SearchResult{
						Chunk: r.chunks[j],
						Score: score,
					}
				}
			}
		}(i, utils.MinInt(i+chunkSize, len(r.chunks)))
	}

	// Close channel after all workers finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var scores []SearchResult
	for result := range resultsChan {
		scores = append(scores, result)
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	if len(scores) > topK {
		scores = scores[:topK]
	}

	return scores
}

func (r *RAGIndex) Stats() string {
	return fmt.Sprintf("chunks: %d, avg_doc_len: %.0f tokens", r.totalDocs, r.avgDocLen)
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var buf strings.Builder

	for _, r := range text {
		if isCJK(r) {
			if buf.Len() > 0 {
				word := buf.String()
				if len(word) > 1 {
					tokens = append(tokens, word)
				}
				buf.Reset()
			}
			tokens = append(tokens, string(r))
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) {
			buf.WriteRune(r)
		} else {
			if buf.Len() > 0 {
				word := buf.String()
				if len(word) > 1 {
					tokens = append(tokens, word)
				}
				buf.Reset()
			}
		}
	}

	if buf.Len() > 0 {
		word := buf.String()
		if len(word) > 1 {
			tokens = append(tokens, word)
		}
	}

	return tokens
}

func isCJK(r rune) bool {
	return unicode.In(r, unicode.Han, unicode.Hangul, unicode.Katakana, unicode.Hiragana)
}

func chunkText(text string, chunkSize int) []string {
	paragraphs := strings.Split(text, "\n")

	var chunks []string
	var current strings.Builder
	currentLen := 0

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		paraLen := utf8.RuneCountInString(para)
		if currentLen > 0 && currentLen+paraLen > chunkSize {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
				currentLen = 0
			}
			if paraLen > chunkSize*2 {
				subChunks := splitLongParagraph(para, chunkSize)
				chunks = append(chunks, subChunks...)
				continue
			}
		}

		current.WriteString(para)
		current.WriteString("\n\n")
		currentLen += paraLen
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	if len(chunks) == 0 {
		if text != "" {
			chunks = append(chunks, text)
		}
	}

	return chunks
}

func splitLongParagraph(text string, chunkSize int) []string {
	runes := []rune(text)
	var chunks []string
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}
	return chunks
}


