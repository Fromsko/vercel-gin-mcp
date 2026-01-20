package tools

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

// DocFile 文档文件
type DocFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// DocsResult 文档下载结果
type DocsResult struct {
	RepoURL string    `json:"repo_url"`
	Owner   string    `json:"owner"`
	Repo    string    `json:"repo"`
	Files   []DocFile `json:"files"`
	Count   int       `json:"count"`
}

// DownhubOptions 下载选项
type DownhubOptions struct {
	RepoURL    string   // GitHub 仓库 URL
	DocsPath   string   // 文档路径（可选，如 "docs"）
	Extensions []string // 文件扩展名过滤（默认 .md, .txt）
	MaxFiles   int      // 最大文件数（防止过多）
}

// DefaultOptions 默认选项
func DefaultOptions() *DownhubOptions {
	return &DownhubOptions{
		Extensions: []string{".md", ".txt"},
		MaxFiles:   50,
	}
}

// Downhub 文档下载器
type Downhub struct {
	opts *DownhubOptions
}

// NewDownhub 创建下载器
func NewDownhub() *Downhub {
	return &Downhub{opts: DefaultOptions()}
}

// URL 设置仓库 URL
func (d *Downhub) URL(url string) *Downhub {
	d.opts.RepoURL = url
	return d
}

// Path 设置文档路径
func (d *Downhub) Path(path string) *Downhub {
	d.opts.DocsPath = path
	return d
}

// Extensions 设置文件扩展名
func (d *Downhub) Extensions(exts ...string) *Downhub {
	d.opts.Extensions = exts
	return d
}

// MaxFiles 设置最大文件数
func (d *Downhub) MaxFiles(n int) *Downhub {
	d.opts.MaxFiles = n
	return d
}

// Fetch 执行下载
func (d *Downhub) Fetch() (*DocsResult, error) {
	if d.opts.RepoURL == "" {
		return nil, fmt.Errorf("repository URL is required")
	}

	result := &DocsResult{
		RepoURL: d.opts.RepoURL,
		Files:   []DocFile{},
	}

	// 解析 owner 和 repo
	if strings.Contains(d.opts.RepoURL, "github.com/") {
		parts := strings.Split(d.opts.RepoURL, "github.com/")
		if len(parts) > 1 {
			repoParts := strings.Split(strings.TrimSuffix(parts[1], ".git"), "/")
			if len(repoParts) >= 2 {
				result.Owner = repoParts[0]
				result.Repo = repoParts[1]
			}
		}
	}

	// 克隆仓库到内存
	r, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:   d.opts.RepoURL,
		Depth: 1,
	})
	if err != nil {
		return nil, fmt.Errorf("clone failed: %w", err)
	}

	// 获取 HEAD
	ref, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("get HEAD failed: %w", err)
	}

	// 获取 commit
	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return nil, fmt.Errorf("get commit failed: %w", err)
	}

	// 获取 tree
	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get tree failed: %w", err)
	}

	// 遍历文件
	err = tree.Files().ForEach(func(f *object.File) error {
		if len(result.Files) >= d.opts.MaxFiles {
			return nil
		}

		if d.shouldInclude(f.Name) {
			content, err := f.Contents()
			if err != nil {
				return nil
			}
			result.Files = append(result.Files, DocFile{
				Path:    f.Name,
				Content: content,
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk tree failed: %w", err)
	}

	result.Count = len(result.Files)
	return result, nil
}

// shouldInclude 判断文件是否应该包含
func (d *Downhub) shouldInclude(filename string) bool {
	// 检查路径前缀
	if d.opts.DocsPath != "" {
		if !strings.HasPrefix(filename, d.opts.DocsPath+"/") && filename != d.opts.DocsPath {
			return false
		}
	}

	// 检查扩展名
	ext := strings.ToLower(filepath.Ext(filename))
	for _, e := range d.opts.Extensions {
		if ext == e {
			return true
		}
	}
	return false
}

// ToMarkdown 将结果转换为 Markdown 格式
func (r *DocsResult) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s/%s\n\n", r.Owner, r.Repo))
	sb.WriteString(fmt.Sprintf("共 %d 个文档文件\n\n", r.Count))

	for _, f := range r.Files {
		sb.WriteString(fmt.Sprintf("## %s\n\n", f.Path))
		sb.WriteString("```\n")
		// 限制内容长度
		content := f.Content
		if len(content) > 2000 {
			content = content[:2000] + "\n... (内容已截断)"
		}
		sb.WriteString(content)
		sb.WriteString("\n```\n\n")
	}

	return sb.String()
}

// QuickDownload 快速下载（便捷函数）
func QuickDownload(repoURL string) (*DocsResult, error) {
	return NewDownhub().URL(repoURL).Fetch()
}

// QuickDownloadPath 快速下载指定路径（便捷函数）
func QuickDownloadPath(repoURL, docsPath string) (*DocsResult, error) {
	return NewDownhub().URL(repoURL).Path(docsPath).Fetch()
}
