// 一次性重灌主题和文章。运行后数据留在 MySQL 中。
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/singhZhang/blog-0001/service/internal/article"
	"github.com/singhZhang/blog-0001/service/internal/category"
	"github.com/singhZhang/blog-0001/service/internal/config"
	"github.com/singhZhang/blog-0001/service/internal/db"
)

func main() {
	cfg := config.Load()
	conn, err := db.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ctx := context.Background()
	if _, err := conn.ExecContext(ctx, `DELETE FROM articles`); err != nil {
		log.Fatal(err)
	}
	if _, err := conn.ExecContext(ctx, `DELETE FROM categories`); err != nil {
		log.Fatal(err)
	}

	categories := category.NewMySQLStore(conn)
	for _, item := range themeList() {
		if _, err := categories.Create(ctx, item); err != nil {
			log.Fatalf("主题 %s: %v", item.Name, err)
		}
	}

	posts, err := articleList()
	if err != nil {
		log.Fatal(err)
	}
	articles := article.NewMySQLStore(conn)
	for _, item := range posts {
		if _, err := articles.Create(ctx, item); err != nil {
			log.Fatalf("文章 %s: %v", item.Slug, err)
		}
	}
	if err := db.MarkContentSeeded(ctx, conn); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("主题 %d，文章 %d\n", len(themeList()), len(posts))
}
