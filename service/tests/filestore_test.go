package tests

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/singhZhang/blog-0001/service/internal/article"
)

// writeArticle 在目录下写入一篇 Markdown 文章，文件名与 slug 之间没有对应关系。
func writeArticle(t *testing.T, dir, name, front, body string) {
	t.Helper()
	content := "---\n" + front + "\n---\n" + body
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试文章 %s: %v", name, err)
	}
}

// frontmatter 生成一份合法的元数据块，slug 与发布日期由调用方指定。
func frontmatter(slug, publishedAt string) string {
	return `id: "001"
slug: "` + slug + `"
title: "第一篇"
excerpt: "摘要"
publishedAt: "` + publishedAt + `"
category: "Engineering"
readMinutes: 5
tags:
  - "Go"`
}

func TestNewFileStoreLoadsAndSorts(t *testing.T) {
	dir := t.TempDir()
	// 故意乱序写入，验证加载后按发布日期倒序、同日按 slug 升序。
	writeArticle(t, dir, "same-day-b.md", frontmatter("same-day-b", "2026-01-01"), "同日发布、slug 靠后")
	writeArticle(t, dir, "older.md", frontmatter("aaa-old", "2026-01-01"), "同日发布、slug 靠前")
	writeArticle(t, dir, "newer.md", frontmatter("newer", "2026-03-01"), "最新发布")

	store, err := article.NewFileStore(dir)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	if got := store.Count(); got != 3 {
		t.Fatalf("文章数量 = %d, 期望 3", got)
	}

	rows, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("列出文章: %v", err)
	}
	var slugs []string
	for _, a := range rows {
		slugs = append(slugs, a.Slug)
	}
	// 03-01 的排最前；同在 01-01 的两篇按 slug 升序。
	want := []string{"newer", "aaa-old", "same-day-b"}
	if strings.Join(slugs, ",") != strings.Join(want, ",") {
		t.Fatalf("排序结果 = %v, 期望 %v", slugs, want)
	}
}

func TestNewFileStoreIgnoresNonMarkdown(t *testing.T) {
	dir := t.TempDir()
	writeArticle(t, dir, "post.md", frontmatter("first-post", "2026-01-01"), "正文")
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("忽略我"), 0o644); err != nil {
		t.Fatalf("写入干扰文件: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "drafts"), 0o755); err != nil {
		t.Fatalf("创建子目录: %v", err)
	}

	store, err := article.NewFileStore(dir)
	if err != nil {
		t.Fatalf(".txt 与子目录不应影响加载: %v", err)
	}
	if store.Count() != 1 {
		t.Fatalf("文章数量 = %d, 期望 1", store.Count())
	}
}

func TestNewFileStoreCRLF(t *testing.T) {
	dir := t.TempDir()
	content := strings.ReplaceAll(
		"---\n"+frontmatter("first-post", "2026-01-01")+"\n---\n第一段\n第二段", "\n", "\r\n")
	if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("写入测试文章: %v", err)
	}

	store, err := article.NewFileStore(dir)
	if err != nil {
		t.Fatalf("CRLF 行尾的文件应能加载: %v", err)
	}
	a, err := store.GetBySlug(context.Background(), "first-post")
	if err != nil {
		t.Fatalf("查询文章: %v", err)
	}
	if a.Title != "第一篇" {
		t.Fatalf("标题 = %q, 期望 %q", a.Title, "第一篇")
	}
}

func TestNewFileStoreErrors(t *testing.T) {
	t.Run("目录不存在", func(t *testing.T) {
		if _, err := article.NewFileStore(filepath.Join(t.TempDir(), "missing")); err == nil {
			t.Fatal("期望目录不存在时报错")
		}
	})

	t.Run("空目录", func(t *testing.T) {
		if _, err := article.NewFileStore(t.TempDir()); err == nil {
			t.Fatal("期望空目录时报错")
		}
	})

	t.Run("slug 重复", func(t *testing.T) {
		dir := t.TempDir()
		writeArticle(t, dir, "a.md", frontmatter("first-post", "2026-01-01"), "正文")
		writeArticle(t, dir, "b.md", frontmatter("first-post", "2026-02-01"), "另一篇")
		_, err := article.NewFileStore(dir)
		if err == nil || !strings.Contains(err.Error(), "first-post") {
			t.Fatalf("期望报 slug 重复且指明 slug, 实际: %v", err)
		}
	})

	t.Run("元数据不合法", func(t *testing.T) {
		dir := t.TempDir()
		broken := strings.Replace(frontmatter("first-post", "2026-01-01"), "readMinutes: 5", "readMinutes: 0", 1)
		writeArticle(t, dir, "broken.md", broken, "正文")
		_, err := article.NewFileStore(dir)
		if err == nil || !strings.Contains(err.Error(), "broken.md") {
			t.Fatalf("期望报错且指明文件名, 实际: %v", err)
		}
	})
}

func TestListOmitsContent(t *testing.T) {
	dir := t.TempDir()
	writeArticle(t, dir, "post.md", frontmatter("first-post", "2026-01-01"), "这里是正文内容")

	store, err := article.NewFileStore(dir)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	rows, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("列出文章: %v", err)
	}
	if len(rows) != 1 || rows[0].Content != "" {
		t.Fatalf("列表不应包含正文, 实际: %+v", rows)
	}
}

func TestGetBySlug(t *testing.T) {
	dir := t.TempDir()
	writeArticle(t, dir, "post.md", frontmatter("first-post", "2026-01-01"), "正文内容")

	store, err := article.NewFileStore(dir)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}

	a, err := store.GetBySlug(context.Background(), "first-post")
	if err != nil {
		t.Fatalf("查询存在的文章: %v", err)
	}
	if a.Content != "正文内容" {
		t.Fatalf("详情应包含正文, 实际 %q", a.Content)
	}

	if _, err := store.GetBySlug(context.Background(), "missing"); !errors.Is(err, article.ErrNotFound) {
		t.Fatalf("查询不存在的文章应返回 ErrNotFound, 实际: %v", err)
	}
}

// TestNewFileStoreFrontmatterDelimiters 通过公开接口验证 frontmatter 分隔线的容错。
// splitFrontMatter 是包内私有函数，测试集中存放后只能借助精心构造的文件间接覆盖。
func TestNewFileStoreFrontmatterDelimiters(t *testing.T) {
	t.Run("带 BOM", func(t *testing.T) {
		dir := t.TempDir()
		// BOM 是不可见控制字符，必须用转义写进字符串字面量。
		content := "\uFEFF---\n" + frontmatter("first-post", "2026-01-01") + "\n---\n正文"
		if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("写入测试文章: %v", err)
		}
		store, err := article.NewFileStore(dir)
		if err != nil {
			t.Fatalf("带 BOM 的文件应能加载: %v", err)
		}
		if _, err := store.GetBySlug(context.Background(), "first-post"); err != nil {
			t.Fatalf("查询文章: %v", err)
		}
	})

	t.Run("缺少开始分隔线", func(t *testing.T) {
		dir := t.TempDir()
		content := frontmatter("first-post", "2026-01-01") + "\n---\n正文"
		if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("写入测试文章: %v", err)
		}
		_, err := article.NewFileStore(dir)
		if err == nil || !strings.Contains(err.Error(), "frontmatter") {
			t.Fatalf("期望缺少开始分隔线时报错且指明 frontmatter, 实际: %v", err)
		}
	})

	t.Run("缺少结束分隔线", func(t *testing.T) {
		dir := t.TempDir()
		content := "---\n" + frontmatter("first-post", "2026-01-01") + "\n正文"
		if err := os.WriteFile(filepath.Join(dir, "post.md"), []byte(content), 0o644); err != nil {
			t.Fatalf("写入测试文章: %v", err)
		}
		_, err := article.NewFileStore(dir)
		if err == nil || !strings.Contains(err.Error(), "frontmatter") {
			t.Fatalf("期望缺少结束分隔线时报错且指明 frontmatter, 实际: %v", err)
		}
	})
}

func TestNewFileStoreEmptyTags(t *testing.T) {
	dir := t.TempDir()
	noTags := strings.Split(frontmatter("first-post", "2026-01-01"), "tags:")[0]
	writeArticle(t, dir, "no-tags.md", noTags, "正文")

	store, err := article.NewFileStore(dir)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	rows, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("列出文章: %v", err)
	}
	if len(rows) != 1 || rows[0].Tags == nil || len(rows[0].Tags) != 0 {
		t.Fatalf("无 tags 的文章应序列化为空切片而非 null, 实际: %#v", rows[0].Tags)
	}
}
