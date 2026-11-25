// Package article 提供博客文章的领域模型与基于 Markdown 文件的存储实现。
package article

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Article 的字段与前端 src/types/blog.ts 中的 Article 接口保持一致，
// json 标签用于接口响应，yaml 标签用于 Markdown 文件的 frontmatter。
type Article struct {
	ID           string   `json:"id" yaml:"id"`
	Slug         string   `json:"slug" yaml:"slug"`
	Title        string   `json:"title" yaml:"title"`
	Excerpt      string   `json:"excerpt" yaml:"excerpt"`
	PublishedAt  string   `json:"publishedAt" yaml:"publishedAt"` // YYYY-MM-DD
	Category     string   `json:"category" yaml:"category"`
	ReadMinutes  int      `json:"readMinutes" yaml:"readMinutes"`
	Tags         []string `json:"tags" yaml:"tags"`
	FeaturedCode string   `json:"featuredCode,omitempty" yaml:"featuredCode,omitempty"`

	// Content 是 Markdown 正文，来自 frontmatter 之后的文件体。
	// 列表接口不返回该字段，详情接口返回。
	Content string `json:"content,omitempty" yaml:"-"`
}

// slugPattern 约束 slug 为小写字母、数字与连字符的组合，与前端路由参数兼容。
var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Validate 校验文章元数据的完整性，返回第一个不满足的规则。
func (a Article) Validate() error {
	switch {
	case a.ID == "":
		return fmt.Errorf("缺少 id")
	case !slugPattern.MatchString(a.Slug):
		return fmt.Errorf("slug %q 只允许小写字母、数字和连字符", a.Slug)
	case a.Title == "":
		return fmt.Errorf("缺少 title")
	case a.Excerpt == "":
		return fmt.Errorf("缺少 excerpt")
	case !isValidDate(a.PublishedAt):
		return fmt.Errorf("publishedAt %q 必须是 YYYY-MM-DD 格式的合法日期", a.PublishedAt)
	case a.Category == "":
		return fmt.Errorf("缺少 category")
	case a.ReadMinutes <= 0:
		return fmt.Errorf("readMinutes 必须大于 0")
	default:
		return nil
	}
}

// isValidDate 要求日期严格符合 YYYY-MM-DD 且真实存在，
// 格式化结果与原值不同的写法（如 2026-1-8）会被拒绝。
func isValidDate(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

// ErrInvalid 表示文章元数据校验不通过，接口层据此返回 400。
var ErrInvalid = errors.New("invalid article")

// NewID 生成文章 ID：8 字节随机数的十六进制表示。
// 管理接口新增文章时由服务端生成，不信任请求体传入的 ID。
func NewID() string {
	buf := make([]byte, 8)
	// crypto/rand.Read 自 Go 1.24 起保证填满且不返回错误，失败会直接 panic。
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// EstimateReadMinutes 按正文长度估算阅读时长（分钟）。
// 中文技术文的舒适阅读速度约每分钟 400 字，Markdown 标记和代码块
// 会拉低实际速度，估算结果向下取整且至少 1 分钟。
func EstimateReadMinutes(content string) int {
	const charsPerMinute = 400
	minutes := utf8.RuneCountInString(content) / charsPerMinute
	if minutes < 1 {
		return 1
	}
	return minutes
}

const excerptRuneLimit = 80

var (
	markdownFence    = regexp.MustCompile("(?s)```.*?```")
	markdownImage    = regexp.MustCompile(`!\[[^\]]*\]\([^)]*\)`)
	markdownLink     = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)
	markdownHTML     = regexp.MustCompile(`<[^>]+>`)
	markdownInline   = regexp.MustCompile("`[^`]*`")
	markdownHeading  = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	markdownQuote    = regexp.MustCompile(`(?m)^>\s?`)
	markdownBullet   = regexp.MustCompile(`(?m)^[-*+]\s+`)
	markdownEmphasis = regexp.MustCompile(`[*_]{1,3}`)
	markdownSpace    = regexp.MustCompile(`\s+`)
)

// ExcerptFromContent 去掉 Markdown 标记后，从正文截取列表摘要。
// 超过 80 个字时以省略号结尾。正文没有可读文字时返回空串。
func ExcerptFromContent(content string) string {
	text := markdownFence.ReplaceAllString(content, " ")
	text = markdownImage.ReplaceAllString(text, " ")
	text = markdownLink.ReplaceAllString(text, "$1")
	text = markdownHTML.ReplaceAllString(text, " ")
	text = markdownInline.ReplaceAllString(text, " ")
	text = markdownHeading.ReplaceAllString(text, "")
	text = markdownQuote.ReplaceAllString(text, "")
	text = markdownBullet.ReplaceAllString(text, "")
	text = markdownEmphasis.ReplaceAllString(text, "")
	text = strings.TrimSpace(markdownSpace.ReplaceAllString(text, " "))
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	if len(runes) <= excerptRuneLimit {
		return text
	}
	return string(runes[:excerptRuneLimit]) + "…"
}
