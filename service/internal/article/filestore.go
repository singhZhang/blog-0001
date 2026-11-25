package article

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// ErrNotFound 表示请求的文章不存在，接口层据此返回 404。
var ErrNotFound = errors.New("article not found")

// ErrDuplicateSlug 表示新增或改名后的 slug 与现有文章冲突，接口层据此返回 409。
var ErrDuplicateSlug = errors.New("duplicate slug")

// FileStore 在启动时一次性读取目录下的 Markdown 文件并常驻内存。
// 读操作直接走内存快照；管理接口的写操作在持锁状态下先落盘再更新内存，
// 读者不会看到中间状态。博客文章数量有限，这样的实现简单且足够快。
type FileStore struct {
	mu       sync.RWMutex
	dir      string
	articles []Article
	bySlug   map[string]Article
}

// NewFileStore 读取 dir 下所有 .md 文件并构建内存索引，任何文件不合法都会导致启动失败，
// 让内容错误在部署阶段暴露而不是在线上以 500 的形式出现。
func NewFileStore(dir string) (*FileStore, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取文章目录 %s: %w", dir, err)
	}

	store := &FileStore{dir: dir, bySlug: make(map[string]Article)}
	for _, entry := range entries {
		// 原子写产生的临时文件模式是 ".tmp-*"，不以 .md 结尾，崩溃残留在这里天然被跳过。
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			continue
		}
		a, err := parseArticleFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("解析 %s: %w", entry.Name(), err)
		}
		if _, exists := store.bySlug[a.Slug]; exists {
			return nil, fmt.Errorf("slug %q 在多个文件中重复", a.Slug)
		}
		store.articles = append(store.articles, a)
		store.bySlug[a.Slug] = a
	}

	if len(store.articles) == 0 {
		return nil, fmt.Errorf("文章目录 %s 中没有可用的 .md 文件", dir)
	}

	sortArticles(store.articles)
	return store, nil
}

// sortArticles 就地排序：发布日期从新到旧；同日发布的按 slug 稳定排序。
func sortArticles(articles []Article) {
	slices.SortStableFunc(articles, func(a, b Article) int {
		if c := strings.Compare(b.PublishedAt, a.PublishedAt); c != 0 {
			return c
		}
		return strings.Compare(a.Slug, b.Slug)
	})
}

// Count 返回文章总数，用于启动日志。
func (s *FileStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.articles)
}

// All 返回包含正文的文章副本，供数据库为空时导入初始内容。
func (s *FileStore) All() []Article {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]Article, len(s.articles))
	copy(rows, s.articles)
	return rows
}

// List 返回按发布日期倒序的文章副本，正文字段被清空以减小响应体积。
func (s *FileStore) List(ctx context.Context) ([]Article, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]Article, 0, len(s.articles))
	for _, a := range s.articles {
		a.Content = ""
		rows = append(rows, a)
	}
	return rows, nil
}

// GetBySlug 返回文章全文；不存在时返回 ErrNotFound。
func (s *FileStore) GetBySlug(ctx context.Context, slug string) (Article, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.bySlug[slug]
	if !ok {
		return Article{}, ErrNotFound
	}
	return a, nil
}

// Create 校验并落盘一篇新文章，返回补全了服务端字段的完整文章。
// ID 与阅读时长由服务端生成，不信任请求体传入的值。
func (s *FileStore) Create(ctx context.Context, input Article) (Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	a := input
	a.ID = NewID()
	a.Tags = normalizeTags(a.Tags)
	a.ReadMinutes = EstimateReadMinutes(a.Content)
	if err := a.Validate(); err != nil {
		return Article{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	if _, exists := s.bySlug[a.Slug]; exists {
		return Article{}, fmt.Errorf("%w: slug %q 已存在", ErrDuplicateSlug, a.Slug)
	}

	if err := s.writeFile(a); err != nil {
		return Article{}, err
	}
	s.insert(a)
	return a, nil
}

// Update 以 slug 定位并整体替换一篇文章；不存在时返回 ErrNotFound。
// ID 保持不变（不允许通过更新修改 ID）；修改 slug 时旧文件在写入新文件成功后删除。
func (s *FileStore) Update(ctx context.Context, slug string, input Article) (Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	old, ok := s.bySlug[slug]
	if !ok {
		return Article{}, ErrNotFound
	}

	a := input
	a.ID = old.ID
	a.Tags = normalizeTags(a.Tags)
	a.ReadMinutes = EstimateReadMinutes(a.Content)
	if err := a.Validate(); err != nil {
		return Article{}, fmt.Errorf("%w: %s", ErrInvalid, err)
	}
	if a.Slug != slug {
		if _, exists := s.bySlug[a.Slug]; exists {
			return Article{}, fmt.Errorf("%w: slug %q 已存在", ErrDuplicateSlug, a.Slug)
		}
	}

	if err := s.writeFile(a); err != nil {
		return Article{}, err
	}
	if a.Slug != slug {
		// 先写后删：删除失败时磁盘上新文件已就位，重试或重启可恢复，数据不丢。
		if err := os.Remove(s.filePath(slug)); err != nil {
			return Article{}, fmt.Errorf("删除旧文件 %s: %w", s.filePath(slug), err)
		}
	}
	s.remove(slug)
	s.insert(a)
	return a, nil
}

// Delete 删除一篇文章；不存在时返回 ErrNotFound。
func (s *FileStore) Delete(ctx context.Context, slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.bySlug[slug]; !ok {
		return ErrNotFound
	}
	if err := os.Remove(s.filePath(slug)); err != nil {
		return fmt.Errorf("删除文件 %s: %w", s.filePath(slug), err)
	}
	s.remove(slug)
	return nil
}

// normalizeTags 保证 tags 序列化成 [] 而不是 null。
func normalizeTags(tags []string) []string {
	if tags == nil {
		return []string{}
	}
	return tags
}

// insert 把文章加入内存索引并保持排序不变量。
func (s *FileStore) insert(a Article) {
	s.articles = append(s.articles, a)
	sortArticles(s.articles)
	s.bySlug[a.Slug] = a
}

// remove 从内存索引中删除指定 slug 的文章。
func (s *FileStore) remove(slug string) {
	delete(s.bySlug, slug)
	s.articles = slices.DeleteFunc(s.articles, func(a Article) bool {
		return a.Slug == slug
	})
}

// filePath 返回文章在磁盘上的路径，文件名与 slug 一致。
func (s *FileStore) filePath(slug string) string {
	return filepath.Join(s.dir, slug+".md")
}

// writeFile 以原子方式把文章写成带 frontmatter 的 Markdown 文件：
// 先写同目录临时文件并落盘，再 rename 到目标名，避免读者看到写了一半的内容。
func (s *FileStore) writeFile(a Article) error {
	data, err := marshalFile(a)
	if err != nil {
		return err
	}

	path := s.filePath(a.Slug)
	// 临时文件模式不带 .md 扩展名：进程崩溃残留的临时文件在下次启动时会被 .md 过滤器跳过。
	tmp, err := os.CreateTemp(s.dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("在 %s 创建临时文件: %w", s.dir, err)
	}
	tmpName := tmp.Name()
	// 成功路径上文件已被 rename 走，这里的 Remove 会命中 not exist，忽略即可。
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("写入 %s: %w", tmpName, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("落盘 %s: %w", tmpName, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭 %s: %w", tmpName, err)
	}
	// CreateTemp 默认权限 0600，改回与源码一致的 0644，保证宿主机挂载场景可读。
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("修改权限 %s: %w", tmpName, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("重命名 %s → %s: %w", tmpName, path, err)
	}
	return nil
}

// marshalFile 生成文章的完整文件内容：frontmatter 之后的文件体是 Markdown 正文。
func marshalFile(a Article) ([]byte, error) {
	// Content 的 yaml 标签是 "-"，不会出现在 frontmatter 中。
	front, err := yaml.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("序列化 frontmatter: %w", err)
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(front)
	b.WriteString("---\n")
	b.WriteString(a.Content)
	b.WriteString("\n")
	return []byte(b.String()), nil
}

// parseArticleFile 解析一个带 YAML frontmatter 的 Markdown 文件。
func parseArticleFile(path string) (Article, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Article{}, err
	}

	front, body, err := splitFrontMatter(string(raw))
	if err != nil {
		return Article{}, err
	}

	var a Article
	if err := yaml.Unmarshal([]byte(front), &a); err != nil {
		return Article{}, fmt.Errorf("解析 frontmatter: %w", err)
	}
	a.Content = strings.TrimSpace(body)
	// 前端会遍历 tags，保证序列化成 [] 而不是 null。
	if a.Tags == nil {
		a.Tags = []string{}
	}

	if err := a.Validate(); err != nil {
		return Article{}, err
	}
	return a, nil
}

// splitFrontMatter 拆出 --- 包围的 frontmatter 与之后的文件体。
func splitFrontMatter(text string) (front, body string, err error) {
	const separator = "---"
	// 兼容带 BOM 的文件。BOM 是不可见控制字符，不能直接写进字符串字面量，用转义表示。
	lines := strings.Split(strings.TrimPrefix(text, "\ufeff"), "\n")
	trimCR := func(s string) string { return strings.TrimSuffix(s, "\r") }

	if trimCR(lines[0]) != separator {
		return "", "", fmt.Errorf("文件必须以 --- 开头的 frontmatter 开始")
	}
	for i := 1; i < len(lines); i++ {
		if trimCR(lines[i]) == separator {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n"), nil
		}
	}
	return "", "", fmt.Errorf("frontmatter 缺少结束分隔线 ---")
}
