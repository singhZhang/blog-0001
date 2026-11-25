// Package tests 集中存放服务的全部单元测试。
// 按项目约定，测试只通过导出接口验证 internal 各包的行为，
// 因此放在独立的 tests 目录，而不是与被测代码同目录。
package tests

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/singhZhang/blog-0001/service/internal/article"
)

// validArticle 返回一份通过全部校验的基准文章，各用例按需篡改单个字段。
func validArticle() article.Article {
	return article.Article{
		ID:          "001",
		Slug:        "understand-query-index-again",
		Title:       "重新理解查询索引",
		Excerpt:     "从 B+ 树与执行计划出发。",
		PublishedAt: "2026-10-12",
		Category:    "Database",
		ReadMinutes: 12,
		Tags:        []string{"MySQL", "Index"},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*article.Article)
		wantErr string
	}{
		{name: "合法文章", wantErr: ""},
		{name: "缺少 id", mutate: func(a *article.Article) { a.ID = "" }, wantErr: "缺少 id"},
		{name: "slug 含大写", mutate: func(a *article.Article) { a.Slug = "Understand-Query" }, wantErr: "slug"},
		{name: "slug 含下划线", mutate: func(a *article.Article) { a.Slug = "understand_query" }, wantErr: "slug"},
		{name: "slug 连字符开头", mutate: func(a *article.Article) { a.Slug = "-understand-query" }, wantErr: "slug"},
		{name: "slug 连字符结尾", mutate: func(a *article.Article) { a.Slug = "understand-query-" }, wantErr: "slug"},
		{name: "slug 连字符重复", mutate: func(a *article.Article) { a.Slug = "understand--query" }, wantErr: "slug"},
		{name: "slug 为空", mutate: func(a *article.Article) { a.Slug = "" }, wantErr: "slug"},
		{name: "缺少 title", mutate: func(a *article.Article) { a.Title = "" }, wantErr: "缺少 title"},
		{name: "缺少 excerpt", mutate: func(a *article.Article) { a.Excerpt = "" }, wantErr: "缺少 excerpt"},
		{name: "日期缺零", mutate: func(a *article.Article) { a.PublishedAt = "2026-1-8" }, wantErr: "publishedAt"},
		{name: "日期不存在", mutate: func(a *article.Article) { a.PublishedAt = "2026-02-30" }, wantErr: "publishedAt"},
		{name: "日期带时间", mutate: func(a *article.Article) { a.PublishedAt = "2026-02-03T00:00:00Z" }, wantErr: "publishedAt"},
		{name: "缺少 category", mutate: func(a *article.Article) { a.Category = "" }, wantErr: "缺少 category"},
		{name: "阅读分钟为零", mutate: func(a *article.Article) { a.ReadMinutes = 0 }, wantErr: "readMinutes"},
		{name: "阅读分钟为负", mutate: func(a *article.Article) { a.ReadMinutes = -5 }, wantErr: "readMinutes"},
		{
			name:    "纯数字 slug 合法",
			mutate:  func(a *article.Article) { a.Slug = "2026" },
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := validArticle()
			if tt.mutate != nil {
				tt.mutate(&a)
			}
			err := a.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("期望通过校验，实际错误: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("期望错误包含 %q，实际通过校验", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("错误信息 %q 未包含 %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestExcerptFromContent(t *testing.T) {
	content := "# 标题\n\n那天下午，监控告警提示延迟升高。\n\n```sql\nSELECT 1;\n```\n\n后面的句子不应该进入摘要。"
	got := article.ExcerptFromContent(content)
	if !strings.HasPrefix(got, "标题 那天下午") {
		t.Fatalf("摘要 = %q，应先去掉标题标记并跳过代码块", got)
	}
	if strings.Contains(got, "SELECT") {
		t.Fatalf("摘要不应包含代码块，实际: %q", got)
	}

	long := strings.Repeat("字", 90)
	got = article.ExcerptFromContent(long)
	if utf8.RuneCountInString(got) != 81 || !strings.HasSuffix(got, "…") {
		t.Fatalf("超长摘要 = %q，期望截到 80 字并加省略号", got)
	}
}
