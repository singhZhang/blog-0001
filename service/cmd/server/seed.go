package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"database/sql"

	"github.com/singhZhang/blog-0001/service/internal/article"
	"github.com/singhZhang/blog-0001/service/internal/category"
	"github.com/singhZhang/blog-0001/service/internal/db"
)

func seedContent(ctx context.Context, conn *sql.DB, articles *article.MySQLStore, categories *category.MySQLStore, contentDir string, logger *slog.Logger) error {
	seeded, err := db.ContentSeeded(ctx, conn)
	if err != nil {
		return fmt.Errorf("读取初始化标记: %w", err)
	}
	if !seeded {
		articleCount, err := articles.CountContext(ctx)
		if err != nil {
			return fmt.Errorf("统计文章: %w", err)
		}
		if articleCount == 0 {
			loaded, err := loadMarkdownArticles(contentDir)
			if err != nil {
				return err
			}
			if len(loaded) > 0 {
				for _, item := range loaded {
					if err := categories.Ensure(ctx, category.Category{Name: item.Category}); err != nil {
						return fmt.Errorf("补齐主题 %s: %w", item.Category, err)
					}
				}
				if err := articles.ImportAll(ctx, loaded); err != nil {
					return fmt.Errorf("导入 Markdown 文章: %w", err)
				}
				logger.Info("已从 Markdown 导入文章", "count", len(loaded), "dir", contentDir)
			}
		}
		if err := db.MarkContentSeeded(ctx, conn); err != nil {
			return fmt.Errorf("记录初始化标记: %w", err)
		}
	}

	rows, err := articles.List(ctx)
	if err != nil {
		return fmt.Errorf("读取文章分类: %w", err)
	}
	for _, item := range rows {
		if err := categories.Ensure(ctx, category.Category{Name: item.Category}); err != nil {
			return fmt.Errorf("补齐主题 %s: %w", item.Category, err)
		}
	}
	return nil
}

func loadMarkdownArticles(dir string) ([]article.Article, error) {
	info, err := os.Stat(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取文章目录 %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("文章目录 %s 不是目录", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取文章目录 %s: %w", dir, err)
	}
	hasMarkdown := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			hasMarkdown = true
			break
		}
	}
	if !hasMarkdown {
		return nil, nil
	}

	store, err := article.NewFileStore(dir)
	if err != nil {
		return nil, fmt.Errorf("解析 Markdown 文章: %w", err)
	}
	return store.All(), nil
}
